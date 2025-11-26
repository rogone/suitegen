package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type ControlFlowNode struct {
	Type string // "if", "else if", "else", "switch-case", "switch-default", "typeswitch-case", "typeswitch-default", "for-if", "range-if"
	Text string // 源码片段
}

type MethodInfo struct {
	Name         string
	Receiver     string
	Params       string
	Results      string
	IsExported   bool
	TestFuncName string // 新增：测试方法名称
	//BranchHints  []ControlFlowNode // 分支提示（用于生成子测试）
	BranchHints []ControlFlowStatement
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
	src, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)
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
					//BranchHints: ExtractControlFlowFromFunc(fset, funcDecl, src),
					BranchHints: ExtractControlFlowGroupedByStatement(funcDecl, fset, src),
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

/*
func extractBranchesFromBody(body *ast.BlockStmt) []Branch {
	var branches []Branch

	// 使用 ast.Inspect 深度遍历 AST
	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt:
			if isNilCheck(x.Cond, "err") {
				branches = append(branches, Branch{Description: "error is not nil", Pos: x.Pos()})
				branches = append(branches, Branch{Description: "error is nil", Pos: x.Pos()})
			} else {
				// 处理 if 条件
				cond := formatNode(x.Cond)
				branches = append(branches, Branch{Description: fmt.Sprintf("when %s is true", cond), Pos: x.Pos()})
				branches = append(branches, Branch{Description: fmt.Sprintf("when %s is false", cond), Pos: x.Pos()})
			}
			// 如果有 else if 或 else，也视为分支（但不重复添加）
			if x.Else != nil {
				if _, ok := x.Else.(*ast.IfStmt); !ok {
					branches = append(branches, Branch{Description: "else branch", Pos: x.Pos()})
				}
			}

		case *ast.SwitchStmt:
			expr := "value"
			if x.Tag != nil {
				expr = formatNode(x.Tag)
			}
			branches = append(branches, Branch{Description: fmt.Sprintf("switch %s: default or unmatched", expr), Pos: x.Pos()})
			// 至少提示一个 case 和 default
			branches = append(branches, Branch{Description: fmt.Sprintf("switch %s: matched case", expr), Pos: x.Pos()})

		case *ast.TypeSwitchStmt:
			//expr := "typed value"
			//if x.Assign != nil {
			//	expr = formatNode(x.Assign)
			//}
			branches = append(branches, Branch{Description: fmt.Sprintf("type switch: matched case"), Pos: x.Pos()})
			branches = append(branches, Branch{Description: fmt.Sprintf("type switch: default"), Pos: x.Pos()})

		case *ast.ForStmt:
			branches = append(branches, Branch{Description: "for loop: iterates at least once", Pos: x.Pos()})
			branches = append(branches, Branch{Description: "for loop: does not iterate", Pos: x.Pos()})

		case *ast.RangeStmt:
			what := formatNode(x.X)
			branches = append(branches, Branch{Description: fmt.Sprintf("range over %s: has elements", what), Pos: x.Pos()})
			branches = append(branches, Branch{Description: fmt.Sprintf("range over %s: empty", what), Pos: x.Pos()})
		}
		return true

	})

	// 去重
	//seen := make(map[string]bool)
	//var unique []Branch
	//for _, b := range branches {
	//	if !seen[b.Description] {
	//		seen[b.Description] = true
	//		unique = append(unique, b)
	//	}
	//}

	//return unique
	return branches
}

// isNilCheck 检查是否是典型的 err != nil 判断
func isNilCheck(expr ast.Expr, varName string) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	ident, ok := bin.X.(*ast.Ident)
	if !ok || ident.Name != varName {
		return false
	}
	if op := bin.Op; op == token.NEQ {
		if nilLit, ok := bin.Y.(*ast.Ident); ok && nilLit.Name == "nil" {
			return true
		}
	}
	return false
}

// formatNode 使用 go/format 将 AST 节点转为字符串（简化处理）
func formatNode(n ast.Expr) string {
	switch x := n.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.BinaryExpr:
		return formatNode(x.X) + " " + x.Op.String() + " " + formatNode(x.Y)
	case *ast.CallExpr:
		return formatNode(x.Fun) + "()"
	case *ast.SelectorExpr:
		return formatNode(x.X) + "." + x.Sel.Name
	default:
		return fmt.Sprintf("%T", n)
	}
}*/

func ExtractControlFlowFromFunc(fset *token.FileSet, funcDecl *ast.FuncDecl, src []byte) []ControlFlowNode {
	var results []ControlFlowNode

	if funcDecl.Body == nil {
		return results
	}

	// 匿名函数用于获取源码文本
	nodeToCode := func(n ast.Node) string {
		start := fset.Position(n.Pos()).Offset
		end := fset.Position(n.End()).Offset
		if start >= 0 && end <= len(src) && start < end {
			return strings.TrimSpace(string(src[start:end]))
		}
		return "<invalid>"
	}

	// 用于标记是否已处理过某个 if 节点（避免重复）
	seen := make(map[ast.Node]bool)

	// 遍历函数体的顶层语句
	for _, stmt := range funcDecl.Body.List {
		switch s := stmt.(type) {
		case *ast.IfStmt:
			// 处理顶层 if
			if !seen[s] {
				extractIfStatements(s, nodeToCode, &results, seen, false)
			}

		case *ast.SwitchStmt:
			// 普通 switch
			for _, cc := range s.Body.List {
				if cas, ok := cc.(*ast.CaseClause); ok {
					if cas.List == nil {
						//results = append(results, ControlFlowNode{
						//	Type: "switch-default",
						//	Text: "default",
						//})
					} else {
						cond := nodeToCode(cas)
						// 去掉 "case ..." 或 "default:" 前缀
						if strings.HasPrefix(cond, "case ") || strings.HasPrefix(cond, "default:") {
							cond = strings.SplitN(cond, ":", 2)[0] // 取 case xxx 这部分
						}
						results = append(results, ControlFlowNode{
							Type: "switch-case",
							Text: cond,
						})
					}
				}
			}

		case *ast.TypeSwitchStmt:
			// 类型 switch
			for _, cc := range s.Body.List {
				if cas, ok := cc.(*ast.CaseClause); ok {
					if cas.List == nil {
						//results = append(results, ControlFlowNode{
						//	Type: "type-default",
						//	Text: "default",
						//})
					} else {
						cond := nodeToCode(cas)
						if strings.HasPrefix(cond, "case ") || strings.HasPrefix(cond, "default:") {
							cond = strings.SplitN(cond, ":", 2)[0]
						}
						results = append(results, ControlFlowNode{
							Type: "type-case",
							Text: cond,
						})
					}
				}
			}

		case *ast.ForStmt:
			// 遍历 for 循环体中的语句（但不再深入嵌套）
			for _, innerStmt := range s.Body.List {
				if ifStmt, ok := innerStmt.(*ast.IfStmt); ok {
					if !seen[ifStmt] {
						extractIfStatements(ifStmt, nodeToCode, &results, seen, true)
					}
				}
			}

		case *ast.RangeStmt:
			// 遍历 range 循环体中的语句
			for _, innerStmt := range s.Body.List {
				if ifStmt, ok := innerStmt.(*ast.IfStmt); ok {
					if !seen[ifStmt] {
						extractIfStatements(ifStmt, nodeToCode, &results, seen, true)
					}
				}
			}
		}
	}

	return results
}

// extractIfStatements 提取单个 if 及其 else if / else 链
// topLevelOnly 表示是否只处理当前 if（不再递归子块）
func extractIfStatements(
	ifStmt *ast.IfStmt,
	nodeToCode func(ast.Node) string,
	results *[]ControlFlowNode,
	seen map[ast.Node]bool,
	isInLoop bool,
) {
	if seen[ifStmt] {
		return
	}
	seen[ifStmt] = true

	kind := "if"
	if isInLoop {
		kind = "loop-if" // 区分 for/range 中的 if
	}

	// 添加 if 条件
	*results = append(*results, ControlFlowNode{
		Type: kind,
		Text: nodeToCode(ifStmt.Cond),
	})

	// 处理 else 分支
	if ifStmt.Else != nil {
		if elseIf, ok := ifStmt.Else.(*ast.IfStmt); ok {
			// 是 else if
			extractIfStatements(elseIf, nodeToCode, results, seen, isInLoop)
		} else {
			// 是 else
			*results = append(*results, ControlFlowNode{
				Type: "else",
				Text: "else",
			})
		}
	}
}
