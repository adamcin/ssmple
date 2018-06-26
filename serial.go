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
	"github.com/rickar/props"
	"os"
	"path/filepath"
	"strings"
)

// the
type Serial interface {
	Load(path string) (map[string]string, error)
	Save(path string, dict *map[string]string) error
}

type PropsSerial struct{}

func (s PropsSerial) Load(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	p := props.NewProperties()
	p.Load(file)

	dict := make(map[string]string, len(p.Names()))
	names := p.Names()
	for i := range names {
		name := names[i]
		dict[name] = p.Get(name)
	}

	return dict, nil
}

func (s PropsSerial) Save(path string, dict *map[string]string) error {
	p := props.NewProperties()
	for key, value := range *dict {
		p.Set(key, value)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return p.Write(file)
}

var serials = make(map[string]Serial, 0)

func GetSerialFor(path string) Serial {
	ext := filepath.Ext(path)
	serial := serials[""]
	if extSerial, ok := serials[ext]; ok {
		serial = extSerial
	}
	return serial
}

func RegisterSerial(serial Serial, exts ...string) error {
	for _, ext := range exts {
		if !strings.HasPrefix(ext, ".") {
			return errors.New("serial must be registered with an extension beginning with '.'")
		}
		if _, found := serials[ext]; found {
			return errors.New("serial already registered for extension " + ext)
		}
		serials[ext] = serial
	}

	return nil
}

func init() {
	serials[""] = PropsSerial{}
}
