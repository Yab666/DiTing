import { store, useScanner } from '../store.js'

export default {
    template: `
    <transition name="fade-slide">
        <div class="audit-view" v-if="store.activeTab === 'audit'">
            <!-- 控制面板 -->
            <section class="control-panel panel">
                <div class="input-group">
                    <i class="ph ph-folder-open" @click="pickFolder" style="cursor: pointer; color: var(--color-accent);"></i>
                    <input type="text" v-model="store.scanPath" @keyup.enter="startScan" :disabled="store.isScanning" placeholder="请输入要扫描的源码文件夹路径...">
                </div>
                <div class="input-group" style="max-width: 280px;">
                    <i class="ph ph-key"></i>
                    <input type="password" v-model="store.apiKey" @change="e => store.setApiKey(e.target.value)" placeholder="DeepSeek API Key (可选)">
                </div>
                <button class="btn btn-primary" @click="startScan" :disabled="store.isScanning">
                    <i class="ph ph-radar" :class="{ spinning: store.isScanning }"></i>
                    {{ store.isScanning ? '审计中...' : '启动审计' }}
                </button>
            </section>

            <!-- 数字化大屏 -->
            <section class="dashboard-grid" v-if="store.isScanning || store.hasScanned">
                <div class="stats-sidebar">
                    <div class="stat-card panel" :class="{ 'warning-glow': store.isScanning }">
                        <i class="ph ph-files"></i>
                        <div class="stat-info">
                            <span class="stat-value">{{ store.scanStats.filesProcessed }}</span>
                            <span class="stat-label">分析文件</span>
                        </div>
                    </div>
                    <div class="stat-card panel critical" :class="{ 'critical-glow': store.results.length > 0 }">
                        <i class="ph ph-bug"></i>
                        <div class="stat-info">
                            <span class="stat-value">{{ store.results.length }}</span>
                            <span class="stat-label">捕获点</span>
                        </div>
                    </div>
                </div>

                <div class="console-panel panel">
                    <div class="panel-header">
                        <span><i class="ph ph-terminal"></i> ENGINE LOGS</span>
                        <div class="status-indicator" :class="{ active: store.isScanning }"></div>
                    </div>
                    <div class="console-body" ref="consoleBody">
                        <div v-for="(log, idx) in store.consoleLogs" :key="idx" class="log-line">
                            <span class="log-cursor">>></span> {{ log }}
                        </div>
                    </div>
                </div>

                <div class="charts-sidebar panel">
                   <div class="chart-box"><canvas id="typeChart"></canvas></div>
                   <div class="chart-box"><canvas id="severityChart"></canvas></div>
                </div>
            </section>

            <!-- 数据表格 -->
            <section class="table-panel panel" v-if="store.hasScanned && store.results.length > 0">
                <div class="filter-bar">
                    <div class="search-box">
                        <i class="ph ph-magnifying-glass"></i>
                        <input type="text" v-model="filterSearch" placeholder="搜索结果...">
                    </div>
                    <select v-model="filterSeverity">
                        <option value="ALL">全部级别</option>
                        <option value="CRITICAL">Critical</option>
                        <option value="MAJOR">Major</option>
                        <option value="MINOR">Minor</option>
                    </select>
                    <div class="text-secondary text-sm">匹配: {{ filteredResults.length }} / {{ store.results.length }}</div>
                </div>

                <table>
                    <thead>
                        <tr>
                            <th class="sortable" @click="toggleSort('Severity')">风险等级</th>
                            <th class="sortable" @click="toggleSort('RuleID')">命中规则</th>
                            <th>文件路径</th>
                            <th>代码快照</th>
                            <th class="text-right">智能研判</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(item, idx) in filteredResults" :key="idx">
                            <td><span class="badge" :class="item.Severity.toLowerCase()">{{ item.Severity }}</span></td>
                            <td class="font-mono">{{ item.RuleID }}</td>
                            <td class="text-muted" :title="item.FilePath">{{ shortenPath(item.FilePath) }} L{{ item.LineNumber }}</td>
                            <td>
                                <div class="popover-container" @mouseenter="loadPreview(item)">
                                    <code class="snippet">{{ item.Content }}</code>
                                    <div class="popover-content">
                                        <div class="popover-header">代码上下文预览</div>
                                        <div class="popover-body" v-html="store.hoverPreviews[item.FilePath + ':' + item.LineNumber] || '加载中...'"></div>
                                    </div>
                                </div>
                            </td>
                            <td class="text-right">
                                <button class="btn-ai" @click="$emit('verify-ai', item, idx)" :disabled="item.aiVerifying">
                                    <i class="ph ph-magic-wand"></i> AI 复核
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </section>
        </div>
    </transition>
    `,
    emits: ['verify-ai'],
    setup(props, { emit }) {
        const { ref, computed, onMounted, nextTick } = Vue
        const consoleBody = ref(null)

        // 图表逻辑
        let typeChartInstance = null
        let severityChartInstance = null

        const initCharts = () => {
             const typeCtx = document.getElementById('typeChart')?.getContext('2d')
             const sevCtx = document.getElementById('severityChart')?.getContext('2d')
             if (!typeCtx || !sevCtx) return

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
            const r = store.results
            const sevData = [
                r.filter(x => x.Severity === 'CRITICAL').length,
                r.filter(x => x.Severity === 'MAJOR').length,
                r.filter(x => x.Severity === 'MINOR').length,
                r.filter(x => x.Severity === 'INFO').length
            ]
            severityChartInstance.data.datasets[0].data = sevData
            severityChartInstance.update()

            const types = [0, 0, 0, 0, 0, 0]
            r.forEach(x => {
                const id = x.RuleID.toLowerCase()
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

        const { startScan } = useScanner(store, updateCharts)

        const filterSeverity = ref('ALL')
        const filterSearch = ref('')
        const sortKey = ref('Severity')
        const sortOrder = ref(-1)

        const filteredResults = computed(() => {
            let list = store.results.filter(item => {
                const matchSev = filterSeverity.value === 'ALL' || item.Severity.toUpperCase() === filterSeverity.value
                const s = filterSearch.value.toLowerCase()
                return matchSev && (!s || item.FilePath.toLowerCase().includes(s) || item.RuleID.toLowerCase().includes(s) || item.Content.toLowerCase().includes(s))
            })
            return list.sort((a, b) => {
                const sevMap = { 'CRITICAL': 4, 'MAJOR': 3, 'MINOR': 2, 'INFO': 1 }
                let valA = sortKey.value === 'Severity' ? (sevMap[a.Severity.toUpperCase()] || 0) : a[sortKey.value]
                let valB = sortKey.value === 'Severity' ? (sevMap[b.Severity.toUpperCase()] || 0) : b[sortKey.value]
                return valA < valB ? -1 * sortOrder.value : (valA > valB ? 1 * sortOrder.value : 0)
            })
        })

        const toggleSort = (key) => {
            if (sortKey.value === key) sortOrder.value *= -1
            else { sortKey.value = key; sortOrder.value = -1 }
        }

        const shortenPath = (path) => {
            if (!path || path.length <= 40) return path
            const parts = path.split(/[\/\\]/)
            return parts.length <= 2 ? path.substring(0, 40) + '...' : `.../${parts[parts.length - 2]}/${parts[parts.length - 1]}`
        }

        const pickFolder = async () => {
            const res = await fetch('/api/ui/pick-folder')
            const data = await res.json()
            if (data.path) store.scanPath = data.path
        }

        const loadPreview = async (item) => {
            const key = `${item.FilePath}:${item.LineNumber}`
            if (store.hoverPreviews[key]) return
            const res = await fetch(`/api/ui/preview?file=${encodeURIComponent(item.FilePath)}&line=${item.LineNumber}&level=${store.previewContextLevel}`)
            const data = await res.json()
            if (data.content) store.hoverPreviews[key] = parseContextToHtml(data.content, item.LineNumber)
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

        onMounted(() => {
            if (store.hasScanned && !typeChartInstance) {
                nextTick(() => { initCharts(); updateCharts(); })
            }
        })

        return { store, startScan, filterSeverity, filterSearch, filteredResults, toggleSort, shortenPath, pickFolder, loadPreview, consoleBody }
    }
}
