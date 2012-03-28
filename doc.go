// This tool will take source code of a go package
// and copy it into the current directory, adjusting
// import paths as necessary.
//
// Example usage:
//   goven github.com/kr/pretty
//
// It takes the named package, clones its code into
// a subdirectory, removes the VCS metadata, and
// rewrites import paths in the current directory tree
// to refer to the new location. For example, vendoring
// github.com/kr/pretty into example.com/foo would
// produce an import path of example.com/foo/github.com/kr/pretty.
package main
