# goven

This tool will take source code of a go package
and copy it into the current directory, adjusting
import paths as necessary.

Example usage:

```bash
$ go get github.com/bmizerany/pat
$ goven github.com/bmizerany/pat
```

This takes the named package, copies its code into
a subdirectory, removes the VCS metadata, and
rewrites import paths in the current directory tree
to refer to the new location. For example, vendoring
`github.com/bmizerany/pat` in `example.com/foo` would
produce an import path of
`example.com/foo/github.com/bmizerany/pat`.

## Install

```bash
$ go get github.com/kr/goven
$ go install github.com/kr/goven
```

## Bugs

- probably doesn't work on Windows

If you use this tool, I suggest doing it on a clean checkout.
Then inspect the output carefully. Caveat lector.
