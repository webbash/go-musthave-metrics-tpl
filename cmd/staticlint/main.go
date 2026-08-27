package main

import (
	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
)

func main() {
	multichecker.Main(
		ErrOsExitAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer, // добавляем анализатор в вызов multichecker
		structtag.Analyzer,
		errcheck.Analyzer,
		bodyclose.Analyzer,
	)
}
