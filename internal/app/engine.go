package app

import (
	"context"
	"ditting/internal/core"
	"ditting/internal/plugin"
	"ditting/internal/rule"
	"ditting/internal/scanner"
	"ditting/pkg/logger"
	"os"
	"path/filepath"
	"strings"
)

type Engine struct {
	Config  *core.AppConfig
	Scanner scanner.FileScanner
	Logger  logger.Logger
	Verbose bool // 是否打印详细扫描过程

	Parsers map[string]plugin.Parser // 后缀名 -> 解析器
	Matcher *rule.Matcher           // 规则匹配引擎

	// 观察者钩子
	OnProgress func(path string)       // 正在分析某个文件
	OnFound    func(line core.Secret) // 发现了一个泄露点
}

// NewEngine 创建一个基于接口的引擎实例。
func NewEngine(config *core.AppConfig, s scanner.FileScanner, l logger.Logger, verbose bool) *Engine {
	return &Engine{
		Config:  config,
		Scanner: s,
		Logger:  l,
		Verbose: verbose,
		Parsers: make(map[string]plugin.Parser),
	}
}
// RegisterParser 注册一个文件解析插件。
func (e *Engine) RegisterParser(p plugin.Parser) {
	for _, ext := range p.SupportedExtensions() {
		e.Parsers[ext] = p
	}
}

// SetMatcher 设置匹配引擎。
func (e *Engine) SetMatcher(m *rule.Matcher) {
	e.Matcher = m
}

// Run 启动整个扫描和分析流程，并返回找到的所有秘密对象。
func (e *Engine) Run(root string) []core.Secret {
	e.Logger.Info("开始扫描目录: %s", root)

	var allSecrets []core.Secret

	filePaths := make(chan string, 100)
	go func() {
		err := e.Scanner.Scan(root, filePaths)
		if err != nil {
			e.Logger.Error("扫描终止: %v", err)
		}
	}()

	count := 0
	for path := range filePaths {
		count++

		// 触发进度回调
		if e.OnProgress != nil {
			e.OnProgress(path)
		}

		// A. 新增对齐：注入“文件名”虚拟键值对，专门用于命中的 sensitive-files 规则
		if e.Matcher != nil {
			fileName := filepath.Base(path)
			fileKV := plugin.KeyValue{
				Key:   "file",
				Value: fileName,
				Path:  "metadata",
				Line:  0,
			}
			if rule := e.Matcher.Match(fileKV); rule != nil {
				secret := core.Secret{
					RuleID:      rule.ID,
					Description: rule.Description,
					FilePath:    path,
					LineNumber:  0,
					Content:     fileName,
					Severity:    rule.Severity,
				}
				allSecrets = append(allSecrets, secret)
				// 触发发现回调
				if e.OnFound != nil {
					e.OnFound(secret)
				}
			}
		}

		// 1. 获取对应的解析器 (智能调度逻辑)
		p := e.getParserForFile(path)
		if p == nil {
			continue
		}

		if e.Verbose {
			e.Logger.Info("正在分析: %s (解析器: %T)", path, p)
		}

		// 2. 解析文件内容
		kvs, err := p.Parse(context.Background(), path)
		if err != nil {
			e.Logger.Warn("解析失败 [%s]: %v", path, err)
			continue
		}

		// 3. 匹配规则
		if e.Matcher != nil {
			for _, kv := range kvs {
				if e.Verbose {
					e.Logger.Info("  [检查 KV] Key=%s, Value=%s", kv.Key, kv.Value)
				}
				if rule := e.Matcher.Match(kv); rule != nil {
					// 持续在控制台输出警告（为了 CI/CD 实时观察）
					e.Logger.Warn("发现潜在泄露! [文件: %s] [行号: %d] [描述: %s] [匹配值: %s]",
						path, kv.Line, rule.Description, kv.Value)
					
					// 收集进结果池用于最终导出报表
					secret := core.Secret{
						RuleID:      rule.ID,
						Description: rule.Description,
						FilePath:    path,
						LineNumber:  kv.Line,
						Content:     kv.Value,
						Severity:    rule.Severity,
					}
					allSecrets = append(allSecrets, secret)
					// 触发发现回调
					if e.OnFound != nil {
						e.OnFound(secret)
					}
				}
			}
		}
	}

	e.Logger.Info("扫描完成。共分析文件: %d，发现 %d 处泄露。", count, len(allSecrets))
	return allSecrets
}

// getParserForFile 实现了与原版 Whispers 类似的智能插件选择逻辑。
func (e *Engine) getParserForFile(path string) plugin.Parser {
	info, err := os.Stat(path)
	if err != nil || info.Size() < 7 {
		return nil
	}

	name := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	// 处理 .dist 或 .template 后缀
	if ext == ".dist" || ext == ".template" {
		actualName := strings.TrimSuffix(name, ext)
		ext = strings.ToLower(filepath.Ext(actualName))
		if ext == "" {
			ext = actualName // 处理没有后缀的情况
		}
	}

	// 1. 根据文件名和扩展名初步匹配
	switch {
	case ext == ".yaml" || ext == ".yml":
		return e.Parsers[".yaml"]
	case ext == ".json":
		return e.Parsers[".json"]
	case ext == ".xml":
		return e.Parsers[".xml"]
	case strings.HasPrefix(name, ".npmrc"):
		return e.Parsers[".npmrc"]
	case strings.HasPrefix(name, ".pypirc"):
		return e.Parsers[".pypirc"]
	case name == "pip.conf" || name == "pip.ini":
		return e.Parsers["pip.conf"]
	case ext == ".properties":
		return e.Parsers[".properties"]
	case strings.HasSuffix(ext, "sh") || ext == ".env":
		return e.Parsers[".sh"]
	case strings.HasPrefix(name, "Dockerfile"):
		return e.Parsers["Dockerfile"]
	case ext == ".dockercfg" || name == "config.json":
		return e.Parsers[".dockercfg"]
	case strings.HasPrefix(name, ".htpasswd"):
		return e.Parsers[".htpasswd"]
	case ext == ".txt":
		return e.Parsers[".txt"]
	case strings.HasPrefix(ext, ".htm"):
		return e.Parsers[".html"]
	case ext == ".py" || strings.HasPrefix(ext, ".py"):
		return e.Parsers[".py"]
	case ext == ".js" || ext == ".mjs":
		return e.Parsers[".js"]
	case ext == ".java":
		return e.Parsers[".java"]
	case ext == ".go":
		return e.Parsers[".go"]
	case strings.HasPrefix(ext, ".php"):
		return e.Parsers[".php"]
	case ext == ".conf" || ext == ".cfg" || ext == ".config" || ext == ".ini" || ext == ".credentials" || ext == ".s3cfg":
		// 对齐原版：检查是否是 XML 格式的 Config
		if isXMLFile(path) {
			return e.Parsers[".xml"]
		}
		return e.Parsers[".conf"]
	}

	return nil
}

// isXMLFile 检查文件头是否包含 XML 声明
func isXMLFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	
	buf := make([]byte, 64)
	n, _ := f.Read(buf)
	content := string(buf[:n])
	return strings.HasPrefix(content, "<?xml ")
}
