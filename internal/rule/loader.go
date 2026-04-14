package rule

import (
	"ditting/internal/core"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RuleLoader 负责从文件系统中加载 YAML 格式的扫描规则。
type RuleLoader struct{}

func NewRuleLoader() *RuleLoader {
	return &RuleLoader{}
}

// LoadFromDir 加载指定目录下所有的 .yaml 规则文件。
func (l *RuleLoader) LoadFromDir(dirPath string) ([]*core.Rule, error) {
	var allRules []*core.Rule

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取规则目录出错: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml") {
			rules, err := l.LoadFromFile(filepath.Join(dirPath, file.Name()))
			if err != nil {
				continue // 某个文件报错，打印日志后继续加载其他
			}
			allRules = append(allRules, rules...)
		}
	}

	return allRules, nil
}

// LoadFromFile 解析单个规则文件。
func (l *RuleLoader) LoadFromFile(filePath string) ([]*core.Rule, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 原始格式是 Map 结构： RuleID -> Rule内容
	var rawRules map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &rawRules); err != nil {
		return nil, err
	}

	var results []*core.Rule
	for id, content := range rawRules {
		rule := &core.Rule{ID: id}
		
		if val, ok := content["description"].(string); ok { rule.Description = val }
		if val, ok := content["message"].(string); ok { rule.Message = val }
		if val, ok := content["severity"].(string); ok { rule.Severity = val }
		if val, ok := content["similar"].(float64); ok { rule.Similar = val }

		// 解析 Key 匹配配置
		if keyMap, ok := content["key"].(map[string]interface{}); ok {
			rule.Key = l.parseMatchConfig(keyMap)
		}

		// 解析 Value 匹配配置
		if valMap, ok := content["value"].(map[string]interface{}); ok {
			rule.Value = l.parseMatchConfig(valMap)
		}

		results = append(results, rule)
	}

	return results, nil
}

func (l *RuleLoader) parseMatchConfig(m map[string]interface{}) *core.MatchConfig {
	cfg := &core.MatchConfig{}
	
	if val, ok := m["regex"].(string); ok { cfg.Regex = val }
	if val, ok := m["ignorecase"].(bool); ok { cfg.IgnoreCase = val }
	
	// YAML 数字可能被解析为 int 或 float64
	if val, ok := m["minlen"].(int); ok {
		cfg.MinLen = val
	} else if val, ok := m["minlen"].(float64); ok {
		cfg.MinLen = int(val)
	}

	if val, ok := m["isBase64"].(bool); ok { cfg.IsBase64 = val }
	if val, ok := m["isUri"].(bool); ok { cfg.IsUri = val }
	if val, ok := m["isLuhn"].(bool); ok { cfg.IsLuhn = val }
	if val, ok := m["isAscii"].(bool); ok { cfg.IsAscii = val }

	return cfg
}
