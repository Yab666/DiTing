package main

import (
	"ditting/internal/app"
	"ditting/internal/core"
	"ditting/internal/plugin"
	"ditting/internal/rule"
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

	// 2. 加载规则库
	loader := rule.NewRuleLoader()
	rules, err := loader.LoadFromDir("configs/rules")
	if err != nil {
		l.Error("规则加载失败: %v", err)
		return
	}
	l.Info("成功加载 %d 条过滤规则", len(rules))
	matcher := rule.NewMatcher(rules)

	// 3. 初始化扫描配置
	config := &core.ScanConfig{
		ExcludeFiles: []string{".git", "node_modules", "vendor"},
	}

	// 4. 创建扫描器并注入
	realScanner := scanner.NewScanner(config.ExcludeFiles, l)
	engine := app.NewEngine(config, realScanner, l, *verbose)

	// 5. 注册解析插件
	engine.RegisterParser(plugin.NewYamlParser())
	engine.SetMatcher(matcher)

	// 6. 运行引擎
	engine.Run(*scanPath)
}
