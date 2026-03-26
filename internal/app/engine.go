package app

import (
	"context"
	"ditting/internal/core"
	"ditting/internal/plugin"
	"ditting/internal/rule"
	"ditting/internal/scanner"
	"ditting/pkg/logger"
	"path/filepath"
)

// Engine 是扫描任务的核心控制器。
type Engine struct {
	Config  *core.ScanConfig
	Scanner scanner.FileScanner
	Logger  logger.Logger
	Verbose bool // 是否打印详细扫描过程

	Parsers map[string]plugin.Parser // 后缀名 -> 解析器
	Matcher *rule.Matcher           // 规则匹配引擎
}

// NewEngine 创建一个基于接口的引擎实例。
func NewEngine(config *core.ScanConfig, s scanner.FileScanner, l logger.Logger, verbose bool) *Engine {
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

// Run 启动整个扫描和分析流程。
func (e *Engine) Run(root string) {
	e.Logger.Info("开始扫描目录: %s", root)

	// 创建用于接收文件路径的管道
	filePaths := make(chan string, 100)

	// 在后台启动文件扫描器
	go func() {
		err := e.Scanner.Scan(root, filePaths)
		if err != nil {
			e.Logger.Error("扫描终止: %v", err)
		}
	}()

	// 主循环：处理扫描到的每一个文件
	count := 0
	for path := range filePaths {
		count++
		if e.Verbose {
			e.Logger.Info("正在分析: %s", path)
		}

		// 1. 获取对应的解析器
		ext := filepath.Ext(path)
		parser, ok := e.Parsers[ext]
		if !ok {
			continue // 暂不支持的文件类型，跳过
		}

		// 2. 解析文件内容
		kvs, err := parser.Parse(context.Background(), path)
		if err != nil {
			e.Logger.Warn("解析失败 [%s]: %v", path, err)
			continue
		}

		// 3. 匹配规则
		if e.Matcher != nil {
			for _, kv := range kvs {
				if rule := e.Matcher.Match(kv); rule != nil {
					e.Logger.Warn("发现潜在泄露! [文件: %s] [行号: %d] [描述: %s] [匹配值: %s]",
						path, kv.Line, rule.Description, kv.Value)
				}
			}
		}
	}

	e.Logger.Info("扫描完成。共分析文件: %d", count)
}
