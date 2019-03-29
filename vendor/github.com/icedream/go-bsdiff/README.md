# bsdiff for Go

This wrapper implementation for Golang reuses the existing
C version of bsdiff as provided by @mendsley and wraps it
into a Go package, abstracting away all the cgo work that
would need to be done otherwise.

## Installation

The library and the helper binaries `go-bsdiff` and `go-bspatch` can be installed like this:

    go get -v github.com/icedream/go-bsdiff/...

## Usage in application code

For exact documentation of the library check out [GoDoc](https://godoc.org/github.com/icedream/go-bsdiff).

Library functionality is provided both as a package `bsdiff` containing both
methods `Diff` and `Patch`, or as subpackages `diff` and `patch` which each
only link the wanted functionality.

Below example will generate a patch and apply it again in one go. This code
is not safe against errors but it shows how to use the provided routines:

```go
package main

import (
    "os"
    "github.com/icedream/go-bsdiff"
    // Or use the subpackages to only link what you need:
    //"github.com/icedream/go-bsdiff/diff"
    //"github.com/icedream/go-bsdiff/patch"
)

const (
    oldFilePath = "your_old_file.dat"
    newFilePath = "your_new_file.dat"
    patchFilePath = "the_generated.patch"
)

func generatePatch() error {
    oldFile, _ := os.Open(oldFilePath)
    defer oldFile.Close()
    newFile, _ := os.Open(newFilePath)
    defer newFile.Close()
    patchFile, _ := os.Create(patchFilePath)
    defer patchFile.Close()

    return bsdiff.Diff(oldFile, newFile, patchFile)
}

func applyPatch() error {
    oldFile, _ := os.Open(oldFilePath)
    defer oldFile.Close()
    newFile, _ := os.Create(newFilePath)
    defer newFile.Close()
    patchFile, _ := os.Open(patchFilePath)
    defer patchFile.Close()

    return bsdiff.Patch(oldFile, newFile, patchFile)
}

func main() {
    generatePatch()
    applyPatch()
}
```

## Usage of the tools

The tools `go-bsdiff` and `go-bspatch` both provide a `--help` flag to print
out all information but in their simplest form, they can be used like this:

```sh
# Creates a patch file $the_generated with differences from
# $your_old_file to $your_new_file.
go-bsdiff "$your_old_file" "$your_new_file" "$the_generated"

# Applies a patch file $the_generated on $your_old_file
# and saves the new file to $your_new_file.
go-bspatch "$your_old_file" "$your_new_file" "$the_generated"
```

## Motivation

There is [a Go implementation of an older version of bsdiff called binarydist](https://github.com/kr/binarydist). The original bsdiff tool has since been updated so patches generating using the original tool are no longer compatible with the Go implementation. I don't know what the changes between the versions are and unfortunately I don't have the time to search for these changes and port them over as a pull request, otherwise I'd have done that instead.

Additionally, @mendsley has already done the extra work of rewriting the code to be embeddable in any application code as a library. So why not make use of cgo, which I was going to look into in more detail at some point anyways?
