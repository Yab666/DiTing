package plugin

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// NpmrcParser 提取 .npmrc 中的 _authToken。
type NpmrcParser struct{}

func NewNpmrcParser() *NpmrcParser {
	return &NpmrcParser{}
}

func (p *NpmrcParser) SupportedExtensions() []string {
	return []string{".npmrc"}
}

func (p *NpmrcParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
		if strings.Contains(line, ":_authToken=") {
			parts := strings.Split(line, ":_authToken=")
			results = append(results, KeyValue{
				Key:   "npm authToken",
				Value: strings.TrimSpace(parts[len(parts)-1]),
				Path:  "npmrc",
				Line:  lineNum,
			})
		}
	}
	return results, nil
}
