// Package main includes multichecker for static analysis of the source code.
//
// Static analysis includes following analyzers:
//
//   - Printf
//
//   - Shadow
//
//   - Shift
//
//   - Ineffassign
//
//   - Errcheck
//
//   - Structtag
//
//   - Ossignal
//
//   - Staticcheck analyzers
//
//     To run static analysis build cmd/staticlint directory and run built binary with provided package directory to analyze.
//     Use help flag to learn about analyzer-specific options.
package main

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {

	staticChecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		ineffassign.Analyzer,
		errcheck.Analyzer,
		structtag.Analyzer,
		OsSignalAnalyzer,
	}

	for _, v := range staticcheck.Analyzers {
		staticChecks = append(staticChecks, v.Analyzer)
	}

	multichecker.Main(
		staticChecks...,
	)
}
