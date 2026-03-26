package core

// Secret 代表扫描发现的敏感信息。
type Secret struct {
	RuleID      string // 匹配到的规则 ID
	Description string // 规则描述
	FilePath    string // 文件路径
	LineNumber  int    // 行号
	Content     string // 匹配到的内容摘要
	Severity    string // 严重程度 (如: BLOCKER, CRITICAL, INFO)
}

// Rule 定义了一个扫描规则的匹配逻辑。
type Rule struct {
	ID          string // 规则唯一标识
	Description string // 规则的人类可读描述
	Severity    string // 默认严重程度
	Regex       string // 正则表达式模式
	MinLength   int    // 最小匹配长度
}

// ScanConfig 包含扫描任务的全局配置。
type ScanConfig struct {
	IncludeFiles []string // 包含的文件 Glob 模式
	ExcludeFiles []string // 排除的文件 Glob 模式
	RulesPath    string   // 外部规则文件路径
}
