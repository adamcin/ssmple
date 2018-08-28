/*
 * Copyright 2018 Mark Adamcin
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"path"
	"strings"
)

const KeyIdSuffix = "_SecureStringKeyId"

func findAllParametersForPath(ctx *CmdContext, paramPath string) ([]ssm.Parameter, error) {
	var paramsForPath []ssm.Parameter
	maxResults := int64(10)
	recursive := false
	withDecryption := true

	input := ssm.GetParametersByPathInput{
		MaxResults:     &maxResults,
		Path:           &paramPath,
		WithDecryption: &withDecryption,
		Recursive:      &recursive}

	request := ctx.Ssms.GetParametersByPathRequest(&input)
	pager := request.Paginate()
	for pager.Next() {
		result := pager.CurrentPage()
		if len(result.Parameters) > 0 {
			paramsForPath = append(paramsForPath, result.Parameters...)
		}
	}

	if pager.Err() != nil {
		return paramsForPath, pager.Err()
	} else {
		return paramsForPath, nil
	}
}

// If value is all spaces, subtract a space to reconstruct the original value for export.
func unescapeValueAfterGet(value string) string {
	if len(value) == 0 {
		return value
	}

	runes := []rune(value)
	for _, ru := range runes {
		if ru != rune(' ') {
			return value
		}
	}
	return string(runes[0: len(runes)-1])
}

// If value is the empty string or all spaces, add a space so the value is non-empty for SSM.
func escapeValueBeforePut(value string) string {
	runes := []rune(value)
	for _, ru := range runes {
		if ru != rune(' ') {
			return value
		}
	}

	return value + " "
}

func getParamsPerPath(ctx *CmdContext, paramPath string, storeDict *map[string]string) error {
	filterKey, _ := ssm.ParametersFilterKeyName.MarshalValue()
	filterOption := "Equals"
	paramsForPath, findErr := findAllParametersForPath(ctx, paramPath)
	if findErr != nil {
		return findErr
	}

	for _, param := range paramsForPath {
		name := *param.Name

		if param.Type == ssm.ParameterTypeStringList ||
			(ctx.Prefs.NoStoreSecureString && param.Type == ssm.ParameterTypeSecureString) {
			continue
		}

		if !strings.HasPrefix(name, paramPath+"/") {
			continue
		}

		storeKey := strings.TrimPrefix(name, paramPath+"/")
		(*storeDict)[storeKey] = unescapeValueAfterGet(*param.Value)

		if param.Type == ssm.ParameterTypeSecureString {
			sidecarStoreKey := storeKey + KeyIdSuffix
			input := ssm.DescribeParametersInput{}
			input.ParameterFilters = append(input.ParameterFilters,
				ssm.ParameterStringFilter{
					Key:    &filterKey,
					Option: &filterOption,
					Values: []string{name}})

			result, err := ctx.Ssms.DescribeParametersRequest(&input).Send()
			if err != nil {
				return err
			}

			if len(result.Parameters) > 0 {
				if result.Parameters[0].KeyId != nil {
					(*storeDict)[sidecarStoreKey] = ctx.KmsMap.aliasFor(*result.Parameters[0].KeyId)
				}
			}
		}
	}
	return nil
}

// Build an SSM parameter path or name.
// prefix:   hierarchy levels 0-(N-2)
// filename: hierarchy level N-1 (.properties, .json, or .yaml extensions will be stripped)
// key:      optional, hierarchy level N
// TODO make platform independent (i.e., this won't work on windows)
func buildParameterPath(prefix string, filename string, key string) string {
	dir := prefix
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}

	fn := filename
	if len(filename) == 0 {
		fn = "$"
	}

	base := path.Join(prefix, fn)
	realdir := path.Dir(base)
	realfn := path.Base(base)

	if len(realfn) > 0 && strings.ContainsRune(realfn, '.') {
		fnRunes := []rune(realfn)
		realfn = string(fnRunes[0:strings.LastIndex(realfn, ".")])
	}

	if len(key) > 0 {
		return path.Join(realdir, realfn, key)
	} else {
		return path.Join(realdir, realfn)
	}
}

func getParamsPerFile(ctx *CmdContext, filename string) error {
	prefixes := ctx.Prefs.Prefixes
	store := ctx.Stores[filename]
	for _, prefix := range prefixes {
		paramPath := buildParameterPath(prefix, filename, "")
		if err := getParamsPerPath(ctx, paramPath, &store.Dict); err != nil {
			return err
		}
	}

	if len(store.Dict) > 0 {
		return store.Save()
	}

	return nil
}

func clearParamsPerFile(ctx *CmdContext, filename string, prefix string) error {
	paramPath := buildParameterPath(prefix, filename, "")
	params, findErr := findAllParametersForPath(ctx, paramPath)
	if findErr != nil {
		return findErr
	}
	count := len(params)
	names := make([]string, count)
	i := 0
	for _, param := range params {
		names[i] = *param.Name
		i++
	}

	batchSize := 10
	batches := (count / batchSize) + 1
	for b := 0; b < batches; b++ {
		input := ssm.DeleteParametersInput{}
		if b+1 < batches {
			input.Names = append(input.Names, names[batchSize*b:batchSize*(b+1)]...)
		} else {
			input.Names = append(input.Names, names[batchSize*b:]...)
		}
		if len(input.Names) > 0 {
			if _, err := ctx.Ssms.DeleteParametersRequest(&input).Send(); err != nil {
				return err
			}
		}
	}

	return nil
}

func putParamsPerFile(ctx *CmdContext, filename string, prefix string) error {
	if ctx.Prefs.ClearOnPut {
		if err := clearParamsPerFile(ctx, filename, prefix); err != nil {
			return err
		}
	}

	store := ctx.Stores[filename]
	for key, value := range store.Dict {
		if strings.HasSuffix(key, KeyIdSuffix) {
			continue
		}
		sidecarKeyId := key + KeyIdSuffix
		name := buildParameterPath(prefix, filename, key)

		keyId, isSecure := store.Dict[sidecarKeyId]
		if isSecure && ctx.Prefs.NoPutSecureString {
			continue
		}

		if len(ctx.Prefs.KeyIdPutAll) > 0 {
			isSecure = true
			keyId = ctx.Prefs.KeyIdPutAll
		}

		keyId = ctx.KmsMap.deref(keyId)

		escaped := escapeValueBeforePut(value)
		input := ssm.PutParameterInput{}

		input.Name = &name
		input.Value = &escaped
		input.Overwrite = &ctx.Prefs.OverwritePut

		if isSecure {
			input.KeyId = &keyId
			input.Type = ssm.ParameterTypeSecureString
		} else {
			input.Type = ssm.ParameterTypeString
		}

		_, err := ctx.Ssms.PutParameterRequest(&input).Send()
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteParamsPerFile(ctx *CmdContext, filename string, prefix string) error {
	var names []string
	store := ctx.Stores[filename]
	for key := range store.Dict {
		names = append(names, buildParameterPath(prefix, filename, key))
	}

	paramPath := buildParameterPath(prefix, filename, "")
	allParams, findErr := findAllParametersForPath(ctx, paramPath)
	if findErr != nil {
		return findErr
	}

	var allNames []string
	for _, param := range allParams {
		allNames = append(allNames, *param.Name)
	}

	var toDelete []string
	for _, cand := range names {
		for _, name := range allNames {
			if cand == name {
				toDelete = append(toDelete, cand)
				break
			}
		}
	}

	count := len(toDelete)
	batchSize := 10
	batches := (count / batchSize) + 1
	for b := 0; b < batches; b++ {
		input := ssm.DeleteParametersInput{}
		if b+1 < batches {
			input.Names = append(input.Names, names[batchSize*b:batchSize*(b+1)]...)
		} else {
			input.Names = append(input.Names, names[batchSize*b:]...)
		}
		if len(input.Names) > 0 {
			if _, err := ctx.Ssms.DeleteParametersRequest(&input).Send(); err != nil {
				return err
			}
		}
	}

	return nil
}
