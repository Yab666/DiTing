package ui

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ditting/internal/app"
	"ditting/internal/core"
	"ditting/internal/plugin"
	"ditting/internal/rule"
	"ditting/internal/scanner"
	"ditting/pkg/config"
)

//go:embed web/*
var webAssets embed.FS

// memLogger 捕获扫描过程日志以备不时之需
type memLogger struct{}
func (l *memLogger) Info(format string, args ...interface{})  {}
func (l *memLogger) Warn(format string, args ...interface{})  {}
func (l *memLogger) Error(format string, args ...interface{}) {}

// ScanRequest 前端提交的接口参数
type ScanRequest struct {
	Path string `json:"path"`
}

// StartWebServer 启动本地可交互仪表盘
func StartWebServer(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	// API 路由
	http.HandleFunc("/api/scan", handleScan)
	http.HandleFunc("/api/scan/stream", handleScanStream)
	http.HandleFunc("/api/llm/verify", handleLLMVerify)
	http.HandleFunc("/api/ui/pick-folder", handlePickFolder)
	http.HandleFunc("/api/ui/preview", handlePreview)

	// 静态资源路由（映射 embed.FS 里的 web 目录）
	subFS, err := fs.Sub(webAssets, "web")
	if err != nil {
		fmt.Printf("挂载 Web 资源失败: %v\n", err)
	}
	http.Handle("/", http.FileServer(http.FS(subFS)))

	fmt.Printf("=== 谛听 (DiTing) 可视化仪表盘已启动 ===\n")
	fmt.Printf("请在浏览器中访问: http://%s\n", addr)

	openBrowser("http://" + addr)

	return http.ListenAndServe(addr, nil)
}

// handleScan 处理核心引擎查杀调用
func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "目录路径不能为空", http.StatusBadRequest)
		return
	}

	// 1. 初始化哑日志和配置
	l := &memLogger{}
	appConfig, _ := config.LoadConfig("", req.Path)

	// 2. 加载规则库 (如果不存在则用兜底，这里因为是 Web 版，默认去当前工作目录找 configs/rules)
	loader := rule.NewRuleLoader()
	rulesDir := appConfig.Rules
	if rulesDir == "" {
		rulesDir = "configs/rules"
	}
	rules, err := loader.LoadFromDir(rulesDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("规则引擎加载失败: %v", err), http.StatusInternalServerError)
		return
	}
	matcher := rule.NewMatcher(rules, appConfig)

	// 3. 构建临时引擎
	s := scanner.NewScanner(appConfig.Exclude.Files, l)
	engine := app.NewEngine(appConfig, s, l, false)
	
	// 注册插件
	engine.RegisterParser(plugin.NewYamlParser())
	engine.RegisterParser(plugin.NewJsonParser())
	engine.RegisterParser(plugin.NewPythonParser())
	engine.RegisterParser(plugin.NewGoParser())
	engine.RegisterParser(plugin.NewJavascriptParser())
	engine.RegisterParser(plugin.NewJavaParser())
	engine.RegisterParser(plugin.NewPhpParser())
	engine.RegisterParser(plugin.NewShellParser())
	engine.RegisterParser(plugin.NewConfigParser())
	engine.RegisterParser(plugin.NewPlainTextParser())
	engine.RegisterParser(plugin.NewXmlParser())
	engine.RegisterParser(plugin.NewDockerfileParser())
	engine.RegisterParser(plugin.NewHtmlParser())
	engine.RegisterParser(plugin.NewJpropertiesParser())
	engine.RegisterParser(plugin.NewNpmrcParser())
	engine.RegisterParser(plugin.NewPipParser())
	engine.RegisterParser(plugin.NewPypircParser())
	engine.RegisterParser(plugin.NewDockercfgParser())
	engine.RegisterParser(plugin.NewHtpasswdParser())
	engine.SetMatcher(matcher)

	// 4. 起飞！收集数据
	secrets := engine.Run(req.Path)

	// 5. 返回 JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secrets)
}

// ScanEvent 代表一个实时的扫描事件包
type ScanEvent struct {
	Type string      `json:"type"` // progress, found, done
	Data interface{} `json:"data"`
}

// handleScanStream 实现 SSE 实时流式扫描
func handleScanStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := r.URL.Query().Get("path")
	if path == "" {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", "Missing path")
		return
	}

	flusher, _ := w.(http.Flusher)

	// 1. 初始化引擎与配置
	l := &memLogger{}
	appConfig, _ := config.LoadConfig("", path)
	loader := rule.NewRuleLoader()
	rulesDir := appConfig.Rules
	if rulesDir == "" {
		rulesDir = "configs/rules"
	}
	rules, _ := loader.LoadFromDir(rulesDir)
	matcher := rule.NewMatcher(rules, appConfig)
	s := scanner.NewScanner(appConfig.Exclude.Files, l)
	engine := app.NewEngine(appConfig, s, l, false)

	// 注册插件 (与 handleScan 保持完全一致)
	engine.RegisterParser(plugin.NewYamlParser())
	engine.RegisterParser(plugin.NewJsonParser())
	engine.RegisterParser(plugin.NewPythonParser())
	engine.RegisterParser(plugin.NewGoParser())
	engine.RegisterParser(plugin.NewJavascriptParser())
	engine.RegisterParser(plugin.NewJavaParser())
	engine.RegisterParser(plugin.NewPhpParser())
	engine.RegisterParser(plugin.NewShellParser())
	engine.RegisterParser(plugin.NewConfigParser())
	engine.RegisterParser(plugin.NewPlainTextParser())
	engine.RegisterParser(plugin.NewXmlParser())
	engine.RegisterParser(plugin.NewDockerfileParser())
	engine.RegisterParser(plugin.NewHtmlParser())
	engine.RegisterParser(plugin.NewJpropertiesParser())
	engine.RegisterParser(plugin.NewNpmrcParser())
	engine.RegisterParser(plugin.NewPipParser())
	engine.RegisterParser(plugin.NewPypircParser())
	engine.RegisterParser(plugin.NewDockercfgParser())
	engine.RegisterParser(plugin.NewHtpasswdParser())
	engine.SetMatcher(matcher)

	// 2. 绑定实时钩子
	engine.OnProgress = func(file string) {
		event := ScanEvent{Type: "progress", Data: file}
		payload, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", string(payload))
		flusher.Flush()
	}

	engine.OnFound = func(secret core.Secret) {
		event := ScanEvent{Type: "found", Data: secret}
		payload, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", string(payload))
		flusher.Flush()
	}

	// 3. 开启扫描
	results := engine.Run(path)

	// 4. 发送结束信号
	eventDone := ScanEvent{Type: "done", Data: results}
	payloadDone, _ := json.Marshal(eventDone)
	fmt.Fprintf(w, "data: %s\n\n", string(payloadDone))
	flusher.Flush()
}

// openBrowser 自动唤起用户默认浏览器
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		// ignore
	}
}

type LLMVerifyRequest struct {
	ApiKey       string `json:"api_key"`
	RuleID       string `json:"RuleID"`
	LineNumber   int    `json:"LineNumber"`
	Content      string `json:"Content"`
	FilePath     string `json:"FilePath"`
	ContextLevel int    `json:"ContextLevel"`
}

func extractContext(filePath string, targetLine int, level int) (string, int) {
	// level 1: 0, level 2: 5, level 3: 20, level 4: 50
	radius := 0
	switch level {
	case 2:
		radius = 5
	case 3:
		radius = 20
	case 4:
		radius = 50
	}

	if radius == 0 {
		return "", 0
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", 0
	}
	defer file.Close()

	startLine := targetLine - radius
	if startLine < 1 {
		startLine = 1
	}
	endLine := targetLine + radius

	var sb strings.Builder
	scanner := bufio.NewScanner(file)
	currentLine := 1
	actualCount := 0

	for scanner.Scan() {
		if currentLine >= startLine && currentLine <= endLine {
			if currentLine == targetLine {
				sb.WriteString(fmt.Sprintf("%d: >>> %s <<<\n", currentLine, scanner.Text()))
			} else {
				sb.WriteString(fmt.Sprintf("%d: %s\n", currentLine, scanner.Text()))
			}
			actualCount++
		}
		if currentLine > endLine {
			break
		}
		currentLine++
	}
	return sb.String(), actualCount
}

// LLMVerifyResponse 是 AI 研判的响应包
type LLMVerifyResponse struct {
	Reply        string `json:"reply"`
	ContextMsg   string `json:"context_msg"`
	ContextBlock string `json:"context_block"`
}

type KimiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type KimiRequest struct {
	Model    string        `json:"model"`
	Messages []KimiMessage `json:"messages"`
}

type KimiResponse struct {
	Choices []struct {
		Message KimiMessage `json:"message"`
	} `json:"choices"`
}

// handleLLMVerify 处理真实大模型(DeepSeek)的二次验证
func handleLLMVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req LLMVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ApiKey == "" {
		http.Error(w, "DeepSeek API Key is missing", http.StatusUnauthorized)
		return
	}

	// 动态截取文件上下文
	contextBlock := req.Content
	contextMsg := "（Lv1：极速模式，仅含有单行违规代码）"
	if req.ContextLevel > 1 {
		block, count := extractContext(req.FilePath, req.LineNumber, req.ContextLevel)
		if block != "" {
			contextBlock = block
			contextMsg = fmt.Sprintf("（Lv%d：视野打开，已注入前后共 %d 行代码大盘作为研判支撑）", req.ContextLevel, count)
		}
	}

	// 构造发给 DeepSeek 的 Prompt
	prompt := fmt.Sprintf(`你是一个资深的安全审计专家。现在有一个静态代码扫描工具拦截了一段疑似凭据泄露的代码片段/上下文。
请分析定位行所触发的告警，是真实的敏感信息硬编码（True Positive），还是误报、测试占位符（False Positive）。
文件路径：%s
命中规则：%s
定位行数：%d
%s

代码实况上下文抽取：
%s

请结合上述上下文环境和代码语义，直接给出你的思考过程，最后必须用一句加粗的文字 "结论：确认高危" 或 "结论：属于误报" 结尾。`, req.FilePath, req.RuleID, req.LineNumber, contextMsg, contextBlock)

	kimiReq := KimiRequest{
		Model: "deepseek-chat", // 使用 DeepSeek 的模型
		Messages: []KimiMessage{
			{Role: "system", Content: "你是一个资深的实战派网络安全专家。你的任务是分析代码中的硬编码机密泄露。请根据提供的代码上下文，准确判断该风险是【真实泄露】还是【误报】。如果是真实泄露，请给出详细的危害说明，并必须附带一份 Markdown 格式的【建议修复方案】代码块。"},
			{Role: "user", Content: prompt},
		},
	}

	reqBytes, _ := json.Marshal(kimiReq)
	httpReq, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", strings.NewReader(string(reqBytes)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	httpReq.Close = true // 防治底层连接复用导致的突然 EOF
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.ApiKey)
	// 增加 User-Agent 伪装，防止某些云 WAF 因识别到 Go-http-client 而直接强制掐断 TCP 连接（报 EOF）
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) DiTing-Scanner/1.0")

	// 增加显式超时，防止一直挂起
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, "Failed to contact DeepSeek API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("DeepSeek 返回了错误 (Status %d): %s", resp.StatusCode, string(bodyBytes)), http.StatusInternalServerError)
		return
	}

	var kimiResp KimiResponse
	if err := json.NewDecoder(resp.Body).Decode(&kimiResp); err != nil {
		http.Error(w, "Failed to decode DeepSeek response", http.StatusInternalServerError)
		return
	}

	if len(kimiResp.Choices) == 0 {
		http.Error(w, "Empty response from DeepSeek API", http.StatusInternalServerError)
		return
	}

	aiReply := kimiResp.Choices[0].Message.Content

	// 返回结果给前端
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LLMVerifyResponse{
		Reply:        aiReply,
		ContextMsg:   contextMsg,
		ContextBlock: contextBlock,
	})
}

// handlePickFolder 唤起系统原生文件夹选择对话框 (支持 Windows)
func handlePickFolder(w http.ResponseWriter, r *http.Request) {
	if runtime.GOOS != "windows" {
		http.Error(w, "目前该功能仅支持 Windows 系统", http.StatusNotImplemented)
		return
	}

	// 利用 PowerShell 唤起原生的 FolderBrowserDialog
	// 关键：强制设置输出编码为 UTF8，并使用 TopMost 窗体强制将对话框置顶
	psScript := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = '请选择要扫描的源码文件夹'; $t = New-Object System.Windows.Forms.Form; $t.TopMost = $true; if($f.ShowDialog($t) -eq 'OK'){ $f.SelectedPath }`
	
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	out, err := cmd.Output()
	
	selectedPath := strings.TrimSpace(string(out))
	if err != nil || selectedPath == "" {
		// 返回空路径表示取消选择
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"path": ""})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": selectedPath})
}

// handlePreview 提供轻量级的实时代码快照预览 (无需 AI)
func handlePreview(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	lineStr := r.URL.Query().Get("line")
	levelStr := r.URL.Query().Get("level") // 新增：等级参数
	
	lineNum, _ := strconv.Atoi(lineStr)
	level, err := strconv.Atoi(levelStr)
	if err != nil || level <= 0 {
		level = 2 // 默认前后约 5 行
	}

	if filePath == "" || lineNum <= 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// 统一获取前后 N 行作为“轻量级”预览
	context, _ := extractContext(filePath, lineNum, level)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"content": context,
	})
}
