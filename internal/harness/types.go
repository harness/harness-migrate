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

package harness

import "github.com/harness/harness-migrate/internal/types/enum"

type (
	// Pipeline defines a pipeline.
	Pipeline struct {
		Name          string `json:"name"`
		Identifier    string `json:"identifier"`
		Version       int    `json:"version"`
		Numofstages   int    `json:"numOfStages"`
		Createdat     int64  `json:"createdAt"`
		Lastupdatedat int64  `json:"lastUpdatedAt"`
	}

	// Project defines a project.
	Project struct {
		Orgidentifier string   `json:"orgIdentifier"`
		Identifier    string   `json:"identifier"`
		Name          string   `json:"name"`
		Color         string   `json:"color,omitempty"`
		Modules       []string `json:"modules,omitempty"`
		Description   string   `json:"description,omitempty"`
	}

	// Org defines an organization.
	Org struct {
		ID   string `json:"identifier"`
		Name string `json:"name"`
		Desc string `json:"description"`
	}

	// Resource defines a base resource.
	Resource struct {
		Type string      `json:"type"`
		Spec interface{} `json:"spec"`
	}

	// Repository defines a resository.
	Repository struct {
		Identifier    string `json:"identifier"`
		ParentID      int64  `json:"parent_id"`
		Description   string `json:"description"`
		IsPublic      bool   `json:"is_public"`
		DefaultBranch string `json:"default_branch"`
		GitURL        string `json:"git_url"`
	}

	// RepoSettings defines general repository settings which are externally accessible
	RepoSettings struct {
		FileSizeLimit *int64 `json:"file_size_limit"`
	}
)

//
// Secret types
//

const (
	// DefaultSecretType defines the default secret type.
	DefaultSecretType = "SecretText"

	// DefaultSecretType defines the default secret value type.
	DefaultSecretValueType = "Inline"

	// DefaultSecretManager defines the default secret manager.
	DefaultSecretManager = "harnessSecretManager"
)

type (
	Secret struct {
		Name              string      `json:"name"`
		Identifier        string      `json:"identifier"`
		Orgidentifier     string      `json:"orgIdentifier,omitempty"`
		Projectidentifier string      `json:"projectIdentifier,omitempty"`
		Description       string      `json:"description,omitempty"`
		Type              string      `json:"type"` // SecretText
		Spec              *SecretText `json:"spec"`
	}

	SecretText struct {
		Value   *string `json:"value"`
		Type    string  `json:"valueType"`               // Inline
		Manager string  `json:"secretManagerIdentifier"` // harnessSecretManager
	}
)

//
// Connector types
//

const (
	ConnectorTypeGithub = "Github"
	ConnectorTypeGitlab = "Gitlab"
)

type (
	// Connector defines a connector.
	Connector struct {
		Name              string `json:"name"`
		Identifier        string `json:"identifier"`
		Orgidentifier     string `json:"orgIdentifier,omitempty"`
		Projectidentifier string `json:"projectIdentifier,omitempty"`
		Version           int    `json:"version,omitempty"`
		Numofstages       int    `json:"numOfStages,omitempty"`
		Createdat         int64  `json:"createdAt,omitempty"`
		Lastupdatedat     int64  `json:"lastUpdatedAt,omitempty"`

		Type string      `json:"type"` // Gitlab, Github
		Spec interface{} `json:"spec"`
	}

	// ConnectorGitlab defines a Gitlab connector.
	ConnectorGitlab struct {
		Executeondelegate bool      `json:"executeOnDelegate"`
		Type              string    `json:"type"` // Account
		URL               string    `json:"url"`
		Validationrepo    string    `json:"validationRepo,omitempty"`
		Authentication    *Resource `json:"authentication"`
		Apiaccess         *Resource `json:"apiAccess"`
	}

	ConnectorDocker struct {
		ExecuteOnDelegate bool      `json:"executeOnDelegate"`
		DockerRegistryURL string    `json:"dockerRegistryUrl"`
		ProviderType      string    `json:"providerType"`
		Authentication    *Resource `json:"auth"`
	}

	// ConnectorGithub defines a Github connector.
	ConnectorGithub struct {
		Executeondelegate bool      `json:"executeOnDelegate"`
		Type              string    `json:"type"` // Account
		URL               string    `json:"url"`
		Validationrepo    string    `json:"validationRepo,omitempty"`
		Authentication    *Resource `json:"authentication"`
		Apiaccess         *Resource `json:"apiAccess"`
	}

	// ConnectorToken defines connector credentials.
	ConnectorToken struct {
		Username string `json:"username,omitempty"`
		Tokenref string `json:"tokenRef,omitempty"`
	}
)

//
// Error Types
//

// Error represents an API codeerror response.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns the codeerror message.
func (e *Error) Error() string {
	return e.Message
}

//
// Response envelopes
//

type (
	// Response envelope for the Pipeline type.
	pipelineEnvelope struct {
		Status string    `json:"status"`
		Data   *Pipeline `json:"data"`
	}

	// Response envelope for the Project type
	projectEnvelope struct {
		Status string `json:"status"`
		Data   *struct {
			Project        *Project `json:"project"`
			Createdat      int64    `json:"createdAt"`
			Lastmodifiedat int64    `json:"lastModifiedAt"`
			Harnessmanaged bool     `json:"harnessManaged"`
		} `json:"data"`
	}

	// Request envelope for the Project type
	projectCreateEnvelope struct {
		Project *Project `json:"project"`
	}

	// Response envelope for the Org type
	orgEnvelope struct {
		Status string `json:"status"`
		Data   *struct {
			Organization   *Org  `json:"organization"`
			Createdat      int64 `json:"createdAt"`
			Lastmodifiedat int64 `json:"lastModifiedAt"`
			Harnessmanaged bool  `json:"harnessManaged"`
		} `json:"data"`
	}

	// Request envelope for the Org type
	orgCreateEnvelope struct {
		Org *Org `json:"organization"`
	}

	// Response envelope for the Connector type
	connectorEnvelope struct {
		Status string `json:"status"`
		Data   *struct {
			Connector      *Connector `json:"connector"`
			Createdat      int64      `json:"createdAt"`
			Lastmodifiedat int64      `json:"lastModifiedAt"`
			Harnessmanaged bool       `json:"harnessManaged"`
		} `json:"data"`
	}

	// Request envelope for the Connector type
	connectorCreateEnvelope struct {
		Connector *Connector `json:"connector"`
	}

	// Response envelope for the Secret type
	secretEnvelope struct {
		Status string `json:"status"`
		Data   *struct {
			Secret         *Secret `json:"secret"`
			Createdat      int64   `json:"createdAt"`
			Lastmodifiedat int64   `json:"lastModifiedAt"`
			Harnessmanaged bool    `json:"harnessManaged"`
		} `json:"data"`
	}

	// Request envelope for the Secret type
	secretCreateEnvelope struct {
		Secret *Secret `json:"secret"`
	}

	// CreateRepositoryInput defines a repo creation request input.
	CreateRepositoryInput struct {
		Identifier    string `json:"identifier"`
		DefaultBranch string `json:"default_branch"`
		IsPublic      bool   `json:"is_public"`
	}

	// CreateGitnessRepositoryInput defines a repo creation request input for gitness.
	CreateGitnessRepositoryInput struct {
		CreateRepositoryInput
		ParentRef string `json:"parent_ref"`
	}

	// CreateRepositoryForMigrateInput defines a repo creation request input for migration.
	CreateRepositoryForMigrateInput struct {
		Identifier    string `json:"identifier"`
		DefaultBranch string `json:"default_branch"`
		IsPublic      bool   `json:"is_public"`
		ParentRef     string `json:"parent_ref"`
	}

	// UpdateRepositoryStateInput defines a repo update state request input.
	UpdateRepositoryStateInput struct {
		State enum.RepoState `json:"state"`
	}

	UpdateDefaultBranchInput struct {
		Name string `json:"name"`
	}
)
