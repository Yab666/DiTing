package main

import (
	"ditting/internal/app"
	"ditting/internal/plugin"
	"ditting/internal/report"
	"ditting/internal/rule"
	"ditting/internal/scanner"
	"ditting/pkg/config"
	"flag"
	"fmt"
	"strings"
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
	configPath := flag.String("config", "", "全局配置文件的路径（可选，默认优先寻找扫描目录下的 .ditingrc）")
	format := flag.String("format", "console", "报告输出格式 (console, json, csv)")
	output := flag.String("output", "report", "报告输出文件路径（扩展名根据格式自动追加）")
	verbose := flag.Bool("v", false, "是否显示详细扫描进度")
	flag.Parse()

	fmt.Println("=== 谛听 (DiTing) 静态扫描工具启动 ===")

	// 1. 初始化日志组件
	l := &ConsoleLogger{}

	// 2. 靠基础设施层加载配置 (大脑)
	// 这个加载器会自动按照优先级去寻找设置的忽略项
	appConfig, err := config.LoadConfig(*configPath, *scanPath)
	if err != nil {
		l.Error("配置加载出现异常，将使用纯底线默认配置: %v", err)
	}

	// 3. 加载规则库 (根据用户的 config.yaml 或者 默认指定的规则文件夹加载)
	loader := rule.NewRuleLoader()
	rules, err := loader.LoadFromDir(appConfig.Rules)
	if err != nil {
		l.Error("规则加载失败: %v", err)
		return
	}
	l.Info("成功加载 %d 条过滤规则", len(rules))
	matcher := rule.NewMatcher(rules, appConfig)

	// 4. 初始化引擎
	// 将配置里的 `exclude.files` 作为基础忽略参数给扫描器，扫描器在遍历文件夹时直接忽略 `.git`, `node_modules` 等。
	realScanner := scanner.NewScanner(appConfig.Exclude.Files, l)
	engine := app.NewEngine(appConfig, realScanner, l, *verbose)

	// 5. 注册解析插件
	engine.RegisterParser(plugin.NewYamlParser())
	engine.RegisterParser(plugin.NewJsonParser())
	engine.RegisterParser(plugin.NewPythonParser())
	engine.RegisterParser(plugin.NewShellParser())
	engine.RegisterParser(plugin.NewConfigParser())
	engine.RegisterParser(plugin.NewPlainTextParser())
	engine.RegisterParser(plugin.NewXmlParser())
	engine.RegisterParser(plugin.NewDockerfileParser())
	engine.RegisterParser(plugin.NewJavascriptParser())
	engine.RegisterParser(plugin.NewPhpParser())
	engine.RegisterParser(plugin.NewGoParser())
	engine.RegisterParser(plugin.NewJavaParser())
	engine.RegisterParser(plugin.NewHtmlParser())
	engine.RegisterParser(plugin.NewJpropertiesParser())
	engine.RegisterParser(plugin.NewNpmrcParser())
	engine.RegisterParser(plugin.NewPipParser())
	engine.RegisterParser(plugin.NewPypircParser())
	engine.RegisterParser(plugin.NewDockercfgParser())
	engine.RegisterParser(plugin.NewHtpasswdParser())
	engine.SetMatcher(matcher)

	// 6. 运行引擎并收集结果
	secrets := engine.Run(*scanPath)

	// 7. 生成报告
	if len(secrets) > 0 && *format != "console" {
		var r report.Reporter
		outPath := *output

		switch *format {
		case "json":
			r = &report.JsonReporter{}
			if !strings.HasSuffix(outPath, ".json") {
				outPath += ".json"
			}
		case "csv":
			r = &report.CsvReporter{}
			if !strings.HasSuffix(outPath, ".csv") {
				outPath += ".csv"
			}
		default:
			l.Warn("未知的导出格式 %s，将跳过生成报告文件", *format)
		}

		if r != nil {
			if err := r.Generate(secrets, outPath); err != nil {
				l.Error("生成报告失败: %v", err)
			} else {
				l.Info("已成功将扫描报告导出至: %s", outPath)
			}
		}
	} else if len(secrets) == 0 {
		l.Info("棒极了！本次扫描未发现任何高危敏感信息。")
	}
}
