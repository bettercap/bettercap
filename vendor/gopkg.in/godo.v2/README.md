**Documentation is WIP**

# godo

[godoc](https://godoc.org/github.com/mgutz/godo/v2)

godo is a task runner and file watcher for golang in the spirit of
rake, gulp.

To install

    go get -u gopkg.in/godo.v2/cmd/godo

## Godofile

Godo runs `Gododir/main.go`.

As an example, create a file **Gododir/main.go** with this content

```go
package main

import (
    "fmt"
    do "gopkg.in/godo.v2"
)

func tasks(p *do.Project) {
    do.Env = `GOPATH=.vendor::$GOPATH`

    p.Task("default", do.S{"hello", "build"}, nil)

    p.Task("hello", nil, func(c *do.Context) {
        name := c.Args.AsString("name", "n")
        if name == "" {
            c.Bash("echo Hello $USER!")
        } else {
            fmt.Println("Hello", name)
        }
    })

    p.Task("assets?", nil,  func(c *do.Context) {
        // The "?" tells Godo to run this task ONLY ONCE regardless of
        // how many tasks depend on it. In this case watchify watches
        // on its own.
	    c.Run("watchify public/js/index.js d -o dist/js/app.bundle.js")
    }).Src("public/**/*.{css,js,html}")

    p.Task("build", do.S{"views", "assets"}, func(c *do.Context) {
        c.Run("GOOS=linux GOARCH=amd64 go build", do.M{"$in": "cmd/server"})
    }).Src("**/*.go")

    p.Task("server", do.S{"views", "assets"}, func(c *do.Context) {
        // rebuilds and restarts when a watched file changes
        c.Start("main.go", do.M{"$in": "cmd/server"})
    }).Src("server/**/*.go", "cmd/server/*.{go,json}").
       Debounce(3000)

    p.Task("views", nil, func(c *do.Context) {
        c.Run("razor templates")
    }).Src("templates/**/*.go.html")
}

func main() {
    do.Godo(tasks)
}
```

To run "server" task from parent dir of `Gododir/`

    godo server

To rerun "server" and its dependencies whenever any of their watched files change

    godo server --watch

To run the "default" task which runs "hello" and "build"

    godo

Task names may add a "?" suffix to execute only once even when watching

```go
// build once regardless of number of dependents
p.Task("assets?", nil, func(*do.Context) { })
```

Task dependencies

    do.S{} or do.Series{} - dependent tasks to run in series
    do.P{} or do.Parallel{} - dependent tasks to run in parallel

    For example, do.S{"clean", do.P{"stylesheets", "templates"}, "build"}


### Task Option Funcs

*   Task#Src() - specify watch paths or the src files for Task#Dest()

        Glob patterns

            /**/   - match zero or more directories
            {a,b}  - match a or b, no spaces
            *      - match any non-separator char
            ?      - match a single non-separator char
            **/    - match any directory, start of pattern only
            /**    - match any in this directory, end of pattern only
            !      - removes files from result set, start of pattern only

*   Task#Dest(globs ...string) - If globs in Src are newer than Dest, then
    the task is run

*   Task#Desc(description string) - Set task's description in usage.

*   Task#Debounce(duration time.Duration) - Disallow a task from running until duration
    has elapsed.

*   Task#Deps(names ...interface{}) - Can be `S, Series, P, Parallel, string`


### Task CLI Arguments

Task CLI arguments follow POSIX style flag convention
(unlike go's built-in flag package). Any command line arguments
succeeding `--` are passed to each task. Note, arguments before `--`
are reserved for `godo`.

As an example,

```go
p.Task("hello", nil, func(c *do.Context) {
    // "(none)" is the default value
    msg := c.Args.MayString("(none)", "message", "msg", "m")
    var name string
    if len(c.Args.NonFlags()) == 1 {
        name = c.Args.NonFlags()[0]
    }
    fmt.Println(msg, name)
})
```

running

```sh
# prints "(none)"
godo hello

# prints "Hello dude" using POSIX style flags
godo hello -- dude --message Hello
godo hello -- dude --msg Hello
godo hello -- -m Hello dude
```

Args functions are categorized as

*  `Must*` - Argument must be set by user or panic.

    ```go
c.Args.MustInt("number", "n")
```

*  `May*` - If argument is not set, default to first value.

    ```go
// defaults to 100
c.Args.MayInt(100, "number", "n")
```

*  `As*` - If argument is not set, default to zero value.

    ```go
// defaults to 0
c.Args.AsInt("number", "n")
```


## Modularity and Namespaces

A project may include other tasks functions with `Project#Use`. `Use` requires a namespace to
prevent task name conflicts with existing tasks.

```go
func buildTasks(p *do.Project) {
    p.Task("default", S{"clean"}, nil)

    p.Task("clean", nil, func(*do.Context) {
        fmt.Println("build clean")
    })
}

func tasks(p *do.Project) {
    p.Use("build", buildTasks)

    p.Task("clean", nil, func(*do.Context) {
        fmt.Println("root clean")
    })

    p.Task("build", do.S{"build:default"}, func(*do.Context) {
        fmt.Println("root clean")
    })
}
```

Running `godo build:.` or `godo build` results in output of `build clean`. Note that
it uses the `clean` task in its namespace not the `clean` in the parent project.

The special name `build:.` is alias for `build:default`.

Task dependencies that start with `"/"` are relative to the parent project and
may be called referenced from sub projects.

## godobin

`godo` compiles `Godofile.go` to `godobin-VERSION` (`godobin-VERSION.exe` on Windows) whenever
`Godofile.go` changes. The binary file is built into the same directory as
`Godofile.go` and should be ignored by adding the path `godobin*` to `.gitignore`.

## Exec functions

All of these functions accept a `map[string]interface{}` or `M` for
options. Option keys that start with `"$"` are reserved for `godo`.
Other fields can be used as context for template.

### Bash

Bash functions uses the bash executable and may not run on all OS.

Run a bash script string. The script can be multiline line with continutation.

```go
c.Bash(`
    echo -n $USER
    echo some really long \
        command
`)
```

Bash can use Go templates

```go
c.Bash(`echo -n {{.name}}`, do.M{"name": "mario", "$in": "cmd/bar"})
```

Run a bash script and capture STDOUT and STDERR.

```go
output, err := c.BashOutput(`echo -n $USER`)
```

### Run

Run `go build` inside of cmd/app and set environment variables.

```go
c.Run(`GOOS=linux GOARCH=amd64 go build`, do.M{"$in": "cmd/app"})
```

Run can use Go templates

```go
c.Run(`echo -n {{.name}}`, do.M{"name": "mario", "$in": "cmd/app"})
```

Run and capture STDOUT and STDERR

```go
output, err := c.RunOutput("whoami")
```

### Start

Start an async command. If the executable has suffix ".go" then it will be "go install"ed then executed.
Use this for watching a server task.

```go
c.Start("main.go", do.M{"$in": "cmd/app"})
```

Godo tracks the process ID of started processes to restart the app gracefully.

### Inside

To run many commands inside a directory, use `Inside` instead of the `$in` option.
`Inside` changes the working directory.

```go
do.Inside("somedir", func() {
    do.Run("...")
    do.Bash("...")
})
```

## User Input

To get plain string

```go
user := do.Prompt("user: ")
```

To get password

```go
password := do.PromptPassword("password: ")
```

## Godofile Run-Time Environment

### From command-line

Environment variables may be set via key-value pairs as arguments to
godo. This feature was added to facilitate users on Windows.

```sh
godo NAME=mario GOPATH=./vendor hello
```

### From source code

To specify whether to inherit from parent's process environment,
set `InheritParentEnv`. This setting defaults to true

```go
do.InheritParentEnv = false
```

To specify the base environment for your tasks, set `Env`.
Separate with whitespace or newlines.

```go
do.Env = `
    GOPATH=.vendor::$GOPATH
    PG_USER=mario
`
```

Functions can add or override environment variables as part of the command string.
Note that environment variables are set before the executable similar to a shell;
however, the `Run` and `Start` functions do not use a shell.

```go
p.Task("build", nil, func(c *do.Context) {
    c.Run("GOOS=linux GOARCH=amd64 go build" )
})
```

The effective environment for exec functions is: `parent (if inherited) <- do.Env <- func parsed env`

Paths should use `::` as a cross-platform path list separator. On Windows `::` is replaced with `;`.
On Mac and linux `::` is replaced with `:`.

### From godoenv file

For special circumstances where the GOPATH needs to be set before building the Gododir,
use `Gododir/godoenv` file.

TIP: Create `Gododir/godoenv` when using a dependency manager like `godep` that necessitates
changing `$GOPATH`

```
# Gododir/godoenv
GOPATH=$PWD/cmd/app/Godeps/_workspace::$GOPATH
```
