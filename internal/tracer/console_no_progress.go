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

package tracer

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type console_no_progress struct {
	time     time.Time
	logLevel LogLevel
}

// Log logs a message to the console.
func (c *console_no_progress) Log(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Println("")
}

// LogError logs an error message to the console.
func (c *console_no_progress) LogError(format string, args ...interface{}) {
	// make sure only print the error message if passed
	modifiedFormat := strings.ReplaceAll(format, "%w", "%v")
	fmt.Println()
	color.Red(modifiedFormat, args...)
}

func NewNoProgress(level LogLevel) *console_no_progress {
	return &console_no_progress{logLevel: level}
}

func (c *console_no_progress) Start(format string, args ...interface{}) {
	c.time = time.Now()
	c.Log(format, args...)
}

func (c *console_no_progress) Stop(format string, args ...interface{}) {
	// make sure only print the error message if passed
	modifiedFormat := strings.ReplaceAll(format, "%w", "%v")
	withTime := modifiedFormat + fmt.Sprintf(" [in %.1f sec]", float32(time.Now().Sub(c.time).Seconds()))
	c.Log(withTime, args...)
	c.Log("")
}

func (c *console_no_progress) Debug() Tracer {
	if c.logLevel == LogLevelDebug {
		return c
	}
	return none{}
}

func (c *console_no_progress) Close() {
}

func (c *console_no_progress) WithLevel(level LogLevel) {
	c.logLevel = level
}
