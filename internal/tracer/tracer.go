// Copyright 2023 Harness Inc. All rights reserved.

package tracer

// Tracer defines tracing methods.
type Tracer interface {
	// Start the named tracer.
	Start(format string, args ...interface{})

	// Stop the named tracer.
	Stop(format string, args ...interface{})

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
func (*none) Close()                                   {}
