# globals
A command-line tool to help you identify and reduce global state on your source code.
It reports mutable names such as of global variables (except errors) and init functions.

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
  -only-init
    	report init functions (default true)
  -skip-errors
    	omit global variables of type error (default true)
  -skip-tests
    	omit analyzing test files (default true)
  -vars
    	report global variables (default true)
```

## Example output
```shell
$ globals
main.go:8: init function
main.go:12: init function
main.go:14: var Enabled
main.go:16: var Configuration
```
