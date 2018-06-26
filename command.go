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
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"log"
	"os"
)

type CmdContext struct {
	Prefs  ParsedArgs
	Stores map[string]*FileStore
	Ssms   *ssm.SSM
	KmsMap KmsMap
}

func requireDir(dir string, mkdir bool) (os.FileInfo, error) {
	fi, fierr := os.Stat(dir)
	if fierr != nil {
		if os.IsNotExist(fierr) && mkdir {
			if mkerr := os.MkdirAll(dir, os.FileMode(int(0755))); mkerr != nil {
				return nil, mkerr
			} else {
				return os.Stat(dir)
			}
		} else {
			return nil, fierr
		}
	}

	if !fi.IsDir() {
		return fi, errors.New("File exists and is not a directory " + dir)
	} else {
		return fi, nil
	}
}

func doGet(ctx *CmdContext) {
	_, fierr := requireDir(ctx.Prefs.ConfDir, true)
	if fierr != nil {
		log.Fatalf("Failed to create conf dir %s. reason: %s", ctx.Prefs.ConfDir, fierr)
	}

	for _, filename := range ctx.Prefs.Filenames {
		if err := getParamsPerFile(ctx, filename); err != nil {
			log.Fatalf("Failed to get parameters for filename %s. reason: %s\n", filename, err)
		}
	}
}

func doPut(ctx *CmdContext) {
	if len(ctx.Prefs.Prefixes) != 1 {
		log.Fatal("put command requires exactly one -s/--starts-with argument.")
	}

	prefix := ctx.Prefs.Prefixes[0]
	for _, filename := range ctx.Prefs.Filenames {
		if err := putParamsPerFile(ctx, filename, prefix); err != nil {
			log.Fatalf("Failed to put parameters from filename %s to prefix %s. reason: %s\n", filename, prefix, err)
		}
	}
}

func doDelete(ctx *CmdContext) {
	if len(ctx.Prefs.Prefixes) != 1 {
		log.Fatal("delete command requires exactly one -s/--starts-with argument.")
	}

	for _, filename := range ctx.Prefs.Filenames {
		deleteParamsPerFile(ctx, filename, ctx.Prefs.Prefixes[0])
	}
}

func doClear(ctx *CmdContext) {
	if len(ctx.Prefs.Prefixes) != 1 {
		log.Fatal("clear command requires exactly one -s/--starts-with argument.")
	}

	for _, filename := range ctx.Prefs.Filenames {
		clearParamsPerFile(ctx, filename, ctx.Prefs.Prefixes[0])
	}
}
