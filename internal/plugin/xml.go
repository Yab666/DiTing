package plugin

import (
	"context"
	"encoding/xml"
	"io"
	"os"
	"regexp"
	"strings"
)

// XmlParser 实现了针对 XML 文件的深度扫描逻辑。
// 对齐原版 whispers/plugins/xml.py 的功能。
type XmlParser struct {
	uriParser *UriParser
	reUri     *regexp.Regexp
}

func NewXmlParser() *XmlParser {
	return &XmlParser{
		uriParser: NewUriParser(),
		reUri:     regexp.MustCompile(`(?i)(http|ftp|smtp|scp|ssh|jdbc[:\w\d]*|s3)s?://?.+`),
	}
}

// SupportedExtensions 返回支持的后缀。
func (p *XmlParser) SupportedExtensions() []string {
	return []string{".xml"}
}

// lineTrackingReader 包装 io.Reader 以跟踪当前行号
type lineTrackingReader struct {
	r      io.Reader
	line   int
	offset int
}

func (lr *lineTrackingReader) Read(p []byte) (n int, err error) {
	n, err = lr.r.Read(p)
	for i := 0; i < n; i++ {
		if p[i] == '\n' {
			lr.line++
		}
	}
	lr.offset += n
	return n, err
}

// Parse 执行 XML 解析。
func (p *XmlParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []KeyValue
	lr := &lineTrackingReader{r: file, line: 1}
	decoder := xml.NewDecoder(lr)

	// 用于跟踪节点路径
	var breadcrumbs []string
	// 用于处理 <key>... <value>... 兄弟节点配对
	var lastKey string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		currentLine := lr.line

		switch t := token.(type) {
		case xml.StartElement:
			tag := t.Name.Local
			breadcrumbs = append(breadcrumbs, tag)
			path := strings.Join(breadcrumbs, ".")

			// 1. 扫描属性: <elem key="value">
			for _, attr := range t.Attr {
				attrKey := attr.Name.Local
				attrVal := attr.Value

				results = append(results, KeyValue{
					Key:   attrKey,
					Value: attrVal,
					Path:  path + "@" + attrKey,
					Line:  currentLine,
				})

				// 2. 扫描属性中的 URI
				if p.reUri.MatchString(attrVal) {
					uriKVs := p.uriParser.ParseURI(attrVal)
					for _, kv := range uriKVs {
						kv.Line = currentLine
						kv.Path = path + "@" + attrKey + ".uri"
						results = append(results, kv)
					}
				}
			}

		case xml.EndElement:
			if len(breadcrumbs) > 0 {
				breadcrumbs = breadcrumbs[:len(breadcrumbs)-1]
			}

		case xml.CharData:
			if len(breadcrumbs) == 0 {
				continue
			}
			text := strings.TrimSpace(string(t))
			if text == "" {
				continue
			}

			tag := breadcrumbs[len(breadcrumbs)-1]
			path := strings.Join(breadcrumbs, ".")

			// 3. 记录标签本身的值: <key>value</key>
			results = append(results, KeyValue{
				Key:   tag,
				Value: text,
				Path:  path,
				Line:  currentLine,
			})

			// 4. 处理嵌入式 KV: <elem>key=value</elem>
			if strings.Contains(text, "=") {
				parts := strings.SplitN(text, "=", 2)
				if len(parts) == 2 {
					results = append(results, KeyValue{
						Key:   strings.TrimSpace(parts[0]),
						Value: strings.TrimSpace(parts[1]),
						Path:  path + ".embedded",
						Line:  currentLine,
					})
				}
			}

			// 5. 处理兄弟节点配对策略: <key>name</key><value>string</value>
			lowerTag := strings.ToLower(tag)
			if lowerTag == "key" {
				lastKey = text
			} else if lowerTag == "value" && lastKey != "" {
				results = append(results, KeyValue{
					Key:   lastKey,
					Value: text,
					Path:  path + ".pair",
					Line:  currentLine,
				})
				lastKey = "" // 使用后重置
			}
		}
	}

	return results, nil
}
