package core

// Secret 代表扫描发现的敏感信息结果。
type Secret struct {
	RuleID      string
	Description string
	FilePath    string
	LineNumber  int
	Content     string
}

// Rule 定义了一个扫描规则的属性。
type Rule struct {
	ID          string
	Description string
	Pattern     string // 正则表达式或特定的解析逻辑
}
