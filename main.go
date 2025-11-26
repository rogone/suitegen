package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	sourceFile string
	//destFile    string
	dest        string
	force       bool
	appendToEnd bool
	help        bool
)

func init() {
	flag.StringVar(&sourceFile, "src", "", "source file name")
	flag.StringVar(&dest, "o", "", "destination file name, default is source file like src_test.go")
	flag.BoolVar(&force, "f", false, "force to generate, will force existing files, dangerous")
	//flag.BoolVar(&appendToEnd, "appendToEnd", true, "append to end of test file")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&help, "h", false, "show help")
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if sourceFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	checkSourceFile(sourceFile)

	if dest == "" {
		dest = destFile(sourceFile)
	}
	fmt.Println("使用目标文件名:", dest)

	err := processFile(sourceFile, dest)
	if err != nil {
		fmt.Printf("处理错误：%v", err)
	}
}

func checkSourceFile(sourceFile string) {
	info, err := os.Stat(sourceFile)
	if os.IsNotExist(err) {
		fmt.Printf("文件不存在:%s\n", sourceFile)
		flag.Usage()
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("发生错误:%v\n", err)
		flag.Usage()
		os.Exit(1)
	} else {
		if !info.Mode().IsRegular() {
			fmt.Println("文件普通文件")
			flag.Usage()
			os.Exit(1)
		}
	}
}

func destFile(destFile string) string {
	absPath, err := filepath.Abs(sourceFile)
	if err != nil {
		log.Fatalf("Invalid file path: %v", err)
	}

	dir := filepath.Dir(absPath)
	base := filepath.Base(absPath)

	// 写入 _test.go 文件
	testFileName := strings.TrimSuffix(base, ".go") + "_test.go"
	return filepath.Join(dir, testFileName)
}

func processFile(inputFile, outputFile string) error {
	structs, packageName, err := parseGoFile(inputFile)
	if err != nil {
		return err
	}

	if len(structs) == 0 {
		return fmt.Errorf("文件中未找到结构体定义")
	}

	return generateTestFiles(structs, packageName, outputFile)
}
