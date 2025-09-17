package main

import (
	"go/types"
	"path/filepath"
)

func (f *Finder) processTypeInScope(
	obj types.Object, dirPath string, pkg *types.Package,
) {
	typeName, ok := obj.(*types.TypeName)
	if !ok {
		return
	}

	namedType, ok := typeName.Type().(*types.Named)
	if !ok {
		return
	}

	if !f.isStructType(namedType) {
		return
	}

	if !f.typeImplementsInterface(namedType) {
		return
	}

	impl := f.createImplementation(dirPath, pkg, typeName)
	f.results = append(f.results, impl)
}

func (f *Finder) isStructType(namedType *types.Named) bool {
	_, ok := namedType.Underlying().(*types.Struct)

	return ok
}

func (f *Finder) createImplementation(
	dirPath string, pkg *types.Package, typeName *types.TypeName,
) Implementation {
	relPath, _ := filepath.Rel(".", dirPath)
	packagePath := filepath.Join(f.modulePath, relPath)
	packagePath = filepath.ToSlash(packagePath)

	return Implementation{
		Package:     pkg.Name(),
		Struct:      typeName.Name(),
		PackagePath: packagePath,
	}
}
