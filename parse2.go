package main

import (
	"go/ast"
	"go/token"
	"strings"
)

// Branch 表示一个分支项
type Branch struct {
	Type string // "if", "else if", "else", "case", "default"
	Text string // 源码表达式或标签
}

// ControlFlowStatement 表示一条控制流语句及其所有分支
type ControlFlowStatement struct {
	Type      string // "if", "switch", "type-switch", "for-if", "range-if"
	FullStmt  string
	StartLine int
	Branches  []Branch // 分支列表
}

// ExtractControlFlowGroupedByStatement 提取函数体内控制流结构，分组返回
func ExtractControlFlowGroupedByStatement(
	funcDecl *ast.FuncDecl,
	fset *token.FileSet,
	src []byte,
) []ControlFlowStatement {
	var results []ControlFlowStatement

	if funcDecl.Body == nil {
		return results
	}

	// 用于获取 AST 节点对应的源码
	nodeToCode := func(n ast.Node) string {
		start := fset.Position(n.Pos()).Offset
		end := fset.Position(n.End()).Offset
		if start >= 0 && end <= len(src) && start < end {
			return strings.TrimSpace(string(src[start:end]))
		}
		return "<invalid>"
	}

	// seen 记录已处理的 if 节点，避免重复（尤其 else if 链）
	seen := make(map[ast.Node]bool)

	// 遍历顶层语句
	for _, stmt := range funcDecl.Body.List {
		switch s := stmt.(type) {
		case *ast.IfStmt:
			if !seen[s] {
				stmts := extractIfStatement(s, nodeToCode, seen, false, fset)
				results = append(results, stmts)
			}

		case *ast.SwitchStmt:
			results = append(results, extractSwitchStatement(s, nodeToCode, fset))

		case *ast.TypeSwitchStmt:
			results = append(results, extractTypeSwitchStatement(s, nodeToCode, fset))

		case *ast.ForStmt:
			// 提取 for 循环体内第一层的 if
			for _, inner := range s.Body.List {
				if ifStmt, ok := inner.(*ast.IfStmt); ok {
					if !seen[ifStmt] {
						stmts := extractIfStatement(ifStmt, nodeToCode, seen, true, fset)
						stmts.Type = "for-if" // 区别来源
						results = append(results, stmts)
					}
				}
			}

		case *ast.RangeStmt:
			// 提取 range 循环体内第一层的 if
			for _, inner := range s.Body.List {
				if ifStmt, ok := inner.(*ast.IfStmt); ok {
					if !seen[ifStmt] {
						stmts := extractIfStatement(ifStmt, nodeToCode, seen, true, fset)
						stmts.Type = "range-if"
						results = append(results, stmts)
					}
				}
			}
		}
	}

	return results
}

func extractIfStatement(
	ifStmt *ast.IfStmt,
	nodeToCode func(ast.Node) string,
	seen map[ast.Node]bool,
	isInLoop bool,
	fset *token.FileSet, // 新增：需要 fset 获取位置
) ControlFlowStatement {
	var branches []Branch
	seen[ifStmt] = true

	// 提取完整的 if 开头语句: "if ..."
	full := nodeToCode(ifStmt)
	// 截断到第一个 { 之前（去掉 body）
	bodyStart := strings.Index(full, "{")
	var fullStmt string
	if bodyStart > 0 {
		fullStmt = strings.TrimSpace(full[:bodyStart])
	} else {
		fullStmt = full // 没有 {（罕见），保留全部
	}

	// 获取起始行号
	startLine := fset.Position(ifStmt.Pos()).Line

	// 添加 if 条件
	branches = append(branches, Branch{
		Type: "if",
		Text: nodeToCode(ifStmt.Cond),
	})

	// 处理 else / else if 链 ...
	if ifStmt.Else != nil {
		if elseIf, ok := ifStmt.Else.(*ast.IfStmt); ok {
			curr := elseIf
			for curr != nil {
				if !seen[curr] {
					seen[curr] = true
					branches = append(branches, Branch{
						Type: "else if",
						Text: nodeToCode(curr.Cond),
					})
				}
				if curr.Else != nil {
					if next, ok := curr.Else.(*ast.IfStmt); ok {
						curr = next
						continue
					}
				}
				branches = append(branches, Branch{Type: "else", Text: "else"})
				break
			}
		} else {
			branches = append(branches, Branch{Type: "else", Text: "else"})
		}
	}

	typ := "if"
	if isInLoop {
		typ = "loop-if"
	}

	return ControlFlowStatement{
		Type:      typ,
		FullStmt:  fullStmt,
		StartLine: startLine,
		Branches:  branches,
	}
}
func extractSwitchStatement(sw *ast.SwitchStmt, nodeToCode func(ast.Node) string, fset *token.FileSet) ControlFlowStatement {
	var branches []Branch

	// 提取 switch 完整语句
	full := nodeToCode(sw)
	bodyStart := strings.Index(full, "{")
	var fullStmt string
	if bodyStart > 0 {
		fullStmt = strings.TrimSpace(full[:bodyStart])
	} else {
		fullStmt = full
	}
	startLine := fset.Position(sw.Pos()).Line

	for _, cc := range sw.Body.List {
		if cas, ok := cc.(*ast.CaseClause); ok {
			if cas.List == nil {
				branches = append(branches, Branch{Type: "default", Text: "default"})
			} else {
				expr := strings.SplitN(nodeToCode(cas), ":", 2)[0]
				expr = strings.TrimPrefix(strings.TrimSpace(expr), "case ")
				branches = append(branches, Branch{Type: "case", Text: expr})
			}
		}
	}

	return ControlFlowStatement{
		Type:      "switch",
		FullStmt:  fullStmt,
		StartLine: startLine,
		Branches:  branches,
	}
}
func extractTypeSwitchStatement(tsw *ast.TypeSwitchStmt, nodeToCode func(ast.Node) string, fset *token.FileSet) ControlFlowStatement {
	var branches []Branch

	full := nodeToCode(tsw)
	bodyStart := strings.Index(full, "{")
	var fullStmt string
	if bodyStart > 0 {
		fullStmt = strings.TrimSpace(full[:bodyStart])
	} else {
		fullStmt = full
	}
	startLine := fset.Position(tsw.Pos()).Line

	for _, cc := range tsw.Body.List {
		if cas, ok := cc.(*ast.CaseClause); ok {
			if cas.List == nil {
				branches = append(branches, Branch{Type: "default", Text: "default"})
			} else {
				expr := strings.SplitN(nodeToCode(cas), ":", 2)[0]
				expr = strings.TrimPrefix(strings.TrimSpace(expr), "case ")
				branches = append(branches, Branch{Type: "case", Text: expr})
			}
		}
	}

	return ControlFlowStatement{
		Type:      "type-switch",
		FullStmt:  fullStmt,
		StartLine: startLine,
		Branches:  branches,
	}
}
