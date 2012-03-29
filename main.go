package main

import (
	"errors"
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var godir, imp string

var (
	fetch   = flag.Bool("fetch", true, "fetch the code")
	rewrite = flag.Bool("rewrite", true, "rewrite include paths")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FLAGS] <package>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var err error

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	imp = flag.Arg(0)

	godir, err = lookupDir()
	if err != nil {
		log.Fatal(err)
	}

	if *fetch {
		err = os.RemoveAll(imp)
		if err != nil {
			log.Fatal(err)
		}

		err = run("git", "clone", "git://"+imp+".git", imp)
		if err != nil {
			log.Fatal(err)
		}

		err = os.RemoveAll(imp + "/.git")
		if err != nil {
			log.Fatal(err)
		}
	}

	if *rewrite {
		err = filepath.Walk(".", mangle)
		if err != nil {
			log.Fatal(err)
		}

		err = run("go", "fmt")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func lookupDir() (string, error) {
	top := os.Getenv("GOPATH")
	if top == "" {
		return "", errors.New("missing GOPATH")
	}

	dot, err := os.Getwd()
	if err != nil {
		return "", err
	}

	top = top + "/src/"
	if strings.HasPrefix(dot, top) {
		return dot[len(top):], nil
	}

	return "", errors.New("cwd not found in GOPATH")
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func mangle(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
	}

	if !info.IsDir() && strings.HasSuffix(path, ".go") {
		err = mangleFile(path)
		if err != nil {
			log.Print(err)
		}
	}
	return nil
}

func mangleFile(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	var changed bool
	for _, s := range f.Imports {
		path, err := strconv.Unquote(s.Path.Value)
		if err != nil {
			return err // can't happen
		}
		if strings.HasPrefix(path, imp) {
			s.Path.Value = strconv.Quote(godir + "/" + path)
			changed = true
		}
	}

	if !changed {
		return nil
	}

	wpath := path + ".temp"
	w, err := os.Create(wpath)
	if err != nil {
		return err
	}

	err = printer.Fprint(w, fset, f)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return os.Rename(wpath, path)
}
