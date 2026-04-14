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

// MatchConfig 定义了针对 Key 或 Value 的具体匹配配置。
type MatchConfig struct {
	Regex      string // 正则表达式
	IgnoreCase bool   // 是否忽略大小写
	MinLen     int    // 最小长度
	IsBase64   bool   // 是否进行 Base64 检查
	IsUri      bool   // 是否进行 URI 检查
	IsLuhn     bool   // 是否进行 Luhn 校验 (如信用卡号)
	IsAscii    bool   // 是否要求必须是 Ascii
}

// Rule 定义了一个扫描规则的完整匹配逻辑。
type Rule struct {
	ID          string       // 规则唯一标识
	Description string       // 规则的人类可读描述
	Message     string       // 命中后展示的消息
	Severity    string       // 严重程度 (BLOCKER, CRITICAL, MAJOR, MINOR, INFO)
	Similar     float64      // 相似度阈值 (如果 Key 和 Value 太像，则可能是占位符，默认 0.3)
	Key         *MatchConfig // 针对键名的匹配配置
	Value       *MatchConfig // 针对键值的匹配配置
}

// AppConfig 包含地听扫描引擎的全局配置，可由 .ditingrc 或 CLI 覆盖。
type AppConfig struct {
	Include IncludeConfig `yaml:"include"`
	Exclude ExcludeConfig `yaml:"exclude"`
	Rules   string        `yaml:"rules"` // 外部规则目录路径
}

type IncludeConfig struct {
	Files []string `yaml:"files"` // 如 ["**/*.go", "**/*.json"]
}

type ExcludeConfig struct {
	Files  []string `yaml:"files"`  // 如 [".git", "node_modules", "vendor"]
	Keys   []string `yaml:"keys"`   // 要排除的键名正则表达式
	Values []string `yaml:"values"` // 要排除的键值正则表达式
	Paths  []string `yaml:"paths"`  // 要排除的具体 JSON/YAML 路径 (面包屑)
}
