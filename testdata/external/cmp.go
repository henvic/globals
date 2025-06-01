package main

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
)

var ReportThis = true

func main() {
	dontReportThis := true
	if dontReportThis {
		fmt.Println(cmp.Diff("a", "b"))
	}
}
