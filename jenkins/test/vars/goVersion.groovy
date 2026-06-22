#!/usr/bin/env groovy

// Single source of truth for Go version across all Jenkins pipelines.
// Must match the `toolchain` directive in api/go.mod (e.g. "toolchain go1.25.1").
// Installing the exact toolchain version means Go never needs to auto-download
// anything — behaviour is identical to a local developer environment.
// Update this whenever the `toolchain` directive in api/go.mod is updated.
def call() {
    return '1.25.1'
}
