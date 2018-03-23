# iprange

[![GoDoc](https://godoc.org/github.com/malfunkt/iprange?status.svg)](https://godoc.org/github.com/malfunkt/iprange)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)]()
[![Build Status](https://travis-ci.org/malfunkt/iprange.svg?branch=master)](https://travis-ci.org/malfunkt/iprange)

`iprange` is a library you can use to parse IPv4 addresses from a string in the `nmap` format.

It takes a string, and returns a list of `Min`-`Max` pairs, which can then be expanded and normalized automatically by the package.

## Supported Formats

`iprange` supports the following formats:

* `10.0.0.1`
* `10.0.0.0/24`
* `10.0.0.*`
* `10.0.0.1-10`
* `10.0.0.1, 10.0.0.5-10, 192.168.1.*, 192.168.10.0/24`

## Usage

```go
package main

import (
	"log"

	"github.com/malfunkt/iprange"
)

func main() {
	list, err := iprange.ParseList("10.0.0.1, 10.0.0.5-10, 192.168.1.*, 192.168.10.0/24")
	if err != nil {
		log.Printf("error: %s", err)
	}
	log.Printf("%+v", list)

	rng := list.Expand()
	log.Printf("%s", rng)
}
```
