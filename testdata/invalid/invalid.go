package invalid

import (
	"errors"
	"fmt"
)

var (
	X          = map[string]int{"foo": 123}
	Exported   string
	unexported string
	_          int8 // Verify that the type int8 is defined.
)

var x = errors.X // undefined

func init() {
	fmt.Println("another init")
}
