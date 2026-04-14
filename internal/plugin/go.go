package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
)

// GoParser 实现了针对 Go 源码的变量赋值解析。
type GoParser struct {
	reAssign *regexp.Regexp
}

func NewGoParser() *GoParser {
	return &GoParser{
		// 匹配 KEY := "VALUE" 或 KEY = "VALUE"
		reAssign: regexp.MustCompile(`([a-zA-Z_]\w*)\s*(?::=|=)\s*['"]([^'"]+)['"]`),
	}
}

func (p *GoParser) SupportedExtensions() []string {
	return []string{".go"}
}

func (p *GoParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
				Path:  "go",
				Line:  lineNum,
			})
		}
	}
	return results, nil
}
