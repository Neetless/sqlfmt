package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/Neetless/sqlfmt/parser"
	printer "github.com/Neetless/sqlfmt/printer"
	"github.com/Neetless/sqlfmt/token"
)

const (
	exitSuccess int = iota
	exitError
)

type formatter struct {
	fset *token.FileSet
	out  io.Writer
}

func main() {
	var outputFilename string
	flag.StringVar(&outputFilename, "o", "", "-o=FILE\twrite documents to FILE")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("requires input source.")
	}

	var fmter formatter

	if outputFilename != "" {
	} else {
		fmter.out = os.Stdout
	}

	fmter.fset = token.NewFileSet()

	code := sqlfmtMain(fmter)
	os.Exit(code)
}

func sqlfmtMain(fmter formatter) int {

	for _, arg := range flag.Args() {
		stmt, err := parser.ParseFile(fmter.fset, arg, nil)
		if err != nil {
			log.Println(err)
			return exitError
		}

		if err := printer.Fprint(fmter.out, fmter.fset, stmt); err != nil {
			log.Println(err)
			return exitError
		}
	}

	return exitSuccess
}
