// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements typechecking of call and selector expressions.

package types

import (
	"github.com/faiface/generics/go/ast"
	"github.com/faiface/generics/go/token"
)

func mapVar(mapping map[*TypeParam]Type, v *Var, visited map[Type]Type) *Var {
	if v == nil {
		return nil
	}
	return NewVar(
		v.Pos(),
		v.Pkg(),
		v.Name(),
		mapType(mapping, v.Type(), visited),
	)
}

func mapType(mapping map[*TypeParam]Type, x Type, visited map[Type]Type) (mapped Type) {
	if visited[x] != nil {
		return visited[x]
	}

	switch x := x.(type) {
	case nil, *Basic:
		return x

	case *Array:
		array := &Array{
			len: x.len,
		}
		visited[x] = array
		array.elem = mapType(mapping, x.elem, visited)
		return array

	case *Slice:
		slice := &Slice{}
		visited[x] = slice
		slice.elem = mapType(mapping, x.elem, visited)
		return slice

	case *Struct:
		str := &Struct{}
		visited[x] = str
		for i, f := range x.fields {
			str.fields = append(str.fields, NewVar(
				f.Pos(),
				f.Pkg(),
				f.Name(),
				mapType(mapping, f.Type(), visited),
			))
			if x.tags != nil {
				str.tags = append(str.tags, x.tags[i])
			}
		}
		return str

	case *Pointer:
		ptr := &Pointer{}
		visited[x] = ptr
		ptr.base = mapType(mapping, x.base, visited)
		return ptr

	case *Tuple:
		if x == nil {
			return NewTuple()
		}
		var vars []*Var
		for _, v := range x.vars {
			vars = append(vars, mapVar(mapping, v, visited))
		}
		return NewTuple(vars...)

	case *Signature:
		return &Signature{
			scope:    x.scope,
			recv:     mapVar(mapping, x.recv, visited),
			params:   mapType(mapping, x.params, visited).(*Tuple),
			results:  mapType(mapping, x.results, visited).(*Tuple),
			variadic: x.variadic,
		}

	case *Interface:
		iface := &Interface{}
		visited[x] = iface
		for _, meth := range x.methods {
			meth := *meth
			meth.typ = mapType(mapping, meth.typ, visited)
			iface.methods = append(iface.methods, &meth)
		}
		for _, emb := range x.embeddeds {
			iface.embeddeds = append(iface.embeddeds, mapType(mapping, emb, visited).(*Named))
		}
		for _, meth := range x.allMethods {
			meth := *meth
			meth.typ = mapType(mapping, meth.typ, visited)
			iface.allMethods = append(iface.allMethods, &meth)
		}
		return iface

	case *Map:
		m := &Map{}
		visited[x] = m
		m.key = mapType(mapping, x.key, visited)
		m.elem = mapType(mapping, x.elem, visited)
		return m

	case *Chan:
		ch := &Chan{
			dir: x.dir,
		}
		visited[x] = ch
		ch.elem = mapType(mapping, x.elem, visited)
		return ch

	case *Named:
		assert(x.NumParams() == 0)
		return x

	case *Instance:
		var args []Type
		for _, arg := range x.args {
			args = append(args, mapType(mapping, arg, visited))
		}
		return NewInstance(x.named, args)

	case *TypeParam:
		if typ, ok := mapping[x]; ok {
			return typ
		}
		return x

	default:
		panic("unreachable")
	}
}

func (check *Checker) call(x *operand, e *ast.CallExpr) exprKind {
	check.exprOrType(x, e.Fun)

	switch x.mode {
	case invalid:
		check.use(e.Args...)
		x.mode = invalid
		x.expr = e
		return statement

	case typexpr:
		if named, ok := x.typ.(*Named); ok && named.NumParams() > 0 {
			// generic type instantiation
			x.typ = check.typ(nil, e, false)
			x.mode = typexpr
			return instance
		}

		// conversion
		T := x.typ
		x.mode = invalid
		switch n := len(e.Args); n {
		case 0:
			check.errorf(e.Rparen, "missing argument in conversion to %s", T)
		case 1:
			check.expr(x, e.Args[0])
			if x.mode != invalid {
				check.conversion(x, T)
			}
		default:
			check.errorf(e.Args[n-1].Pos(), "too many arguments in conversion to %s", T)
		}
		x.expr = e
		return conversion

	case builtin:
		id := x.id
		if !check.builtin(x, e, id) {
			x.mode = invalid
		}
		x.expr = e
		// a non-constant result implies a function call
		if x.mode != invalid && x.mode != constant_ {
			check.hasCallOrRecv = true
		}
		return predeclaredFuncs[id].kind

	default:
		// function/method call
		sig, _ := x.typ.Underlying().(*Signature)
		if sig == nil {
			check.invalidOp(x.pos(), "cannot call non-function %s", x)
			x.mode = invalid
			x.expr = e
			return statement
		}

		var mapping map[*TypeParam]Type
		if len(sig.typeParams) != 0 {
			mapping = make(map[*TypeParam]Type)
		}

		// unnamed generic type parameters
		for i := 0; i < len(sig.unnamed); i++ {
			typ := check.typ(nil, e.Args[i], false)
			if !assignableToTypeParam(typ, sig.unnamed[i]) {
				check.errorf(x.pos(), "value of type %v does not satisfy the restrictions of %v", typ, sig.unnamed[i])
				x.mode = invalid
				return statement
			}
			mapping[sig.unnamed[i]] = typ
		}

		arg, n, _ := unpack(func(x *operand, i int) { check.multiExpr(x, e.Args[len(sig.unnamed)+i]) }, len(e.Args)-len(sig.unnamed), false)
		if arg != nil {
			check.arguments(mapping, x, e, sig, arg, n)
		} else {
			x.mode = invalid
		}

		// determine result
		switch sig.results.Len() {
		case 0:
			x.mode = novalue
		case 1:
			x.mode = value
			x.typ = mapType(mapping, sig.results.vars[0].typ, make(map[Type]Type)) // unpack tuple
		default:
			x.mode = value
			x.typ = mapType(mapping, sig.results, make(map[Type]Type))
		}

		x.expr = e
		check.hasCallOrRecv = true

		if m := check.GenericCalls; len(sig.typeParams) > 0 && m != nil {
			m[e] = &GenericCall{
				NumUnnamed: len(sig.unnamed),
				Mapping:    mapping,
			}
		}

		return statement
	}
}

// use type-checks each argument.
// Useful to make sure expressions are evaluated
// (and variables are "used") in the presence of other errors.
// The arguments may be nil.
func (check *Checker) use(arg ...ast.Expr) {
	var x operand
	for _, e := range arg {
		// The nil check below is necessary since certain AST fields
		// may legally be nil (e.g., the ast.SliceExpr.High field).
		if e != nil {
			check.rawExpr(&x, e, nil)
		}
	}
}

// useLHS is like use, but doesn't "use" top-level identifiers.
// It should be called instead of use if the arguments are
// expressions on the lhs of an assignment.
// The arguments must not be nil.
func (check *Checker) useLHS(arg ...ast.Expr) {
	var x operand
	for _, e := range arg {
		// If the lhs is an identifier denoting a variable v, this assignment
		// is not a 'use' of v. Remember current value of v.used and restore
		// after evaluating the lhs via check.rawExpr.
		var v *Var
		var v_used bool
		if ident, _ := unparen(e).(*ast.Ident); ident != nil {
			// never type-check the blank name on the lhs
			if ident.Name == "_" {
				continue
			}
			if _, obj := check.scope.LookupParent(ident.Name, token.NoPos); obj != nil {
				// It's ok to mark non-local variables, but ignore variables
				// from other packages to avoid potential race conditions with
				// dot-imported variables.
				if w, _ := obj.(*Var); w != nil && w.pkg == check.pkg {
					v = w
					v_used = v.used
				}
			}
		}
		check.rawExpr(&x, e, nil)
		if v != nil {
			v.used = v_used // restore v.used
		}
	}
}

// useGetter is like use, but takes a getter instead of a list of expressions.
// It should be called instead of use if a getter is present to avoid repeated
// evaluation of the first argument (since the getter was likely obtained via
// unpack, which may have evaluated the first argument already).
func (check *Checker) useGetter(get getter, n int) {
	var x operand
	for i := 0; i < n; i++ {
		get(&x, i)
	}
}

// A getter sets x as the i'th operand, where 0 <= i < n and n is the total
// number of operands (context-specific, and maintained elsewhere). A getter
// type-checks the i'th operand; the details of the actual check are getter-
// specific.
type getter func(x *operand, i int)

// unpack takes a getter get and a number of operands n. If n == 1, unpack
// calls the incoming getter for the first operand. If that operand is
// invalid, unpack returns (nil, 0, false). Otherwise, if that operand is a
// function call, or a comma-ok expression and allowCommaOk is set, the result
// is a new getter and operand count providing access to the function results,
// or comma-ok values, respectively. The third result value reports if it
// is indeed the comma-ok case. In all other cases, the incoming getter and
// operand count are returned unchanged, and the third result value is false.
//
// In other words, if there's exactly one operand that - after type-checking
// by calling get - stands for multiple operands, the resulting getter provides
// access to those operands instead.
//
// If the returned getter is called at most once for a given operand index i
// (including i == 0), that operand is guaranteed to cause only one call of
// the incoming getter with that i.
//
func unpack(get getter, n int, allowCommaOk bool) (getter, int, bool) {
	if n != 1 {
		// zero or multiple values
		return get, n, false
	}
	// possibly result of an n-valued function call or comma,ok value
	var x0 operand
	get(&x0, 0)
	if x0.mode == invalid {
		return nil, 0, false
	}

	if t, ok := x0.typ.(*Tuple); ok {
		// result of an n-valued function call
		return func(x *operand, i int) {
			x.mode = value
			x.expr = x0.expr
			x.typ = t.At(i).typ
		}, t.Len(), false
	}

	if x0.mode == mapindex || x0.mode == commaok {
		// comma-ok value
		if allowCommaOk {
			a := [2]Type{x0.typ, Typ[UntypedBool]}
			return func(x *operand, i int) {
				x.mode = value
				x.expr = x0.expr
				x.typ = a[i]
			}, 2, true
		}
		x0.mode = value
	}

	// single value
	return func(x *operand, i int) {
		if i != 0 {
			unreachable()
		}
		*x = x0
	}, 1, false
}

// arguments checks argument passing for the call with the given signature.
// The arg function provides the operand for the i'th argument.
func (check *Checker) arguments(mapping map[*TypeParam]Type, x *operand, call *ast.CallExpr, sig *Signature, arg getter, n int) {
	if call.Ellipsis.IsValid() {
		// last argument is of the form x...
		if !sig.variadic {
			check.errorf(call.Ellipsis, "cannot use ... in call to non-variadic %s", call.Fun)
			check.useGetter(arg, n)
			return
		}
		if len(call.Args) == 1 && n > 1 {
			// f()... is not permitted if f() is multi-valued
			check.errorf(call.Ellipsis, "cannot use ... with %d-valued %s", n, call.Args[0])
			check.useGetter(arg, n)
			return
		}
	}

	// evaluate arguments
	for i := 0; i < n; i++ {
		arg(x, i)
		if x.mode != invalid {
			var ellipsis token.Pos
			if i == n-1 && call.Ellipsis.IsValid() {
				ellipsis = call.Ellipsis
			}
			check.argument(mapping, call.Fun, sig, i, x, ellipsis)
		}
	}

	// check argument count
	if sig.variadic {
		// a variadic function accepts an "empty"
		// last argument: count one extra
		n++
	}
	if n < sig.params.Len() {
		check.errorf(call.Rparen, "too few arguments in call to %s", call.Fun)
		// ok to continue
	}
}

// argument checks passing of argument x to the i'th parameter of the given signature.
// If ellipsis is valid, the argument is followed by ... at that position in the call.
func (check *Checker) argument(mapping map[*TypeParam]Type, fun ast.Expr, sig *Signature, i int, x *operand, ellipsis token.Pos) {
	check.singleValue(x)
	if x.mode == invalid {
		return
	}

	n := sig.params.Len()

	// determine parameter type
	var typ Type
	switch {
	case i < n:
		typ = sig.params.vars[i].typ
	case sig.variadic:
		typ = sig.params.vars[n-1].typ
		if debug {
			if _, ok := typ.(*Slice); !ok {
				check.dump("%s: expected unnamed slice type, got %s", sig.params.vars[n-1].Pos(), typ)
			}
		}
	default:
		check.errorf(x.pos(), "too many arguments")
		return
	}

	if ellipsis.IsValid() {
		// argument is of the form x... and x is single-valued
		if i != n-1 {
			check.errorf(ellipsis, "can only use ... with matching parameter")
			return
		}
		if _, ok := x.typ.Underlying().(*Slice); !ok && x.typ != Typ[UntypedNil] { // see issue #18268
			check.errorf(x.pos(), "cannot use %s as parameter of type %s", x, typ)
			return
		}
	} else if sig.variadic && i >= n-1 {
		// use the variadic parameter slice's element type
		typ = typ.(*Slice).elem
	}

	check.assignment(mapping, x, typ, check.sprintf("argument to %s", fun))
}

func (check *Checker) selector(x *operand, e *ast.SelectorExpr) {
	// these must be declared before the "goto Error" statements
	var (
		obj      Object
		index    []int
		indirect bool
	)

	sel := e.Sel.Name
	// If the identifier refers to a package, handle everything here
	// so we don't need a "package" mode for operands: package names
	// can only appear in qualified identifiers which are mapped to
	// selector expressions.
	if ident, ok := e.X.(*ast.Ident); ok {
		_, obj := check.scope.LookupParent(ident.Name, check.pos)
		if pname, _ := obj.(*PkgName); pname != nil {
			assert(pname.pkg == check.pkg)
			check.recordUse(ident, pname)
			pname.used = true
			pkg := pname.imported
			exp := pkg.scope.Lookup(sel)
			if exp == nil {
				if !pkg.fake {
					check.errorf(e.Pos(), "%s not declared by package %s", sel, pkg.name)
				}
				goto Error
			}
			if !exp.Exported() {
				check.errorf(e.Pos(), "%s not exported by package %s", sel, pkg.name)
				// ok to continue
			}
			check.recordUse(e.Sel, exp)

			// Simplified version of the code for *ast.Idents:
			// - imported objects are always fully initialized
			switch exp := exp.(type) {
			case *Const:
				assert(exp.Val() != nil)
				x.mode = constant_
				x.typ = exp.typ
				x.val = exp.val
			case *TypeName:
				x.mode = typexpr
				x.typ = exp.typ
			case *Var:
				x.mode = variable
				x.typ = exp.typ
			case *Func:
				x.mode = value
				x.typ = exp.typ
			case *Builtin:
				x.mode = builtin
				x.typ = exp.typ
				x.id = exp.id
			default:
				check.dump("unexpected object %v", exp)
				unreachable()
			}
			x.expr = e
			return
		}
	}

	check.exprOrType(x, e.X)
	if x.mode == invalid {
		goto Error
	}

	obj, index, indirect, _ = LookupFieldOrMethod(x.typ, x.mode == variable, check.pkg, sel)
	if obj == nil {
		switch {
		case index != nil:
			// TODO(gri) should provide actual type where the conflict happens
			check.invalidOp(e.Pos(), "ambiguous selector %s", sel)
		case indirect:
			check.invalidOp(e.Pos(), "%s is not in method set of %s", sel, x.typ)
		default:
			check.invalidOp(e.Pos(), "%s has no field or method %s", x, sel)
		}
		goto Error
	}

	if x.mode == typexpr {
		// method expression
		m, _ := obj.(*Func)
		if m == nil {
			check.invalidOp(e.Pos(), "%s has no method %s", x, sel)
			goto Error
		}

		check.recordSelection(e, MethodExpr, x.typ, m, index, indirect)

		// the receiver type becomes the type of the first function
		// argument of the method expression's function type
		var params []*Var
		sig := m.typ.(*Signature)
		if sig.params != nil {
			params = sig.params.vars
		}
		x.mode = value
		x.typ = &Signature{
			params:   NewTuple(append([]*Var{NewVar(token.NoPos, check.pkg, "", x.typ)}, params...)...),
			results:  sig.results,
			variadic: sig.variadic,
		}

		check.addDeclDep(m)

	} else {
		// regular selector
		switch obj := obj.(type) {
		case *Var:
			check.recordSelection(e, FieldVal, x.typ, obj, index, indirect)
			if x.mode == variable || indirect {
				x.mode = variable
			} else {
				x.mode = value
			}
			x.typ = obj.typ

		case *Func:
			// TODO(gri) If we needed to take into account the receiver's
			// addressability, should we report the type &(x.typ) instead?
			check.recordSelection(e, MethodVal, x.typ, obj, index, indirect)

			if debug {
				// Verify that LookupFieldOrMethod and MethodSet.Lookup agree.
				typ := x.typ
				if x.mode == variable {
					// If typ is not an (unnamed) pointer or an interface,
					// use *typ instead, because the method set of *typ
					// includes the methods of typ.
					// Variables are addressable, so we can always take their
					// address.
					if _, ok := typ.(*Pointer); !ok && !IsInterface(typ) {
						typ = &Pointer{base: typ}
					}
				}
				// If we created a synthetic pointer type above, we will throw
				// away the method set computed here after use.
				// TODO(gri) Method set computation should probably always compute
				// both, the value and the pointer receiver method set and represent
				// them in a single structure.
				// TODO(gri) Consider also using a method set cache for the lifetime
				// of checker once we rely on MethodSet lookup instead of individual
				// lookup.
				mset := NewMethodSet(typ)
				if m := mset.Lookup(check.pkg, sel); m == nil || m.obj != obj {
					check.dump("%s: (%s).%v -> %s", e.Pos(), typ, obj.name, m)
					check.dump("%s\n", mset)
					panic("method sets and lookup don't agree")
				}
			}

			x.mode = value

			// remove receiver
			sig := *obj.typ.(*Signature)
			sig.recv = nil
			x.typ = &sig

			check.addDeclDep(obj)

		default:
			unreachable()
		}
	}

	// everything went well
	x.expr = e
	return

Error:
	x.mode = invalid
	x.expr = e
}
