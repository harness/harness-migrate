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

// Package types provides common types for data export and import.

package types

const (
	MultiPartFileField = "file"
	MultiPartDataField = "data"
)

type (
	// RepositoriesImportInput is input object for importing repo(s).
	RepositoriesImportInput struct {
		RequestId string `json:"request_id"`
	}

	// RepositoriesImportOutput is output object for importing repo(s).
	RepositoriesImportOutput struct {
		RequestId string `json:"request_id"`
		Users     struct {
			NotPresent []string `json:"not_present"`
		} `json:"users,omitempty"`
	}

	// RepositoryUsersImportInput is object for creating/rejecting user invite during repo(s) import.
	RepositoryUsersImportInput struct {
		Emails        []string `json:"emails"`
		MapToImporter bool     `json:"map_to_importer"`
	}

	// RepositoryUsersImportOutput is object for reporting back user import.
	RepositoryUsersImportOutput struct {
		Success      bool   `json:"success"`
		ErrorMessage string `json:"error_message"`
	}

	// RepositoryImportStatus is object for reporting back status of import.
	RepositoryImportStatus struct {
		Status RepoImportStatus `json:"status"`
	}
)

type RepoImportStatus string

const (
	RepoImportStatusComplete   RepoImportStatus = "complete"
	RepoImportStatusInProgress RepoImportStatus = "in_progress"
)
