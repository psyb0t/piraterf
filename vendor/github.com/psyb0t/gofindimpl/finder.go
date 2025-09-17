package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type Implementation struct {
	Package     string `json:"package"`
	Struct      string `json:"struct"`
	PackagePath string `json:"packagePath"`
}

type Finder struct {
	fset             *token.FileSet
	interfaceName    string
	interfaceMethods []string
	modulePath       string
	results          []Implementation
	config           *types.Config
}

type noopImporter struct{}

func (imp *noopImporter) Import(path string) (*types.Package, error) {
	// Return a minimal package for any import
	// This allows type checking to proceed without actually resolving imports
	return types.NewPackage(path, path), nil
}

func NewFinder(interfaceName string) *Finder {
	fset := token.NewFileSet()
	config := &types.Config{
		Importer: &noopImporter{},
		Error: func(_ error) {
			// Ignore type checking errors for incomplete packages
		},
	}

	return &Finder{
		fset:          fset,
		interfaceName: interfaceName,
		results:       make([]Implementation, 0),
		config:        config,
	}
}

func (f *Finder) validateGoModRoot() error {
	if _, err := os.Stat("./go.mod"); os.IsNotExist(err) {
		return ErrGoModNotFound
	}

	return nil
}

func (f *Finder) loadModulePath() error {
	content, err := os.ReadFile("./go.mod")
	if err != nil {
		return fmt.Errorf(
			"failed to read go.mod: %w",
			err,
		)
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			f.modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module"))

			return nil
		}
	}

	return ErrNoModuleDeclaration
}

func (f *Finder) parseInterface(filePath string) error {
	file, err := parser.ParseFile(f.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf(
			"failed to parse interface file %s: %w",
			filePath,
			err,
		)
	}

	found := false

	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}

		if ts, ok := n.(*ast.TypeSpec); ok {
			if ts.Name.Name == f.interfaceName {
				if iface, ok := ts.Type.(*ast.InterfaceType); ok {
					f.interfaceMethods = f.getInterfaceMethods(iface)
					found = true

					return false
				}
			}
		}

		return true
	})

	if !found {
		return fmt.Errorf("%w '%s' in %s",
			ErrInterfaceNotFound, f.interfaceName, filePath)
	}

	return nil
}

func (f *Finder) getInterfaceMethods(iface *ast.InterfaceType) []string {
	var methods []string

	for _, method := range iface.Methods.List {
		if len(method.Names) > 0 {
			methods = append(methods, method.Names[0].Name)
		}
	}

	return methods
}

func (f *Finder) scanDirectory(searchDir string) error {
	logrus.Debugf("Starting scan of directory: %s", searchDir)

	err := filepath.Walk(
		searchDir,
		func(
			path string,
			info os.FileInfo,
			err error,
		) error {
			if err != nil {
				logrus.WithError(err).
					Debugf("Walk error for path: %s", path)

				return err
			}

			logrus.WithFields(logrus.Fields{
				"path":  path,
				"isDir": info.IsDir(),
				"name":  info.Name(),
			}).Debug("Walking path")

			if !info.IsDir() || strings.HasPrefix(info.Name(), ".") {
				return nil
			}

			if info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}

			logrus.Debugf("Analyzing directory: %s", path)
			f.analyzeDirectory(path)

			return nil
		})
	if err != nil {
		return fmt.Errorf(
			"failed to scan directory: %w",
			err,
		)
	}

	return nil
}

func (f *Finder) analyzeDirectory(dirPath string) {
	logrus.Debugf("Analyzing directory: %s", dirPath)

	files, err := f.parsePackageFiles(dirPath)
	if err != nil {
		logrus.WithError(err).
			Debugf("Error parsing files in: %s", dirPath)

		return
	}

	if len(files) == 0 {
		logrus.Debugf("No files found in: %s", dirPath)

		return
	}

	logrus.WithField("count", len(files)).
		Debugf("Found files in: %s", dirPath)

	pkg, err := f.typeCheckPackage(files)
	if err != nil {
		logrus.WithError(err).
			Debugf("Type check failed for: %s", dirPath)

		return
	}

	logrus.Debugf("Successfully type-checked package: %s", pkg.Name())
	f.findImplementationsInTypedPackage(dirPath, pkg)
}

func (f *Finder) parsePackageFiles(dirPath string) ([]*ast.File, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read directory: %w",
			err,
		)
	}

	files := make([]*ast.File, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") ||
			strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())

		file, err := parser.ParseFile(f.fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		files = append(files, file)
	}

	return files, nil
}

func (f *Finder) typeCheckPackage(files []*ast.File) (*types.Package, error) {
	if len(files) == 0 {
		return nil, ErrNoFilesToTypeCheck
	}

	pkgName := files[0].Name.Name

	pkg, err := f.config.Check(pkgName, f.fset, files, nil)
	if err != nil {
		// Try to continue even if type checking fails
		logrus.WithError(err).Debug("Type checking had errors (continuing anyway)")
	}

	if pkg == nil {
		return nil, fmt.Errorf(
			"type checking failed completely: %w",
			err,
		)
	}

	return pkg, nil
}

func (f *Finder) findImplementationsInTypedPackage(
	dirPath string, pkg *types.Package,
) {
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		f.processTypeInScope(obj, dirPath, pkg)
	}
}

func (f *Finder) typeImplementsInterface(namedType *types.Named) bool {
	if len(f.interfaceMethods) == 0 {
		return false
	}

	// Check both value and pointer method sets
	valueMethodSet := types.NewMethodSet(namedType)
	pointerMethodSet := types.NewMethodSet(types.NewPointer(namedType))

	foundMethods := make(map[string]bool)

	// Add methods from both sets
	for i := 0; i < valueMethodSet.Len(); i++ {
		method := valueMethodSet.At(i)
		foundMethods[method.Obj().Name()] = true
	}

	for i := 0; i < pointerMethodSet.Len(); i++ {
		method := pointerMethodSet.At(i)
		foundMethods[method.Obj().Name()] = true
	}

	for _, requiredMethod := range f.interfaceMethods {
		if !foundMethods[requiredMethod] {
			return false
		}
	}

	return true
}

func (f *Finder) getResults() []Implementation {
	return f.results
}
