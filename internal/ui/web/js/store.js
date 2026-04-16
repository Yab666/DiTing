const { reactive, ref, nextTick, computed } = Vue

export const store = reactive({
    // 状态
    activeTab: 'home',
    scanPath: '',
    apiKey: localStorage.getItem('diting_deepseek_apikey') || '',
    isScanning: false,
    hasScanned: false,
    results: [],
    scanStats: { filesProcessed: 0, duration: 0 },
    consoleLogs: [],
    
    // UI 控制
    showAiModal: false,
    aiMessages: [],
    isAiThinking: false,
    hoverPreviews: {},
    previewContextLevel: Number(localStorage.getItem('diting_preview_level')) || 2,
    contextLevel: Number(localStorage.getItem('diting_context_level')) || 2,

    // 方法
    goAudit() {
        this.activeTab = 'audit'
    },
    
    setApiKey(key) {
        this.apiKey = key
        localStorage.setItem('diting_deepseek_apikey', key)
    },

    clearPreviews() {
        this.hoverPreviews = {}
        localStorage.setItem('diting_preview_level', this.previewContextLevel)
    }
})

// 抽离核心扫描逻辑
export function useScanner(store, updateChartsCallback) {
    const startScan = () => {
        if (!store.scanPath || store.isScanning) return

        store.isScanning = true
        store.hasScanned = false
        store.results = []
        store.consoleLogs = []
        store.scanStats = { filesProcessed: 0, duration: 0 }

        const startTime = Date.now()
        const timer = setInterval(() => {
            if (store.isScanning) {
                store.scanStats.duration = ((Date.now() - startTime) / 1000).toFixed(1)
            } else {
                clearInterval(timer)
            }
        }, 100)

        const eventSource = new EventSource(`/api/scan/stream?path=${encodeURIComponent(store.scanPath)}`)

        eventSource.onmessage = (event) => {
            const payload = JSON.parse(event.data)
            if (payload.type === 'progress') {
                store.scanStats.filesProcessed++
                store.consoleLogs.push(`正在扫描: ${payload.data}`)
                if (store.consoleLogs.length > 50) store.consoleLogs.shift()
            } else if (payload.type === 'found') {
                store.results.push(payload.data)
                if (updateChartsCallback) updateChartsCallback()
            } else if (payload.type === 'done') {
                eventSource.close()
                store.isScanning = false
                store.hasScanned = true
                store.results = payload.data
                if (updateChartsCallback) updateChartsCallback()
                store.consoleLogs.push(`✅ 扫描任务完成！共检出 ${store.results.length} 处隐患。`)
            }
        }

        eventSource.onerror = (err) => {
            eventSource.close()
            store.isScanning = false
            console.error("SSE Error:", err)
        }
    }

    return { startScan }
}
