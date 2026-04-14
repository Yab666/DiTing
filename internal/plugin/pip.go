package plugin

import (
	"bufio"
	"context"
	"net/url"
	"os"
	"strings"
)

// PipParser 提取 pip 配置中 URL 包含的密码。
type PipParser struct{}

func NewPipParser() *PipParser {
	return &PipParser{}
}

func (p *PipParser) SupportedExtensions() []string {
	return []string{"pip.conf", "pip.ini"}
}

func (p *PipParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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
		if !strings.Contains(line, "http") {
			continue
		}

		parts := strings.Split(line, "=")
		uStr := strings.TrimSpace(parts[len(parts)-1])
		u, err := url.Parse(uStr)
		if err == nil && u.User != nil {
			pass, ok := u.User.Password()
			if ok {
				results = append(results, KeyValue{
					Key:   "pip password",
					Value: pass,
					Path:  "pip",
					Line:  lineNum,
				})
			}
		}
	}
	return results, nil
}
