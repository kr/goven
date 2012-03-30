// This tool will take source code of a go package
// and copy it into the current directory, adjusting
// import paths as necessary.
//
// Example usage:
//   go get github.com/bmizerany/pat
//   goven github.com/bmizerany/pat
//
// It takes the named package, copies its files from the
// usual place in GOPATH into the current directory,
// removes the VCS metadata, and rewrites import paths
// in the current directory tree to refer to the new location.
// For example, vendoring github.com/bmizerany/pat into
// example.com/foo would produce an import path of
// example.com/foo/github.com/bmizerany/pat.
package main
