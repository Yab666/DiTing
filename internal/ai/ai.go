package ai

// Analyzer 提供与大语言模型交互的能力。
// 它用于对初步发现的可疑内容进行二次分析，以排除误报。
type Analyzer struct {
	// TODO: 包含 API 密钥和厂商（如 GPT, DeepSeek）的配置
}

// FilterResult 使用 AI 判断该扫描结果是否为真实威胁。
func (a *Analyzer) FilterResult(content string) bool {
	// 调用 LLM API 的核心逻辑
	return true
}
