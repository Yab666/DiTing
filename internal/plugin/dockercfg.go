package plugin

import (
	"context"
	"encoding/json"
	"os"
)

// Dockercfg 结构用于解析 Docker 认证配置。
type Dockercfg struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}

// DockercfgParser 提取 .dockercfg 或 config.json 中的认证 Token。
type DockercfgParser struct{}

func NewDockercfgParser() *DockercfgParser {
	return &DockercfgParser{}
}

func (p *DockercfgParser) SupportedExtensions() []string {
	return []string{".dockercfg", "config.json"}
}

func (p *DockercfgParser) Parse(ctx context.Context, filePath string) ([]KeyValue, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Dockercfg
	if err := json.Unmarshal(content, &config); err != nil {
		// 如果不是合法的 Docker JSON，直接忽略（可能是其他项目的 config.json）
		return nil, nil
	}

	var results []KeyValue
	for _, auth := range config.Auths {
		if auth.Auth != "" {
			results = append(results, KeyValue{
				Key:   "Dockercfg",
				Value: auth.Auth,
				Path:  "dockercfg",
				Line:  0,
			})
		}
	}
	return results, nil
}
