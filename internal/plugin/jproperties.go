package plugin

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// JpropertiesParser 实现了针对 Java .properties 文件的解析。
type JpropertiesParser struct{}

func NewJpropertiesParser() *JpropertiesParser {
	return &JpropertiesParser{}
}

func (p *JpropertiesParser) SupportedExtensions() []string {
	return []string{".properties"}
}

func (p *JpropertiesParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			results = append(results, KeyValue{
				Key:   strings.ReplaceAll(strings.TrimSpace(parts[0]), ".", "_"),
				Value: strings.TrimSpace(parts[1]),
				Path:  "properties",
				Line:  lineNum,
			})
		} else if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			results = append(results, KeyValue{
				Key:   strings.ReplaceAll(strings.TrimSpace(parts[0]), ".", "_"),
				Value: strings.TrimSpace(parts[1]),
				Path:  "properties",
				Line:  lineNum,
			})
		}
	}
	return results, nil
}
