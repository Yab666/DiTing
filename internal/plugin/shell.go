package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
)

// ShellParser 实现了针对 Shell 脚本及环境变量文件 (.sh, .env) 的解析逻辑。
// 对齐原版 whispers/plugins/shell.py 的实现。
type ShellParser struct {
	reKeyValue *regexp.Regexp
	reCurlUser *regexp.Regexp
}

func NewShellParser() *ShellParser {
	return &ShellParser{
		// 识别简单的 Key=Value 配对
		reKeyValue: regexp.MustCompile(`^([^=]+)=(.+)$`),
		// 针对 curl 命令提取用户凭据 (-u, --user)
		reCurlUser: regexp.MustCompile(`(?i)(?:-u|--user)\s+['"]?([^'"]+)['"]?`),
	}
}

// SupportedExtensions 返回支持的后缀。
func (p *ShellParser) SupportedExtensions() []string {
	return []string{".sh", ".bash", ".zsh", ".env", "bashrc", "zshrc"}
}

// Parse 执行 Shell/Env 解析。
func (p *ShellParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []KeyValue
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var buffer strings.Builder

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 忽略空行
		if trimmedLine == "" {
			continue
		}

		// 处理跨行指令 (对齐原版 line.endswith("\\"))
		if strings.HasSuffix(trimmedLine, "\\") {
			buffer.WriteString(strings.TrimSuffix(trimmedLine, "\\") + " ")
			continue
		}

		// 拼接完成或单行逻辑
		fullLine := buffer.String() + trimmedLine
		buffer.Reset()

		// 忽略完全是注释的行 (原版逻辑处理：if line.startswith("#"))
		if strings.HasPrefix(strings.TrimSpace(fullLine), "#") {
			continue
		}

		// 1. 尝试键值对提取 (对齐原版 item.split("="))
		if matches := p.reKeyValue.FindStringSubmatch(fullLine); len(matches) > 0 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			if value != "" {
				results = append(results, KeyValue{
					Key:   key,
					Value: value,
					Path:  "shell",
					Line:  lineNum,
				})
			}
		}

		// 2. 针对 curl 命令进行专项解析 (对齐原版 curl 方法)
		if strings.Contains(strings.ToLower(fullLine), "curl") {
			if matches := p.reCurlUser.FindStringSubmatch(fullLine); len(matches) > 0 {
				creds := matches[1]
				if strings.Contains(creds, ":") {
					parts := strings.SplitN(creds, ":", 2)
					results = append(results, KeyValue{
						Key:   "cURL_Password",
						Value: parts[1],
						Path:  "curl",
						Line:  lineNum,
					})
				}
			}
		}
	}

	return results, nil
}
