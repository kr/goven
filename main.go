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
	"runtime"
	"strconv"
	"strings"
)

var curgodir, imp string

var (
	copy    = flag.Bool("copy", true, "copy the code")
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

	pkgdir := which(imp)
	if pkgdir == "" {
		log.Fatal("could not find package")
	}

	curgodir, err = lookupDir()
	if err != nil {
		log.Fatal(err)
	}

	if *copy {
		impdir := filepath.Clean(imp)

		err = os.RemoveAll(impdir)
		if err != nil {
			log.Fatal(err)
		}

		err = os.MkdirAll(impdir, 0770)
		if err != nil {
			log.Fatal(err)
		}

		if runtime.GOOS == "windows" {
			err = run("xcopy", "/D", "/E", "/C", "/R", "/I", "/K", "/Y", "/F", "/Q", filepath.Clean(pkgdir)+".", impdir)
		} else {
			err = run("cp", "-r", pkgdir+"/.", impdir)
		}
		if err != nil {
			log.Fatal(err)
		}

		scmdirs := []string{".git", ".hg", ".bzr"}
		for _, scmdir := range scmdirs {
			err = os.RemoveAll(filepath.Join(impdir, scmdir))
			if err != nil {
				log.Fatal(err)
			}
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

func which(pkg string) string {
	for _, top := range strings.Split(os.Getenv("GOPATH"), ":") {
		dir := filepath.Join(top, "src", pkg)
		_, err := os.Stat(dir)
		if err == nil {
			return filepath.Clean(dir)
		}
		p := err.(*os.PathError)
		if !os.IsNotExist(p.Err) {
			log.Print(err)
		}
	}
	return ""
}

func lookupDir() (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", errors.New("missing GOPATH")
	}

	dot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dot = filepath.ToSlash(dot)

	items := strings.Split(gopath, string(filepath.ListSeparator))
	for _, top := range items {
		top = filepath.ToSlash(filepath.Join(top, "src"))
		if strings.HasPrefix(dot, top) {
			return dot[len(top):], nil
		}
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
		if strings.HasPrefix(filepath.ToSlash(path), imp) {
			s.Path.Value = strconv.Quote(filepath.Join(curgodir, path))
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

	os.Remove(path)

	return os.Rename(wpath, path)
}
