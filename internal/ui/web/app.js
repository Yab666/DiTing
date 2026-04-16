const { createApp, ref, computed, onMounted, nextTick } = Vue

createApp({
    setup() {
        const scanPath = ref('')
        const isScanning = ref(false)
        const hasScanned = ref(false)
        const results = ref([])
        const scanStats = ref({ filesProcessed: 0, duration: 0 })
        const consoleLogs = ref([])
        const consoleBody = ref(null)

        // 导航状态 (解耦：首页 vs 审计中心)
        const activeTab = ref('home')
        const goAudit = () => activeTab.value = 'audit'

        // 图表实例
        let typeChartInstance = null
        let severityChartInstance = null

        // 初始化图表
        const initCharts = () => {
            const typeCtx = document.getElementById('typeChart').getContext('2d')
            const sevCtx = document.getElementById('severityChart').getContext('2d')

            typeChartInstance = new Chart(typeCtx, {
                type: 'radar',
                data: {
                    labels: ['Password', 'Secret', 'URI', 'File', 'Key', 'Other'],
                    datasets: [{
                        label: '漏洞分布',
                        data: [0, 0, 0, 0, 0, 0],
                        fill: true,
                        backgroundColor: 'rgba(59, 130, 246, 0.2)',
                        borderColor: 'rgb(59, 130, 246)',
                        pointBackgroundColor: 'rgb(59, 130, 246)',
                    }]
                },
                options: {
                    scales: { r: { grid: { color: '#262626' }, ticks: { display: false } } },
                    plugins: { legend: { labels: { color: '#ededed', font: { size: 10 } } } }
                }
            })

            severityChartInstance = new Chart(sevCtx, {
                type: 'doughnut',
                data: {
                    labels: ['Critical', 'Major', 'Minor', 'Info'],
                    datasets: [{
                        data: [0, 0, 0, 0],
                        backgroundColor: ['#ef4444', '#f59e0b', '#3b82f6', '#737373'],
                        borderWidth: 0
                    }]
                },
                options: {
                    cutout: '70%',
                    plugins: { legend: { position: 'right', labels: { color: '#ededed', font: { size: 10 } } } }
                }
            })
        }

        const updateCharts = () => {
            if (!typeChartInstance || !severityChartInstance) return

            // 更新严重程度
            const sevData = [
                results.value.filter(r => r.Severity === 'CRITICAL').length,
                results.value.filter(r => r.Severity === 'MAJOR').length,
                results.value.filter(r => r.Severity === 'MINOR').length,
                results.value.filter(r => r.Severity === 'INFO').length
            ]
            severityChartInstance.data.datasets[0].data = sevData
            severityChartInstance.update()

            // 更新类型分布 (基于 RuleID 关键字简单分类)
            const types = [0, 0, 0, 0, 0, 0] // Password, Secret, URI, File, Key, Other
            results.value.forEach(r => {
                const id = r.RuleID.toLowerCase()
                if (id.includes('password')) types[0]++
                else if (id.includes('secret')) types[1]++
                else if (id.includes('uri')) types[2]++
                else if (id.includes('file')) types[3]++
                else if (id.includes('key')) types[4]++
                else types[5]++
            })
            typeChartInstance.data.datasets[0].data = types
            typeChartInstance.update()
        }

        const startScan = () => {
            if (!scanPath.value || isScanning.value) return

            isScanning.value = true
            hasScanned.value = false
            results.value = []
            consoleLogs.value = []
            scanStats.value = { filesProcessed: 0, duration: 0 }

            nextTick(() => {
                if (!typeChartInstance) initCharts()
                updateCharts()
            })

            const startTime = Date.now()
            const timer = setInterval(() => {
                if (isScanning.value) {
                    scanStats.value.duration = ((Date.now() - startTime) / 1000).toFixed(1)
                } else {
                    clearInterval(timer)
                }
            }, 100)

            // 使用 SSE 开启实时双向流
            const eventSource = new EventSource(`/api/scan/stream?path=${encodeURIComponent(scanPath.value)}`)

            eventSource.onmessage = (event) => {
                const payload = JSON.parse(event.data)
                
                if (payload.type === 'progress') {
                    scanStats.value.filesProcessed++
                    consoleLogs.value.push(`正在扫描: ${payload.data}`)
                    if (consoleLogs.value.length > 50) consoleLogs.value.shift()
                    
                    nextTick(() => {
                        if (consoleBody.value) consoleBody.value.scrollTop = consoleBody.value.scrollHeight
                    })
                } else if (payload.type === 'found') {
                    results.value.push(payload.data)
                    updateCharts()
                } else if (payload.type === 'done') {
                    eventSource.close()
                    isScanning.value = false
                    hasScanned.value = true
                    results.value = payload.data
                    updateCharts()
                    consoleLogs.value.push(`✅ 扫描任务圆满完成！共检出 ${results.value.length} 处隐患。`)
                }
            }

            eventSource.onerror = (err) => {
                eventSource.close()
                isScanning.value = false
                console.error("SSE Error:", err)
            }
        }

        const totalCritical = computed(() => results.value.filter(r => ['CRITICAL', 'BLOCKER'].includes(r.Severity.toUpperCase())).length)
        const totalMajor = computed(() => results.value.filter(r => r.Severity.toUpperCase() === 'MAJOR').length)

        // 筛选与排序逻辑
        const filterSeverity = ref('ALL')
        const filterSearch = ref('')
        const sortKey = ref('Severity') // 默认按严重程度排
        const sortOrder = ref(-1) // -1 是降序

        const sevMap = { 'CRITICAL': 4, 'BLOCKER': 4, 'MAJOR': 3, 'MINOR': 2, 'INFO': 1 }

        const filteredResults = computed(() => {
            let list = results.value.filter(item => {
                const matchSev = filterSeverity.value === 'ALL' || item.Severity.toUpperCase() === filterSeverity.value
                const s = filterSearch.value.toLowerCase()
                const matchText = !s || item.FilePath.toLowerCase().includes(s) || 
                                 item.RuleID.toLowerCase().includes(s) || 
                                 item.Content.toLowerCase().includes(s)
                return matchSev && matchText
            })

            return list.sort((a, b) => {
                let valA = a[sortKey.value]
                let valB = b[sortKey.value]
                
                // 特殊处理严重程度权重排序
                if (sortKey.value === 'Severity') {
                    valA = sevMap[valA.toUpperCase()] || 0
                    valB = sevMap[valB.toUpperCase()] || 0
                }

                if (valA < valB) return -1 * sortOrder.value
                if (valA > valB) return 1 * sortOrder.value
                return 0
            })
        })

        const toggleSort = (key) => {
            if (sortKey.value === key) {
                sortOrder.value *= -1
            } else {
                sortKey.value = key
                sortOrder.value = -1
            }
        }

        // 悬浮预览逻辑
        const hoverPreviews = ref({})
        const previewContextLevel = ref(Number(localStorage.getItem('diting_preview_level')) || 2)
        
        const clearPreviews = () => {
            hoverPreviews.value = {}
            localStorage.setItem('diting_preview_level', previewContextLevel.value)
        }

        const loadPreview = async (item) => {
            const key = `${item.FilePath}:${item.LineNumber}`
            if (hoverPreviews.value[key]) return
            
            try {
                const res = await fetch(`/api/ui/preview?file=${encodeURIComponent(item.FilePath)}&line=${item.LineNumber}&level=${previewContextLevel.value}`)
                const data = await res.json()
                if (data.content) {
                    hoverPreviews.value[key] = parseContextToHtml(data.content, item.LineNumber)
                }
            } catch (e) {
                hoverPreviews.value[key] = '<div style="color:#ef4444;padding:10px">无法加载预览</div>'
            }
        }

        const shortenPath = (path) => {
            if (!path) return ''
            if (path.length <= 40) return path
            const parts = path.split(/[\/\\]/)
            return parts.length <= 2 ? path.substring(0, 40) + '...' : `.../${parts[parts.length - 2]}/${parts[parts.length - 1]}`
        }

        // AI 逻辑
        const showAiModal = ref(false)
        const aiMessages = ref([])
        const isAiThinking = ref(false)
        const chatBody = ref(null)
        const pickFolderLoading = ref(false)
        const apiKey = ref(localStorage.getItem('diting_deepseek_apikey') || '')
        const contextLevel = ref(Number(localStorage.getItem('diting_context_level')) || 2)

        const closeAiModal = () => showAiModal.value = false
        const scrollToBottom = () => setTimeout(() => { if (chatBody.value) chatBody.value.scrollTop = chatBody.value.scrollHeight }, 50)

        const pickFolder = async () => {
            if (pickFolderLoading.value) return
            pickFolderLoading.value = true
            try {
                const res = await fetch('/api/ui/pick-folder')
                const data = await res.json()
                if (data.path) scanPath.value = data.path
            } catch (e) {
                console.error("Picker error:", e)
            } finally {
                pickFolderLoading.value = false
            }
        }

        const parseContextToHtml = (block, vulnLine) => {
            if (!block) return ''
            const lines = block.split('\n')
            let html = '<div class="code-viewport"><div class="viewport-title">源文件审计快照</div><div class="viewport-body">'
            lines.forEach(line => {
                if (!line.trim() && line === '') return
                const match = line.match(/^\s*(\d+): (.*)$/)
                if (match) {
                    const num = parseInt(match[1])
                    const content = match[2]
                    const isVuln = num === vulnLine
                    // 使用紧凑格式，防止 white-space: pre 引入意外空格
                    html += `<div class="code-line ${isVuln ? 'vuln-highlight' : ''}">` +
                            `<div class="line-num">${num}</div>` +
                            `<div class="code-content">${content.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</div></div>`
                }
            })
            html += '</div></div>'
            return html
        }

        const verifyWithAI = async (item, idx) => {
            if (!apiKey.value.trim()) return alert("请先输入 API Key！")
            localStorage.setItem('diting_deepseek_apikey', apiKey.value.trim())
            localStorage.setItem('diting_context_level', contextLevel.value)
            
            item.aiStatus = '诊断中...'
            showAiModal.value = true
            aiMessages.value = [
                { role: 'ai', content: `你好！我是 **谛听 AI (DeepSeek)**。正在对 <code>${item.RuleID}</code> 进行深度研判...` }
            ]
            isAiThinking.value = true
            scrollToBottom()

            try {
                const res = await fetch('/api/llm/verify', {
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
                })
                const result = await res.json()
                isAiThinking.value = false

                // 1. 插入代码上下文预览气泡
                if (result.context_block) {
                    aiMessages.value.push({ 
                        role: 'ai', 
                        content: parseContextToHtml(result.context_block, item.LineNumber) 
                    })
                }

                // 2. 插入 AI 分析结果
                const replyHtml = result.reply
                    .replace(/### (.*)/g, '<h3 style="color:#a78bfa;margin:10px 0">$1</h3>')
                    .replace(/结论：(.*)/g, '<div class="remediation-status"><span class="fix-badge">最终结论</span><br>结论：$1</div>')
                    .replace(/\n/g, '<br>')
                    .replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')

                aiMessages.value.push({ role: 'ai', content: replyHtml })
                
                // 3. 如果是高危且有修复建议需求，引导按钮 (此处展示修复按钮)
                if (result.reply.includes('结论：确认高危')) {
                    aiMessages.value.push({ 
                        role: 'ai', 
                        isAction: true,
                        content: `<button class="btn-fix-request" onclick="window.requestFix('${item.RuleID}', ${item.LineNumber})"><i class="ph ph-lightbulb"></i> 获取一键修复建议方案</button>`
                    })
                }

                item.aiStatus = result.reply.includes('属于误报') ? '确认误报' : '确认高危'
                scrollToBottom()
            } catch (err) {
                isAiThinking.value = false
                item.aiStatus = '请求失败'
                aiMessages.value.push({ role: 'ai', content: '❌ 请求失败，请检查网络或 API Key。' })
            }
        }

        // 挂载全局方法供 HTML 里的 onclick 调用
        window.requestFix = (ruleId, line) => {
            aiMessages.value.push({ role: 'user', content: `请针对第 ${line} 行的 ${ruleId} 漏洞给出具体的代码修复方案。` })
            isAiThinking.value = true
            scrollToBottom()
            // 模拟 AI 生成修复方案
            setTimeout(() => {
                isAiThinking.value = false
                aiMessages.value.push({ 
                    role: 'ai', 
                    content: `<div class="remediation-status">
                        <span class="fix-badge">建议修复方案</span><br>
                        检测到该处使用了硬编码凭据。推荐做法：<br>
                        1. 将敏感信息提取到本地 <code>.env</code> 文件中。<br>
                        2. 在代码中使用环境变量读取：<br>
                        <pre><code>// 修复示例\npassword := os.Getenv("DB_PASSWORD")</code></pre>
                    </div>`
                })
                scrollToBottom()
            }, 1000)
        }

        return {
            scanPath, isScanning, hasScanned, results, scanStats, consoleLogs, consoleBody,
            startScan, totalCritical, totalMajor, shortenPath, verifyWithAI,
            showAiModal, aiMessages, isAiThinking, closeAiModal, chatBody, apiKey, contextLevel, pickFolder,
            filteredResults, filterSeverity, filterSearch, sortKey, toggleSort, hoverPreviews, loadPreview,
            previewContextLevel, clearPreviews, activeTab, goAudit
        }
    }
}).mount('#app')
