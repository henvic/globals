# globals
A command-line tool to help you identify and reduce global state on your Go source code (global variables and init() functions).

By default this tool doesn't report global variables for errors and regular expressions (regex).

## Install

```go
go install github.com/henvic/globals@latest
```

## Usage
Run the tool in your project directory to list all global variables:

```shell
$ globals ./...
```

Some flags are available:

```shell
$ globals -h
Usage of globals:
  -include-errors
    	don't omit global variables of type error
  -include-regexp
    	don't omit global variables of type *regexp.Regexp (regular expressions)
  -test
    	indicates whether test files should be analyzed, too (default true)
  -inits
    	report init functions (default true)
  -vars
    	report global variables (default true)
```

## Example output

```shell
$ globals ./...
main.go:8: init function
main.go:12: init function
main.go:14: var Enabled
main.go:16: var Configuration
```
