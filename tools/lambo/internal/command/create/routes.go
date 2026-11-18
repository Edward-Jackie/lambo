package create

import (
	"go/ast"
	"go/token"
	"path"
	"strconv"
	"strings"
)

func routeString(expr ast.Expr, literals map[string]string) (string, bool) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		literal, err := strconv.Unquote(value.Value)
		return literal, value.Kind == token.STRING && err == nil
	case *ast.Ident:
		literal, ok := literals[value.Name]
		return literal, ok
	case *ast.BinaryExpr:
		if value.Op == token.ADD {
			left, leftOK := routeString(value.X, literals)
			right, rightOK := routeString(value.Y, literals)
			return left + right, leftOK && rightOK
		}
	}
	return "", false
}

func collectGoRoutes(file *ast.File, values map[string]string, names map[string]bool) {
	factories := make(map[string]string)
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		alias := path.Base(importPath)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		switch {
		case importPath == "github.com/gin-gonic/gin":
			factories[alias] = "gin"
		case strings.HasSuffix(importPath, "/internal/platform/httpserver"):
			factories[alias] = "httpserver"
		}
	}
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		groups := make(map[string]string)
		for _, param := range function.Type.Params.List {
			if isGinRouterType(param.Type, factories) {
				for _, name := range param.Names {
					groups[name.Name] = ""
				}
			}
		}
		literals := make(map[string]string, len(values))
		for name, value := range values {
			literals[name] = value
		}
		bind := func(name *ast.Ident, expr ast.Expr) {
			literal, ok := routeString(expr, literals)
			if ok {
				literals[name.Name] = literal
			}
			prefix, ok := routeGroup(expr, groups, literals, factories)
			if ok {
				groups[name.Name] = prefix
			}
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			switch item := node.(type) {
			case *ast.AssignStmt:
				for index, rhs := range item.Rhs {
					if index < len(item.Lhs) {
						if name, ok := item.Lhs[index].(*ast.Ident); ok {
							bind(name, rhs)
						}
					}
				}
			case *ast.ValueSpec:
				for index, value := range item.Values {
					bind(item.Names[index], value)
				}
			case *ast.CallExpr:
				addGoRoute(item, groups, literals, factories, names)
			}
			return true
		})
	}
}

func isGinRouterType(expr ast.Expr, factories map[string]string) bool {
	if pointer, ok := expr.(*ast.StarExpr); ok {
		expr = pointer.X
	}
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name, ok := selector.X.(*ast.Ident)
	if !ok || factories[name.Name] != "gin" {
		return false
	}
	switch selector.Sel.Name {
	case "Engine", "RouterGroup", "IRouter", "IRoutes":
		return true
	}
	return false
}

func routeGroup(expr ast.Expr, groups, literals, factories map[string]string) (string, bool) {
	switch item := expr.(type) {
	case *ast.Ident:
		prefix, ok := groups[item.Name]
		return prefix, ok
	case *ast.CallExpr:
		selector, ok := item.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		if name, ok := selector.X.(*ast.Ident); ok {
			factory := factories[name.Name]
			if (factory == "gin" && (selector.Sel.Name == "New" || selector.Sel.Name == "Default")) ||
				(factory == "httpserver" && selector.Sel.Name == "NewEngine") {
				return "", true
			}
		}
		if selector.Sel.Name == "Group" && len(item.Args) > 0 {
			base, baseOK := routeGroup(selector.X, groups, literals, factories)
			if prefix, ok := routeString(item.Args[0], literals); baseOK && ok {
				return joinRoute(base, prefix), true
			}
		}
	}
	return "", false
}

var httpMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}

func addGoRoute(call *ast.CallExpr, groups, literals, factories map[string]string, names map[string]bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || len(call.Args) == 0 {
		return
	}
	prefix, knownRouter := routeGroup(selector.X, groups, literals, factories)
	if !knownRouter {
		return
	}
	methods, routeArg := goRouteMethods(selector.Sel.Name, call.Args, literals)
	if len(methods) == 0 {
		return
	}
	suffix, ok := routeString(call.Args[routeArg], literals)
	if !ok {
		return
	}
	route := joinRoute(prefix, suffix)
	segments := strings.Split(route, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "*") {
			segments[i] = segment[:1]
		}
	}
	for _, method := range methods {
		names[apiPathPrefix+method+" "+strings.Join(segments, "/")] = true
	}
}

func goRouteMethods(method string, args []ast.Expr, literals map[string]string) ([]string, int) {
	switch method {
	case "Any":
		return httpMethods, 0
	case "Handle":
		if len(args) < 2 {
			return nil, 0
		}
		return []string{goHTTPMethod(args[0], literals)}, 1
	case "Match":
		if len(args) < 2 {
			return nil, 0
		}
		list, ok := args[0].(*ast.CompositeLit)
		if !ok {
			return nil, 0
		}
		methods := make([]string, 0, len(list.Elts))
		for _, entry := range list.Elts {
			methods = append(methods, goHTTPMethod(entry, literals))
		}
		return methods, 1
	}
	for _, candidate := range httpMethods {
		if method == candidate {
			return []string{method}, 0
		}
	}
	return nil, 0
}

func goHTTPMethod(expr ast.Expr, literals map[string]string) string {
	if value, ok := routeString(expr, literals); ok {
		return value
	}
	if selector, ok := expr.(*ast.SelectorExpr); ok && strings.HasPrefix(selector.Sel.Name, "Method") {
		return strings.ToUpper(strings.TrimPrefix(selector.Sel.Name, "Method"))
	}
	return ""
}

func joinRoute(prefix, suffix string) string {
	if suffix == "" && prefix != "" {
		return prefix
	}
	full := path.Join("/", prefix, suffix)
	if full != "/" && strings.HasSuffix(suffix, "/") {
		return full + "/"
	}
	return full
}
