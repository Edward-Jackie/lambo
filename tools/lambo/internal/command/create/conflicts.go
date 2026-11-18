package create

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const apiPathPrefix = "API 路径 "

// checkGoConflicts 检查包级名称和生成路由，避免 DTO 与其他业务模块发生冲突。
func checkGoConflicts(root string, files []outputFile) error {
	for _, file := range files {
		if filepath.Ext(file.Path) != ".go" {
			continue
		}
		if err := safePath(root, file.Path); err != nil {
			return err
		}
		names, err := goDeclarations(file.Path, file.Data)
		if err != nil {
			return err
		}
		dir := filepath.Dir(file.Path)
		dirs := []string{dir}
		if filepath.Base(dir) == "handler" {
			for _, domain := range []string{"business", "system"} {
				other := filepath.Join("backend", "internal", domain, "handler")
				if other != dir {
					dirs = append(dirs, other)
				}
			}
		}
		for _, current := range dirs {
			if err := checkGoDir(root, current, file.Path, names, current == dir); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkGoDir(root, dir, target string, names map[string]bool, samePackage bool) error {
	entries, err := os.ReadDir(filepath.Join(root, dir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() || filepath.ToSlash(path) == target || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			return err
		}
		existing, err := goDeclarations(path, data)
		if err != nil {
			return err
		}
		for name := range existing {
			if names[name] && (samePackage || strings.HasPrefix(name, apiPathPrefix)) {
				return fmt.Errorf("%s 已由 %s 使用，拒绝生成 %s", name, path, target)
			}
		}
	}
	return nil
}

func goDeclarations(path string, data []byte) (map[string]bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, data, 0)
	if err != nil {
		return nil, fmt.Errorf("解析 Go 声明 %s：%w", path, err)
	}
	names := make(map[string]bool)
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.FuncDecl:
			if value.Recv == nil {
				names["Go 标识 "+value.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range value.Specs {
				switch value := spec.(type) {
				case *ast.TypeSpec:
					names["Go 标识 "+value.Name.Name] = true
				case *ast.ValueSpec:
					for i, name := range value.Names {
						if name.Name == "_" {
							continue
						}
						names["Go 标识 "+name.Name] = true
						if strings.HasSuffix(name.Name, "APIPath") && i < len(value.Values) {
							if literal, ok := value.Values[i].(*ast.BasicLit); ok && literal.Kind == token.STRING {
								path, err := strconv.Unquote(literal.Value)
								if err == nil {
									names[apiPathPrefix+path] = true
								}
							}
						}
					}
				}
			}
		}
	}
	return names, nil
}
