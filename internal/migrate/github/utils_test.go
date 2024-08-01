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
