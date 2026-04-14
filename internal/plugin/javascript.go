package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
)

// JavascriptParser 实现了针对 JS 源码的变量赋值解析。
type JavascriptParser struct {
	reAssign *regexp.Regexp
}

func NewJavascriptParser() *JavascriptParser {
	return &JavascriptParser{
		// 匹配 const/let/var KEY = "VALUE"
		reAssign: regexp.MustCompile(`(?m)(?:const|let|var)\s+([a-zA-Z_]\w*)\s*=\s*['"]([^'"]+)['"]`),
	}
}

func (p *JavascriptParser) SupportedExtensions() []string {
	return []string{".js", ".mjs", ".jsx", ".ts", ".tsx"}
}

func (p *JavascriptParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
		if matches := p.reAssign.FindStringSubmatch(line); len(matches) > 0 {
			results = append(results, KeyValue{
				Key:   matches[1],
				Value: matches[2],
				Path:  "javascript",
				Line:  lineNum,
			})
		}
	}
	return results, nil
}
