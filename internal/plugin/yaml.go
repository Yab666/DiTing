package plugin

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// YamlParser 实现了针对 YAML 文件的深度解析逻辑。
// 它支持自动化预处理以兼容常见的模板语法，并能进行深度结构化遍历。
type YamlParser struct {
	// 预编译常用的正则表达式，用于识别模板标签和干扰项
	reJinja2  *regexp.Regexp
	reTags    *regexp.Regexp
	reComment *regexp.Regexp
	reUri     *regexp.Regexp // 用于识别字符串是否包含 URI
	uriParser *UriParser     // 嵌套的 URI 解析器
}

// NewYamlParser 创建并初始化解析器实例。
func NewYamlParser() *YamlParser {
	return &YamlParser{
		// 识别常见的模板占位符语法，如 {{ var }}
		reJinja2: regexp.MustCompile(`.*(\[)?\{\{.*\}\}(\])?.*`),
		// 识别嵌套的代码片段标签，如 <% %> 或 {% %}
		reTags: regexp.MustCompile(`(?s)[<{]%.*?%[}>]`),
		// 识别标准的全行注释内容
		reComment: regexp.MustCompile(`(?m)^#.*$`),
		// 识别常见的协议头链接 (同步原版配置)
		reUri:     regexp.MustCompile(`(?i)(http|ftp|smtp|scp|ssh|jdbc[:\w\d]*|s3)s?://?.+`),
		uriParser: NewUriParser(),
	}
}

// SupportedExtensions 返回该解析器支持的文件扩展名。
func (p *YamlParser) SupportedExtensions() []string {
	return []string{".yaml", ".yml"}
}

// preprocess 执行文件内容的预清洗工作。
// 主要通过正则处理以支持非标准的 YAML 语法（如未加引号的模板占位符）。
func (p *YamlParser) preprocess(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过 YAML 的文档起始符
		if strings.HasPrefix(line, "---") {
			continue
		}

		// 处理未被引号包裹的模板语法，通过自动补全引号来确保解析器兼容
		if p.reJinja2.MatchString(line) {
			line = strings.ReplaceAll(line, "\"", "'")
			line = strings.ReplaceAll(line, "{{", "\"{{")
			line = strings.ReplaceAll(line, "}}", "}}\"")
		}
		builder.WriteString(line + "\n")
	}

	content := builder.String()
	// 清除多行代码标签干扰项
	content = p.reTags.ReplaceAllString(content, "")
	// 清除注释内容，聚焦于真实数据
	content = p.reComment.ReplaceAllString(content, "")

	return content, nil
}

// Parse 执行文件的完整解析流程。
// 流程包含：文件读取 -> 逻辑清洗 -> 树状建模 -> 键值对提取。
func (p *YamlParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	// 1. 同步进行内容预处理
	cleanedContent, err := p.preprocess(filePath)
	if err != nil {
		return nil, err
	}

	// 2. 将内容反序列化为节点树，以保留精确的行号信息
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(cleanedContent), &node); err != nil {
		// 回退机制：对于极特殊情况，尝试直接读取原始文件
		data, _ := os.ReadFile(filePath)
		if err := yaml.Unmarshal(data, &node); err != nil {
			return nil, err
		}
	}

	var results []KeyValue
	if len(node.Content) > 0 {
		root := node.Content[0]
		// 3. 针对云平台配置格式（如 CloudFormation）进行专项提取
		p.parseSpecialFormats(root, &results)
		// 4. 执行深度递归遍历提取键值对
		p.traverse(root, "", &results)
	}

	return results, nil
}

// parseSpecialFormats 针对特定行业标准（如 IaC 模板）的参数段进行专项识别。
func (p *YamlParser) parseSpecialFormats(node *yaml.Node, results *[]KeyValue) {
	if node.Kind != yaml.MappingNode {
		return
	}

	// 识别是否为带有 Parameters 说明的模板文件
	var paramsNode *yaml.Node
	isTemplate := false
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		if key == "AWSTemplateFormatVersion" || key == "Resources" {
			isTemplate = true
		}
		if key == "Parameters" {
			paramsNode = node.Content[i+1]
		}
	}

	// 如果符合模板特征，则提取 Parameters 中的 Default 敏感值
	if isTemplate && paramsNode != nil && paramsNode.Kind == yaml.MappingNode {
		for i := 0; i < len(paramsNode.Content); i += 2 {
			paramKey := paramsNode.Content[i].Value
			paramValNode := paramsNode.Content[i+1]
			if paramValNode.Kind == yaml.MappingNode {
				for j := 0; j < len(paramValNode.Content); j += 2 {
					if paramValNode.Content[j].Value == "Default" {
						*results = append(*results, KeyValue{
							Key:   paramKey,
							Value: paramValNode.Content[j+1].Value,
							Path:  "Parameters." + paramKey + ".Default",
							Line:  paramValNode.Content[j+1].Line,
						})
					}
				}
			}
		}
	}
}

// traverse 核心递归逻辑：将 YAML 的树状结构扁平化为可供匹配的键值对。
func (p *YamlParser) traverse(node *yaml.Node, currentPath string, results *[]KeyValue) {
	switch node.Kind {
	case yaml.MappingNode:
		// 状态标记，用于识别特定语义下的键值组合（如 key: pass, value: 123）
		hasKey, hasValue := false, false
		var keyVal, valueVal string
		
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			key := keyNode.Value
			if key == "key" { hasKey = true; keyVal = valNode.Value }
			if key == "value" { hasValue = true; valueVal = valNode.Value }

			newPath := key
			if currentPath != "" {
				newPath = currentPath + "." + key
			}

			// 处理带标签的内容（例如云定义中的 !Ref 等标签）
			// 注意：排除 yaml.v3 的标准类型标签，比如 !!str, !!int 等
			valText := valNode.Value
			if valNode.Tag != "" && !strings.HasPrefix(valNode.Tag, "!!") {
				valText = valNode.Tag + " " + valText
			}

			*results = append(*results, KeyValue{
				Key:   key,
				Value: valText,
				Path:  newPath,
				Line:  valNode.Line,
			})

			p.traverse(valNode, newPath, results)
		}

		// 复合语义识别：如果单层 Mapping 中同时显式包含 key 和 value，则提取该组合
		if hasKey && hasValue {
			*results = append(*results, KeyValue{
				Key:   keyVal,
				Value: valueVal,
				Path:  currentPath + ".explicit_pair",
				Line:  node.Line,
			})
		}

	case yaml.SequenceNode:
		// 列表项目通常沿用父级上下文路径
		for _, itemNode := range node.Content {
			p.traverse(itemNode, currentPath, results)
		}

	case yaml.ScalarNode:
		// 1. 标量反解：解析字符串中嵌套的 key=value 模式
		if strings.Contains(node.Value, "=") {
			p.extractEmbeddedKV(node.Value, currentPath, node.Line, results)
		}

		// 2. URI 专项提取：解析连接字符串中的敏感信息 (同步原版逻辑)
		if p.reUri.MatchString(node.Value) {
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
func (p *YamlParser) extractEmbeddedKV(text, path string, line int, results *[]KeyValue) {
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
