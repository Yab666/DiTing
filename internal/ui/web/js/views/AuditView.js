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

            <!-- 数字化大屏 (始终显示，修复图表挂载时找不到DOM的问题) -->
            <section class="dashboard-grid">
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
                    <div class="stat-card panel warning">
                        <i class="ph ph-warning-diamond"></i>
                        <div class="stat-info">
                            <span class="stat-value">{{ criticalCount }}</span>
                            <span class="stat-label">致命风险</span>
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

                <div class="charts-sidebar panel" style="display: flex; flex-direction: column;">
                    <div class="panel-header" style="padding-bottom: 12px; border-bottom: 1px solid rgba(255,255,255,0.05); margin-bottom: 16px;">
                        <span style="font-size: 13px; font-weight: 600; color: #e5e5e5; display: flex; align-items: center; gap: 8px;">
                            <i class="ph-duotone ph-chart-donut" style="color: var(--color-accent); font-size: 16px;"></i> 风险构成分析
                        </span>
                    </div>
                    
                    <!-- 极客风 2x2 统计网格 -->
                    <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 10px; padding: 0 16px; margin-bottom: 24px;">
                        <div style="background: rgba(239, 68, 68, 0.05); border: 1px solid rgba(239, 68, 68, 0.15); border-radius: 8px; padding: 12px; display: flex; flex-direction: column; align-items: center; justify-content: center;">
                            <span style="color: rgba(255,255,255,0.5); font-size: 10px; letter-spacing: 1px; margin-bottom: 4px;">CRITICAL</span>
                            <span style="color: #f87171; font-family: 'JetBrains Mono', monospace; font-size: 20px; font-weight: 700; text-shadow: 0 0 10px rgba(239,68,68,0.5);">{{ criticalCount }}</span>
                        </div>
                        <div style="background: rgba(245, 158, 11, 0.05); border: 1px solid rgba(245, 158, 11, 0.15); border-radius: 8px; padding: 12px; display: flex; flex-direction: column; align-items: center; justify-content: center;">
                            <span style="color: rgba(255,255,255,0.5); font-size: 10px; letter-spacing: 1px; margin-bottom: 4px;">MAJOR</span>
                            <span style="color: #fbbf24; font-family: 'JetBrains Mono', monospace; font-size: 20px; font-weight: 700; text-shadow: 0 0 10px rgba(245,158,11,0.5);">{{ majorCount }}</span>
                        </div>
                        <div style="background: rgba(59, 130, 246, 0.05); border: 1px solid rgba(59, 130, 246, 0.15); border-radius: 8px; padding: 12px; display: flex; flex-direction: column; align-items: center; justify-content: center;">
                            <span style="color: rgba(255,255,255,0.5); font-size: 10px; letter-spacing: 1px; margin-bottom: 4px;">MINOR</span>
                            <span style="color: #60a5fa; font-family: 'JetBrains Mono', monospace; font-size: 20px; font-weight: 700; text-shadow: 0 0 10px rgba(59,130,246,0.5);">{{ minorCount }}</span>
                        </div>
                        <div style="background: rgba(255, 255, 255, 0.02); border: 1px solid rgba(255, 255, 255, 0.05); border-radius: 8px; padding: 12px; display: flex; flex-direction: column; align-items: center; justify-content: center;">
                            <span style="color: rgba(255,255,255,0.5); font-size: 10px; letter-spacing: 1px; margin-bottom: 4px;">INFO</span>
                            <span style="color: #a3a3a3; font-family: 'JetBrains Mono', monospace; font-size: 20px; font-weight: 700;">{{ infoCount }}</span>
                        </div>
                    </div>

                    <!-- 动态图表区域 (居中饱满的环形图) -->
                    <div style="flex: 1; padding: 0 16px 16px 16px; display: flex; flex-direction: column; justify-content: center; position: relative;">
                        <div style="position: absolute; top: 0; left: 16px; right: 16px; border-top: 1px dashed rgba(255,255,255,0.1);"></div>
                        <div class="chart-box" style="position: relative; width: 100%; height: 160px; margin-top: 16px;">
                            <canvas id="severityChart"></canvas>
                        </div>
                    </div>
                </div>
            </section>

            <!-- 数据表格区 (强制显示，避免布局坍塌) -->
            <section class="table-panel panel">
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
                        <tr v-if="filteredResults.length === 0">
                            <td colspan="5" style="text-align: center; padding: 30px; color: var(--text-secondary);">
                                <i class="ph ph-coffee" style="font-size: 24px; margin-bottom: 8px; display: block;"></i>
                                暂无审计数据...
                            </td>
                        </tr>
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
        let severityChartInstance = null

        const initCharts = () => {
             const sevCtx = document.getElementById('severityChart')?.getContext('2d')
             if (!sevCtx) return

             severityChartInstance = new Chart(sevCtx, {
                 type: 'doughnut',
                 data: {
                     labels: ['Critical', 'Major', 'Minor', 'Info'],
                     datasets: [{
                         data: [0, 0, 0, 0],
                         backgroundColor: ['#ef4444', '#f59e0b', '#3b82f6', '#404040'],
                         borderWidth: 0,
                         hoverOffset: 4
                     }]
                 },
                 options: {
                     responsive: true,
                     maintainAspectRatio: false,
                     cutout: '75%',
                     plugins: { 
                         legend: { 
                             display: false
                         },
                         tooltip: {
                             backgroundColor: 'rgba(0,0,0,0.8)',
                             titleFont: { size: 13 },
                             bodyFont: { size: 12, family: "'JetBrains Mono', monospace" },
                             padding: 10,
                             cornerRadius: 8
                         }
                     },
                     layout: { padding: 0 }
                 }
             })
        }

        const updateCharts = () => {
            if (!severityChartInstance) return
            const r = store.results
            const sevData = [
                r.filter(x => x.Severity === 'CRITICAL').length,
                r.filter(x => x.Severity === 'MAJOR').length,
                r.filter(x => x.Severity === 'MINOR').length,
                r.filter(x => x.Severity === 'INFO').length
            ]
            severityChartInstance.data.datasets[0].data = sevData
            severityChartInstance.update()
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
            // 只要进入审计页，就立即初始化图表框架，使用双重 nextTick 确保 DOM 完全渲染
            nextTick(() => { 
                setTimeout(() => {
                    initCharts(); 
                    if (store.results.length > 0) updateCharts();
                }, 100)
            })
        })

        // 计算属性：用于右侧统计表
        const criticalCount = computed(() => store.results.filter(x => x.Severity === 'CRITICAL').length)
        const majorCount = computed(() => store.results.filter(x => x.Severity === 'MAJOR').length)
        const minorCount = computed(() => store.results.filter(x => x.Severity === 'MINOR').length)
        const infoCount = computed(() => store.results.filter(x => x.Severity === 'INFO').length)

        return { store, startScan, filterSeverity, filterSearch, filteredResults, toggleSort, shortenPath, pickFolder, loadPreview, consoleBody, criticalCount, majorCount, minorCount, infoCount }
    }
}
