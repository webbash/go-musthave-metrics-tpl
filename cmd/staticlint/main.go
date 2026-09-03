package main

import (
	"github.com/gostaticanalysis/nilerr"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1000"
)

func main() {
	analyzers := []*analysis.Analyzer{
		ErrOsExitAnalyzer,

		// Passes analyzers.
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,

		// Custom analyzers.
		bodyclose.Analyzer,
		nilerr.Analyzer,
	}

	// All analyzers from staticcheck.Analyzers.
	for _, analyzer := range staticcheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// ST1000 — Check for docs declaration.
	analyzers = append(analyzers, st1000.SCAnalyzer.Analyzer)

	multichecker.Main(analyzers...)
}
