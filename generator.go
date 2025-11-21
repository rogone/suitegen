
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// 更新测试模板以支持未导出方法
const testTemplate = `package {{.PackageName}}

import (
	"testing"
	"github.com/stretchr/testify/suite"
)

type {{.StructName}}TestSuite struct {
	suite.Suite
	// 这里可以添加测试所需的字段
}

// SetupAllSuite 在所有测试套件开始前运行
func (suite *{{.StructName}}TestSuite) SetupAllSuite() {
	// 在所有测试套件开始前运行
}

// TearDownAllSuite 在所有测试套件结束后运行
func (suite *{{.StructName}}TestSuite) TearDownAllSuite() {
	// 在所有测试套件结束后运行
}

// SetupTestSuite 在当前测试套件开始前运行
func (suite *{{.StructName}}TestSuite) SetupTestSuite() {
	// 在当前测试套件开始前运行
}

// TearDownTestSuite 在当前测试套件结束后运行
func (suite *{{.StructName}}TestSuite) TearDownTestSuite() {
	// 在当前测试套件结束后运行
}

// SetupSubTest 在每个子测试开始前运行
func (suite *{{.StructName}}TestSuite) SetupSubTest() {
	// 在每个子测试开始前运行
}

// TearDownSubTest 在每个子测试结束后运行
func (suite *{{.StructName}}TestSuite) TearDownSubTest() {
	// 在每个子测试结束后运行
}

// BeforeTest 在每个测试方法开始前运行
func (suite *{{.StructName}}TestSuite) BeforeTest(suiteName, testName string) {
	// 在每个测试方法开始前运行
}

// AfterTest 在每个测试方法结束后运行
func (suite *{{.StructName}}TestSuite) AfterTest(suiteName, testName string) {
	// 在每个测试方法结束后运行
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
{{if .IsExported}}
func (suite *{{$.StructName}}TestSuite) Test{{.Name}}() {
	// TODO: 实现 {{.Name}} 方法的测试用例
	// 参数: {{.Params}}
	// 返回值: {{.Results}}
	suite.T().Logf("测试 {{.Name}} 方法")
}
{{else}}
func (suite *{{$.StructName}}TestSuite) Test_{{.Name}}() {
	// TODO: 实现未导出的 {{.Name}} 方法的测试用例
	// 参数: {{.Params}}
	// 返回值: {{.Results}}
	suite.T().Logf("测试未导出的 {{.Name}} 方法")
}
{{end}}
{{end}}

func Test{{.StructName}}Suite(t *testing.T) {
	suite.Run(t, new({{.StructName}}TestSuite))
}
`

// 更新方法信息结构体，添加测试方法名称字段
type MethodInfo struct {
	Name          string
	Receiver      string
	Params        string
	Results       string
	IsExported    bool
	TestFuncName  string // 新增：测试方法名称
}

// 更新方法信息生成逻辑
func processMethodInfo(method MethodInfo) MethodInfo {
	if method.IsExported {
		method.TestFuncName = "Test" + method.Name
	} else {
		method.TestFuncName = "Test_" + method.Name
	}
	return method
}

func generateTestFiles(structs []StructInfo, packageName, outputDir string) error {
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	for _, structInfo := range structs {
		// 处理所有方法的测试方法名称
		for i := range structInfo.Methods {
			structInfo.Methods[i] = processMethodInfo(structInfo.Methods[i])
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
<code_end>

主要更新内容：

1. **命名规则区分**：导出的方法使用`TestMethodName`格式，未导出的方法使用`Test_methodName`格式 

2. **条件模板生成**：使用`{{if .IsExported}}`和`{{else}}`来区分生成不同的测试方法

3. **测试方法名称字段**：新增`TestFuncName`字段来存储格式化后的测试方法名称

4. **方法信息处理**：添加`processMethodInfo`函数来统一处理测试方法的命名逻辑

该实现确保了测试代码的规范性和一致性，符合Go语言测试的最佳实践 
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// 更新测试模板以支持未导出方法
const testTemplate = `package {{.PackageName}}

import (
	"testing"
	"github.com/stretchr/testify/suite"
)

type {{.StructName}}TestSuite struct {
	suite.Suite
	// 这里可以添加测试所需的字段
}

// SetupAllSuite 在所有测试套件开始前运行
func (suite *{{.StructName}}TestSuite) SetupAllSuite() {
	// 在所有测试套件开始前运行
}

// TearDownAllSuite 在所有测试套件结束后运行
func (suite *{{.StructName}}TestSuite) TearDownAllSuite() {
	// 在所有测试套件结束后运行
}

// SetupTestSuite 在当前测试套件开始前运行
func (suite *{{.StructName}}TestSuite) SetupTestSuite() {
	// 在当前测试套件开始前运行
}

// TearDownTestSuite 在当前测试套件结束后运行
func (suite *{{.StructName}}TestSuite) TearDownTestSuite() {
	// 在当前测试套件结束后运行
}

// SetupSubTest 在每个子测试开始前运行
func (suite *{{.StructName}}TestSuite) SetupSubTest() {
	// 在每个子测试开始前运行
}

// TearDownSubTest 在每个子测试结束后运行
func (suite *{{.StructName}}TestSuite) TearDownSubTest() {
	// 在每个子测试结束后运行
}

// BeforeTest 在每个测试方法开始前运行
func (suite *{{.StructName}}TestSuite) BeforeTest(suiteName, testName string) {
	// 在每个测试方法开始前运行
}

// AfterTest 在每个测试方法结束后运行
func (suite *{{.StructName}}TestSuite) AfterTest(suiteName, testName string) {
	// 在每个测试方法结束后运行
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
{{if .IsExported}}
func (suite *{{$.StructName}}TestSuite) Test{{.Name}}() {
	// TODO: 实现 {{.Name}} 方法的测试用例
	// 参数: {{.Params}}
	// 返回值: {{.Results}}
	suite.T().Logf("测试 {{.Name}} 方法")
}
{{else}}
func (suite *{{$.StructName}}TestSuite) Test_{{.Name}}() {
	// TODO: 实现未导出的 {{.Name}} 方法的测试用例
	// 参数: {{.Params}}
	// 返回值: {{.Results}}
	suite.T().Logf("测试未导出的 {{.Name}} 方法")
}
{{end}}
{{end}}

func Test{{.StructName}}Suite(t *testing.T) {
	suite.Run(t, new({{.StructName}}TestSuite))
}
`

// 更新方法信息结构体，添加测试方法名称字段
type MethodInfo struct {
	Name          string
	Receiver      string
	Params        string
	Results       string
	IsExported    bool
	TestFuncName  string // 新增：测试方法名称
}

// 更新方法信息生成逻辑
func processMethodInfo(method MethodInfo) MethodInfo {
	if method.IsExported {
		method.TestFuncName = "Test" + method.Name
	} else {
		method.TestFuncName = "Test_" + method.Name
	}
	return method
}

func generateTestFiles(structs []StructInfo, packageName, outputDir string) error {
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	for _, structInfo := range structs {
		// 处理所有方法的测试方法名称
		for i := range structInfo.Methods {
			structInfo.Methods[i] = processMethodInfo(structInfo.Methods[i])
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
<code_end>

主要更新内容：

1. **命名规则区分**：导出的方法使用`TestMethodName`格式，未导出的方法使用`Test_methodName`格式 

2. **条件模板生成**：使用`{{if .IsExported}}`和`{{else}}`来区分生成不同的测试方法

3. **测试方法名称字段**：新增`TestFuncName`字段来存储格式化后的测试方法名称

4. **方法信息处理**：添加`processMethodInfo`函数来统一处理测试方法的命名逻辑

该实现确保了测试代码的规范性和一致性，符合Go语言测试的最佳实践 
package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func generateTestFiles(structs []StructInfo, packageName, outputDir string) error {
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	for _, structInfo := range structs {
		// 处理所有方法的测试方法名称
		for i := range structInfo.Methods {
			structInfo.Methods[i] = processMethodInfo(structInfo.Methods[i])
		}

		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_test.go", strings.ToLower(structInfo.Name)))
		
		// 检查文件是否存在
		if _, err := os.Stat(outputFile); err == nil {
			// 文件存在，执行增量更新
			if err := updateExistingTestFile(outputFile, structInfo, packageName, tmpl); err != nil {
				return fmt.Errorf("更新现有测试文件失败: %v", err)
			}
			fmt.Printf("更新测试文件: %s\n", outputFile)
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
		StructName:  structInfo.Name,
		Methods:     structInfo.Methods,
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
	var insertIndex int
	var foundTestSuite bool

	for i, line := range lines {
		if strings.Contains(line, "func Test"+structInfo.Name+"Suite(t *testing.T) {") {
			insertIndex = i
			foundTestSuite = true
			break
		}
	}

	if !foundTestSuite {
		return fmt.Errorf("在现有文件中未找到测试套件入口函数")
	}

	// 生成新的测试方法代码
	var newTestMethods strings.Builder
	data := TemplateData{
		PackageName: packageName,
		StructName:  structInfo.Name,
		Methods:     newMethods,
	}
	
	// 只生成测试方法部分
	for _, method := range newMethods {
		if method.IsExported {
			newTestMethods.WriteString(fmt.Sprintf(`
func (suite *%sTestSuite) Test%s() {
	// TODO: 实现 %s 方法的测试用例
	// 参数: %s
	// 返回值: %s
	suite.T().Logf("测试 %s 方法")
}
`, structInfo.Name, method.Name, method.Name, method.Params, method.Results, method.Name))
		} else {
			newTestMethods.WriteString(fmt.Sprintf(`
func (suite *%sTestSuite) Test_%s() {
	// TODO: 实现未导出的 %s 方法的测试用例
	// 参数: %s
	// 返回值: %s
	suite.T().Logf("测试未导出的 %s 方法")
}
`, structInfo.Name, method.Name, method.Name, method.Params, method.Results, method.Name))
		}
	}

	// 插入新的测试方法
	newLines := make([]string, 0)
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, newTestMethods.String())
	newLines = append(newLines, lines[insertIndex:]...)

	// 写回文件
	return os.WriteFile(outputFile, []byte(strings.Join(newLines, "\n")), 0644)
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

	return existingMethods, nil
}
<code_end>

主要功能特点：
- 增量更新机制：检查现有测试文件，只添加缺失的测试方法
- 智能解析：解析现有测试文件中的方法签名，避免重复生成
- 精确插入：在测试套件入口函数前插入新测试方法
- 保持完整性：保留原有测试方法的实现逻辑和注释
- 错误处理：完善的错误检查和恢复机制

该实现确保了测试代码的持续集成和版本管理的友好性，同时避免了重复代码和冲突问题。