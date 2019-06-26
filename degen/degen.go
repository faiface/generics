package degen

import (
	"local/generics/go/ast"
	"local/generics/go/token"
)

func degenTypeSpec(cfg *config, spec *ast.TypeSpec) bool {
	if len(spec.Params) > 0 {
		panic("cannot degenerate a generic type")
	}

	typ, changedTyp := degenNode(cfg, spec.Type)

	cfg.output.Decls = append(cfg.output.Decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{&ast.TypeSpec{
			Name:   spec.Name,
			Assign: spec.Assign,
			Type:   typ.(ast.Expr),
		}},
	})

	return changedTyp
}

func degenFuncDecl(cfg *config, fdecl *ast.FuncDecl) bool {
	if len(fdecl.TypeParams) > 0 {
		panic("cannot degenerate a generic function")
	}

	recv, changedRecv := degenNode(cfg, fdecl.Recv)
	typ, changedType := degenNode(cfg, fdecl.Type)
	body, changedBody := degenNode(cfg, fdecl.Body)

	cfg.output.Decls = append(cfg.output.Decls, &ast.FuncDecl{
		Recv: recv.(*ast.FieldList),
		Name: fdecl.Name,
		Type: typ.(*ast.FuncType),
		Body: body.(*ast.BlockStmt),
	})

	return changedBody || changedRecv || changedType
}

func degenStmtList(cfg *config, stmts []ast.Stmt) ([]ast.Stmt, bool) {
	var degenStmts []ast.Stmt
	changed := false
	for _, stmt := range stmts {
		degenStmt, didChange := degenNode(cfg, stmt)
		degenStmts = append(degenStmts, degenStmt.(ast.Stmt))
		changed = changed || didChange
	}
	return degenStmts, changed
}

func degenExprList(cfg *config, exprs []ast.Expr) ([]ast.Expr, bool) {
	var degenExprs []ast.Expr
	changed := false
	for _, expr := range exprs {
		degenExpr, changedExpr := degenNode(cfg, expr)
		degenExprs = append(degenExprs, degenExpr.(ast.Expr))
		changed = changed || changedExpr
	}
	return degenExprs, changed
}

func degenNode(cfg *config, node ast.Node) (ast.Node, bool) {
	switch node := node.(type) {
	default:
		return node, false
	case
		*ast.CommentGroup, *ast.BadExpr, *ast.Ident, *ast.BasicLit,
		*ast.SelectorExpr, *ast.BadStmt, *ast.EmptyStmt,
		*ast.BranchStmt, *ast.ImportSpec, *ast.BadDecl:
		return node, false

	case *ast.Field:
		degenType, changedType := degenNode(cfg, node.Type)
		return &ast.Field{
			Names: node.Names,
			Type:  degenType.(ast.Expr),
			Tag:   node.Tag,
		}, changedType

	case *ast.FieldList:
		if node == nil {
			return (*ast.FieldList)(nil), false
		}
		var degenList []*ast.Field
		changed := false
		for _, f := range node.List {
			degenF, changedF := degenNode(cfg, f)
			degenList = append(degenList, degenF.(*ast.Field))
			changed = changed || changedF
		}
		return &ast.FieldList{
			List: degenList,
		}, changed

	case *ast.Ellipsis:
		degenExpr, changedExpr := degenNode(cfg, node.Elt)
		return &ast.Ellipsis{
			Elt: degenExpr.(ast.Expr),
		}, changedExpr

	case *ast.FuncLit:
		var (
			degenType, changedType = degenNode(cfg, node.Type)
			degenBody, changedBody = degenNode(cfg, node.Body)
		)
		return &ast.FuncLit{
			Type: degenType.(*ast.FuncType),
			Body: degenBody.(*ast.BlockStmt),
		}, changedType || changedBody

	case *ast.CompositeLit:
		var (
			degenType, changedType = degenNode(cfg, node.Type)
			degenElts, changedElts = degenExprList(cfg, node.Elts)
		)
		return &ast.CompositeLit{
			Type: degenType.(ast.Expr),
			Elts: degenElts,
		}, changedType || changedElts

	case *ast.ParenExpr:
		degenX, changed := degenNode(cfg, node.X)
		return &ast.ParenExpr{
			X: degenX.(ast.Expr),
		}, changed

	case *ast.IndexExpr:
		var (
			degenX, changedX         = degenNode(cfg, node.X)
			degenIndex, changedIndex = degenNode(cfg, node.Index)
		)
		return &ast.IndexExpr{
			X:     degenX.(ast.Expr),
			Index: degenIndex.(ast.Expr),
		}, changedX || changedIndex

	case *ast.SliceExpr:
		var (
			degenX, changedX       = degenNode(cfg, node.X)
			degenLow, changedLow   = degenNode(cfg, node.Low)
			degenHigh, changedHigh = degenNode(cfg, node.High)
			degenMax, changedMax   = degenNode(cfg, node.Max)
		)
		return &ast.SliceExpr{
			X:      degenX.(ast.Expr),
			Low:    maybeNil(degenLow),
			High:   maybeNil(degenHigh),
			Max:    maybeNil(degenMax),
			Slice3: node.Slice3,
		}, changedX || changedLow || changedHigh || changedMax

	case *ast.TypeAssertExpr:
		var (
			degenX, changedX       = degenNode(cfg, node.X)
			degenType, changedType = degenNode(cfg, node.Type)
		)
		return &ast.TypeAssertExpr{
			X:    degenX.(ast.Expr),
			Type: degenType.(ast.Expr),
		}, changedX || changedType

	case *ast.CallExpr:
		var (
			degenFun, changedFun   = degenNode(cfg, node.Fun)
			degenArgs, changedArgs = degenExprList(cfg, node.Args)
		)

		genericInstance, isInstance := cfg.info.GenericInstances[node]
		if isInstance {
			typeSpec := degenFun.(*ast.Ident).Obj.Decl.(*ast.TypeSpec)
			instName := instTypeSpec(cfg, genericInstance, typeSpec, node)
			return &ast.Ident{
				Name: instName,
			}, true
		}

		genericCall, isCall := cfg.info.GenericCalls[node]
		if isCall {
			funcDecl := degenFun.(*ast.Ident).Obj.Decl.(*ast.FuncDecl)
			instName := instFuncDecl(cfg, genericCall, funcDecl)
			return &ast.CallExpr{
				Fun:      &ast.Ident{Name: instName},
				Args:     degenArgs[genericCall.NumUnnamed:],
				Ellipsis: node.Ellipsis,
			}, true
		}

		return &ast.CallExpr{
			Fun:      degenFun.(ast.Expr),
			Args:     degenArgs,
			Ellipsis: node.Ellipsis,
		}, changedFun || changedArgs

	case *ast.StarExpr:
		degenX, changedX := degenNode(cfg, node.X)
		return &ast.StarExpr{
			X: degenX.(ast.Expr),
		}, changedX

	case *ast.UnaryExpr:
		degenX, changedX := degenNode(cfg, node.X)
		return &ast.UnaryExpr{
			Op: node.Op,
			X:  degenX.(ast.Expr),
		}, changedX

	case *ast.BinaryExpr:
		var (
			degenX, changedX = degenNode(cfg, node.X)
			degenY, changedY = degenNode(cfg, node.Y)
		)
		return &ast.BinaryExpr{
			X:  degenX.(ast.Expr),
			Op: node.Op,
			Y:  degenY.(ast.Expr),
		}, changedX || changedY

	case *ast.KeyValueExpr:
		var (
			degenKey, changedKey     = degenNode(cfg, node.Key)
			degenValue, changedValue = degenNode(cfg, node.Value)
		)
		return &ast.KeyValueExpr{
			Key:   degenKey.(ast.Expr),
			Value: degenValue.(ast.Expr),
		}, changedKey || changedValue

	case *ast.ArrayType:
		var (
			degenLen, changedLen = degenNode(cfg, node.Len)
			degenElt, changedElt = degenNode(cfg, node.Elt)
		)
		return &ast.ArrayType{
			Len: maybeNil(degenLen),
			Elt: degenElt.(ast.Expr),
		}, changedLen || changedElt

	case *ast.StructType:
		degenFields, changedFields := degenNode(cfg, node.Fields)
		return &ast.StructType{
			Fields:     degenFields.(*ast.FieldList),
			Incomplete: node.Incomplete,
		}, changedFields

	case *ast.FuncType:
		var (
			degenParams, changedParams   = degenNode(cfg, node.Params)
			degenResults, changedResults = degenNode(cfg, node.Results)
		)
		return &ast.FuncType{
			Params:  degenParams.(*ast.FieldList),
			Results: degenResults.(*ast.FieldList),
		}, changedParams || changedResults

	case *ast.InterfaceType:
		degenMethods, changedMethods := degenNode(cfg, node.Methods)
		return &ast.InterfaceType{
			Methods:    degenMethods.(*ast.FieldList),
			Incomplete: node.Incomplete,
		}, changedMethods

	case *ast.MapType:
		var (
			degenKey, changedKey     = degenNode(cfg, node.Key)
			degenValue, changedValue = degenNode(cfg, node.Value)
		)
		return &ast.MapType{
			Key:   degenKey.(ast.Expr),
			Value: degenValue.(ast.Expr),
		}, changedKey || changedValue

	case *ast.ChanType:
		degenValue, changedValue := degenNode(cfg, node.Value)
		return &ast.ChanType{
			Dir:   node.Dir,
			Value: degenValue.(ast.Expr),
		}, changedValue

	case *ast.TypeParam:
		panic("unexpected type parameter")

	case *ast.DeclStmt:
		degenDecl, changedDecl := degenNode(cfg, node.Decl)
		return &ast.DeclStmt{
			Decl: degenDecl.(ast.Decl),
		}, changedDecl

	case *ast.LabeledStmt:
		degenStmt, changedStmt := degenNode(cfg, node.Stmt)
		return &ast.LabeledStmt{
			Label: node.Label,
			Stmt:  degenStmt.(ast.Stmt),
		}, changedStmt

	case *ast.ExprStmt:
		degenX, changedX := degenNode(cfg, node.X)
		return &ast.ExprStmt{
			X: degenX.(ast.Expr),
		}, changedX

	case *ast.SendStmt:
		var (
			degenChan, changedChan   = degenNode(cfg, node.Chan)
			degenValue, changedValue = degenNode(cfg, node.Value)
		)
		return &ast.SendStmt{
			Chan:  degenChan.(ast.Expr),
			Value: degenValue.(ast.Expr),
		}, changedChan || changedValue

	case *ast.IncDecStmt:
		degenX, changedX := degenNode(cfg, node.X)
		return &ast.IncDecStmt{
			X:   degenX.(ast.Expr),
			Tok: node.Tok,
		}, changedX

	case *ast.AssignStmt:
		var (
			degenLhs, changedLhs = degenExprList(cfg, node.Lhs)
			degenRhs, changedRhs = degenExprList(cfg, node.Rhs)
		)
		return &ast.AssignStmt{
			Lhs: degenLhs,
			Tok: node.Tok,
			Rhs: degenRhs,
		}, changedLhs || changedRhs

	case *ast.GoStmt:
		degenCall, changedCall := degenNode(cfg, node.Call)
		return &ast.GoStmt{
			Call: degenCall.(*ast.CallExpr),
		}, changedCall

	case *ast.DeferStmt:
		degenCall, changedCall := degenNode(cfg, node.Call)
		return &ast.DeferStmt{
			Call: degenCall.(*ast.CallExpr),
		}, changedCall

	case *ast.ReturnStmt:
		degenResults, changedResults := degenExprList(cfg, node.Results)
		return &ast.ReturnStmt{
			Results: degenResults,
		}, changedResults

	case *ast.BlockStmt:
		degenList, changedList := degenStmtList(cfg, node.List)
		return &ast.BlockStmt{
			List: degenList,
		}, changedList

	case *ast.IfStmt:
		var (
			degenInit, changedInit = degenNode(cfg, node.Init)
			degenCond, changedCond = degenNode(cfg, node.Cond)
			degenBody, changedBody = degenNode(cfg, node.Body)
			degenElse, changedElse = degenNode(cfg, node.Else)
		)
		return &ast.IfStmt{
			Init: maybeNilStmt(degenInit),
			Cond: degenCond.(ast.Expr),
			Body: degenBody.(*ast.BlockStmt),
			Else: maybeNilStmt(degenElse),
		}, changedInit || changedCond || changedBody || changedElse

	case *ast.CaseClause:
		var (
			degenList, changedList = degenExprList(cfg, node.List)
			degenBody, changedBody = degenStmtList(cfg, node.Body)
		)
		return &ast.CaseClause{
			List: degenList,
			Body: degenBody,
		}, changedList || changedBody

	case *ast.SwitchStmt:
		var (
			degenInit, changedInit = degenNode(cfg, node.Init)
			degenTag, changedTag   = degenNode(cfg, node.Tag)
			degenBody, changedBody = degenNode(cfg, node.Body)
		)
		return &ast.SwitchStmt{
			Init: maybeNilStmt(degenInit),
			Tag:  degenTag.(ast.Expr),
			Body: degenBody.(*ast.BlockStmt),
		}, changedInit || changedTag || changedBody

	case *ast.TypeSwitchStmt:
		var (
			degenInit, changedInit     = degenNode(cfg, node.Init)
			degenAssign, changedAssign = degenNode(cfg, node.Assign)
			degenBody, changedBody     = degenNode(cfg, node.Body)
		)
		return &ast.TypeSwitchStmt{
			Init:   maybeNilStmt(degenInit),
			Assign: degenAssign.(ast.Stmt),
			Body:   degenBody.(*ast.BlockStmt),
		}, changedInit || changedAssign || changedBody

	case *ast.CommClause:
		var (
			degenComm, changedComm = degenNode(cfg, node.Comm)
			degenBody, changedBody = degenStmtList(cfg, node.Body)
		)
		return &ast.CommClause{
			Comm: degenComm.(ast.Stmt),
			Body: degenBody,
		}, changedComm || changedBody

	case *ast.SelectStmt:
		degenBody, changedBody := degenNode(cfg, node.Body)
		return &ast.SelectStmt{
			Body: degenBody.(*ast.BlockStmt),
		}, changedBody

	case *ast.ForStmt:
		var (
			degenInit, changedInit = degenNode(cfg, node.Init)
			degenCond, changedCond = degenNode(cfg, node.Cond)
			degenPost, changedPost = degenNode(cfg, node.Post)
			degenBody, changedBody = degenNode(cfg, node.Body)
		)
		return &ast.ForStmt{
			Init: maybeNilStmt(degenInit),
			Cond: maybeNil(degenCond),
			Post: maybeNilStmt(degenPost),
			Body: degenBody.(*ast.BlockStmt),
		}, changedInit || changedCond || changedPost || changedBody

	case *ast.RangeStmt:
		var (
			degenKey, changedKey     = degenNode(cfg, node.Key)
			degenValue, changedValue = degenNode(cfg, node.Value)
			degenX, changedX         = degenNode(cfg, node.X)
			degenBody, changedBody   = degenNode(cfg, node.Body)
		)
		return &ast.RangeStmt{
			Key:   maybeNil(degenKey),
			Value: maybeNil(degenValue),
			Tok:   node.Tok,
			X:     degenX.(ast.Expr),
			Body:  degenBody.(*ast.BlockStmt),
		}, changedKey || changedValue || changedX || changedBody

	case *ast.ValueSpec:
		var (
			degenType, changedType     = degenNode(cfg, node.Type)
			degenValues, changedValues = degenExprList(cfg, node.Values)
		)
		return &ast.ValueSpec{
			Names:  node.Names,
			Type:   maybeNil(degenType),
			Values: degenValues,
		}, changedType || changedValues

	case *ast.TypeSpec:
		if len(node.Params) != 0 {
			panic("cannot degenerate a generic type")
		}
		degenType, changedType := degenNode(cfg, node.Type)
		return &ast.TypeSpec{
			Name:   node.Name,
			Assign: node.Assign,
			Type:   degenType.(ast.Expr),
		}, changedType

	case *ast.GenDecl:
		var degenSpecs []ast.Spec
		changed := false
		for _, spec := range node.Specs {
			degenSpec, changedSpec := degenNode(cfg, spec)
			degenSpecs = append(degenSpecs, degenSpec.(ast.Spec))
			changed = changed || changedSpec
		}
		return &ast.GenDecl{
			Tok:    node.Tok,
			Lparen: node.Lparen,
			Specs:  degenSpecs,
			Rparen: node.Rparen,
		}, changed

	case *ast.FuncDecl:
		panic("unexpected function declaration")
	}
}
