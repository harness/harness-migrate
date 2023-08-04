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

// Tracer defines tracing methods.
type Tracer interface {
	// Start the named tracer.
	Start(format string, args ...interface{})

	// Stop the named tracer.
	Stop(format string, args ...interface{})

	// Log a message.
	Log(format string, args ...interface{})

	// Close the tracer.
	Close()
}

// Default returns the default tracer (noop)
func Default() Tracer {
	return new(none)
}

// none implements a noop tracer.
type none struct{}

func (*none) Start(format string, args ...interface{}) {}
func (*none) Stop(format string, args ...interface{})  {}
func (*none) Log(format string, args ...interface{})   {}
func (*none) Close()                                   {}
