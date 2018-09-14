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

// simple serializer interface for load/save operations between maps
// and different file formats.
type Serial interface {
	Load(path string) (map[string]string, error)
	Save(path string, dict *map[string]string) error
}

// the default Serial implementation writes to files in Java .properties format,
// which can also support OSGi configs and shell variable declaration files.
type PropsSerial struct{}

// Read a key-value map from a file specified by path.
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

// Write a key-value map to a file specified by path.
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

// private global map of file extensions to Serial implementations.
var serials = make(map[string]Serial, 0)

// Retrieve the appropriate serializer for the given path.
// Returns the PropsSerial by default for unregistered extensions.
func GetSerialFor(path string) Serial {
	ext := filepath.Ext(path)
	serial := serials[""]
	if extSerial, ok := serials[ext]; ok {
		serial = extSerial
	}
	return serial
}

// Register a serializer implementation for one or more extensions.
// Each provided value in exts must begin with a period. An error will
// be thrown if an attempt is made to register for an extension that has
// already been registered.
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

// add the default PropsSerial impl
func init() {
	serials[""] = PropsSerial{}
}
