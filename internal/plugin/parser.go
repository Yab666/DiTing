package plugin

import "context"

// KeyValue 代表从文件中提取出的一个键值对。
// 无论文件是 YAML, JSON 还是 XML，最终都会被拆解为这种通用的结构。
type KeyValue struct {
	Key    string // 键名 (比如: password)
	Value  string // 键值 (比如: 123456)
	Path   string // 在文件内部的路径 (比如: database.auth.password)
	Line   int    // 所在行号
}

// Parser 是所有文件解析插件必须实现的接口。
type Parser interface {
	// Parse 解析文件内容并返回所有的键值对。
	// 使用 context 支持超时的安全退出。
	Parse(ctx context.Context, filePath string) ([]KeyValue, error)
	
	// SupportedExtensions 返回该插件支持的文件后缀名 (如: []string{".yaml", ".yml"})。
	SupportedExtensions() []string
}
