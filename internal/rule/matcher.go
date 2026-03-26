package rule

import (
	"ditting/internal/core"
	"ditting/internal/plugin"
	"regexp"
)

// Matcher 负责将提取出的键值对与规则进行匹配。
type Matcher struct {
	Rules []*core.Rule
}

// NewMatcher 创建一个新的匹配器。
func NewMatcher(rules []*core.Rule) *Matcher {
	return &Matcher{Rules: rules}
}

// Match 检查一个键值对是否命中了任何规则。
// 如果匹配成功，返回对应的 Rule，否则返回 nil。
func (m *Matcher) Match(kv plugin.KeyValue) *core.Rule {
	for _, rule := range m.Rules {
		// 1. 先尝试匹配键名 (Key)
		// 原版 Whispers 很多规则是基于键名的（比如看到 key 叫 password 就报警）
		if m.matchString(rule.Regex, kv.Key) {
			return rule
		}

		// 2. 再尝试匹配键值 (Value)
		// 比如检测 Base64 格式、硬编码的 API Key 格式等
		if m.matchString(rule.Regex, kv.Value) {
			return rule
		}
	}
	return nil
}

// matchString 执行正则匹配逻辑。
func (m *Matcher) matchString(pattern, text string) bool {
	if pattern == "" || text == "" {
		return false
	}
	
	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return false
	}
	return matched
}
