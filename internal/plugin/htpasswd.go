package plugin

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// HtpasswdParser 提取 .htpasswd 中的哈希值。
type HtpasswdParser struct{}

func NewHtpasswdParser() *HtpasswdParser {
	return &HtpasswdParser{}
}

func (p *HtpasswdParser) SupportedExtensions() []string {
	return []string{".htpasswd"}
}

func (p *HtpasswdParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				results = append(results, KeyValue{
					Key:   "htpasswd Hash",
					Value: strings.TrimSpace(parts[1]),
					Path:  "htpasswd",
					Line:  lineNum,
				})
			}
		}
	}
	return results, nil
}
