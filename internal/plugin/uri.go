package plugin

import (
	"context"
	"net/url"
)

// UriParser 专门负责解析 URI/URL 字符串并提取其中的敏感信息。
type UriParser struct{}

func NewUriParser() *UriParser {
	return &UriParser{}
}

// SupportedExtensions URI 解析器通常不直接处理文件后缀，而是处理内容。
func (p *UriParser) SupportedExtensions() []string {
	return []string{}
}

// ParseURI 将字符串解析为键值对（例如密码、查询参数）。
func (p *UriParser) ParseURI(text string) []KeyValue {
	var results []KeyValue

	u, err := url.Parse(text)
	if err != nil {
		return nil
	}

	// 1. 提取密码
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			results = append(results, KeyValue{
				Key:   "URI_Password",
				Value: password,
			})
		}
	}

	// 2. 提取查询参数
	query := u.Query()
	for k, vs := range query {
		for _, v := range vs {
			results = append(results, KeyValue{
				Key:   k,
				Value: v,
			})
		}
	}

	return results
}

// Parse 实现通用接口（此处主要调用 ParseURI）。
func (p *UriParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	return nil, nil // URI 插件目前主要供其他插件调用
}
