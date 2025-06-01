package main_test

import (
	"fmt"
	"testing"

	"github.com/henvic/globals/testdata/main"
)

var Foo = 1234

func TestNotMain(t *testing.T) {
	fmt.Println(main.Main)
}
