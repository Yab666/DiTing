import { store } from '../store.js'

export default {
    template: `
    <transition name="fade-slide">
        <section class="landing-hero" v-if="store.activeTab === 'home'">
            <img src="assets/logo.png" class="hero-logo">
            <h1>DiTing Security Radar</h1>
            <p class="hero-subtitle">
                基于 Go 并发引擎与 DeepSeek AI 驱动的下一代隐私挖掘与秘密审计平台。
                <br>精准、高速、零误报。
            </p>
            <div class="hero-actions">
                <button class="btn btn-hero" @click="store.goAudit()">
                    <i class="ph ph-rocket-launch"></i> 立即进入审计中心
                </button>
                <div class="hero-badges">
                    <span><i class="ph ph-lightning"></i> 极速并发</span>
                    <span><i class="ph ph-brain"></i> AI 降噪</span>
                    <span><i class="ph ph-shield-check"></i> 开箱即用</span>
                </div>
            </div>
        </section>
    </transition>
    `,
    setup() {
        return { store }
    }
}
