// Package donotreport contains only code that shall never be reported.
package donotreport

func ErrFoo() string {
	return "foo"
}

func Something() (int, string, float64, error) {
	var err error

	var (
		a          int = 32
		NotAGlobal string
		C          float64
	)

	NotAGlobal = "this is not a global variable"

	var init = func() {}
	init()

	return a, NotAGlobal, C, err
}
