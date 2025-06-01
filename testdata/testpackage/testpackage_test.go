package testpackage_test

import (
	"testing"

	"github.com/henvic/globals/testdata/testpackage"
)

var Foo = 1234

func TestSeePackageName(t *testing.T) {
	testpackage.Foo()
}
