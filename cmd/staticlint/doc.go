// Package main provides the staticlint multichecker for this project.
//
// The checker is started with a package pattern, for example:
//
//	go run ./cmd/staticlint ./...
//
// The multichecker loads the requested packages, runs every registered
// analyzer, and prints diagnostics with the source position and description.
// A non-zero exit status means that at least one analyzer reported a problem.
//
// The registered analyzers are:
//
//   - osExit: reports a direct os.Exit call in the main function of a main
//     package. Test-generated main packages are ignored.
//   - printf: checks formatting functions and reports mismatched format
//     directives and arguments.
//   - shadow: reports declarations that shadow existing identifiers from an
//     outer scope.
//   - shift: reports invalid or suspicious shift operations, including shifts
//     that overflow the width of their type.
//   - structtag: checks the syntax and consistency of struct field tags.
//   - errcheck: reports returned errors that are silently ignored instead of
//     being handled or explicitly assigned to the blank identifier.
//   - bodyclose: checks that HTTP response bodies are closed after requests.
//   - staticcheck.Analyzers: includes every analyzer in Staticcheck's SA
//     category. These analyzers detect correctness problems and common
//     misuses of the standard library and language features.
//   - ST1000: checks that packages have a correctly formatted package comment.
//
// The SA analyzers are supplied by honnef.co/go/tools/staticcheck. ST1000 is
// supplied by honnef.co/go/tools/stylecheck and is included to cover a class
// other than SA.
package main
