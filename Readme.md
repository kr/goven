# goven

This tool will take source code of a go package
and copy it into the current directory, adjusting
import paths as necessary.

Example usage:

```bash
$ goven github.com/kr/pretty
```

## Install

```bash
$ go get github.com/kr/goven
$ go install github.com/kr/goven
```

## Bugs

- only works with git
- reformats go source code according to package go/printer
- always clones the master branch

If you use this tool, I suggest doing it on a clean checkout.
Then inspect the output carefully. Caveat lector.
