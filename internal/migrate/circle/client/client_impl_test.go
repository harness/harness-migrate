// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

func TestFindOrg(t *testing.T) {
	defer gock.Off()

	gock.New("https://circleci.com").
		Get("/api/private/organization/1af87b45-02e9-467e-be19-e4ae74fc114e").
		Reply(200).
		File("testdata/org_find.json")

	client := New("dummy0d0ac576df34be6a882")
	got, err := client.FindOrg("1af87b45-02e9-467e-be19-e4ae74fc114e")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Org)
	raw, _ := ioutil.ReadFile("testdata/org_find.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindOrgID(t *testing.T) {
	defer gock.Off()

	gock.New("https://circleci.com").
		Get("/api/v2/me/collaborations").
		Reply(200).
		File("testdata/org_search.json")

	client := New("dummy0d0ac576df34be6a882")
	got, err := client.FindOrgID("81700eff-acc7-459b-9c17-2f9e3d937db7")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Org)
	raw, _ := ioutil.ReadFile("testdata/org_search.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindProject(t *testing.T) {
	defer gock.Off()

	gock.New("https://circleci.com").
		Get("/api/v2/project/circleci/216caad37819ng3EFmqdLS/T216caad37819NpwJ368").
		Reply(200).
		File("testdata/project_find.json")

	client := New("dummy0d0ac576df34be6a882")
	got, err := client.FindProject("circleci/216caad37819ng3EFmqdLS/T216caad37819NpwJ368")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Project)
	raw, _ := ioutil.ReadFile("testdata/project_find.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestFindConfig(t *testing.T) {
	defer gock.Off()

	gock.New("https://circleci.com").
		Get("/api/v2/pipeline/1af87b45-02e9-467e-be19-e4ae74fc114e/config").
		Reply(200).
		File("testdata/config_find.json")

	client := New("dummy0d0ac576df34be6a882")
	got, err := client.FindConfig("1af87b45-02e9-467e-be19-e4ae74fc114e")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(Config)
	raw, _ := ioutil.ReadFile("testdata/config_find.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestListProjects(t *testing.T) {
	defer gock.Off()

	gock.New("https://circleci.com").
		Get("/api/private/project").
		MatchParam("organization-id", "386635e9-b87a-4e95-a451-75e3e02b1b93").
		MatchHeader("Circle-Token", "dummy0d0ac576df34be6a882").
		Reply(200).
		File("testdata/project_list.json")

	client := New("dummy0d0ac576df34be6a882")
	got, err := client.ListProjects("386635e9-b87a-4e95-a451-75e3e02b1b93")
	if err != nil {
		t.Error(err)
		return
	}

	want := []*Project{}
	raw, _ := ioutil.ReadFile("testdata/project_list.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestListPipelines(t *testing.T) {
	defer gock.Off()

	gock.New("https://circleci.com").
		Get("/api/v2/project/386635e9-b87a-4e95-a451-75e3e02b1b93/pipeline").
		Reply(200).
		File("testdata/pipeline_list.json")

	client := New("dummy0d0ac576df34be6a882")
	got, err := client.ListPipelines("386635e9-b87a-4e95-a451-75e3e02b1b93")
	if err != nil {
		t.Error(err)
		return
	}

	want := []*Pipeline{}
	raw, _ := ioutil.ReadFile("testdata/pipeline_list.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestListEnvs(t *testing.T) {
	defer gock.Off()

	gock.New("https://circleci.com").
		Get("/api/v2/project/circleci/Gz3a4NZuvL2Eng3EFmqdLS/TaCb55wNToTQJA9NpwJ368/envvar").
		Reply(200).
		File("testdata/env_list.json")

	client := New("dummy0d0ac576df34be6a882")
	got, err := client.ListEnvs("circleci/Gz3a4NZuvL2Eng3EFmqdLS/TaCb55wNToTQJA9NpwJ368")
	if err != nil {
		t.Error(err)
		return
	}

	want := []*Env{}
	raw, _ := ioutil.ReadFile("testdata/env_list.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}
