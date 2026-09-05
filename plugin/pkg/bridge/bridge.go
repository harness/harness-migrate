// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

// Package bridge adapts harness-migrate's engine interfaces to the unified CLI
// runtime. The engine takes its collaborators as parameters — a tracer for
// narration, a context for cancellation — so a plugin command has to supply
// implementations backed by the host's facilities rather than the standalone
// binary's.
//
// Everything here is command-independent and shared by every handler in
// pkg/migrateplugin. Nothing under ../../../internal changes.
package bridge
