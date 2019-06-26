package main

import (
	"bytes"
	"flag"
	"fmt"
	"local/generics/degen"
	"local/generics/go/ast"
	"local/generics/go/importer"
	"local/generics/go/parser"
	"local/generics/go/printer"
	"local/generics/go/token"
	"local/generics/go/types"
	"os"
)

var (
	output   = flag.String("out", "out.go", "output file")
	debug    = flag.Bool("debug", false, "prints intermediate type-checking errors to the standard output and other debug info")
	maxStage = flag.Int("maxstage", -1, "maximum number of stages")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags...] <file>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 || flag.Arg(0) == "" {
		flag.Usage()
		return
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(
		fset,
		flag.Arg(0),
		nil,
		parser.DeclarationErrors,
	)
	if err != nil {
		fail(err)
	}

	typesCfg := types.Config{
		Importer: importer.Default(),
	}
	info := types.Info{
		Types:        make(map[ast.Expr]types.TypeAndValue),
		GenericCalls: make(map[*ast.CallExpr]*types.GenericCall),
	}
	_, err = typesCfg.Check("", fset, []*ast.File{file}, &info)
	if err != nil {
		fail(err)
	}

	// degenerate
	for stage := 1; *maxStage < 0 || stage <= *maxStage; stage++ {
		if *debug {
			fmt.Printf("Stage %d.\n", stage)
		}

		var changed bool
		file, changed = degen.Degen(fset, file, *debug)

		var b bytes.Buffer
		err := printer.Fprint(&b, fset, file)
		if err != nil {
			fail(err)
		}

		fset = token.NewFileSet()
		file, err = parser.ParseFile(
			fset,
			flag.Arg(0),
			&b,
			parser.DeclarationErrors,
		)
		if err != nil {
			fail(err)
		}

		if !changed {
			break
		}
	}

	// filter out generic function declarations
	var decls []ast.Decl
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		default:
			decls = append(decls, decl)

		case *ast.FuncDecl:
			if len(decl.TypeParams) == 0 {
				decls = append(decls, decl)
			}

		case *ast.GenDecl:
			if decl.Tok != token.TYPE {
				decls = append(decls, decl)
				continue
			}

			for _, spec := range decl.Specs {
				spec := spec.(*ast.TypeSpec)

				if len(spec.Params) == 0 {
					decls = append(decls, &ast.GenDecl{
						Tok:   decl.Tok,
						Specs: []ast.Spec{spec},
					})
				}
			}
		}
	}
	file.Decls = decls

	outputFile, err := os.Create(*output)
	if err != nil {
		fail(err)
	}
	defer outputFile.Close()
	printer.Fprint(outputFile, fset, file)
}
