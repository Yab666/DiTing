package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// KimiRequest 兼容 OpenAI 格式的请求体 (DeepSeek 同样适用)
type KimiRequest struct {
	Model    string        `json:"model"`
	Messages []KimiMessage `json:"messages"`
}

type KimiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// KimiResponse 响应体
type KimiResponse struct {
	Choices []struct {
		Message KimiMessage `json:"message"`
	} `json:"choices"`
}

// Analyzer 提供与大语言模型交互的能力。
type Analyzer struct {
	ApiKey string
}

func NewAnalyzer(apiKey string) *Analyzer {
	return &Analyzer{ApiKey: apiKey}
}

// AnalyzeSecret 使用 AI 判断该扫描结果是否为真实威胁并给出建议。
func (a *Analyzer) AnalyzeSecret(filePath, ruleID string, lineNumber int, contextMsg, contextBlock string) (string, error) {
	if a.ApiKey == "" {
		return "", fmt.Errorf("DeepSeek API Key is missing")
	}

	// 构造发给 AI 的 Prompt
	prompt := fmt.Sprintf(`你是一个资深的安全审计专家。现在有一个静态代码扫描工具拦截了一段疑似凭据泄露的代码片段/上下文。
请分析定位行所触发的告警，是真实的敏感信息硬编码（True Positive），还是误报、测试占位符（False Positive）。
文件路径：%s
命中规则：%s
定位行数：%d
%s

代码实况上下文抽取：
%s

请结合上述上下文环境和代码语义，直接给出你的思考过程，最后必须用一句加粗的文字 "结论：确认高危" 或 "结论：属于误报" 结尾。`, filePath, ruleID, lineNumber, contextMsg, contextBlock)

	kimiReq := KimiRequest{
		Model: "deepseek-chat",
		Messages: []KimiMessage{
			{Role: "system", Content: "你是一个资深的实战派网络安全专家。你的任务是分析代码中的硬编码机密泄露。请根据提供的代码上下文，准确判断该风险是【真实泄露】还是【误报】。如果是真实泄露，请给出详细的危害说明，并必须附带一份 Markdown 格式的【建议修复方案】代码块。"},
			{Role: "user", Content: prompt},
		},
	}

	reqBytes, _ := json.Marshal(kimiReq)
	httpReq, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.ApiKey)
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) DiTing-Scanner/1.0")

	client := &http.Client{
		Timeout: 45 * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to contact DeepSeek API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DeepSeek returned error (Status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var kimiResp KimiResponse
	if err := json.NewDecoder(resp.Body).Decode(&kimiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(kimiResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from AI")
	}

	return kimiResp.Choices[0].Message.Content, nil
}
