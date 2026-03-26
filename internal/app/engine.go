package app

import (
	"ditting/internal/core"
	"ditting/internal/scanner"
	"ditting/pkg/logger"
)

// Engine 是扫描任务的核心控制器。
type Engine struct {
	Config  *core.ScanConfig
	Scanner scanner.FileScanner
	Logger  logger.Logger
	Verbose bool // 是否打印详细扫描过程
}

// NewEngine 创建一个基于接口的引擎实例。
func NewEngine(config *core.ScanConfig, s scanner.FileScanner, l logger.Logger, verbose bool) *Engine {
	return &Engine{
		Config:  config,
		Scanner: s,
		Logger:  l,
		Verbose: verbose,
	}
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
		// TODO: 在这里调用插件系统进行内容解析和规则匹配
	}

	e.Logger.Info("扫描完成。共分析文件: %d", count)
}
