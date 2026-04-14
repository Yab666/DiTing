package plugin

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// ConfigParser 实现了针对简单配置文件的解析逻辑 (.conf, .cfg, .ini, .credentials)。
// 对齐原版 whispers/plugins/config.py 的实现。
type ConfigParser struct{}

func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

// SupportedExtensions 返回支持的后缀。
func (p *ConfigParser) SupportedExtensions() []string {
	return []string{".conf", ".cfg", ".config", ".ini", ".credentials", ".s3cfg"}
}

// Parse 执行简单的 Key=Value 提取。
func (p *ConfigParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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

		// 忽略空行和注释 (原版 config.py 虽然没显式跳过 #，但其 split 逻辑隐含要求存在 =)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 对齐原版 logic: if "=" not in line: continue
		if !strings.Contains(line, "=") {
			continue
		}

		// 对齐原版: key, value = line.split("=", 1)
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if value != "" {
			results = append(results, KeyValue{
				Key:   key,
				Value: value,
				Path:  "config",
				Line:  lineNum,
			})
		}
	}

	return results, nil
}
