package degen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/faiface/generics/go/ast"
	"github.com/faiface/generics/go/token"
	"github.com/faiface/generics/go/types"
)

func instTypeSpec(cfg *config, genInst *types.GenericInstance, spec *ast.TypeSpec, expr ast.Expr) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s", spec.Name)

	var typeParams []*types.TypeParam
	for param := range genInst.Mapping {
		typeParams = append(typeParams, param)
	}

	sort.Slice(typeParams, func(i, j int) bool {
		return typeParams[i].Name() < typeParams[j].Name()
	})

	for _, param := range typeParams {
		fmt.Fprintf(&b, "_")
		writeType(&b, genInst.Mapping[param])
	}

	name := b.String()

	if cfg.instantiated[name] {
		return name
	}
	cfg.instantiated[name] = true

	result := &ast.TypeSpec{
		Name:   &ast.Ident{Name: name},
		Assign: spec.Assign,
		Type:   instNode(cfg, genInst.Mapping, spec.Type).(ast.Expr),
	}

	cfg.output.Decls = append(cfg.output.Decls, &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{result},
	})

	// instantiate fitting associated methods
	for _, decl := range cfg.input.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if len(decl.TypeParams) == 0 {
				continue
			}
			if decl.Recv.NumFields() == 0 {
				continue
			}

			_, _, _, mapping := types.LookupFieldOrMethod(
				cfg.info.TypeOf(expr),
				true,
				cfg.info.ObjectOf(decl.Name).Pkg(),
				decl.Name.Name,
			)
			if mapping == nil {
				continue
			}

			instMethodDecl(cfg, mapping, name, decl)
		}
	}

	return name
}

func instFuncDecl(cfg *config, genCall *types.GenericCall, fdecl *ast.FuncDecl) string {
	name := fdecl.Name.Name

	if fdecl.Recv.NumFields() == 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%s", fdecl.Name.Name)

		var typeParams []*types.TypeParam
		for param := range genCall.Mapping {
			typeParams = append(typeParams, param)
		}

		sort.Slice(typeParams, func(i, j int) bool {
			return typeParams[i].Name() < typeParams[j].Name()
		})

		for _, param := range typeParams {
			fmt.Fprintf(&b, "_")
			writeType(&b, genCall.Mapping[param])
		}

		name = b.String()

		if cfg.instantiated[name] {
			return name
		}
		cfg.instantiated[name] = true
	}

	cfg.output.Decls = append(cfg.output.Decls, &ast.FuncDecl{
		Recv: instNode(cfg, genCall.Mapping, fdecl.Recv).(*ast.FieldList),
		Name: &ast.Ident{Name: name},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: instFieldList(
					cfg, genCall.Mapping,
					fdecl.Type.Params.List[genCall.NumUnnamed:],
				),
			},
			Results: instNode(cfg, genCall.Mapping, fdecl.Type.Results).(*ast.FieldList),
		},
		Body: instNode(cfg, genCall.Mapping, fdecl.Body).(*ast.BlockStmt),
	})

	return name
}

func instMethodDecl(cfg *config, mapping map[*types.TypeParam]types.Type, recvName string, fdecl *ast.FuncDecl) {
	var recv ast.Expr = &ast.Ident{
		Name: recvName,
	}
	if _, ok := fdecl.Recv.List[0].Type.(*ast.StarExpr); ok {
		recv = &ast.StarExpr{
			X: recv,
		}
	}

	cfg.output.Decls = append(cfg.output.Decls, &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			&ast.Field{
				Names: fdecl.Recv.List[0].Names,
				Type:  recv,
			},
		}},
		Name: fdecl.Name,
		Type: instNode(cfg, mapping, fdecl.Type).(*ast.FuncType),
		Body: instNode(cfg, mapping, fdecl.Body).(*ast.BlockStmt),
	})
}

func instFieldList(cfg *config, mapping map[*types.TypeParam]types.Type, list []*ast.Field) []*ast.Field {
	var instList []*ast.Field
	for _, field := range list {
		instList = append(instList, &ast.Field{
			Names: field.Names,
			Type:  instNode(cfg, mapping, field.Type).(ast.Expr),
			Tag:   field.Tag,
		})
	}
	return instList
}

func instStmtList(cfg *config, mapping map[*types.TypeParam]types.Type, stmts []ast.Stmt) []ast.Stmt {
	var instStmts []ast.Stmt
	for _, stmt := range stmts {
		instStmts = append(instStmts, instNode(cfg, mapping, stmt).(ast.Stmt))
	}
	return instStmts
}

func instExprList(cfg *config, mapping map[*types.TypeParam]types.Type, exprs []ast.Expr) []ast.Expr {
	var instExprs []ast.Expr
	for _, expr := range exprs {
		instExprs = append(instExprs, instNode(cfg, mapping, expr).(ast.Expr))
	}
	return instExprs
}

func maybeNil(node ast.Node) ast.Expr {
	if node == nil {
		return nil
	}
	return node.(ast.Expr)
}

func maybeNilStmt(node ast.Node) ast.Stmt {
	if node == nil {
		return nil
	}
	return node.(ast.Stmt)
}

func instNode(cfg *config, mapping map[*types.TypeParam]types.Type, node ast.Node) ast.Node {
	switch node := node.(type) {
	default:
		return node

	case
		*ast.Comment, *ast.CommentGroup, *ast.BadExpr, *ast.BasicLit,
		*ast.BadStmt, *ast.EmptyStmt, *ast.BranchStmt, *ast.ImportSpec, *ast.BadDecl:
		return node

	case *ast.Field:
		return &ast.Field{
			Names: node.Names,
			Type:  instNode(cfg, mapping, node.Type).(ast.Expr),
			Tag:   node.Tag,
		}

	case *ast.FieldList:
		if node == nil {
			return (*ast.FieldList)(nil)
		}
		return &ast.FieldList{
			List: instFieldList(cfg, mapping, node.List),
		}

	case *ast.Ident:
		typ, ok := cfg.info.Types[node]
		if !ok {
			return node
		}
		if !typ.IsType() {
			return node
		}
		typeParam, ok := typ.Type.(*types.TypeParam)
		if !ok {
			return node
		}
		replacement, ok := mapping[typeParam]
		if !ok {
			panic("no replacement for a generic type")
		}
		return typeToExpr(replacement)

	case *ast.SelectorExpr:
		return &ast.SelectorExpr{
			X:   instNode(cfg, mapping, node.X).(ast.Expr),
			Sel: node.Sel,
		}

	case *ast.Ellipsis:
		return &ast.Ellipsis{
			Elt: instNode(cfg, mapping, node.Elt).(ast.Expr),
		}

	case *ast.FuncLit:
		return &ast.FuncLit{
			Type: instNode(cfg, mapping, node.Type).(*ast.FuncType),
			Body: instNode(cfg, mapping, node.Body).(*ast.BlockStmt),
		}

	case *ast.CompositeLit:
		return &ast.CompositeLit{
			Type: instNode(cfg, mapping, node.Type).(ast.Expr),
			Elts: instExprList(cfg, mapping, node.Elts),
		}

	case *ast.ParenExpr:
		return &ast.ParenExpr{
			X: instNode(cfg, mapping, node.X).(ast.Expr),
		}

	case *ast.IndexExpr:
		return &ast.IndexExpr{
			X:     instNode(cfg, mapping, node.X).(ast.Expr),
			Index: instNode(cfg, mapping, node.Index).(ast.Expr),
		}

	case *ast.SliceExpr:
		return &ast.SliceExpr{
			X:      instNode(cfg, mapping, node.X).(ast.Expr),
			Low:    maybeNil(instNode(cfg, mapping, node.Low)),
			High:   maybeNil(instNode(cfg, mapping, node.High)),
			Max:    maybeNil(instNode(cfg, mapping, node.Max)),
			Slice3: node.Slice3,
		}

	case *ast.TypeAssertExpr:
		return &ast.TypeAssertExpr{
			X:    instNode(cfg, mapping, node.X).(ast.Expr),
			Type: instNode(cfg, mapping, node.Type).(ast.Expr),
		}

	case *ast.CallExpr:
		return &ast.CallExpr{
			Fun:      instNode(cfg, mapping, node.Fun).(ast.Expr),
			Args:     instExprList(cfg, mapping, node.Args),
			Ellipsis: node.Ellipsis,
		}

	case *ast.StarExpr:
		return &ast.StarExpr{
			X: instNode(cfg, mapping, node.X).(ast.Expr),
		}

	case *ast.UnaryExpr:
		return &ast.UnaryExpr{
			Op: node.Op,
			X:  instNode(cfg, mapping, node.X).(ast.Expr),
		}

	case *ast.BinaryExpr:
		return &ast.BinaryExpr{
			X:  instNode(cfg, mapping, node.X).(ast.Expr),
			Op: node.Op,
			Y:  instNode(cfg, mapping, node.Y).(ast.Expr),
		}

	case *ast.KeyValueExpr:
		return &ast.KeyValueExpr{
			Key:   instNode(cfg, mapping, node.Key).(ast.Expr),
			Value: instNode(cfg, mapping, node.Value).(ast.Expr),
		}

	case *ast.ArrayType:
		return &ast.ArrayType{
			Len: maybeNil(instNode(cfg, mapping, node.Len)),
			Elt: instNode(cfg, mapping, node.Elt).(ast.Expr),
		}

	case *ast.StructType:
		return &ast.StructType{
			Fields:     instNode(cfg, mapping, node.Fields).(*ast.FieldList),
			Incomplete: node.Incomplete,
		}

	case *ast.FuncType:
		return &ast.FuncType{
			Params:  instNode(cfg, mapping, node.Params).(*ast.FieldList),
			Results: instNode(cfg, mapping, node.Results).(*ast.FieldList),
		}

	case *ast.InterfaceType:
		return &ast.InterfaceType{
			Methods:    instNode(cfg, mapping, node.Methods).(*ast.FieldList),
			Incomplete: node.Incomplete,
		}

	case *ast.MapType:
		return &ast.MapType{
			Key:   instNode(cfg, mapping, node.Key).(ast.Expr),
			Value: instNode(cfg, mapping, node.Value).(ast.Expr),
		}

	case *ast.ChanType:
		return &ast.ChanType{
			Dir:   node.Dir,
			Value: instNode(cfg, mapping, node.Value).(ast.Expr),
		}

	case *ast.TypeParam:
		replacement, ok := mapping[cfg.info.TypeOf(node).(*types.TypeParam)]
		if !ok {
			panic("no replacement for a generic type")
		}
		return typeToExpr(replacement)

	case *ast.DeclStmt:
		return &ast.DeclStmt{
			Decl: instNode(cfg, mapping, node.Decl).(ast.Decl),
		}

	case *ast.LabeledStmt:
		return &ast.LabeledStmt{
			Label: node.Label,
			Stmt:  instNode(cfg, mapping, node.Stmt).(ast.Stmt),
		}

	case *ast.ExprStmt:
		return &ast.ExprStmt{
			X: instNode(cfg, mapping, node.X).(ast.Expr),
		}

	case *ast.SendStmt:
		return &ast.SendStmt{
			Chan:  instNode(cfg, mapping, node.Chan).(ast.Expr),
			Value: instNode(cfg, mapping, node.Value).(ast.Expr),
		}

	case *ast.IncDecStmt:
		return &ast.IncDecStmt{
			X:   instNode(cfg, mapping, node.X).(ast.Expr),
			Tok: node.Tok,
		}

	case *ast.AssignStmt:
		return &ast.AssignStmt{
			Lhs: instExprList(cfg, mapping, node.Lhs),
			Tok: node.Tok,
			Rhs: instExprList(cfg, mapping, node.Rhs),
		}

	case *ast.GoStmt:
		return &ast.GoStmt{
			Call: instNode(cfg, mapping, node.Call).(*ast.CallExpr),
		}

	case *ast.DeferStmt:
		return &ast.DeferStmt{
			Call: instNode(cfg, mapping, node.Call).(*ast.CallExpr),
		}

	case *ast.ReturnStmt:
		return &ast.ReturnStmt{
			Results: instExprList(cfg, mapping, node.Results),
		}

	case *ast.BlockStmt:
		return &ast.BlockStmt{
			List: instStmtList(cfg, mapping, node.List),
		}

	case *ast.IfStmt:
		return &ast.IfStmt{
			Init: maybeNilStmt(instNode(cfg, mapping, node.Init)),
			Cond: instNode(cfg, mapping, node.Cond).(ast.Expr),
			Body: instNode(cfg, mapping, node.Body).(*ast.BlockStmt),
			Else: maybeNilStmt(instNode(cfg, mapping, node.Else)),
		}

	case *ast.CaseClause:
		return &ast.CaseClause{
			List: instExprList(cfg, mapping, node.List),
			Body: instStmtList(cfg, mapping, node.Body),
		}

	case *ast.SwitchStmt:
		return &ast.SwitchStmt{
			Init: maybeNilStmt(instNode(cfg, mapping, node.Init)),
			Tag:  instNode(cfg, mapping, node.Tag).(ast.Expr),
			Body: instNode(cfg, mapping, node.Body).(*ast.BlockStmt),
		}

	case *ast.TypeSwitchStmt:
		return &ast.TypeSwitchStmt{
			Init:   maybeNilStmt(instNode(cfg, mapping, node.Init)),
			Assign: instNode(cfg, mapping, node.Assign).(ast.Stmt),
			Body:   instNode(cfg, mapping, node.Body).(*ast.BlockStmt),
		}

	case *ast.CommClause:
		return &ast.CommClause{
			Comm: instNode(cfg, mapping, node.Comm).(ast.Stmt),
			Body: instStmtList(cfg, mapping, node.Body),
		}

	case *ast.SelectStmt:
		return &ast.SelectStmt{
			Body: instNode(cfg, mapping, node.Body).(*ast.BlockStmt),
		}

	case *ast.ForStmt:
		return &ast.ForStmt{
			Init: maybeNilStmt(instNode(cfg, mapping, node.Init)),
			Cond: maybeNil(instNode(cfg, mapping, node.Cond)),
			Post: maybeNilStmt(instNode(cfg, mapping, node.Post)),
			Body: instNode(cfg, mapping, node.Body).(*ast.BlockStmt),
		}

	case *ast.RangeStmt:
		return &ast.RangeStmt{
			Key:   maybeNil(instNode(cfg, mapping, node.Key)),
			Value: maybeNil(instNode(cfg, mapping, node.Value)),
			Tok:   node.Tok,
			X:     instNode(cfg, mapping, node.X).(ast.Expr),
			Body:  instNode(cfg, mapping, node.Body).(*ast.BlockStmt),
		}

	case *ast.ValueSpec:
		return &ast.ValueSpec{
			Names:  node.Names,
			Type:   maybeNil(instNode(cfg, mapping, node.Type)),
			Values: instExprList(cfg, mapping, node.Values),
		}

	case *ast.TypeSpec:
		return &ast.TypeSpec{
			Name: node.Name,
			Type: instNode(cfg, mapping, node.Type).(ast.Expr),
		}

	case *ast.GenDecl:
		var instSpecs []ast.Spec
		for _, spec := range node.Specs {
			instSpecs = append(instSpecs, instNode(cfg, mapping, spec).(ast.Spec))
		}
		return &ast.GenDecl{
			Tok:    node.Tok,
			Lparen: node.Lparen,
			Specs:  instSpecs,
			Rparen: node.Rparen,
		}

	case *ast.FuncDecl:
		panic("unexpected function declaration")
	}
}
