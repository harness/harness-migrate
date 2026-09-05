// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

package bridge

import (
	"fmt"
	"strings"

	"github.com/harness/cli/v3/pkg/console"
	"github.com/harness/cli/v3/pkg/hlog"

	"github.com/harness/harness-migrate/internal/tracer"
)

// NewTracer returns the tracer a migration narrates through: an animated
// progress bar on a terminal, plain lines when output is redirected or when the
// caller asks for them.
//
// It reuses internal/tracer rather than adapting to pkg/console. Core's console
// allows one animated stage per process (a package-level spinnerBusy flag) and
// offers no way to end a stage without printing a success or failure marker,
// while the engine opens Start/Stop pairs nested and from errgroup workers, and
// reports failures through Stop as well. internal/tracer already handles that
// shape, so the only thing worth replacing is the log level — see debugVia.
func NewTracer(noProgress bool) tracer.Tracer {
	if noProgress || !console.IsTTY() {
		return debugVia{tracer.NewNoProgress(tracer.LogLevelInfo)}
	}
	return debugVia{tracer.New()}
}

// debugVia hands a tracer's debug output to hlog.
//
// internal/tracer decides at construction whether Debug() logs or discards,
// which means reading the debug flag — and the host's --debug is a BoolFunc that
// calls hlog.SetDebug() without storing anything. Routing Debug() to hlog puts
// the decision where the state already lives, so nothing has to read the flag.
type debugVia struct {
	tracer.Tracer
}

// Debug returns a tracer whose output hlog drops unless --debug was passed.
func (debugVia) Debug() tracer.Tracer { return debugTracer{} }

// WithLevel is a no-op: hlog owns the level.
func (debugVia) WithLevel(tracer.LogLevel) {}

// debugTracer discards its output unless the host was run with --debug.
type debugTracer struct{}

func (debugTracer) Start(format string, args ...interface{}) {
	hlog.Debug(render(format, args...))
}

func (debugTracer) Stop(format string, args ...interface{}) {
	hlog.Debug(render(format, args...))
}

func (debugTracer) Log(format string, args ...interface{}) {
	hlog.Debug(render(format, args...))
}

func (debugTracer) LogError(format string, args ...interface{}) {
	hlog.Debug(render(format, args...))
}

func (debugTracer) Close() {}

func (d debugTracer) Debug() tracer.Tracer { return d }

func (debugTracer) WithLevel(tracer.LogLevel) {}

// render formats a tracer message. The engine passes error constants containing
// %w, which only fmt.Errorf understands — without the swap they would render as
// %!w(...). internal/tracer's own console does the same substitution.
func render(format string, args ...interface{}) string {
	return fmt.Sprintf(strings.ReplaceAll(format, "%w", "%v"), args...)
}
