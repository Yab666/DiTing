const { createApp, ref, computed } = Vue

createApp({
    setup() {
        const scanPath = ref('')
        const isScanning = ref(false)
        const hasScanned = ref(false)
        const results = ref([])

        const startScan = async () => {
            if (!scanPath.value || isScanning.value) return;

            isScanning.value = true;
            hasScanned.value = false;
            results.value = [];

            try {
                const response = await fetch('/api/scan', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ path: scanPath.value })
                });

                if (!response.ok) {
                    const text = await response.text();
                    alert(`Scan failed: ${text}`);
                    return;
                }

                results.value = await response.json();
                hasScanned.value = true;
            } catch (err) {
                alert(`Error connecting to scanner: ${err.message}`);
            } finally {
                isScanning.value = false;
            }
        }

        const totalCritical = computed(() => {
            return results.value.filter(r => 
                r.Severity.toUpperCase() === 'CRITICAL' || 
                r.Severity.toUpperCase() === 'BLOCKER'
            ).length;
        });

        const totalMajor = computed(() => {
            return results.value.filter(r => r.Severity.toUpperCase() === 'MAJOR').length;
        });

        // Helper to shorten long paths for the UI
        const shortenPath = (path) => {
            if (!path) return '';
            if (path.length <= 40) return path;
            const parts = path.split(/[\/\\]/);
            if (parts.length <= 2) return path.substring(0, 40) + '...';
            return `.../${parts[parts.length - 2]}/${parts[parts.length - 1]}`;
        }

        const showAiModal = ref(false)
        const aiMessages = ref([])
        const isAiThinking = ref(false)
        const chatBody = ref(null)
        const pickFolderLoading = ref(false)
        
        // DeepSeek API Key 状态管理，并从本地存储持久化
        const apiKey = ref(localStorage.getItem('diting_deepseek_apikey') || '')
        const contextLevel = ref(Number(localStorage.getItem('diting_context_level')) || 2)

        const closeAiModal = () => {
            showAiModal.value = false;
        }

        const scrollToBottom = () => {
            setTimeout(() => {
                if (chatBody.value) {
                    chatBody.value.scrollTop = chatBody.value.scrollHeight;
                }
            }, 50);
        }

        const pickFolder = async () => {
            if (pickFolderLoading.value) return;
            pickFolderLoading.value = true;
            try {
                const response = await fetch('/api/ui/pick-folder');
                const data = await response.json();
                if (data.path) {
                    scanPath.value = data.path;
                }
            } catch (e) {
                console.error("无法调起文件夹选择器:", e);
            } finally {
                pickFolderLoading.value = false;
            }
        }

        const verifyWithAI = async (item, idx) => {
            if (!apiKey.value.trim()) {
                alert("请先在上方的控制栏中输入 DeepSeek API Key！");
                return;
            }
            // 保存 apiKey 和下拉策略 到本地
            localStorage.setItem('diting_deepseek_apikey', apiKey.value.trim());
            localStorage.setItem('diting_context_level', contextLevel.value);

            item.aiStatus = '诊断中...';
            showAiModal.value = true;
            
            // 初始化对话
            aiMessages.value = [
                { role: 'ai', content: `你好！我是 **谛听 AI 审查大脑 (Powered by DeepSeek)**。您的探针已设定为 **视野 Lv${contextLevel.value}**，正在向我灌装上下文数据。` },
                { role: 'user', content: `请帮我审查文件 <code>${shortenPath(item.FilePath)}</code> 的第 ${item.LineNumber} 行附近代码<br><br>它命中了 <b>${item.RuleID}</b> 规则，这是真实的泄露还是误报？` }
            ];
            
            isAiThinking.value = true;
            scrollToBottom();

            try {
                // 真实调用 Kimi/DS API 接入点
                const response = await fetch('/api/llm/verify', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        api_key: apiKey.value.trim(),
                        RuleID: item.RuleID,
                        LineNumber: item.LineNumber,
                        Content: item.Content,
                        FilePath: item.FilePath,
                        ContextLevel: Number(contextLevel.value)
                    })
                });
                
                isAiThinking.value = false;
                
                if (!response.ok) {
                    const errMsg = await response.text();
                    aiMessages.value.push({
                        role: 'ai',
                        content: `<span style="color: #ef4444">调用失败: ${errMsg}</span>`
                    });
                    item.aiStatus = '调用失败';
                } else {
                    const result = await response.json();
                    
                    // markdown简单的换行转html处理
                    const formattedReply = result.reply.replace(/\n/g, '<br>');
                    const ctxMsg = result.context_msg ? `<br><br><span style="color:#a78bfa; font-size: 11px;">${result.context_msg}</span>` : '';

                    aiMessages.value.push({ 
                        role: 'ai', 
                        content: formattedReply + ctxMsg
                    });
                    
                    // 尝试从回复中提取结论来更新状态按钮
                    if (result.reply.includes('属于误报') || result.reply.includes('False Positive')) {
                        item.aiStatus = '确认误报';
                    } else if (result.reply.includes('确认高危') || result.reply.includes('True Positive')) {
                        item.aiStatus = '确认极危';
                    } else {
                        item.aiStatus = '审查完毕';
                    }
                }
                scrollToBottom();
            } catch (err) {
                isAiThinking.value = false;
                aiMessages.value.push({
                    role: 'ai',
                    content: `<span style="color: #ef4444">网络请求失败: ${err.message}</span>`
                });
                item.aiStatus = '请求异常';
                scrollToBottom();
            }
        }

        return {
            scanPath,
            isScanning,
            hasScanned,
            results,
            startScan,
            totalCritical,
            totalMajor,
            shortenPath,
            verifyWithAI,
            showAiModal,
            aiMessages,
            isAiThinking,
            closeAiModal,
            chatBody,
            apiKey,
            contextLevel,
            pickFolder
        }
    }
}).mount('#app')
