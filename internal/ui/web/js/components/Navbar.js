import { store } from '../store.js'

export default {
    template: `
    <nav class="navbar panel">
        <div class="nav-brand" @click="store.activeTab = 'home'">
            <img src="assets/logo.png" class="nav-logo">
            <span>DiTing</span>
        </div>
        <div class="nav-links">
            <button :class="{ active: store.activeTab === 'home' }" @click="store.activeTab = 'home'">
                <i class="ph ph-house"></i> 首页
            </button>
            <button :class="{ active: store.activeTab === 'audit' }" @click="store.activeTab = 'audit'">
                <i class="ph ph-shield-check"></i> 审计中心
            </button>
        </div>
    </nav>
    `,
    setup() {
        return { store }
    }
}
