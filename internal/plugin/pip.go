package plugin

import (
	"bufio"
	"context"
	"net/url"
	"os"
	"strings"
)

// PipParser 提取 pip 配置中的敏感信息。
// 它既能识别 Key=Value 形式的密码变量，也能识别 URL 链接中的凭据。
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
		line := strings.TrimSpace(scanner.Text())
		
		// 忽略空行和注释
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 1. 基础 Key-Value 提取 (通用策略)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			results = append(results, KeyValue{
				Key:   key,
				Value: val,
				Path:  "pip",
				Line:  lineNum,
			})

			// 2. 深度 URL 凭据嗅探 (特异性策略)
			if strings.Contains(val, "://") && strings.Contains(val, "@") {
				u, err := url.Parse(val)
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
		}
	}
	return results, nil
}
