package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
)

// PythonParser 使用语义正则表达式模拟 AST 提取逻辑，
// 从 Python 源码中识别变量赋值、字典键值、函数参数及环境变量调用。
type PythonParser struct {
	// 匹配变量赋值: KEY = "VALUE" 或 KEY = 'VALUE'
	reAssign *regexp.Regexp
	// 匹配字典项: "KEY": "VALUE"
	reDict *regexp.Regexp
	// 匹配函数关键字参数: func(KEY="VALUE")
	reKeyword *regexp.Regexp
	// 匹配环境变量调用: os.getenv("KEY", "VALUE")
	reEnv *regexp.Regexp
	// 匹配所有函数调用 (针对危险函数规则)
	reCall *regexp.Regexp
}

// NewPythonParser 创建并初始化 Python 语义解析器。
func NewPythonParser() *PythonParser {
	return &PythonParser{
		reAssign:  regexp.MustCompile(`(?m)^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*['"]([^'"]*)['"]`),
		reDict:    regexp.MustCompile(`['"]([^'"]+)['"]\s*:\s*['"]([^'"]*)['"]`),
		reKeyword: regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*['"]([^'"]*)['"]`),
		reEnv:     regexp.MustCompile(`(?:os\.getenv|environ\.get)\(\s*['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]*)['"])?\s*\)`),
		reCall:    regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9._]*)\s*\(`),
	}
}

// SupportedExtensions 返回支持的后缀。
func (p *PythonParser) SupportedExtensions() []string {
	return []string{".py"}
}

// Parse 执行 Python 源码分析。
func (p *PythonParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []KeyValue
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 忽略注释行
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// 1. 匹配变量赋值 (KEY = "VALUE")
		if matches := p.reAssign.FindStringSubmatch(line); len(matches) > 0 {
			results = append(results, KeyValue{
				Key:   matches[1],
				Value: matches[2],
				Path:  "assignment",
				Line:  lineNum,
			})
		}

		// 2. 匹配字典项 ("KEY": "VALUE")
		if matches := p.reDict.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, m := range matches {
				results = append(results, KeyValue{
					Key:   m[1],
					Value: m[2],
					Path:  "dict",
					Line:  lineNum,
				})
			}
		}

		// 3. 匹配函数环境变量调用 os.getenv("KEY", "DEFAULT")
		if matches := p.reEnv.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, m := range matches {
				// 获取 Key
				results = append(results, KeyValue{
					Key:   m[1],
					Value: "", // 仅作为键名
					Path:  "env",
					Line:  lineNum,
				})
				// 如果有默认值，也作为 Value 进行检查 (对齐原版 logic)
				if m[2] != "" {
					results = append(results, KeyValue{
						Key:   m[1],
						Value: m[2],
						Path:  "env_default",
						Line:  lineNum,
					})
				}
			}
		}

		// 4. 匹配函数调用 (针对 dangerous-functions: key=function, value=调用名)
		// 对齐原版：将所有函数名提取出来供危险函数规则匹配
		if matches := p.reCall.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, m := range matches {
				results = append(results, KeyValue{
					Key:   "function",
					Value: m[1] + "()", // 拼接括号以适配规则引擎中的 regex: ^(eval|exec)\(
					Path:  "call",
					Line:  lineNum,
				})
			}
		}
		
		// 5. 标量特征提取：如果行中包含 = 且不是赋值语句，可能是关键字参数
		if strings.Contains(line, "=") && !p.reAssign.MatchString(line) {
			if matches := p.reKeyword.FindAllStringSubmatch(line, -1); len(matches) > 0 {
				for _, m := range matches {
					results = append(results, KeyValue{
						Key:   m[1],
						Value: m[2],
						Path:  "keyword",
						Line:  lineNum,
					})
				}
			}
		}
	}

	return results, nil
}
