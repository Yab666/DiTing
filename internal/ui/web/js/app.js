import { store } from './store.js'
import Navbar from './components/Navbar.js'
import HomeView from './views/HomeView.js'
import AuditView from './views/AuditView.js'

const { createApp, nextTick } = Vue

const app = createApp({
    components: {
        Navbar,
        HomeView,
        AuditView
    },
    setup() {
        const scrollToBottom = (el) => setTimeout(() => { if (el) el.scrollTop = el.scrollHeight }, 50)

        const verifyWithAI = async (item, idx) => {
            if (!store.apiKey) return alert("请先输入 API Key！")
            
            item.aiStatus = '诊断中...'
            store.showAiModal = true
            store.aiMessages = [{ role: 'ai', content: `正在对 <code>${item.RuleID}</code> 进行研判...` }]
            store.isAiThinking = true

            try {
                const res = await fetch('/api/llm/verify', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        api_key: store.apiKey,
                        RuleID: item.RuleID,
                        LineNumber: item.LineNumber,
                        Content: item.Content,
                        FilePath: item.FilePath,
                        ContextLevel: Number(store.contextLevel)
                    })
                })
                const result = await res.json()
                store.isAiThinking = false

                if (result.context_block) {
                    store.aiMessages.push({ role: 'ai', content: parseContextToHtml(result.context_block, item.LineNumber) })
                }

                const replyHtml = result.reply
                    .replace(/### (.*)/g, '<h3 style="color:#a78bfa;margin:10px 0">$1</h3>')
                    .replace(/结论：(.*)/g, '<div class="remediation-status"><span class="fix-badge">结论</span><br>$1</div>')
                    .replace(/\n/g, '<br>')
                    .replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')

                store.aiMessages.push({ role: 'ai', content: replyHtml })
                item.aiStatus = result.reply.includes('属于误报') ? '确认误报' : '确认高危'
            } catch (err) {
                store.isAiThinking = false
                store.aiMessages.push({ role: 'ai', content: '❌ 请求失败' })
            }
        }

        const parseContextToHtml = (block, vulnLine) => {
            const lines = block.split('\n')
            let html = '<div class="code-viewport"><div class="viewport-title">源文件审计快照</div><div class="viewport-body">'
            lines.forEach(line => {
                const match = line.match(/^\s*(\d+): (.*)$/)
                if (match) {
                    const num = parseInt(match[1]), content = match[2], isVuln = num === vulnLine
                    html += `<div class="code-line ${isVuln ? 'vuln-highlight' : ''}"><div class="line-num">${num}</div><div class="code-content">${content.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</div></div>`
                }
            })
            return html + '</div></div>'
        }

        return { store, verifyWithAI }
    }
})

app.mount('#app')
