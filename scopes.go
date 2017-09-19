package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"

	"golang.org/x/tools/go/loader"
)

func main() {
	var conf loader.Config

	_, err := conf.FromArgs(os.Args[1:], false)
	if err != nil {
		fmt.Println("loader: " + err.Error())
		os.Exit(1)
	}

	prog, err := conf.Load()
	if err != nil {
		fmt.Println("loader: " + err.Error())
		os.Exit(1)
	}

	pass := true

	for _, pkg := range prog.Imported {
		for _, file := range pkg.Files {
			w := &walker{scope: pkg.Pkg.Scope()}
			ast.Walk(w, file)

			for _, err := range w.errors {
				file := prog.Fset.File(err.node.Pos())
				pos := file.PositionFor(err.node.Pos(), true)

				fmt.Printf(
					"%s:%d:%d: accessing outer scope when closer var of same type exists. Did you mean %s?\n",
					stripWd(file.Name()),
					pos.Line,
					pos.Column,
					err.shadowing.inner.Name(),
				)
				pass = false
			}
		}
	}

	if !pass {
		os.Exit(1)
	}
}

type shadowError struct {
	shadowing shadowing
	node      ast.Node
	name      string
}

type shadowing struct {
	outer types.Object
	inner types.Object
}

type walker struct {
	scope        *types.Scope
	typeshadowed []shadowing
	parent       *walker
	errors       []shadowError
}

func (s *walker) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch n := n.(type) {
	case *ast.Ident:
		_, o := s.scope.LookupParent(n.Name, n.Pos())
		for _, shadowed := range s.typeshadowed {
			if o == shadowed.outer {
				s.addError(shadowError{
					shadowing: shadowed,
					node:      n,
					name:      n.Name,
				})
			}
		}

	case *ast.FuncLit:
		scope := s.scope.Innermost(n.Body.Pos())
		typeShadowed := make([]shadowing, len(s.typeshadowed))
		copy(typeShadowed, s.typeshadowed)

		names := []string{}
		for parent := scope.Parent(); parent != nil; parent = parent.Parent() {
			names = append(names, parent.Names()...)
			parent.Parent()
		}

		for _, param := range n.Type.Params.List {
			for _, name := range param.Names {
				paramType := scope.Lookup(name.Name)

				for _, name := range names {
					_, o := scope.Parent().LookupParent(name, n.Pos())
					if o != nil && similar(paramType.Type(), o.Type()) {
						typeShadowed = append(typeShadowed, shadowing{
							outer: o,
							inner: paramType,
						})
					}
				}
			}
		}

		ast.Walk(&walker{
			scope:        scope,
			typeshadowed: typeShadowed,
			parent:       s,
		}, n.Body)

		return nil

	}

	return s
}

func (s *walker) addError(e shadowError) {
	top := s
	for top.parent != nil {
		top = top.parent
	}

	top.errors = append(top.errors, e)
}

func similar(x, y types.Type) bool {
	switch x := x.(type) {
	case *types.Pointer:
		if y, ok := y.(*types.Pointer); ok {
			return similar(x.Elem(), y.Elem())
		}

	case *types.Named:
		if y, ok := y.(*types.Named); ok {
			if x == y {
				return true
			}
		}

		return similar(x.Underlying(), y.Underlying())

	case *types.Struct:
		return types.Identical(x, y)

	case *types.Interface:
		xMethods := methods(x)
		yMethods := methods(y)

		if len(xMethods) == 0 {
			return false
		}

		for _, method := range xMethods {
			yMethod := yMethods.getByName(method.Name())
			if yMethod == nil || !types.Identical(method.Type(), yMethod.Type()) {

				return false
			}
		}

		return true
	}

	return false
}

type funcs []*types.Func

func methods(t types.Type) funcs {
	ret := funcs{}
	switch t := t.(type) {
	case *types.Pointer:
		return methods(t.Elem())
	case *types.Named:
		for i := 0; i < t.NumMethods(); i++ {
			ret = append(ret, t.Method(i))
		}
	case *types.Interface:
		for i := 0; i < t.NumMethods(); i++ {
			ret = append(ret, t.Method(i))
		}
	}

	return ret
}

func (f funcs) getByName(name string) *types.Func {
	for _, fun := range f {
		if fun.Name() == name {
			return fun
		}
	}
	return nil
}

func stripWd(filename string) string {
	wd, err := os.Getwd()
	if err != nil {
		return filename
	}

	if filename[0:len(wd)+1] != wd+"/" {
		return filename

	}

	return filename[len(wd)+1:]
}
