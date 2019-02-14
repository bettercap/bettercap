# minimist

Simple CLI args parser.

Port of [minimist](https://github.com/substack/minimist) to golang

## options

`--a            // a == true`
`--a=foo        // a == "foo"`
`--a foo        // a == "foo"`

`--no-a         // a == false`
`-a             // a == true`
`-ab            // a == true, b == true`
`-ab foo        // a == true, b == "foo"`

# license

MIT
