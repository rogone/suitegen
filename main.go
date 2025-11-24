package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	sourceFile string
	//destFile    string
	destDir     string
	overwrite   bool
	appendToEnd bool
	help        bool
)

func init() {
	flag.StringVar(&sourceFile, "src", "", "source file name")
	flag.StringVar(&destDir, "outputDir", "", "destination file name, default is source file directory")
	flag.BoolVar(&overwrite, "overwrite", false, "overwrite existing files, dangerous")
	flag.BoolVar(&appendToEnd, "appendToEnd", true, "append to end of test file")
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

	if destDir == "" {
		destDir = filepath.Dir(sourceFile)
		fmt.Println("使用源文件目录:", destDir)
	}

	err := processFile(sourceFile, destDir)
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
