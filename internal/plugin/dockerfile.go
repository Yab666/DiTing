package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
)

// DockerfileParser 实现了针对 Dockerfile 指令的扫描逻辑。
// 对齐原版 whispers/plugins/dockerfile.py 并增强了对 ARG 的支持。
type DockerfileParser struct {
	// 识别 ENV KEY=VALUE 或 ENV KEY VALUE
	reEnv *regexp.Regexp
	// 识别 ARG KEY=VALUE 或 ARG KEY VALUE
	reArg *regexp.Regexp
}

func NewDockerfileParser() *DockerfileParser {
	return &DockerfileParser{
		reEnv: regexp.MustCompile(`(?i)^ENV\s+([a-zA-Z_][a-zA-Z0-9_]*)(?:\s+|=)(.+)`),
		reArg: regexp.MustCompile(`(?i)^ARG\s+([a-zA-Z_][a-zA-Z0-9_]*)(?:\s+|=)(.+)`),
	}
}

// SupportedExtensions 返回支持的后缀。
// 注意：Dockerfile 通常没有后缀，此处主要匹配文件名。
func (p *DockerfileParser) SupportedExtensions() []string {
	return []string{"Dockerfile", ".dockerfile"}
}

// Parse 执行 Dockerfile 解析。
func (p *DockerfileParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
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

		// 忽略空行和纯注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 1. 匹配 ENV 指令
		if matches := p.reEnv.FindStringSubmatch(line); len(matches) > 0 {
			results = append(results, KeyValue{
				Key:   matches[1],
				Value: strings.Trim(matches[2], `"'`), // 去掉可能的引号
				Path:  "env",
				Line:  lineNum,
			})
		}

		// 2. 匹配 ARG 指令
		if matches := p.reArg.FindStringSubmatch(line); len(matches) > 0 {
			results = append(results, KeyValue{
				Key:   matches[1],
				Value: strings.Trim(matches[2], `"'`), // 去掉可能的引号
				Path:  "arg",
				Line:  lineNum,
			})
		}
	}

	return results, nil
}
