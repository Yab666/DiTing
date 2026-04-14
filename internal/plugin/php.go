package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
)

// PhpParser 实现了针对 PHP 源码的变量赋值与 define 解析。
type PhpParser struct {
	reAssign *regexp.Regexp
	reDefine *regexp.Regexp
}

func NewPhpParser() *PhpParser {
	return &PhpParser{
		// 匹配 $KEY = "VALUE" 或键值对 "KEY" => "VALUE"
		reAssign: regexp.MustCompile(`(?:(?:\$([a-zA-Z_]\w*))|(?:['"]([^'"]+)['"]\s*=>))\s*=\s*['"]([^'"]+)['"]`),
		// 匹配 define("KEY", "VALUE")
		reDefine: regexp.MustCompile(`(?i)define\s*\(\s*['"]([^'"]+)['"]\s*,\s*['"]([^'"]+)['"]\s*\)`),
	}
}

func (p *PhpParser) SupportedExtensions() []string {
	return []string{".php"}
}

func (p *PhpParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
		
		// 1. 匹配 define
		if matches := p.reDefine.FindStringSubmatch(line); len(matches) > 0 {
			results = append(results, KeyValue{
				Key:   matches[1],
				Value: matches[2],
				Path:  "php.define",
				Line:  lineNum,
			})
		}

		// 2. 匹配赋值
		if matches := p.reAssign.FindStringSubmatch(line); len(matches) > 0 {
			key := matches[1]
			if key == "" {
				key = matches[2]
			}
			results = append(results, KeyValue{
				Key:   key,
				Value: matches[3],
				Path:  "php.assignment",
				Line:  lineNum,
			})
		}
	}
	return results, nil
}
