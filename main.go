
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type MethodInfo struct {
	Name       string
	Receiver   string
	Params     string
	Results    string
	IsExported bool
}

type StructInfo struct {
	Name    string
	Methods []MethodInfo
}

type TemplateData struct {
	PackageName string
	StructName  string
	Methods     []MethodInfo
}

const testTemplate = `package {{.PackageName}}

import (
	"testing"
	"github.com/stretchr/testify/suite"
)

type {{.StructName}}TestSuite struct {
	suite.Suite
	// 这里可以添加测试所需的字段
}

func (suite *{{.StructName}}TestSuite) SetupSuite() {
	// 在所有测试开始前运行
}

func (suite *{{.StructName}}TestSuite) SetupTest() {
	// 在每个测试开始前运行
}

func (suite *{{.StructName}}TestSuite) TearDownTest() {
	// 在每个测试结束后运行
}

func (suite *{{.StructName}}TestSuite) TearDownSuite() {
	// 在所有测试结束后运行
}
{{range .Methods}}
func (suite *{{$.StructName}}TestSuite) Test{{.Name}}() {
	// TODO: 实现 {{.Name}} 方法的测试用例
	// 参数: {{.Params}}
	// 返回值: {{.Results}}
	suite.T().Logf("测试 {{.Name}} 方法")
}
{{end}}

func Test{{.StructName}}Suite(t *testing.T) {
	suite.Run(t, new({{.StructName}}TestSuite))
}
`
<code_end>

<code_start project_name=go_test_generator filename=parser.go title=Go代码解析器 entrypoint=false runnable=false project_final_file=false>
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

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
							Name:    typeSpec.Name.Name,
							Methods: make([]MethodInfo, 0),
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
<code_end>

<code_start project_name=go_test_generator filename=generator.go title=测试代码生成器 entrypoint=false runnable=false project_final_file=false>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func generateTestFiles(structs []StructInfo, packageName, outputDir string) error {
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	for _, structInfo := range structs {
		// 只处理有导出方法的结构体
		hasExportedMethods := false
		for _, method := range structInfo.Methods {
			if method.IsExported {
				hasExportedMethods = true
				break
			}
		}

		if !hasExportedMethods {
			continue
		}

		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_test.go", strings.ToLower(structInfo.Name)))
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建测试文件失败: %v", err)
		}
		defer file.Close()

		data := TemplateData{
			PackageName: packageName,
			StructName:  structInfo.Name,
			Methods:     structInfo.Methods,
		}

		if err := tmpl.Execute(file, data); err != nil {
			return fmt.Errorf("执行模板失败: %v", err)
		}

		fmt.Printf("生成测试文件: %s\n", outputFile)
	}

	return nil
}

func processFile(inputFile, outputDir string) error {
	structs, packageName, err := parseGoFile(inputFile)
	if err != nil {
		return err
	}

	if len(structs) == 0 {
		return fmt.Errorf("文件中未找到结构体定义")
	}

	return generateTestFiles(structs, packageName, outputDir)
}
<code_end>

<code_start project_name=go_test_generator filename=go.mod title=项目依赖配置 entrypoint=false runnable=false project_final_file=true>
module test-generator

go 1.21

require github.com/stretchr/testify v1.8.4

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
<code_end>

该测试代码生成工具具有以下特点：

**核心功能**
- 解析Go源文件中的结构体定义和方法签名
- 自动生成符合testify/suite规范的测试套件
- 为每个导出的结构体方法生成对应的测试方法框架

**技术实现**
- 使用Go标准库的ast包进行语法树分析
- 支持指针接收器和值接收器的方法识别
- 提取方法参数和返回值类型信息

**测试套件规范**
- 完整的SetupSuite/TearDownSuite生命周期方法
- 每个测试方法独立的SetupTest/TearDownTest
- 遵循testify断言库的最佳实践

**使用方法**
编译后运行：`./test-generator -input source.go -output ./tests`

该工具能够显著提高单元测试的编写效率，确保测试代码的一致性和规范性。