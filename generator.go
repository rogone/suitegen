package main

import (
	_ "embed"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"
)

//go:embed template/suite.tmpl
var testTemplate string

//go:embed template/exported_method.tmpl
var exportedMethod string

//go:embed template/nonexported_method.tmpl
var nonexportedMethod string

// 更新方法信息生成逻辑
func processMethodInfo(method MethodInfo) MethodInfo {
	if method.IsExported {
		method.TestFuncName = "Test" + method.Name
	} else {
		method.TestFuncName = "Test_" + method.Name
	}
	return method
}

func generateTestFiles(structs []StructInfo, packageName, outputFile string) error {
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	for _, structInfo := range structs {
		// 处理所有方法的测试方法名称
		for i := range structInfo.Methods {
			structInfo.Methods[i] = processMethodInfo(structInfo.Methods[i])
		}

		//outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_test.go", strings.ToLower(structInfo.Name)))

		if force {
			if err := createNewTestFile(outputFile, structInfo, packageName, tmpl); err != nil {
				return fmt.Errorf("创建测试文件失败: %v", err)
			}
			fmt.Printf("生成测试文件: %s\n", outputFile)
			continue
		}

		// 检查文件是否存在
		if _, err := os.Stat(outputFile); err == nil {
			// 文件存在，执行增量更新
			//if err := updateExistingTestFile(outputFile, structInfo, packageName, tmpl); err != nil {
			//	return fmt.Errorf("更新现有测试文件失败: %v", err)
			//}
			//fmt.Printf("更新测试文件: %s\n", outputFile)
			fmt.Printf("文件: %s已存在\n", outputFile)
			flag.Usage()
			os.Exit(0)
		} else {
			// 文件不存在，创建新文件
			if err := createNewTestFile(outputFile, structInfo, packageName, tmpl); err != nil {
				return fmt.Errorf("创建测试文件失败: %v", err)
			}
			fmt.Printf("生成测试文件: %s\n", outputFile)
		}
	}

	return nil
}

func createNewTestFile(outputFile string, structInfo StructInfo, packageName string, tmpl *template.Template) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建测试文件失败: %v", err)
	}
	defer file.Close()

	data := TemplateData{
		PackageName: packageName,
		StructInfo:  structInfo,
	}

	return tmpl.Execute(file, data)
}

func updateExistingTestFile(outputFile string, structInfo StructInfo, packageName string, tmpl *template.Template) error {
	// 解析现有测试文件，获取已存在的测试方法
	existingMethods, err := parseExistingTestMethods(outputFile)
	if err != nil {
		return fmt.Errorf("解析现有测试文件失败: %v", err)
	}

	// 过滤掉已存在的测试方法
	var newMethods []MethodInfo
	for _, method := range structInfo.Methods {
		testFuncName := method.TestFuncName
		if !existingMethods[testFuncName] {
			newMethods = append(newMethods, method)
		}
	}

	if len(newMethods) == 0 {
		fmt.Printf("所有测试方法已存在，无需更新: %s\n", outputFile)
		return nil
	}

	// 读取现有文件内容
	content, err := os.ReadFile(outputFile)
	if err != nil {
		return fmt.Errorf("读取现有测试文件失败: %v", err)
	}

	// 找到插入新测试方法的位置
	lines := strings.Split(string(content), "\n")

	// 生成新的测试方法代码
	var newTestMethods strings.Builder

	// 只生成测试方法部分
	for _, method := range newMethods {
		if method.IsExported {
			newTestMethods.WriteString(
				fmt.Sprintf(
					exportedMethod,
					structInfo.Name,
					method.Name,
					method.Name,
					method.Params,
					method.Results,
					method.Name))
		} else {
			newTestMethods.WriteString(
				fmt.Sprintf(
					nonexportedMethod,
					structInfo.Name,
					method.Name,
					method.Name,
					method.Params,
					method.Results,
					method.Name))
		}
	}

	// 插入新的测试方法
	lines = append(lines, newTestMethods.String())

	// 写回文件
	return os.WriteFile(outputFile, []byte(strings.Join(lines, "\n")), 0644)
}

func parseExistingTestMethods(filename string) (map[string]bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析现有测试文件失败: %v", err)
	}

	existingMethods := make(map[string]bool)

	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			// 检查是否是测试套件的方法
			if funcDecl.Recv != nil {
				receiverType := extractReceiverType(funcDecl.Recv)
				if strings.HasSuffix(receiverType, "TestSuite") {
					existingMethods[funcDecl.Name.Name] = true
				}
			}
		}
	}

	return existingMethods, nil
}
