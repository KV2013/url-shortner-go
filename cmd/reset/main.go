package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/KV2013/url-shortner-go/internal/logger"
	"go.uber.org/zap"
)

var basicZero = map[string]string{
	"bool":       "false",
	"string":     `""`,
	"int":        "0",
	"int8":       "0",
	"int16":      "0",
	"int32":      "0",
	"int64":      "0",
	"uint":       "0",
	"uint8":      "0",
	"uint16":     "0",
	"uint32":     "0",
	"uint64":     "0",
	"uintptr":    "0",
	"float32":    "0",
	"float64":    "0",
	"complex64":  "0",
	"complex128": "0",
	"byte":       "0",
	"rune":       "0",
}

type structInfo struct {
	Name    string
	Fields  []*ast.Field
	GenDecl *ast.GenDecl
}

type packageData struct {
	importPath string
	dir        string
	pkgName    string
	structs    []structInfo
	files      []*ast.File
}

type generator struct {
	fset       *token.FileSet
	modulePath string
	moduleRoot string
	known      map[string]bool
	packages   []packageData
	Logger     *zap.Logger
}

func main() {

	Logger, loggerErr := logger.New("info")

	if loggerErr != nil {
		log.Fatal("Ошибка при создании логгера")
	}
	g := &generator{
		fset:   token.NewFileSet(),
		known:  make(map[string]bool),
		Logger: Logger,
	}

	var err error
	g.moduleRoot, g.modulePath, err = findModule()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error finding module:", err)
		os.Exit(1)
	}
	Logger.Debug("Debug", zap.String("moduleRoot", g.moduleRoot), zap.String("modulePath", g.modulePath))

	if err := g.scanAll(); err != nil {
		fmt.Fprintln(os.Stderr, "scan error:", err)
		os.Exit(1)
	}

	if err := g.generateAll(); err != nil {
		fmt.Fprintln(os.Stderr, "generate error:", err)
		os.Exit(1)
	}
}

func findModule() (root, modPath string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	for {
		modFile := filepath.Join(dir, "go.mod")
		data, readErr := os.ReadFile(modFile)
		if readErr == nil {
			for line := range strings.SplitSeq(string(data), "\n") {
				if strings.HasPrefix(line, "module ") {
					return dir, strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func (g *generator) scanAll() error {
	return filepath.WalkDir(g.moduleRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || base == "vendor" || base == "testdata" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		pkg, err := g.scanDir(path)
		if err != nil {
			return nil
		}
		if pkg != nil {
			g.packages = append(g.packages, *pkg)
			for _, s := range pkg.structs {
				g.known[pkg.importPath+"."+s.Name] = true
			}
		}
		return nil
	})
}

func (g *generator) scanDir(dir string) (*packageData, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var goFiles []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, ".gen.go") {
			continue
		}
		goFiles = append(goFiles, filepath.Join(dir, name))
	}
	if len(goFiles) == 0 {
		return nil, nil
	}

	var (
		pkgName string
		structs []structInfo
		files   []*ast.File
	)

	g.Logger.Debug("scanDir 176", zap.Int("goFiles_len", len(goFiles)), zap.String("dir", dir))
	for _, gf := range goFiles {
		g.Logger.Debug("scanDir 178", zap.String("goFile", gf))
		f, err := parser.ParseFile(g.fset, gf, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		if pkgName == "" {
			pkgName = f.Name.Name
		}
		files = append(files, f)

		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				g.Logger.Debug("scanDir 202", zap.String("goFile", gf))
				if !hasGenerateReset(gd.Doc) && !hasGenerateReset(ts.Doc) {
					continue
				}
				structs = append(structs, structInfo{
					Name:    ts.Name.Name,
					Fields:  st.Fields.List,
					GenDecl: gd,
				})
			}
		}
	}
	g.Logger.Debug("scanDir", zap.Int("structs_len", len(structs)))

	if len(structs) == 0 {
		return nil, nil
	}

	relPath, _ := filepath.Rel(g.moduleRoot, dir)
	importPath := g.modulePath + "/" + filepath.ToSlash(relPath)

	return &packageData{
		importPath: importPath,
		dir:        dir,
		pkgName:    pkgName,
		structs:    structs,
		files:      files,
	}, nil
}

func hasGenerateReset(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	return strings.Contains(cg.Text(), "generate:reset")
}

// ---------- code generation ----------

func (g *generator) generateAll() error {
	for _, pkg := range g.packages {
		if err := g.generatePackage(pkg); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) generatePackage(pkg packageData) error {
	var buf bytes.Buffer
	buf.WriteString("// Code generated by reset; DO NOT EDIT.\n\n package " + pkg.pkgName + "\n\n")

	importMap := buildImportMap(pkg.files)

	for _, si := range pkg.structs {
		var method bytes.Buffer
		receiver := firstLower(si.Name)
		fmt.Fprintf(&method, "func (%s *%s) Reset() {\n", receiver, si.Name)
		fmt.Fprintf(&method, "\tif %s == nil {\n\t\treturn\n\t}\n", receiver)

		for _, field := range si.Fields {
			if len(field.Names) == 0 {
				continue
			}
			for _, name := range field.Names {
				code := g.fieldResetCode(receiver+"."+name.Name, field.Type, pkg.importPath, importMap)
				if code != "" {
					method.WriteString("\t" + code + "\n")
				}
			}
		}

		method.WriteString("}\n")
		buf.WriteString(method.String())
		buf.WriteString("\n")
	}

	out := buf.String()

	formatted, err := format.Source([]byte(out))
	if err != nil {
		return fmt.Errorf("formatting %s: %w\n%s", pkg.dir, err, out)
	}

	genFile := filepath.Join(pkg.dir, "reset.gen.go")
	if err := os.WriteFile(genFile, formatted, 0644); err != nil {
		return err
	}
	fmt.Println("generated", genFile)
	return nil
}

func (g *generator) fieldResetCode(fullName string, typeExpr ast.Expr, thisPkg string, importMap map[string]string) string {
	switch t := typeExpr.(type) {
	case *ast.Ident:
		return g.identReset(fullName, t.Name, thisPkg)
	case *ast.StarExpr:
		return g.pointerReset(fullName, t.X, thisPkg, importMap)
	case *ast.ArrayType:
		return g.arrayReset(fullName, t)
	case *ast.MapType:
		return fmt.Sprintf("clear(%s)", fullName)
	case *ast.SelectorExpr:
		return g.selectorReset(fullName, t, importMap)
	case *ast.FuncType, *ast.ChanType:
		return fmt.Sprintf("%s = nil", fullName)
	}
	return ""
}

func (g *generator) identReset(fullName, typeName string, thisPkg string) string {
	if zero, ok := basicZero[typeName]; ok {
		return fmt.Sprintf("%s = %s", fullName, zero)
	}
	fullType := thisPkg + "." + typeName
	if g.known[fullType] {
		return fmt.Sprintf("%s.Reset()", fullName)
	}
	return ""
}

func (g *generator) pointerReset(fullName string, inner ast.Expr, thisPkg string, importMap map[string]string) string {
	code := g.underPointerReset(fullName, inner, thisPkg, importMap)
	if code == "" {
		return ""
	}
	return fmt.Sprintf("if %s != nil {\n\t\t%s\n\t}", fullName, code)
}

func (g *generator) underPointerReset(fullName string, inner ast.Expr, thisPkg string, importMap map[string]string) string {
	deref := "*" + fullName

	switch t := inner.(type) {
	case *ast.Ident:
		if zero, ok := basicZero[t.Name]; ok {
			return fmt.Sprintf("%s = %s", deref, zero)
		}
		fullType := thisPkg + "." + t.Name
		if g.known[fullType] {
			return fmt.Sprintf("%s.Reset()", fullName)
		}
		return ""
	case *ast.ArrayType:
		if t.Len == nil {
			return fmt.Sprintf("%s = (%s)[:0]", deref, deref)
		}
		arrZero := formatNode(g.fset, t) + "{}"
		return fmt.Sprintf("%s = %s", deref, arrZero)
	case *ast.MapType:
		return fmt.Sprintf("clear(%s)", deref)
	case *ast.SelectorExpr:
		fullType := g.resolveSelector(t, importMap)
		if fullType != "" && g.known[fullType] {
			return fmt.Sprintf("%s.Reset()", fullName)
		}
		return ""
	}
	return ""
}

func (g *generator) arrayReset(fullName string, t *ast.ArrayType) string {
	if t.Len == nil {
		return fmt.Sprintf("%s = %s[:0]", fullName, fullName)
	}
	arrZero := formatNode(g.fset, t) + "{}"
	return fmt.Sprintf("%s = %s", fullName, arrZero)
}

func (g *generator) selectorReset(fullName string, t *ast.SelectorExpr, importMap map[string]string) string {
	fullType := g.resolveSelector(t, importMap)
	if fullType != "" && g.known[fullType] {
		return fmt.Sprintf("%s.Reset()", fullName)
	}
	return ""
}

func (g *generator) resolveSelector(t *ast.SelectorExpr, importMap map[string]string) string {
	pkgIdent, ok := t.X.(*ast.Ident)
	if !ok {
		return ""
	}
	importPath, ok := importMap[pkgIdent.Name]
	if !ok {
		return ""
	}
	return importPath + "." + t.Sel.Name
}

func buildImportMap(files []*ast.File) map[string]string {
	m := make(map[string]string)
	for _, f := range files {
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			var alias string
			if imp.Name != nil {
				alias = imp.Name.Name
			} else {
				alias = filepath.Base(path)
			}
			m[alias] = path
		}
	}
	return m
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "<unknown>"
	}
	return buf.String()
}

func firstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1])
}
