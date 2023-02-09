// Copyright 2023 Harness Inc. All rights reserved.

package harness

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
)
