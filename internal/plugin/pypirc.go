package plugin

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// PypircParser 提取 .pypirc 中的密码。
type PypircParser struct{}

func NewPypircParser() *PypircParser {
	return &PypircParser{}
}

func (p *PypircParser) SupportedExtensions() []string {
	return []string{".pypirc"}
}

func (p *PypircParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
		if strings.Contains(line, "password:") {
			parts := strings.Split(line, "password:")
			results = append(results, KeyValue{
				Key:   "PyPI password",
				Value: strings.TrimSpace(parts[len(parts)-1]),
				Path:  "pypirc",
				Line:  lineNum,
			})
		}
	}
	return results, nil
}
