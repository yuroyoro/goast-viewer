package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

type Ast struct {
	Label    string            `json:"label"`
	Pos      int               `json:"pos"`
	End      int               `json:"end"`
	Attrs    map[string]string `json:"attrs"`
	Children []*Ast            `json:"children"`
}

type AstConverter interface {
	ToAst() *Ast
}

func Parse(filename string, source string) (a *Ast, dump string, err error) {

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, filename, source, parser.ParseComments)

	// Print the AST.
	var bf bytes.Buffer
	ast.Fprint(&bf, fset, f, func(string, reflect.Value) bool {
		return true
	})

	a, err = BuildAst("", f)
	if err != nil {
		return nil, "", err
	}
	return a, string(bf.Bytes()), nil
}

func BuildAst(prefix string, n any) (astobj *Ast, err error) {
	v := reflect.ValueOf(n)
	t := v.Type()

	a := Ast{Label: Label(prefix, n), Attrs: map[string]string{}, Children: []*Ast{}}

	if node, ok := n.(ast.Node); ok {
		a.Pos = int(node.Pos())
		a.End = int(node.End())
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if v.IsValid() == false {
		return nil, nil
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:

		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)

			child, err := BuildAst(fmt.Sprintf("%d", i), f.Interface())
			if err != nil {
				return nil, err
			}
			a.Children = append(a.Children, child)
		}
	case reflect.Map:
		for _, kv := range v.MapKeys() {
			f := v.MapIndex(kv)

			child, err := BuildAst(fmt.Sprintf("%v", kv.Interface()), f.Interface())
			if err != nil {
				return nil, err
			}
			a.Children = append(a.Children, child)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			fo := f
			name := t.Field(i).Name

			if f.Kind() == reflect.Ptr {
				f = f.Elem()
			}

			if f.IsValid() == false {
				continue
			}

			if _, ok := v.Interface().(ast.Object); !ok && f.Kind() == reflect.Interface {

				switch f.Interface().(type) {
				case ast.Decl, ast.Expr, ast.Node, ast.Spec, ast.Stmt:

					child, err := BuildAst(name, f.Interface())
					if err != nil {
						return nil, err
					}
					a.Children = append(a.Children, child)
					continue
				}
			}

			switch f.Kind() {
			case reflect.Struct, reflect.Array, reflect.Slice, reflect.Map:
				child, err := BuildAst(name, fo.Interface())
				if err != nil {
					return nil, err
				}
				a.Children = append(a.Children, child)

			default:
				a.Attrs[name] = fmt.Sprintf("%v", f.Interface())
			}
		}
	}

	return &a, nil
}

func Label(prefix string, n any) string {

	var bf bytes.Buffer

	if prefix != "" {
		fmt.Fprintf(&bf, "%s : ", prefix)
	}
	fmt.Fprintf(&bf, "%T", n)

	v := reflect.ValueOf(n)
	t := v.Type()

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if v.IsValid() == false {
		return ""
	}

	switch v.Kind() {

	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		fmt.Fprintf(&bf, "(len = %d)", v.Len())

	case reflect.Struct:
		if v.Kind() == reflect.Struct {
			fs := []string{}
			for i := 0; i < v.NumField(); i++ {
				f := v.Field(i)
				name := t.Field(i).Name
				switch name {
				case "Name", "Kind", "Tok", "Op":
					fs = append(fs, fmt.Sprintf("%s: %v", name, f.Interface()))
				}
			}
			if len(fs) > 0 {
				fmt.Fprintf(&bf, " (%s)", strings.Join(fs, ", "))
			}
		}
	default:
		fmt.Fprintf(&bf, " : %s", n)
	}
	return string(bf.Bytes())
}
