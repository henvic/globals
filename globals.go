package main

import (
	"flag"
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

type options struct {
	inits  bool
	vars   bool
	errors bool
	regexp bool
}

func main() {
	singlechecker.Main(NewAnalyzer())
}

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "globals",
		Doc:  "reports global variables and init functions",
		Flags: func() flag.FlagSet {
			fs := flag.NewFlagSet("globals", flag.ExitOnError)
			fs.Bool("inits", true, "report init functions")
			fs.Bool("vars", true, "report global variables")
			fs.Bool("include-errors", false, "don't omit global variables of type error")
			fs.Bool("include-regexp", false, "don't omit global variables of type *regexp.Regexp (regular expressions)")
			return *fs
		}(),
		Run: run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	opts := options{
		inits:  pass.Analyzer.Flags.Lookup("inits").Value.(flag.Getter).Get().(bool),
		vars:   pass.Analyzer.Flags.Lookup("vars").Value.(flag.Getter).Get().(bool),
		errors: pass.Analyzer.Flags.Lookup("include-errors").Value.(flag.Getter).Get().(bool),
		regexp: pass.Analyzer.Flags.Lookup("include-regexp").Value.(flag.Getter).Get().(bool),
	}

	for _, file := range pass.Files {
		// Skip generated files
		if isGeneratedFile(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				// Global variables.
				if d.Tok == token.VAR && opts.vars {
					for _, spec := range d.Specs {
						vs, ok := spec.(*ast.ValueSpec)
						if !ok {
							continue
						}
						if embeds(vs) {
							continue
						}
						for _, name := range vs.Names {
							if name.Name == "_" {
								continue
							}
							obj := pass.TypesInfo.Defs[name]
							if obj == nil {
								continue
							}
							typ := obj.Type()
							if typ == nil {
								continue
							}
							if !opts.errors {
								// Skip variables of type error.
								if typ.String() == "error" ||
									types.Implements(typ, errorInterface) ||
									types.Implements(types.NewPointer(typ), errorInterface) {
									continue
								}
							}
							// Skip variables of type *regexp.Regexp.
							if !opts.regexp && typ.String() == "*regexp.Regexp" {
								continue
							}
							pass.Reportf(name.Pos(), "var %s", name.Name)
						}
					}
				}
			case *ast.FuncDecl:
				// Init functions.
				if opts.inits && d.Recv == nil && d.Name.Name == "init" {
					pass.Reportf(d.Name.Pos(), "init function")
				}
			}
		}
	}
	return nil, nil
}

// embeds checks if a comment group contains a go:embed directive.
func embeds(vs *ast.ValueSpec) bool {
	if vs.Doc != nil {
		for _, c := range vs.Doc.List {
			if strings.HasPrefix(c.Text, "//go:embed") {
				return true
			}
		}
	}
	return false
}

// Source of the Regex comes from the Go standard library: https://go-review.googlesource.com/c/go/+/283633
var generatedRegexp = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)

// isGeneratedFile returns true if the *ast.File has a "Code generated" comment near the top.
// See https://github.com/golang/go/issues/13560
func isGeneratedFile(f *ast.File) bool {
	for i, cg := range f.Comments {
		if i > 5 {
			break // Only check the first few comment groups
		}
		for _, c := range cg.List {
			if generatedRegexp.MatchString(c.Text) {
				return true
			}
		}
	}
	return false
}

// errorInterface is the type of the built-in error interface.
var errorInterface = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
