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

import "strings"

// Option configures a Digital Ocean provider option.
type Option func(*gitnessClient)

// WithAddress returns an option to set the base address.
func WithAddress(address string) Option {
	return func(p *gitnessClient) {
		p.address = strings.TrimSuffix(address, "/")
	}
}

// WithTracing returns an option to enable tracing.
func WithTracing(tracing bool) Option {
	return func(p *gitnessClient) {
		p.tracing = tracing
	}
}
