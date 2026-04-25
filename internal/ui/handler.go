package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"ditting/internal/ai"
	"ditting/internal/app"
	"ditting/internal/core"
	"ditting/internal/plugin"
	"ditting/internal/rule"
	"ditting/internal/scanner"
	"ditting/pkg/config"
)

// memLogger 捕获扫描过程日志以备不时之需
type memLogger struct{}

func (l *memLogger) Info(format string, args ...interface{})  {}
func (l *memLogger) Warn(format string, args ...interface{})  {}
func (l *memLogger) Error(format string, args ...interface{}) {}

// ScanRequest 前端提交的接口参数
type ScanRequest struct {
	Path string `json:"path"`
}

// ScanEvent 代表一个实时的扫描事件包
type ScanEvent struct {
	Type string      `json:"type"` // progress, found, done
	Data interface{} `json:"data"`
}

type LLMVerifyRequest struct {
	ApiKey       string `json:"api_key"`
	RuleID       string `json:"RuleID"`
	LineNumber   int    `json:"LineNumber"`
	Content      string `json:"Content"`
	FilePath     string `json:"FilePath"`
	ContextLevel int    `json:"ContextLevel"`
}

// LLMVerifyResponse 是 AI 研判的响应包
type LLMVerifyResponse struct {
	Reply        string `json:"reply"`
	ContextMsg   string `json:"context_msg"`
	ContextBlock string `json:"context_block"`
}

// Handler 处理来自 Web UI 的 HTTP 请求。
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// HandleScan 处理核心引擎查杀调用
func (h *Handler) HandleScan(w http.ResponseWriter, r *http.Request) {
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

	l := &memLogger{}
	appConfig, _ := config.LoadConfig("", req.Path)
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

	s := scanner.NewScanner(appConfig.Exclude.Files, l)
	engine := app.NewEngine(appConfig, s, l, false)
	h.registerParsers(engine)
	engine.SetMatcher(matcher)

	secrets := engine.Run(req.Path)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secrets)
}

// HandleScanStream 实现 SSE 实时流式扫描
func (h *Handler) HandleScanStream(w http.ResponseWriter, r *http.Request) {
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
	h.registerParsers(engine)
	engine.SetMatcher(matcher)

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

	results := engine.Run(path)

	eventDone := ScanEvent{Type: "done", Data: results}
	payloadDone, _ := json.Marshal(eventDone)
	fmt.Fprintf(w, "data: %s\n\n", string(payloadDone))
	flusher.Flush()
}

// HandleLLMVerify 处理大模型二次验证
func (h *Handler) HandleLLMVerify(w http.ResponseWriter, r *http.Request) {
	var req LLMVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contextBlock := req.Content
	contextMsg := "（Lv1：极速模式，仅含有单行违规代码）"
	if req.ContextLevel > 1 {
		block, count := h.extractContext(req.FilePath, req.LineNumber, req.ContextLevel)
		if block != "" {
			contextBlock = block
			contextMsg = fmt.Sprintf("（Lv%d：视野打开，已注入前后共 %d 行代码作为研判支撑）", req.ContextLevel, count)
		}
	}

	analyzer := ai.NewAnalyzer(req.ApiKey)
	reply, err := analyzer.AnalyzeSecret(req.FilePath, req.RuleID, req.LineNumber, contextMsg, contextBlock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LLMVerifyResponse{
		Reply:        reply,
		ContextMsg:   contextMsg,
		ContextBlock: contextBlock,
	})
}

// HandlePickFolder 弹出系统文件夹选择框
func (h *Handler) HandlePickFolder(w http.ResponseWriter, r *http.Request) {
	script := `
	Add-Type -AssemblyName System.Windows.Forms
	$f = New-Object System.Windows.Forms.FolderBrowserDialog
	$f.Description = "请选择要扫描的代码目录"
	$f.ShowNewFolderButton = $false
	$topMostForm = New-Object System.Windows.Forms.Form
	$topMostForm.TopMost = $true
	$result = $f.ShowDialog($topMostForm)
	if ($result -eq "OK") { Write-Host $f.SelectedPath }
	$topMostForm.Dispose()
	`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Failed to pick folder", http.StatusInternalServerError)
		return
	}

	selectedPath := strings.TrimSpace(string(output))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": selectedPath})
}

// HandlePreview 获取代码文件预览
func (h *Handler) HandlePreview(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	lineStr := r.URL.Query().Get("line")
	line, _ := strconv.Atoi(lineStr)

	if filePath == "" || line == 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	content, _ := h.extractContext(filePath, line, 2)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}

func (h *Handler) registerParsers(e *app.Engine) {
	e.RegisterParser(plugin.NewYamlParser())
	e.RegisterParser(plugin.NewJsonParser())
	e.RegisterParser(plugin.NewPythonParser())
	e.RegisterParser(plugin.NewGoParser())
	e.RegisterParser(plugin.NewJavascriptParser())
	e.RegisterParser(plugin.NewJavaParser())
	e.RegisterParser(plugin.NewPhpParser())
	e.RegisterParser(plugin.NewShellParser())
	e.RegisterParser(plugin.NewConfigParser())
	e.RegisterParser(plugin.NewPlainTextParser())
	e.RegisterParser(plugin.NewXmlParser())
	e.RegisterParser(plugin.NewDockerfileParser())
	e.RegisterParser(plugin.NewHtmlParser())
	e.RegisterParser(plugin.NewJpropertiesParser())
	e.RegisterParser(plugin.NewNpmrcParser())
	e.RegisterParser(plugin.NewPipParser())
	e.RegisterParser(plugin.NewPypircParser())
	e.RegisterParser(plugin.NewDockercfgParser())
	e.RegisterParser(plugin.NewHtpasswdParser())
}

func (h *Handler) extractContext(filePath string, targetLine int, level int) (string, int) {
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

func OpenBrowser(url string) {
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
