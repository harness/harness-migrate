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

package gitimporter

import (
	"strings"

	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/tracer"
)

// Importer imports data from gitlab to Harness.
type Importer struct {
	Harness *harness.Client

	HarnessSpace string
	HarnessToken string

	ZipFileLocation string

	Tracer tracer.Tracer

	RequestId string
}

func NewImporter(space, token, location, requestId string, tracer tracer.Tracer) *Importer {
	spaceSplit := strings.Split(space, "/")
	client := harness.New(spaceSplit[0], token)

	return &Importer{
		Harness:         &client,
		HarnessSpace:    space,
		HarnessToken:    token,
		ZipFileLocation: location,
		Tracer:          tracer,
		RequestId:       requestId,
	}
}
