# goven

This tool will take source code of a go package
and copy it into the current directory, adjusting
import paths as necessary.

Example usage:

```bash
$ goven github.com/bmizerany/pat
```

This takes the named package, clones its code into
a subdirectory, removes the VCS metadata, and
rewrites import paths in the current directory tree
to refer to the new location. For example, vendoring
`github.com/bmizerany/pat` into `example.com/foo` would
produce an import path of
`example.com/foo/github.com/bmizerany/pat`.

## Install

```bash
$ go get github.com/kr/goven
$ go install github.com/kr/goven
```

## Bugs

- only works with git
- always clones the default branch (usually master)
- does not handle GOPATH with more than one entry
- needs better error checking and argument handling
- does not attempt to handle updating an already-vendored package
- probably doesn't work on Windows

If you use this tool, I suggest doing it on a clean checkout.
Then inspect the output carefully. Caveat lector.
