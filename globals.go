package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type options struct {
	inits      bool
	vars       bool
	skipErrors bool
	skipTests  bool
}

func main() {
	var opts options
	flag.BoolVar(&opts.vars, "vars", true, "report global variables")
	flag.BoolVar(&opts.inits, "only-init", true, "report init functions")
	flag.BoolVar(&opts.skipErrors, "skip-errors", true, "omit global variables of type error")
	flag.BoolVar(&opts.skipTests, "skip-tests", true, "omit analyzing test files")
	flag.Parse()

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	var path string
	args := flag.Args()
	switch len(args) {
	case 0:
		path = workingDir
	case 1:
		path = args[0]
		if path == "./..." || path == "." {
			path = workingDir
		}
	default:
		fmt.Fprintln(os.Stderr, "Usage: globals [path]")
		os.Exit(2)
	}

	pkgCache := map[string]*typecheck{}
	extCache := map[string]*typecheck{}
	if err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip vendor and testdata directories.
			if d.Name() == "vendor" || d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if opts.skipTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		if filepath.Ext(path) == ".go" {
			if err := processFile(pkgCache, extCache, path, workingDir, opts); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// typecheck is a parsed and type-checked Go package.
type typecheck struct {
	Files   []*ast.File
	FileSet *token.FileSet
	Info    *types.Info

	source    map[string]*ast.File
	generated map[string]bool
}

func newTypecheck() *typecheck {
	return &typecheck{
		Files:   []*ast.File{},
		FileSet: token.NewFileSet(),
		Info: &types.Info{
			Defs: make(map[*ast.Ident]types.Object),
			Uses: make(map[*ast.Ident]types.Object),
		},
		source:    map[string]*ast.File{},
		generated: map[string]bool{},
	}
}

func (tc *typecheck) File(name string) (*ast.File, bool) {
	if f, ok := tc.source[name]; ok {
		return f, tc.generated[name]
	}
	return nil, false
}

func (tc *typecheck) Check(dir string) error {
	conf := types.Config{
		Importer: importer.ForCompiler(tc.FileSet, "source", nil),
	}
	_, err := conf.Check(dir, tc.FileSet, tc.Files, tc.Info)
	if err != nil {
		return err
	}
	return nil
}

func (tc *typecheck) AddFile(filename string) error {
	f, err := parser.ParseFile(tc.FileSet, filename, nil, 0)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}
	tc.Files = append(tc.Files, f)
	tc.source[filename] = f
	tc.generated[filename] = isGeneratedFile(filename)
	return nil
}

func (tc *typecheck) analyze(filename, workingDir string, opts options) {
	file, gen := tc.File(filename)
	if gen || file == nil {
		return
	}
	for _, decl := range file.Decls {
		if opts.inits {
			tc.analyzeInit(decl, workingDir) // init() functions
		}
		if opts.vars {
			tc.analyzeGlobalVar(decl, workingDir, opts) // global variables
		}
	}
}

func (tc *typecheck) analyzeGlobalVar(decl ast.Decl, workingDir string, opts options) {
	// Global variables
	genDecl, ok := decl.(*ast.GenDecl)
	if !ok || genDecl.Tok != token.VAR {
		return
	}
	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range valueSpec.Names {
			if name.String() == "_" {
				break
			}
			if opts.skipErrors {
				// Skip variables of type error.
				obj := tc.Info.Defs[name]
				if obj != nil {
					typ := obj.Type()
					if typ != nil {
						// Directly error type.
						if typ.String() == "error" {
							break
						}
						// Implements error interface
						if types.Implements(typ, errorInterface) || types.Implements(types.NewPointer(typ), errorInterface) {
							break
						}
					}
				}

			}
			report(tc.FileSet, name, "var", "", workingDir)
		}
	}
}

// analyzeInit checks if the given declaration is an init function and reports it.
func (tc *typecheck) analyzeInit(decl ast.Decl, workingDir string) {
	funcDecl, ok := decl.(*ast.FuncDecl)
	if ok && funcDecl.Recv == nil && funcDecl.Name.Name == "init" {
		report(tc.FileSet, funcDecl.Name, "", "function", workingDir)
	}
}

func isExternalPackageTest(filename string) bool {
	f, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.PackageClauseOnly)
	return err == nil && f.Name != nil && strings.HasSuffix(f.Name.Name, "_test")
}

// parseAndTypeCheck parses all .go files in dir and returns the files, fsets, and info.
func parseAndTypeCheck(dir string, skipTests bool) (pkgTC *typecheck, extTC *typecheck, err error) {
	pkgTC = newTypecheck()
	extTC = newTypecheck()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}

		// Note: If buggy behavior happens after changes, check if we forgot to replace this with extTC
		// if the file is an external test package.
		tc := pkgTC

		filename := filepath.Join(dir, entry.Name())
		if strings.HasSuffix(entry.Name(), "_test.go") {
			if skipTests {
				continue
			}

			// Check if the file is part of an external test package.
			if isExternalPackageTest(filename) {
				tc = extTC // If it's an external test package, use the separate external typecheck.
			}
		}

		if err := tc.AddFile(filename); err != nil {
			return nil, nil, err
		}
	}

	if err := pkgTC.Check(dir); err != nil {
		return nil, nil, err
	}
	if err := extTC.Check(dir); err != nil {
		return nil, nil, fmt.Errorf("test package: %w", err)
	}
	return pkgTC, extTC, nil
}

// Source of the Regex comes from the Go standard library: https://go-review.googlesource.com/c/go/+/283633
var generatedRegexp = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)

// isGeneratedFile returns true if the file has a "Code generated" comment near the top.
// See https://github.com/golang/go/issues/13560
func isGeneratedFile(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 10 && scanner.Scan(); i++ {
		if generatedRegexp.Match(scanner.Bytes()) {
			return true
		}
	}
	return false
}

// errorInterface is the type of the built-in error interface.
var errorInterface = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

func processFile(pkgCache, extCache map[string]*typecheck, filename, workingDir string, opts options) error {
	var err error
	dir := filepath.Dir(filename)
	pkgTC, pkgOK := pkgCache[dir]
	extTC, extOK := extCache[dir]
	if !pkgOK || !extOK {
		pkgTC, extTC, err = parseAndTypeCheck(dir, opts.skipTests)
		if err != nil {
			return err
		}
		pkgCache[dir] = pkgTC
		extCache[dir] = extTC
	}
	if pkgTC != nil {
		pkgTC.analyze(filename, workingDir, opts)
	}
	if extTC != nil {
		extTC.analyze(filename, workingDir, opts)
	}
	return nil
}

func report(fset *token.FileSet, name *ast.Ident, prefix, suffix, workingDir string) {
	position := fset.Position(name.Pos())
	path := position.Filename
	// Try to use a relative path unless it escapes the working directory.
	rel, err := filepath.Rel(workingDir, path)
	if err == nil && !strings.Contains(rel, "..") {
		path = rel
	}
	res := path + ":" + strconv.Itoa(position.Line) + ": "
	if prefix != "" {
		res += prefix + " "
	}
	res += name.Name
	if suffix != "" {
		res += " " + suffix
	}
	fmt.Fprintln(os.Stderr, res)
}
