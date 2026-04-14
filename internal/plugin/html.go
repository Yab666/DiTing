package plugin

import (
	"context"
	"os"
	"regexp"
	"strings"
)

// HtmlParser 实现了针对 HTML 注释的提取逻辑。
type HtmlParser struct {
	reComment *regexp.Regexp
}

func NewHtmlParser() *HtmlParser {
	return &HtmlParser{
		// 识别 HTML 多行注释 <!-- ... -->
		reComment: regexp.MustCompile(`(?s)<!--(.*?)-->`),
	}
}

func (p *HtmlParser) SupportedExtensions() []string {
	return []string{".html", ".htm", ".xhtml"}
}

func (p *HtmlParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []KeyValue
	matches := p.reComment.FindAllStringSubmatch(string(content), -1)
	for _, m := range matches {
		text := strings.TrimSpace(m[1])
		if text != "" {
			results = append(results, KeyValue{
				Key:   "comment",
				Value: text,
				Path:  "html",
				Line:  0, // 全局匹配暂不精确计算行号
			})
		}
	}
	return results, nil
}
