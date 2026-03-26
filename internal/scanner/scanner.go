package scanner

import (
	"ditting/pkg/logger"
	"os"
	"path/filepath"
	"strings"
)

// FileScanner 定义了扫描器必须实现的功能接口。
type FileScanner interface {
	Scan(root string, results chan<- string) error
}

// Scanner 负责遍历文件系统并发现待扫描的文件。
type Scanner struct {
	ExcludePatterns []string
	Logger          logger.Logger
}

// NewScanner 创建并返回一个新的扫描器实例。
func NewScanner(exclude []string, l logger.Logger) *Scanner {
	return &Scanner{
		ExcludePatterns: exclude,
		Logger:          l,
	}
}

// Scan 递归遍历指定路径，并将符合条件的文件路径发送到 results 管道中。
func (s *Scanner) Scan(root string, results chan<- string) error {
	defer close(results)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 如果是权限错误，记录警告并继续扫描其他文件
			if os.IsPermission(err) {
				s.Logger.Warn("权限不足，跳过路径: %s", path)
				return nil
			}
			s.Logger.Error("遍历出错 [%s]: %v", path, err)
			return err
		}

		// 检查路径是否应该被排除
		if s.shouldExclude(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理常规文件
		if !info.IsDir() && info.Mode().IsRegular() {
			results <- path
		}

		return nil
	})
}

// shouldExclude 检查给定路径是否匹配任何排除模式。
func (s *Scanner) shouldExclude(path string) bool {
	for _, pattern := range s.ExcludePatterns {
		// 简单的子字符串匹配，后续可以扩展为真正的 Glob 匹配
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}
