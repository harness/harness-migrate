// Copyright 2023 Harness Inc. All rights reserved.

package gitlab

import (
	"os"

	"golang.org/x/exp/slog"
)

// helper function creates a logger
func createLogger(debug bool) slog.Logger {
	opts := new(slog.HandlerOptions)
	if debug {
		opts.Level = slog.DebugLevel
	}
	return slog.New(
		opts.NewTextHandler(os.Stdout),
	)
}
