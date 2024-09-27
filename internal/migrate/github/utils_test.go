// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestExtractHunkInfo(t *testing.T) {
	want := []string{
		"@@ -0,0 +35 @@",
		"@@ -115,4 +124,20 @@",
		"@@ -146,2 +145,0 @@",
		"@@ -123,2 +137 @@",
		"@@ -9 +8,0 @@",
		"@@ -13,0 +30 @@",
		"@@ -11,0 +15,5 @@",
		"@@ -11,0 +21,4 @@",
		"@@ -16,0 +16,4 @@",
	}
	var c []*codeComment
	raw, _ := ioutil.ReadFile("testdata/pr_comments.json")
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	for i, c := range c {
		res, _ := extractHunkInfo(c)
		if res != want[i] {
			t.Errorf("got = %v, want %v", res, want[i])
		}
	}
}
