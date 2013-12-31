package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
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

var (
	copy       = flag.Bool("copy", true, "copy the code")
	rewrite    = flag.Bool("rewrite", true, "rewrite include paths")
	verbose    = flag.Bool("verbose", false, "notes what is being done")
	prefixPath = flag.String("prefix", "", "subdirectory to put files in, e.g. 'third_party'")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FLAGS] <package>\n", os.Args[0])
	flag.PrintDefaults()
}

func logf(format string, v ...interface{}) {
	if *verbose {
		fmt.Printf(format+"\n", v...)
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("goven: ")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	// Package to import.
	impName := flag.Arg(0)

	// Package to import relative to GOPATH.
	impDir := packageToDir(impName)
	if impDir == "" {
		log.Fatalf("Could not find package %s", impName)
	}
	logf("Found package path: %s", impDir)

	// Current directory relative to GOPATH.
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	relativeName, err := dirToPackage(cwd)
	if err != nil {
		log.Fatal(err)
	}

	destDir := impName
	if *prefixPath != "" {
		destDir = filepath.Join(*prefixPath, destDir)
		relativeName = relativeName + "/" + *prefixPath
	}
	logf("Using relative path: %s", destDir)

	if *copy {
		logf("Copying %s into %s", impDir, destDir)
		if err = os.RemoveAll(destDir); err != nil {
			log.Fatal(err)
		}
		if err = os.MkdirAll(destDir, 0770); err != nil {
			log.Fatal(err)
		}
		// TODO(maruel): Copy manually so dot directories are not copied in the
		// first place.
		if err = run("cp", "-r", impDir+"/.", destDir); err != nil {
			log.Fatal(err)
		}

		for _, scmdir := range []string{".git", ".hg", ".bzr"} {
			if err = os.RemoveAll(filepath.Join(destDir, scmdir)); err != nil {
				log.Fatal(err)
			}
		}
	}

	if *rewrite {
		logf("Adding prefix %s to %s", relativeName, impName)
		callback := func(p string, info os.FileInfo, err error) error {
			return mangle(p, impName, relativeName, info, err)
		}
		if err = filepath.Walk(".", callback); err != nil {
			log.Fatal(err)
		}
		if err = run("go", "fmt"); err != nil {
			log.Fatal(err)
		}
	}
}

// Retrieves the directory containing the package to import.
func packageToDir(pkg string) string {
	for _, top := range strings.Split(os.Getenv("GOPATH"), ":") {
		dir := top + "/src/" + pkg
		_, err := os.Stat(dir)
		if err == nil {
			return dir
		}
		if p := err.(*os.PathError); !os.IsNotExist(p.Err) {
			log.Print(err)
		}
	}
	return ""
}

// Returns the relative import path to this directory.
func dirToPackage(dir string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", errors.New("missing GOPATH")
	}

	items := strings.Split(gopath, ":")
	for _, top := range items {
		top = top + "/src/"
		if strings.HasPrefix(dir, top) {
			logf("Found GOPATH %s", top)
			return dir[len(top):], nil
		}
		if top, err := filepath.EvalSymlinks(top); err == nil {
			// EvalSymlinks() removes trailing /.
			top += "/"
			if strings.HasPrefix(dir, top) {
				logf("Found GOPATH %s", top)
				return dir[len(top):], nil
			}
		}
	}

	return "", fmt.Errorf("%s not found in GOPATH\nGOPATH=%s", dir, gopath)
}

func run(name string, args ...string) error {
	logf("run(%s, %s)", name, strings.Join(args, ", "))
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func mangle(filePath, impName, relativeName string, info os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
	}

	if !info.IsDir() && strings.HasSuffix(filePath, ".go") {
		if err = mangleFileImports(filePath, impName, relativeName); err != nil {
			log.Print(err)
		}
	}
	return nil
}

// Mangles imports in a file.
func mangleFileImports(filePath, impName, relativeName string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("Failed to parse %s: %s", filePath, err)
	}
	if changed, err := mangleImports(f, impName, relativeName); changed || err != nil {
		return err
	}
	wpath := filePath + ".temp"
	w, err := os.Create(wpath)
	if err != nil {
		return fmt.Errorf("Failed to create %s: %s", wpath, err)
		return err
	}
	if err = printer.Fprint(w, fset, f); err != nil {
		return fmt.Errorf("Failed to write %s: %s", wpath, err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("Failed to close %s: %s", wpath, err)
	}
	logf("Modified %s", filePath)
	return os.Rename(wpath, filePath)
}

func mangleImports(f *ast.File, impName, relativeName string) (bool, error) {
	var changed bool
	for _, s := range f.Imports {
		path, err := strconv.Unquote(s.Path.Value)
		if err != nil {
			return false, err
		}
		if strings.HasPrefix(path, impName) {
			s.Path.Value = strconv.Quote(relativeName + "/" + path)
			changed = true
		}
	}
	return changed, nil
}
