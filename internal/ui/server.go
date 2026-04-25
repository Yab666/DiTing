package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed web/*
var webAssets embed.FS

// StartWebServer 启动本地可交互仪表盘
func StartWebServer(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	h := NewHandler()

	// API 路由 - 绑定到 Handler 成员方法
	http.HandleFunc("/api/scan", h.HandleScan)
	http.HandleFunc("/api/scan/stream", h.HandleScanStream)
	http.HandleFunc("/api/llm/verify", h.HandleLLMVerify)
	http.HandleFunc("/api/ui/pick-folder", h.HandlePickFolder)
	http.HandleFunc("/api/ui/preview", h.HandlePreview)

	// 静态资源路由（映射 embed.FS 里的 web 目录）
	subFS, err := fs.Sub(webAssets, "web")
	if err != nil {
		fmt.Printf("挂载 Web 资源失败: %v\n", err)
	}
	http.Handle("/", http.FileServer(http.FS(subFS)))

	fmt.Printf("=== 谛听 (DiTing) 可视化仪表盘已启动 ===\n")
	fmt.Printf("请在浏览器中访问: http://%s\n", addr)

	OpenBrowser("http://" + addr)

	return http.ListenAndServe(addr, nil)
}
