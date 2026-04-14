package rule

import (
	"ditting/internal/core"
	"ditting/internal/plugin"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2"
)

// Matcher 负责将提取出的键值对与规则进行深度匹配。
type Matcher struct {
	Rules  []*core.Rule
	Config *core.AppConfig
	// 预编译用于 isUri 检查的正则
	uriRegex *regexp2.Regexp
	// 预编译用于全局免扫描的正则 (路径/键名/键值)
	excludePathRegexes  []*regexp2.Regexp
	excludeKeyRegexes   []*regexp2.Regexp
	excludeValueRegexes []*regexp2.Regexp
}

// compileRegexList 辅助方法，用于编译字符数组到正则数组
func compileRegexList(patterns []string) []*regexp2.Regexp {
	var regexes []*regexp2.Regexp
	for _, pattern := range patterns {
		if re, err := regexp2.Compile(pattern, regexp2.IgnoreCase); err == nil {
			regexes = append(regexes, re)
		}
	}
	return regexes
}

// NewMatcher 创建一个新的匹配器。
func NewMatcher(rules []*core.Rule, config *core.AppConfig) *Matcher {
	m := &Matcher{
		Rules:    rules,
		Config:   config,
		// 使用与原版一致的 URI 识别正则
		uriRegex: regexp2.MustCompile(`(?i)(http|ftp|smtp|scp|ssh|jdbc[:\w\d]*|s3)s?://?.+`, 0),
	}

	// 预编译配置中的排查规则（支持正则）
	if config != nil {
		m.excludePathRegexes = compileRegexList(config.Exclude.Paths)
		m.excludeKeyRegexes = compileRegexList(config.Exclude.Keys)
		m.excludeValueRegexes = compileRegexList(config.Exclude.Values)
	}

	return m
}

// matchAny 辅助方法，检查文本是否匹配正则数组中的任何一项
func matchAny(text string, regexes []*regexp2.Regexp) bool {
	if text == "" || len(regexes) == 0 {
		return false
	}
	for _, re := range regexes {
		if match, _ := re.MatchString(text); match {
			return true
		}
	}
	return false
}

// IsExcluded 测试当前 KV 是否被用户全局配置明确排除。
func (m *Matcher) IsExcluded(kv plugin.KeyValue) bool {
	if matchAny(kv.Path, m.excludePathRegexes) {
		return true
	}
	if matchAny(kv.Key, m.excludeKeyRegexes) {
		return true
	}
	if matchAny(kv.Value, m.excludeValueRegexes) {
		return true
	}
	return false
}

// Match 检查一个键值对是否命中了任何规则。
func (m *Matcher) Match(kv plugin.KeyValue) *core.Rule {
	// 1. 全局配置排除检查 (对齐 Whispers secrets.py -> is_excluded)
	if m.IsExcluded(kv) {
		return nil
	}
	// 2. 如果是动态变量或占位符，直接跳过 (逻辑对齐 is_static)
	if !m.IsStatic(kv.Key, kv.Value) {
		return nil
	}

	for _, rule := range m.Rules {
		if m.CheckRule(rule, kv) {
			return rule
		}
	}
	return nil
}

// CheckRule 验证单个规则是否匹配给定的键值对。
func (m *Matcher) CheckRule(rule *core.Rule, kv plugin.KeyValue) bool {
	// 1. 相似度检查
	threshold := rule.Similar
	if threshold == 0 {
		threshold = 0.7
	}
	ratio := m.similarityRatio(kv.Key, kv.Value)
	if ratio >= threshold {
		fmt.Printf("[DEBUG] Rule %s similarity rejected: ratio %f >= threshold %f\n", rule.ID, ratio, threshold)
		return false
	}

	// 2. Key 匹配
	if rule.Key != nil {
		if !m.matchConfig(rule.Key, kv.Key) {
			fmt.Printf("[DEBUG] Rule %s key rejected: KV_Key=%s Regex=%s\n", rule.ID, kv.Key, rule.Key.Regex)
			return false
		}
	}

	// 3. Value 匹配
	if rule.Value != nil {
		if !m.matchConfig(rule.Value, kv.Value) {
			fmt.Printf("[DEBUG] Rule %s value rejected: KV_Value=%s\n", rule.ID, kv.Value)
			return false
		}
	}

	return true
}

// IsStatic 检查给定的值是否是静态硬编码的秘密（而非动态变量或占位符）。
// 对应原版 whispers/secrets.py 中的 is_static 逻辑。
func (m *Matcher) IsStatic(key, value string) bool {
	if value == "" || value == "null" {
		return false
	}

	// 过滤常见的动态变量格式
	if strings.HasPrefix(value, "$") && !strings.Contains(value[1:], "$") {
		return false // e.g. $VAR
	}
	if strings.Contains(value, "{{") && strings.Contains(value, "}}") {
		return false // e.g. {{.SECRET}}
	}
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		return false // e.g. ${SECRET}
	}
	if strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
		return false // e.g. <PLACEHOLDER>
	}

	// 过滤 IaC 引用 (CloudFormation 等)
	if strings.HasPrefix(value, "!") {
		return false // e.g. !Ref, !Sub
	}

	// 如果 Key 和 Value 完全一致，或者是 Value 以 Key 结尾（如 "my_password": "my_password"）
	// 通常是测试代码或占位符
	sKey := strings.ToLower(strings.TrimSpace(key))
	sValue := strings.ToLower(strings.TrimSpace(value))
	if sKey != "" && (sKey == sValue || strings.HasSuffix(sValue, sKey)) {
		return false
	}

	return true
}

// similarityRatio 计算两个字符串的简单相似度比例 (0.0 - 1.0)。
func (m *Matcher) similarityRatio(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1.0
	}

	matches := 0
	aChars := make(map[rune]int)
	for _, char := range a {
		aChars[char]++
	}

	for _, char := range b {
		if count, ok := aChars[char]; ok && count > 0 {
			matches++
			aChars[char]--
		}
	}

	return 2.0 * float64(matches) / float64(len(a)+len(b))
}

// matchConfig 根据配置验证字符串是否合规。
func (m *Matcher) matchConfig(cfg *core.MatchConfig, text string) bool {
	// A. 长度检查
	if cfg.MinLen > 0 && len(text) < cfg.MinLen {
		fmt.Printf("DEBUG: MinLen check failed: len(%s)=%d < %d\n", text, len(text), cfg.MinLen)
		return false
	}

	// B. 正则匹配
	if cfg.Regex != "" {
		var opts regexp2.RegexOptions = 0
		if cfg.IgnoreCase {
			opts = regexp2.IgnoreCase
		}
		
		re, err := regexp2.Compile(cfg.Regex, opts)
		if err != nil {
			return false
		}
		
		match, err := re.MatchString(text)
		if err != nil || !match {
			fmt.Printf("DEBUG: Regex check failed for %s with regex %s\n", text, cfg.Regex)
			return false
		}
	}

	// C. IsBase64 检查
	if cfg.IsBase64 {
		_, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return false
		}
	}

	// D. IsUri 检查
	isUriText, _ := m.uriRegex.MatchString(text)
	if cfg.IsUri && !isUriText {
		return false
	}

	// E. IsAscii 检查
	if cfg.IsAscii {
		for _, r := range text {
			if r > unicode.MaxASCII {
				fmt.Printf("DEBUG: Ascii check failed for %s\n", text)
				return false
			}
		}
	}

	// F. Luhn 检查
	if cfg.IsLuhn {
		if !m.luhnCheck(text) {
			return false
		}
	}

	return true
}

// luhnCheck 实现 Luhn 算法校验。
func (m *Matcher) luhnCheck(number string) bool {
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")
	if len(number) < 2 {
		return false
	}
	sum := 0
	shouldDouble := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if shouldDouble {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		shouldDouble = !shouldDouble
	}
	return sum%10 == 0
}
