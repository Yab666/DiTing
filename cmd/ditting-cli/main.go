package main

import (
	"ditting/internal/app"
	"ditting/internal/core"
	"ditting/internal/scanner"
	"flag"
	"fmt"
)

// ConsoleLogger 是一个简单的控制台日志实现。
type ConsoleLogger struct{}

func (l *ConsoleLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}
func (l *ConsoleLogger) Warn(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}
func (l *ConsoleLogger) Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

func main() {
	// 0. 定义命令行参数
	scanPath := flag.String("path", "../tests/fixtures", "要扫描的目录路径")
	verbose := flag.Bool("v", false, "是否显示详细扫描进度")
	flag.Parse()

	fmt.Println("=== 谛听 (DiTing) 静态扫描工具启动 ===")

	// 1. 初始化日志组件
	l := &ConsoleLogger{}

	// 2. 初始化扫描配置
	config := &core.ScanConfig{
		ExcludeFiles: []string{".git", "node_modules", "vendor"},
	}

	// 3. 创建真实的扫描器零件
	realScanner := scanner.NewScanner(config.ExcludeFiles, l)

	// 4. 将所有依赖“注入”到引擎中
	engine := app.NewEngine(config, realScanner, l, *verbose)

	// 5. 运行引擎
	engine.Run(*scanPath)
}
