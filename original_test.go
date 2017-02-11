package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"testing"
)

func TestForDelve(t *testing.T) {
	// prepare fset and source.
	fset := token.NewFileSet()
	source := []byte(`package main
import "fmt"

func main() {
	a := 10 * 20 - 5 * 7
	fmt.Println(a)
}`)

	// test run for delve debug
	parser.ParseFile(fset, "example_test.go", source, parser.Trace)
}

func TestOriginalPackages(t *testing.T) {
	ef := func(_ token.Position, msg string) {
		fmt.Println(msg)
	}

	var s scanner.Scanner
	fset := token.NewFileSet()

	source := []byte(`package main
import "fmt"

func main() {
	var a *int
	a = 2 * 2
	var c , d   string
	fmt.Println(a)
	c = "a"
	b = c
	fmt.Println(b)
}`)

	s.Init(fset.AddFile("./testdata/test.sql", fset.Base(), len(source)),
		source,
		ef,
		scanner.ScanComments)
	fmt.Println("Scan token example.")
	for {
		pos, tok, lit := s.Scan()
		fmt.Printf("%v, %v, %v\n", pos, tok, lit)
		if tok == token.EOF {
			break
		}
	}
	printAst()

	confirmAstPos()
}

func printAst() {
	fset := token.NewFileSet() // positions are relative to fset

	source := []byte(`package main
import "fmt"

func main() {
	
	
	var a *int
	
	
	a=2*2
fmt.Println(a)

	var b  ,  c string
	b = ""
	c = b
	
	fmt.Println(c)

}`)
	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, "example_test.go", source, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	v := f.Decls[1].(*ast.FuncDecl)
	varNode := v.Body.List[3].(*ast.DeclStmt)
	v2 := varNode.Decl.(*ast.GenDecl)
	bNode := v2.Specs[0].(*ast.ValueSpec).Names[0]
	cNode := v2.Specs[0].(*ast.ValueSpec).Names[1]

	fmt.Printf("b.Pos %d, b.End %d, c.Pos %d, c.End %d\n", bNode.Pos(), bNode.End(), cNode.Pos(), cNode.End())

	fmt.Printf("fset after parse file: %v\n", fset)

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	fmt.Printf("%s\n", buf.String())
}

func confirmAstPos() {
	fset := token.NewFileSet() // positions are relative to fset

	source := []byte(`package main

import "fmt"

func main() {
	
	
	var a *int
	
	
	a=2*2
fmt.Println(a)
}`)
	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, "example_test.go", source, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("ast.File.Pos(): %d, End(): %d\n", f.Pos(), f.End())

	fmt.Printf("Decl[0].Pos(): %d, End(): %d\n", f.Decls[0].Pos(), f.Decls[0].End())
	fmt.Printf("Decl[1].Pos(): %d, End(): %d\n", f.Decls[1].Pos(), f.Decls[1].End())

}
