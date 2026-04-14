package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
)

// JavaParser 实现了针对 Java 源码的变量赋值解析。
type JavaParser struct {
	reAssign *regexp.Regexp
}

func NewJavaParser() *JavaParser {
	return &JavaParser{
		// 匹配 String KEY = "VALUE";
		reAssign: regexp.MustCompile(`(?:[a-zA-Z_]\w*\s+)+([a-zA-Z_]\w*)\s*=\s*['"]([^'"]+)['"]`),
	}
}

func (p *JavaParser) SupportedExtensions() []string {
	return []string{".java", ".groovy", ".kt"}
}

func (p *JavaParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
				Path:  "java",
				Line:  lineNum,
			})
		}
	}
	return results, nil
}
