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

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
)

// Tracer defines tracing methods.
type Tracer interface {
	// Start the named tracer.
	Start(format string, args ...interface{})

	// Stop the named tracer.
	Stop(format string, args ...interface{})

	// Log a message.
	Log(format string, args ...interface{})

	// LogError logs an error message.
	LogError(format string, args ...interface{})

	// Close the tracer.
	Close()

	Debug() Tracer

	// WithLevel is used to set tracer log level at creation
	WithLevel(level LogLevel)
}

// Default returns the default tracer (noop)
func Default() Tracer {
	return new(none)
}

// none implements a noop tracer.
type none struct{}

func (none) Start(format string, args ...interface{})    {}
func (none) Stop(format string, args ...interface{})     {}
func (none) Log(format string, args ...interface{})      {}
func (none) LogError(format string, args ...interface{}) {}
func (none) Close()                                      {}
func (n none) Debug() Tracer {
	return n
}
func (n none) WithLevel(level LogLevel) {}
