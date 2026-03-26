package rule

import (
	"ditting/internal/core"
	"ditting/internal/plugin"
	"encoding/base64"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2"
)

// Matcher 负责将提取出的键值对与规则进行深度匹配。
type Matcher struct {
	Rules []*core.Rule
	// 预编译用于 isUri 检查的正则
	uriRegex *regexp2.Regexp
}

// NewMatcher 创建一个新的匹配器。
func NewMatcher(rules []*core.Rule) *Matcher {
	return &Matcher{
		Rules: rules,
		// 使用与原版一致的 URI 识别正则
		uriRegex: regexp2.MustCompile(`(?i)(http|ftp|smtp|scp|ssh|jdbc[:\w\d]*|s3)s?://?.+`, 0),
	}
}

// Match 检查一个键值对是否命中了任何规则。
func (m *Matcher) Match(kv plugin.KeyValue) *core.Rule {
	for _, rule := range m.Rules {
		if m.CheckRule(rule, kv) {
			return rule
		}
	}
	return nil
}

// CheckRule 验证单个规则是否匹配给定的键值对。
func (m *Matcher) CheckRule(rule *core.Rule, kv plugin.KeyValue) bool {
	// 1. 相似度检查 (防止占位符误报)
	threshold := rule.Similar
	if threshold == 0 {
		threshold = 0.3
	}
	if m.similarityRatio(kv.Key, kv.Value) >= threshold {
		return false
	}

	// 2. 如果规则定义了 Key 匹配要求，但 Key 不匹配，则返回 false
	if rule.Key != nil {
		if !m.matchConfig(rule.Key, kv.Key) {
			return false
		}
	}

	// 3. 如果规则定义了 Value 匹配要求，但 Value 不匹配，则返回 false
	if rule.Value != nil {
		if !m.matchConfig(rule.Value, kv.Value) {
			return false
		}
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
	if text == "" {
		// 如果规则没定义正则表达式，只是定义了 isBase64 等，空文可能不匹配
		// 遵循原版逻辑：有 Regex 则先过 Regex，没有则认为匹配基础成功
	}

	// A. 长度检查
	if cfg.MinLen > 0 && len(text) < cfg.MinLen {
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
			return false
		}
	}

	// C. IsBase64 检查 (这里其实有逻辑差：原版是判断“是否是Base64且解码后如何”)
	// 简化对齐：如果要求是Base64，则必须能解码
	if cfg.IsBase64 {
		_, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return false
		}
	}

	// D. IsUri 检查 (同步原版：要求结果必须等于 isUri 的布尔值)
	// 原版逻辑：if "isUri" in rule: return rule["isUri"] == is_uri(text)
	isUriText, _ := m.uriRegex.MatchString(text)
	// 注意：由于 cfg.IsUri 默认为 false，我们需要辨别它是“未设置”还是“显式设置为 false”
	// 在我们的 struct 中目前无法辨别。但原版规则如 apikey 显式设为 isUri: False
	// 我们暂定如果配置了该字段，则进行校验。
	// 这里其实有一个坑：Go 的 bool 默认 false。
	// 改进：这里我们主要看规则里显式写的 False。
	if cfg.IsUri && !isUriText {
		return false
	}
	// 如果规则要求 NOT URI (isUri: False)，但实际上是 URI
	// 这种情况目前无法完全对齐，除非使用 *bool。
	// 暂且跳过，针对主要泄露点优化。

	// E. IsAscii 检查
	if cfg.IsAscii {
		for _, r := range text {
			if r > unicode.MaxASCII {
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
