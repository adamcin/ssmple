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
	"os"
	"path/filepath"
)

// a FileStore struct accumulates the parameter state for each file.
// the FileStore is seri/deseried from the Path using a Serial object
// matched to the extension of the path.
type FileStore struct {
	Path string
	Dict map[string]string
}

// the *FileStore.Load() function encapsulates the input/output of the
// associated Serial
func (fs *FileStore) Load() error {
	serial := GetSerialFor(fs.Path)
	dict, err := serial.Load(fs.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	} else {
		fs.Dict = dict
		return nil
	}
}

func (fs *FileStore) Save() error {
	serial := GetSerialFor(fs.Path)
	return serial.Save(fs.Path, &fs.Dict)
}

func NewFileStore(confDir string, filename string) FileStore {
	path := filepath.Join(confDir, filename)
	dict := make(map[string]string, 0)

	store := FileStore{
		Path: path,
		Dict: dict}

	return store
}
