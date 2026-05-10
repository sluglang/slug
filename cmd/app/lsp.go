package main

import (
	"os"
	"slug/internal/lexer"
	"slug/internal/lsp"
	"slug/internal/parser"
	"slug/internal/semantic"
)

func runLanguageServer(typeCheck bool, typeCheckTrace bool) error {
	analyze := func(path, src string) ([]string, []string) {
		l := lexer.New(src)
		p := parser.New(l, path, src)
		program := p.ParseProgram()
		errs := append([]string{}, p.Errors()...)
		if !typeCheck {
			return errs, nil
		}
		semErrs, semWarns := semantic.AnalyzeWithOptions(path, src, program, semantic.AnalyzeOptions{
			EnableTypeCheck: typeCheck,
			TypeCheckTrace:  typeCheckTrace,
			TraceWriter:     os.Stderr,
		})
		errOut := append(errs, semErrs...)
		return errOut, semWarns
	}
	server := lsp.NewServer(os.Stdin, os.Stdout, analyze)
	return server.Run()
}
