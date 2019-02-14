# gosexy/to

*Convenient* functions for converting values between common Go datatypes. For
Go 1.1+.

This package ignores errors and allows quick-and-dirty conversions between Go
datatypes.  When any conversion seems unreasonable a [zero value][3] is used as
fallback.

If you're not working with human provided data, fuzzy input or if you'd rather
not ignore any error in your program, you should better use the standard Go
packages for conversion, such as [strconv][4], [fmt][5] or even [standard
conversion][6] they may be better suited for the task.

[![Build Status](https://travis-ci.org/gosexy/to.png)](https://travis-ci.org/gosexy/to)

## Installation

```sh
go get -u menteslibres.net/gosexy/to
```

## Usage

Import the package

```go
import "menteslibres.net/gosexy/to"
```

Use the available `to` functions to convert a `float64` into a `string`:

```go
// "1.23"
s := to.String(1.23)
```

Or a `bool` into `string`:

```go
// "true"
s := to.String(true)
```

What about the other way around? `string` to `float64` and `string` to `bool`.

```go
// 1.23
f := to.Float64("1.23")

// true
b := to.Bool("true")
```

Note that this package only provides `to.Uint64()`, `to.Int64()` and
`to.Float64()` but no `to.Uint8()`, `to.Uint()` or `to.Float32()` functions, if
you'd like to produce a `float32` instead of a `float64` you'd first use
`to.Float64()` and then cast the output using `float32()`.

```go
f32 := float32(to.Float64("12.34"))
```

There is another important function, `to.Convert()` that accepts any value
(`interface{}`) as first argument and also a `reflect.Kind`, as second, that
defines the data type the first argument will be converted to, this is also
the only function that returns an `error` value.

```go
val, err := to.Convert("12345", reflect.Int64)
```

Date formats and durations are also handled, you can use many fuzzy date formats
and they would be converted into `time.Time` values.

```go
timeVal = to.Time("2012-03-24")
timeVal = to.Time("Mar 24, 2012")

durationVal := to.Duration("12s37ms")
```

Now, an important question: how fast is this library compared to standard
methods, like the `fmt` or `strconv` packages?

It is, of course, a little slower that `strconv` methods but it is faster than
`fmt`, so it provides an acceptable speed for most projects. You can test it by
yourself:

```sh
$ go test -test.bench=.
PASS
BenchmarkFmtIntToString           5000000               547 ns/op
BenchmarkFmtFloatToString         2000000               914 ns/op
BenchmarkStrconvIntToString      10000000               142 ns/op
BenchmarkStrconvFloatToString     1000000              1155 ns/op
BenchmarkIntToString             10000000               325 ns/op
BenchmarkFloatToString            2000000               873 ns/op
BenchmarkIntToBytes              10000000               198 ns/op
BenchmarkBoolToString            50000000                48.0 ns/op
BenchmarkFloatToBytes             2000000               773 ns/op
BenchmarkIntToBool                5000000               403 ns/op
BenchmarkStringToTime             1000000              1063 ns/op
BenchmarkConvert                 10000000               199 ns/op
ok      menteslibres.net/gosexy/to      27.670s
```

See the [docs][1] for a full reference of all the available `to` methods.

## License

This is Open Source released under the terms of the MIT License:

> Copyright (c) 2013-2014 José Carlos Nieto, https://menteslibres.net/xiam
>
> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:
>
> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

[1]: http://godoc.org/menteslibres.net/gosexy/to
[2]: https://menteslibres.net/gosexy/to
[3]: http://golang.org/ref/spec#The_zero_value
[4]: http://golang.org/pkg/strconv/
[5]: http://golang.org/pkg/fmt/
[6]: http://golang.org/ref/spec#Conversions
