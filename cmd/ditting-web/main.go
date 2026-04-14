package main

import (
	"ditting/internal/ui"
	"flag"
	"fmt"
	"os"
)

func main() {
	port := flag.Int("port", 8080, "Web 控制台的访问端口")
	flag.Parse()

	if err := ui.StartWebServer(*port); err != nil {
		fmt.Printf("Web 控制台启动失败: %v\n", err)
		os.Exit(1)
	}
}
