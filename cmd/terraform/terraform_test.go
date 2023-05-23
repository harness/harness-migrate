package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRun(t *testing.T) {

	pipelineTypes := []string{"docker", "kubernetes"}

	for _, pipelineType := range pipelineTypes {

		tests, err := filepath.Glob("testdata/examples/" + pipelineType + "/*.json")
		if err != nil {
			t.Error(err)
			return
		}

		for _, test := range tests {
			testDir, testFile := filepath.Split(test)
			t.Run(test, func(t *testing.T) {
				c := terraformCommand{}

				dockerTerraformCommand := terraformCommand{
					input:           test,
					account:         "DNVIhrzCr9SnPHMQUEvRspB",
					endpoint:        "https://app.harness.io/gateway",
					providerSource:  "harness/harness",
					providerVersion: "0.19.1",
					output:          testDir + "output/" + testFile + ".tf",
					organization:    "exampleOrg",
					dockerConn:      "exampleDockerConn",
					downgrade:       true,
					orgSecrets:      true,
				}

				kubernetesTerraformCommand := terraformCommand{
					input:           test,
					account:         "DNVIhrzCr9SnPHMQUEvRspB",
					endpoint:        "https://app.harness.io/gateway",
					providerSource:  "harness/harness",
					providerVersion: "0.19.1",
					output:          testDir + "output/" + testFile + ".tf",
					organization:    "exampleOrg",
					dockerConn:      "exampleDockerConn",
					kubeName:        "exampleNamespace",
					kubeConn:        "exampleKubeConn",
					downgrade:       true,
					orgSecrets:      true,
				}

				if pipelineType == "docker" {
					c = dockerTerraformCommand
				}
				if pipelineType == "kubernetes" {
					c = kubernetesTerraformCommand
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
}
