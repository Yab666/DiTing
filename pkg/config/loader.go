package config

import (
	"ditting/internal/core"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig 负责从指定位置加载用户配置。
// 它会根据优先级查找配置：
// 1. 如果用户显式传入 configPath，就用这个。
// 2. 否则，如果扫描目标目录 rootPath 存在 `.ditingrc`，就用这个。
// 3. 将加载后的配置与默认安全配置（排除 .git, node_modules) 合并。
func LoadConfig(configPath string, rootPath string) (*core.AppConfig, error) {
	// 初始化硬编码防呆底线（防止小白没挂配置导致遍历全盘挂掉）
	cfg := &core.AppConfig{
		Include: core.IncludeConfig{
			Files: []string{"**/*"},
		},
		Exclude: core.ExcludeConfig{
			Files: []string{".git", "node_modules", "vendor", ".idea", ".vscode"},
		},
		Rules: "configs/rules",
	}

	actualPath := ""
	if configPath != "" {
		actualPath = configPath
	} else {
		// 探测扫描目录下有没有 .ditingrc
		rcPath := filepath.Join(rootPath, ".ditingrc")
		if _, err := os.Stat(rcPath); err == nil {
			actualPath = rcPath
		}
	}

	if actualPath != "" {
		data, err := os.ReadFile(actualPath)
		if err == nil {
			var userCfg core.AppConfig
			if err := yaml.Unmarshal(data, &userCfg); err == nil {
				// 简单的合并策略（用户配置追加或覆盖默认）
				if len(userCfg.Include.Files) > 0 {
					cfg.Include.Files = userCfg.Include.Files
				}
				if len(userCfg.Exclude.Files) > 0 {
					cfg.Exclude.Files = append(cfg.Exclude.Files, userCfg.Exclude.Files...)
				}
				cfg.Exclude.Keys = userCfg.Exclude.Keys
				cfg.Exclude.Values = userCfg.Exclude.Values
				cfg.Exclude.Paths = userCfg.Exclude.Paths
				
				if userCfg.Rules != "" {
					cfg.Rules = userCfg.Rules
				}
			}
		}
	}

	return cfg, nil
}
