package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// JsonParser 实现了针对 JSON 文件的解析逻辑。
// 为了获得更好的行号支持，它利用 YAML 解析器（JSON 是 YAML 的子集）进行节点遍历。
type JsonParser struct {
	reCommentLine *regexp.Regexp
	reCommentEnd  *regexp.Regexp
	uriParser     *UriParser // 嵌套的 URI 解析器，用于对齐原版逻辑
}

// NewJsonParser 创建并初始化 JSON 解析器。
func NewJsonParser() *JsonParser {
	return &JsonParser{
		// 识别以 // 开头的单行注释
		reCommentLine: regexp.MustCompile(`^\s*//.*`),
		// 识别带空格间隔的行尾 // 注释 (避开 https:// 等链接)
		reCommentEnd: regexp.MustCompile(`\s+//.*$`),
		uriParser:    NewUriParser(),
	}
}

// SupportedExtensions 返回支持的后缀。
func (p *JsonParser) SupportedExtensions() []string {
	return []string{".json"}
}

// preprocess 执行 JSON 的清洗（去除注释以兼容非标 JSON）。
func (p *JsonParser) preprocess(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 1. 跳过 // 开头的行
		if p.reCommentLine.MatchString(line) {
			continue
		}

		// 2. 清除行尾的 // 注释 (对齐原版 re.sub(r" // ?.*$", "", line))
		line = p.reCommentEnd.ReplaceAllString(line, "")

		builder.WriteString(line + "\n")
	}

	return builder.String(), nil
}

// Parse 执行 JSON 解析流程。
func (p *JsonParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	// 1. 预处理清洗注释
	cleanedContent, err := p.preprocess(filePath)
	if err != nil {
		return nil, err
	}

	// 2. 利用 yaml.v3 解析以保留行号
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(cleanedContent), &node); err != nil {
		return nil, err
	}

	var results []KeyValue
	if len(node.Content) > 0 {
		// JSON 通常只有一个根 Node
		p.traverse(node.Content[0], "", &results)
	}

	return results, nil
}

// traverse 核心递归逻辑：与 YAML 解析器逻辑对齐。
func (p *JsonParser) traverse(node *yaml.Node, currentPath string, results *[]KeyValue) {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			key := keyNode.Value
			newPath := key
			if currentPath != "" {
				newPath = currentPath + "." + key
			}

			// 记录当前键值对
			*results = append(*results, KeyValue{
				Key:   key,
				Value: valNode.Value,
				Path:  newPath,
				Line:  valNode.Line,
			})

			// 递归遍历值
			p.traverse(valNode, newPath, results)
		}

	case yaml.SequenceNode:
		for _, itemNode := range node.Content {
			p.traverse(itemNode, currentPath, results)
		}

	case yaml.ScalarNode:
		// 标量反解：解析字符串中嵌套的 key=value 模式 (对齐原版)
		if strings.Contains(node.Value, "=") {
			p.extractEmbeddedKV(node.Value, currentPath, node.Line, results)
		}
		
		// URI 提取支持
		// 对齐原版：如果值匹配 URI 模式，进行拆解
		// 此处正则表达式与 YamlParser 保持一致
		reUri := regexp.MustCompile(`(?i)(http|ftp|smtp|scp|ssh|jdbc[:\w\d]*|s3)s?://?.+`)
		if reUri.MatchString(node.Value) {
			uriKVs := p.uriParser.ParseURI(node.Value)
			for _, kv := range uriKVs {
				kv.Path = currentPath + ".uri"
				kv.Line = node.Line
				*results = append(*results, kv)
			}
		}
	}
}

// extractEmbeddedKV 处理标量字符串内部的子键值对。
func (p *JsonParser) extractEmbeddedKV(text, path string, line int, results *[]KeyValue) {
	parts := strings.SplitN(text, "=", 2)
	if len(parts) == 2 {
		*results = append(*results, KeyValue{
			Key:   strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
			Path:  path,
			Line:  line,
		})
	}
}
