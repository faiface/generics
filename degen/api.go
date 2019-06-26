package degen

import (
	"fmt"
	"local/generics/go/ast"
	"local/generics/go/importer"
	"local/generics/go/token"
	"local/generics/go/types"
)

func Degen(fset *token.FileSet, input *ast.File, debug bool) (output *ast.File, changed bool) {
	typesCfg := &types.Config{
		Importer: importer.Default(),
	}
	info := &types.Info{
		Types:            make(map[ast.Expr]types.TypeAndValue),
		Defs:             make(map[*ast.Ident]types.Object),
		Uses:             make(map[*ast.Ident]types.Object),
		GenericCalls:     make(map[*ast.CallExpr]*types.GenericCall),
		GenericInstances: make(map[*ast.CallExpr]*types.GenericInstance),
	}
	_, err := typesCfg.Check("", fset, []*ast.File{input}, info)
	if err != nil && debug {
		fmt.Println(err)
	}

	output = &ast.File{
		Name:    input.Name,
		Imports: input.Imports,
	}

	cfg := &config{
		info:         info,
		instantiated: make(map[string]bool),
		input:        input,
		output:       output,
	}

	for _, decl := range input.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv.NumFields() == 0 {
				cfg.instantiated[decl.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					cfg.instantiated[spec.Name.Name] = true
				}
			}
		}
	}

	for _, decl := range input.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if len(decl.TypeParams) > 0 {
				output.Decls = append(output.Decls, decl)
				continue
			}
			changed = degenFuncDecl(cfg, decl) || changed

		case *ast.GenDecl:
			if decl.Tok != token.TYPE {
				output.Decls = append(output.Decls, decl)
				continue
			}

			for _, spec := range decl.Specs {
				spec := spec.(*ast.TypeSpec)

				if len(spec.Params) > 0 {
					output.Decls = append(output.Decls, decl)
					continue
				}

				changed = degenTypeSpec(cfg, spec) || changed
			}

		default:
			output.Decls = append(output.Decls, decl)
		}
	}

	return output, changed
}

type config struct {
	info         *types.Info
	instantiated map[string]bool
	input        *ast.File
	output       *ast.File
}
