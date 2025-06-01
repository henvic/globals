# globals
A command-line tool to help you identify and reduce global state on your source code.
It reports mutable names such as of global variables (except errors) and init functions.

## Install

```go
go install github.com/henvic/globals@latest
```

## Usage
Run the tool in your project directory to list all global variables:

```sh
globals ./...
```

## Example output
```shell
$ globals
main.go:8: init function
main.go:12: init function
main.go:14: var Enabled
main.go:16: var Configuration
```
