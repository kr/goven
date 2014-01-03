package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestMangleImports(t *testing.T) {
	content := "package main; import \"foo\""
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "bar.go", content, parser.ImportsOnly)
	if 1 != len(f.Imports) {
		t.Fatalf("Oops: %s", f.Imports)
	}
	if f.Imports[0].Path.Value != "\"foo\"" {
		t.Fatal("Oops: %s", f.Imports[0].Path.Value)
	}
	changed, err := mangleImports(f, "foo", "bar")
	if !changed {
		t.Fatal("Oops")
	}
	if err != nil {
		t.Fatalf("Oops: %s", err)
	}
	if 1 != len(f.Imports) {
		t.Fatalf("Oops: %s", f.Imports)
	}
	if f.Imports[0].Path.Value != "\"bar/foo\"" {
		t.Fatal("Oops: %s", f.Imports[0].Path.Value)
	}
}

func TestMangleImportsRename(t *testing.T) {
	content := "package main; import foo2 \"foo\""
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "bar.go", content, parser.ImportsOnly)
	if 1 != len(f.Imports) {
		t.Fatalf("Oops: %s", f.Imports)
	}
	if f.Imports[0].Name.Name != "foo2" {
		t.Fatal("Oops: %s", f.Imports[0].Path.Value)
	}
	if f.Imports[0].Path.Value != "\"foo\"" {
		t.Fatal("Oops: %s", f.Imports[0].Path.Value)
	}
	changed, err := mangleImports(f, "foo", "bar")
	if !changed {
		t.Fatal("Oops")
	}
	if err != nil {
		t.Fatalf("Oops: %s", err)
	}
	if 1 != len(f.Imports) {
		t.Fatalf("Oops: %s", f.Imports)
	}
	if f.Imports[0].Path.Value != "\"bar/foo\"" {
		t.Fatal("Oops: %s", f.Imports[0].Path.Value)
	}
	if f.Imports[0].Name.Name != "foo2" {
		t.Fatal("Oops: %s", f.Imports[0].Path.Value)
	}
}
