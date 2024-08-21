// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Report struct {
	name   string
	report map[string]int
	errors map[string]*Error
}

type Error struct {
	error map[string]string
}

func Init(name string) *Report {
	return &Report{
		name:   name,
		report: make(map[string]int),
		errors: make(map[string]*Error),
	}
}

// ReportMetric to report metric for a type
func (r *Report) ReportMetric(typ string, value int) {
	r.report[typ] = value
}

// ReportError can be used to report error for a typ and key for that type with an error msg
// If a key is reported twice it will be overwritten.
func (r *Report) ReportError(typ string, key string, error string) {
	m, ok := r.errors[typ]
	if ok {
		m.error[key] = error
	}
	r.errors[typ] = &Error{error: make(map[string]string)}
	m = r.errors[typ]
	m.error[key] = error
}

func (r *Report) ReportErrors(typ string, key string, errors []string) {
	if len(errors) == 0 {
		return
	}

	m, ok := r.errors[typ]
	if ok {
		m.error[key] = strings.Join(errors, ",")
	}
	r.errors[typ] = &Error{error: make(map[string]string)}
	m = r.errors[typ]
	m.error[key] = strings.Join(errors, ",")
}
func (r *Report) PublishReport() {
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	fmt.Println("")
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Type", "Success", "Error(s)"}, rowConfigAutoMerge)
	var rows []table.Row
	for k, v := range r.report {
		errorCount := 0
		val, ok := r.errors[k]
		if ok {
			errorCount = len(val.error)
		}
		rows = append(rows, table.Row{k, v, errorCount})
	}
	t.SetOutputMirror(os.Stdout)
	t.AppendRows(rows)
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignHeader: text.AlignCenter},
	})
	t.Render()
}
