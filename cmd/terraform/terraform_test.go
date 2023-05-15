package terraform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRun(t *testing.T) {
	tests, err := filepath.Glob("testdata/examples/*.json")
	if err != nil {
		t.Error(err)
		return
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			c := terraformCommand{
				input:  test,
				output: "testdata/output.tf",
				// ... other necessary fields ...
			}

			err = c.run(nil)
			if err != nil {
				t.Error(err)
				return
			}

			// Read the output file
			got := map[string]interface{}{}
			outputData, err := os.ReadFile(c.output)
			if err != nil {
				t.Error(err)
				return
			}
			if err := json.Unmarshal(outputData, &got); err != nil {
				t.Error(err)
				return
			}

			// Read the golden file
			want := map[string]interface{}{}
			goldenData, err := os.ReadFile(test + ".golden")
			if err != nil {
				t.Error(err)
				return
			}
			if err := json.Unmarshal(goldenData, &want); err != nil {
				t.Error(err)
				return
			}

			// Compare the output to the golden file
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("Unexpected conversion result")
				t.Log(diff)
			}
		})
	}
}
