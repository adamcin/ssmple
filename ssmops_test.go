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

import "testing"

func assertBuildParameterPath(t *testing.T, prefix string, filename string, key string, expected string) {
	if result := buildParameterPath(prefix, filename, key);
		result != expected {
		t.Errorf("parameter path not constructed correctly. expected: %s, actual: %s\n", expected, result)
	}
}

func TestBuildParameterPath(t *testing.T) {
	assertBuildParameterPath(t, "/alpha/beta", "one/file.properties", "myprop",
		"/alpha/beta/one/file/myprop")
	assertBuildParameterPath(t, "/alpha/beta", "one/file", "myprop",
		"/alpha/beta/one/file/myprop")
	assertBuildParameterPath(t, "/alpha/beta", "./one/file.properties", "myprop",
		"/alpha/beta/one/file/myprop")
	assertBuildParameterPath(t, "/alpha/beta", "../one/file.properties", "myprop",
		"/alpha/one/file/myprop")
	assertBuildParameterPath(t, "/alpha/beta", "two/one/file.properties", "myprop",
		"/alpha/beta/two/one/file/myprop")
	assertBuildParameterPath(t, "/alpha/beta", "two/./file.properties", "myprop",
		"/alpha/beta/two/file/myprop")
	assertBuildParameterPath(t, "/alpha/beta", "two/../file.properties", "myprop",
		"/alpha/beta/file/myprop")
	assertBuildParameterPath(t, "/alpha/beta", "../two/one/./file.properties", "myprop",
		"/alpha/two/one/file/myprop")
}
