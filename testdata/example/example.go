package example

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

var (
	X          = map[string]int{"foo": 123}
	Exported   string
	unexported string
	_          int8 // Verify that the type int8 is defined.

	IAmABadError            = errors.New("bad error not starting with Err")
	ErrNotError             func()
	ErrFoo                  = errors.New("foo error")
	ErrBar                  error
	ErrPointer                    = &ErrorPointer{}
	ErrBaz                  error = errors.New("baz")
	ErrComplex                    = fmt.Errorf("complex error: %w", ErrFoo)
	ErrFake                       = 123
	ErrFool                 error = ErrorFool{}
	ErrMoreFool                   = ErrorFool{}
	ErrFew, ErrMulti              = ErrorFool{}, ErrBar
	ErrUninitializedPointer *ErrorPointer
)

const (
	ExportedConst   = "exported"
	unexportedConst = "unexported"
)

var (
	ExportedAnonymous   = func() {}
	unexportedAnonymous = func() {}
)

type ErrorPointer struct{}

func (*ErrorPointer) Error() string {
	return "pointer error"
}

type ErrorFool struct{}

func (ErrorFool) Error() string {
	return "fool error"
}

func ExportedFunc() {}

func UnexportedFunc() {}

func Something() {
	rand.IntN(10)
}

func init() {
	fmt.Println("another init")
}
