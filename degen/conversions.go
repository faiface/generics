package degen

import (
	"fmt"
	"io"
	"local/generics/go/ast"
	"local/generics/go/token"
	"local/generics/go/types"
)

func tupleToFieldList(t *types.Tuple) *ast.FieldList {
	var fields ast.FieldList
	for i := 0; i < t.Len(); i++ {
		v := t.At(i)
		field := &ast.Field{
			Type: typeToExpr(v.Type()),
		}
		if v.Name() != "" {
			field.Names = []*ast.Ident{{Name: v.Name()}}
		}
		fields.List = append(fields.List, field)
	}
	return &fields
}

func typeToExpr(t types.Type) ast.Expr {
	switch t := t.(type) {
	case nil:
		return &ast.BadExpr{}

	case *types.Basic:
		return &ast.Ident{
			Name: t.Name(),
		}

	case *types.Array:
		return &ast.ArrayType{
			Len: &ast.BasicLit{
				Kind:  token.INT,
				Value: fmt.Sprint(t.Len()),
			},
			Elt: typeToExpr(t.Elem()),
		}

	case *types.Slice:
		return &ast.ArrayType{
			Elt: typeToExpr(t.Elem()),
		}

	case *types.Struct:
		var fields ast.FieldList
		for i := 0; i < t.NumFields(); i++ {
			v := t.Field(i)
			field := &ast.Field{
				Type: typeToExpr(v.Type()),
			}
			if v.Name() != "" {
				field.Names = []*ast.Ident{{Name: v.Name()}}
			}
			fields.List = append(fields.List, field)
		}
		return &ast.StructType{
			Fields: &fields,
		}

	case *types.Pointer:
		return &ast.StarExpr{
			X: typeToExpr(t.Elem()),
		}

	case *types.Tuple:
		return &ast.BadExpr{}

	case *types.Signature:
		if len(t.TypeParams()) > 0 {
			return &ast.BadExpr{}
		}
		return &ast.FuncType{
			Params:  tupleToFieldList(t.Params()),
			Results: tupleToFieldList(t.Results()),
		}

	case *types.Interface:
		var methods ast.FieldList
		for i := 0; i < t.NumMethods(); i++ {
			meth := t.Method(i)
			methods.List = append(methods.List, &ast.Field{
				Names: []*ast.Ident{{Name: meth.Name()}},
				Type:  typeToExpr(meth.Type()),
			})
		}
		return &ast.InterfaceType{
			Methods: &methods,
		}

	case *types.Map:
		return &ast.MapType{
			Key:   typeToExpr(t.Key()),
			Value: typeToExpr(t.Elem()),
		}

	case *types.Chan:
		dir := map[types.ChanDir]ast.ChanDir{
			types.SendRecv: ast.SEND | ast.RECV,
			types.SendOnly: ast.SEND,
			types.RecvOnly: ast.RECV,
		}
		return &ast.ChanType{
			Dir:   dir[t.Dir()],
			Value: typeToExpr(t.Elem()),
		}

	case *types.Named:
		return &ast.Ident{
			Name: t.Obj().Name(),
		}

	case *types.TypeParam:
		return &ast.BadExpr{}

	default:
		return &ast.BadExpr{}
	}
}

func writeType(w io.Writer, t types.Type) {
	switch t := t.(type) {
	case nil:
		fmt.Fprintf(w, "bad")

	case *types.Basic:
		fmt.Fprintf(w, "%s", t.Name())

	case *types.Array:
		fmt.Fprintf(w, "array_%d_", t.Len())
		writeType(w, t.Elem())

	case *types.Slice:
		fmt.Fprintf(w, "slice_")
		writeType(w, t.Elem())

	case *types.Struct:
		fmt.Fprint(w, "struct_")
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			fmt.Fprintf(w, "%s_", field.Name())
			writeType(w, field.Type())
			fmt.Fprintf(w, "_")
		}
		fmt.Fprintf(w, "end")

	case *types.Pointer:
		fmt.Fprintf(w, "ptr_")
		writeType(w, t.Elem())

	case *types.Tuple:
		fmt.Fprintf(w, "bad")

	case *types.Signature:
		if len(t.TypeParams()) > 0 {
			fmt.Fprintf(w, "bad")
			return
		}
		fmt.Fprint(w, "func_")
		for i := 0; i < t.Params().Len(); i++ {
			param := t.Params().At(i)
			writeType(w, param.Type())
			fmt.Fprintf(w, "_")
		}
		fmt.Fprint(w, "to_")
		for i := 0; i < t.Results().Len(); i++ {
			result := t.Results().At(i)
			writeType(w, result.Type())
			fmt.Fprintf(w, "_")
		}
		fmt.Fprintf(w, "end")

	case *types.Interface:
		fmt.Fprintf(w, "interface_")
		for i := 0; i < t.NumMethods(); i++ {
			meth := t.Method(i)
			fmt.Fprintf(w, "%s_", meth.Name())
			writeType(w, meth.Type())
			fmt.Fprintf(w, "_")
		}
		fmt.Fprintf(w, "end")

	case *types.Map:
		fmt.Fprintf(w, "map_")
		writeType(w, t.Key())
		fmt.Fprintf(w, "_")
		writeType(w, t.Elem())

	case *types.Chan:
		dir := map[types.ChanDir]string{
			types.SendRecv: "both",
			types.SendOnly: "send",
			types.RecvOnly: "recv",
		}
		fmt.Fprintf(w, "chan_%s_", dir[t.Dir()])
		writeType(w, t.Elem())

	case *types.Named:
		fmt.Fprintf(w, "%s", t.Obj().Name())

	case *types.TypeParam:
		fmt.Fprintf(w, "bad")

	default:
		fmt.Fprintf(w, "bad")
	}
}
