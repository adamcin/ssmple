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
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"strings"
)

type KmsMap struct {
	aliasesToKeys map[string]string
	keysToAliases map[string]string
}

func (ka KmsMap) deref(alias string) string {
	var fqAlias string

	if strings.HasPrefix(alias, "alias/") {
		fqAlias = alias
	} else {
		fqAlias = "alias/" + alias
	}

	if val, ok := ka.aliasesToKeys[fqAlias]; ok {
		return val
	} else {
		return fqAlias
	}
}

func (ka KmsMap) aliasFor(keyId string) string {
	if val, ok := ka.keysToAliases[keyId]; ok {
		return val
	} else {
		return keyId
	}
}

func buildAliasList(kmss *kms.KMS, kmsMap *KmsMap) error {
	request := kmss.ListAliasesRequest(nil)
	result, err := request.Send()
	if err != nil {
		return err
	} else {
		for _, entry := range result.Aliases {
			if entry.TargetKeyId != nil && entry.AliasName != nil {
				kmsMap.aliasesToKeys[*entry.AliasName] = *entry.TargetKeyId
				kmsMap.keysToAliases[*entry.TargetKeyId] = *entry.AliasName
			}
		}
		return nil
	}
}
