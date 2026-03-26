package plugin

// Parser 是所有文件解析器必须实现的接口。
// 不同的解析器（YAML, Python, JSON 等）通过实现该接口来提取文件中的潜在秘密。
type Parser interface {
	// Name 返回解析器的名称。
	Name() string
	
	// Parse 解析文件内容并返回发现的潜在密钥。
	Parse(content []byte) ([]string, error)
}
