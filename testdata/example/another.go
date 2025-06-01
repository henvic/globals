package example

import (
	"fmt"
	"math/rand/v2"
)

func init() {
	fmt.Println("first init")
}

func init() {
	fmt.Println("second init")
}

func AnotherFunction() {
	rand.IntN(10)
}
