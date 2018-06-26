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
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type YamlSerial struct{}

func (s YamlSerial) Load(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(file)
	var m map[string]interface{}
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	dict := make(map[string]string)
	for k, v := range m {
		switch v.(type) {
		case string:
			dict[k] = v.(string)
		case []interface{}, map[string]interface{}:
			return nil, errors.New("nested arrays and objects are not supported. json key " + k)
		default:
			dict[k] = fmt.Sprintf("%v", v)
		}
	}

	return dict, nil
}

func (s YamlSerial) Save(path string, dict *map[string]string) error {
	file, err := os.Create(path)

	if err != nil {
		return err
	}

	enc := yaml.NewEncoder(file)
	return enc.Encode(*dict)
}

func init() {
	RegisterSerial(YamlSerial{}, ".yml", ".yaml")
}
