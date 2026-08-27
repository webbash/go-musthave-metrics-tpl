package main

import (
	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1000"
)

func main() {
	analyzers := []*analysis.Analyzer{
		ErrOsExitAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		bodyclose.Analyzer,
	}

	// All analyzers from staticcheck.Analyzers.
	for _, analyzer := range staticcheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// ST1000 — Check for docs declaration.
	analyzers = append(analyzers, st1000.SCAnalyzer.Analyzer)

	multichecker.Main(analyzers...)
}
