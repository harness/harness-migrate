package terraform

import (
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
		testDir, testFile := filepath.Split(test)
		t.Run(test, func(t *testing.T) {
			c := terraformCommand{
				input:          test,
				output:         testDir + "output/" + testFile + ".out",
				harnessAccount: "DNVIhrzCr9SnPHMQUEvRspB",
				harnessOrg:     "exampleOrg",
				dockerConn:     "exampleDockerConn",
				downgrade:      true,
			}

			// Run the terraform config generation
			err = c.run(nil)
			if err != nil {
				t.Error(err)
				return
			}

			// Read the output file
			got, err := os.ReadFile(c.output)
			if err != nil {
				t.Error(err)
				return
			}

			// Read the golden file
			want, err := os.ReadFile(test + ".golden")
			if err != nil {
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
