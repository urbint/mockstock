package generator

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"
)

type Parser struct {
	file *ast.File
	path string

	pkg *types.Package
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(path string) error {
	dir := filepath.Dir(path)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var astFiles []*ast.File

	fset := token.NewFileSet()

	for _, fi := range files {
		if filepath.Ext(fi.Name()) != ".go" {
			continue
		}

		fpath := filepath.Join(dir, fi.Name())
		// Parse the file containing this very example
		// but stop after processing the imports.
		f, err := parser.ParseFile(fset, fpath, nil, 0)
		if err != nil {
			return err
		}

		if fi.Name() == filepath.Base(path) {
			p.file = f
		}

		astFiles = append(astFiles, f)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Type-check a package consisting of this file.
	// Type information for the imported packages
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.
	conf := types.Config{Importer: importer.Default()}
	conf.DisableUnusedImportCheck = true
	conf.IgnoreFuncBodies = true
	pkg, err := conf.Check(abs, fset, astFiles, nil)
	if err != nil {
		return err
	}

	p.path = abs
	p.pkg = pkg
	return nil
}

func (p *Parser) Find(name string) (*Interface, error) {
	obj := p.pkg.Scope().Lookup(name)
	if obj == nil {
		return nil, ErrNotInterface
	}

	typ := obj.Type().(*types.Named)

	name = typ.Obj().Name()

	iface := typ.Underlying().(*types.Interface).Complete()

	return &Interface{name, p.path, p.file, iface}, nil
}

/*
func (p *Parser) FindOld(name string) (*Interface, error) {
	for _, decl := range p.file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if typespec, ok := spec.(*ast.TypeSpec); ok {
					if typespec.Name.Name == name {
						if iface, ok := typespec.Type.(*ast.InterfaceType); ok {
							return &Interface{name, p.path, p.file, iface}, nil
						} else {
							return nil, ErrNotInterface
						}
					}
				}
			}
		}
	}
	return nil, nil
}
*/

type Interface struct {
	Name string
	Path string
	File *ast.File
	Type *types.Interface
}

func (p *Parser) Interfaces() []*Interface {
	var ifaces []*Interface

	scope := p.pkg.Scope()

	for _, name := range scope.Names() {
		obj := p.pkg.Scope().Lookup(name)
		if obj == nil {
			continue
		}

		typ, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}

		name = typ.Obj().Name()

		iface, ok := typ.Underlying().(*types.Interface)
		if !ok {
			continue
		}

		ifaces = append(ifaces, &Interface{name, p.path, p.file, iface.Complete()})
	}

	return ifaces
}
