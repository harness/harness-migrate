// Copyright 2023 Harness Inc. All rights reserved.

package client

import "strings"

// Option configures a Digital Ocean provider option.
type Option func(*client)

// WithAddress returns an option to set the base address.
func WithAddress(address string) Option {
	return func(p *client) {
		p.address = strings.TrimSuffix(address, "/")
	}
}

// WithTracing returns an option to enable tracing.
func WithTracing(tracing bool) Option {
	return func(p *client) {
		p.tracing = tracing
	}
}
