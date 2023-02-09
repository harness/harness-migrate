// Copyright 2023 Harness Inc. All rights reserved.

package harness

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

func TestFindOrg(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/ng/api/organizations/default").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		Reply(200).
		File("testdata/find_org.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	got, err := client.FindOrg("default")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Org)
	raw, _ := ioutil.ReadFile("testdata/find_org.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindProject(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/ng/api/projects/playground").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		MatchParam("orgIdentifier", "default").
		Reply(200).
		File("testdata/find_project.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	got, err := client.FindProject("default", "playground")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Project)
	raw, _ := ioutil.ReadFile("testdata/find_project.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindPipeline(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/pipeline/api/pipelines/summary/testpipeline").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		MatchParam("orgIdentifier", "default").
		MatchParam("projectIdentifier", "playground").
		Reply(200).
		File("testdata/find_pipeline.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	got, err := client.FindPipeline("default", "playground", "testpipeline")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Pipeline)
	raw, _ := ioutil.ReadFile("testdata/find_pipeline.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindSecret(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/ng/api/v2/secrets/password").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		MatchParam("orgIdentifier", "default").
		MatchParam("projectIdentifier", "playground").
		Reply(200).
		File("testdata/find_secret.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	got, err := client.FindSecret("default", "playground", "password")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Secret)
	raw, _ := ioutil.ReadFile("testdata/find_secret.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindSecretOrg(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/ng/api/v2/secrets/password").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		MatchParam("orgIdentifier", "default").
		Reply(200).
		File("testdata/find_secret.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	got, err := client.FindSecretOrg("default", "password")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Secret)
	raw, _ := ioutil.ReadFile("testdata/find_secret.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindConnector(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/ng/api/connectors/gitlab").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		MatchParam("orgIdentifier", "default").
		MatchParam("projectIdentifier", "playground").
		Reply(200).
		File("testdata/find_connector.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	got, err := client.FindConnector("default", "playground", "gitlab")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Connector)
	raw, _ := ioutil.ReadFile("testdata/find_connector.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindConnectorOrg(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/ng/api/connectors/gitlab").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		MatchParam("orgIdentifier", "default").
		Reply(200).
		File("testdata/find_connector.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	got, err := client.FindConnectorOrg("default", "gitlab")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Connector)
	raw, _ := ioutil.ReadFile("testdata/find_connector.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindConnectorNull(t *testing.T) {
	defer gock.Off()

	gock.New("https://app.harness.io").
		Get("/gateway/ng/api/connectors/nonexistent").
		MatchParam("accountIdentifier", "gVcEoNyqQNKbigC_hA3JqA").
		MatchParam("orgIdentifier", "default").
		MatchParam("projectIdentifier", "playground").
		Reply(200).
		File("testdata/find_connector_not_found.json")

	client := New("gVcEoNyqQNKbigC_hA3JqA", "dummy0d0ac576df34be6a882")
	_, err := client.FindConnector("default", "playground", "nonexistent")
	if err == nil {
		t.Errorf("Want not found error, got no error")
	}
}
