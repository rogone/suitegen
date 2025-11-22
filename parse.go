package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type MethodInfo struct {
	Name         string
	Receiver     string
	Params       string
	Results      string
	IsExported   bool
	TestFuncName string // 新增：测试方法名称
}

type StructInfo struct {
	Name       string
	IsExported bool
	Methods    []MethodInfo
}

type TemplateData struct {
	PackageName string
	StructInfo
}

func parseGoFile(filename string) ([]StructInfo, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("解析文件失败: %v", err)
	}

	structs := make([]StructInfo, 0)
	packageName := node.Name.Name

	// 收集所有结构体类型
	structTypes := make(map[string]bool)
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						structTypes[typeSpec.Name.Name] = true
						structs = append(structs, StructInfo{
							Name:       typeSpec.Name.Name,
							Methods:    make([]MethodInfo, 0),
							IsExported: ast.IsExported(typeSpec.Name.Name),
						})
					}
				}
			}
		}
	}

	// 收集结构体的方法
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Recv != nil {
			receiverType := extractReceiverType(funcDecl.Recv)
			if structTypes[receiverType] {
				methodInfo := MethodInfo{
					Name:       funcDecl.Name.Name,
					Receiver:   receiverType,
					Params:     extractParams(funcDecl.Type.Params),
					Results:    extractResults(funcDecl.Type.Results),
					IsExported: ast.IsExported(funcDecl.Name.Name),
				}

				// 找到对应的结构体并添加方法
				for i := range structs {
					if structs[i].Name == receiverType {
						structs[i].Methods = append(structs[i].Methods, methodInfo)
						break
					}
				}
			}
		}
	}

	return structs, packageName, nil
}

func extractReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	switch t := field.Type.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func extractParams(params *ast.FieldList) string {
	if params == nil || len(params.List) == 0 {
		return "无参数"
	}

	var paramStrs []string
	for _, param := range params.List {
		if len(param.Names) > 0 {
			for _, name := range param.Names {
				paramStrs = append(paramStrs, fmt.Sprintf("%s %s", name.Name, getTypeString(param.Type)))
			}
		} else {
			paramStrs = append(paramStrs, getTypeString(param.Type))
		}
	}

	return strings.Join(paramStrs, ", ")
}

func extractResults(results *ast.FieldList) string {
	if results == nil || len(results.List) == 0 {
		return "无返回值"
	}

	var resultStrs []string
	for _, result := range results.List {
		resultStrs = append(resultStrs, getTypeString(result.Type))
	}

	return strings.Join(resultStrs, ", ")
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	default:
		return "interface{}"
	}
}
