//go:build tools

package tools

/*
 * list of modules required to build and run all tests
 * see: https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
 */

import (
	_ "github.com/daixiang0/gci"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "mvdan.cc/gofumpt"
)
