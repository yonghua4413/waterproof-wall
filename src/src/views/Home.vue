<template>
  <div>
    <section class="hero">
      <h1>安全可靠的滑块验证码平台</h1>
      <p>多租户架构，HMAC 签名验票，一行代码接入，为你的应用保驾护航</p>
      <div class="hero-actions">
        <button class="btn btn-primary" @click="onStart">立即开始</button>
      </div>
    </section>

    <section id="features" class="features">
      <h2>核心特性</h2>
      <div class="features-grid">
        <div class="feat"><h3>🔐 HMAC 签名验票</h3><p>Ticket 使用 HMAC-SHA256 签名，防止伪造和重放攻击，保障验证安全性</p></div>
        <div class="feat"><h3>🏢 多租户隔离</h3><p>每个应用独立 appId/appKey/appSecret，数据完全隔离，互不影响</p></div>
        <div class="feat"><h3>🧠 行为分析</h3><p>轨迹滑动速度、加速度分析，精准区分人机，提升防护效果</p></div>
        <div class="feat"><h3>📦 一行代码接入</h3><p>前端 SDK 仅需一行 script 标签引入，调用 initCaptcha 即可完成接入</p></div>
        <div class="feat"><h3>🔑 密钥安全</h3><p>App Secret 仅在创建和重置时展示一次，数据库加密存储</p></div>
        <div class="feat"><h3>📊 用量统计</h3><p>按日统计请求数、通过率、验票数，一目了然掌握验证服务状态</p></div>
      </div>
    </section>

    <section class="integration">
      <h2>快速接入</h2>
      <h3 style="font-size:16px;margin-bottom:12px;color:var(--text2)">前端</h3>
      <div class="code-block">
        <pre v-html="codeExampleFront"></pre>
      </div>
    </section>

    <section class="faq">
      <h2>常见问题</h2>
      <div class="faq-list">
        <div class="faq-item">
          <h3>什么是滑块验证码？</h3>
          <p>滑块验证码是一种人机验证方式，用户通过拖动滑块完成拼图来证明自己是真人，而非自动化脚本。相比传统字符验证码，滑块验证码用户体验更好、安全性更高。</p>
        </div>
        <div class="faq-item">
          <h3>HMAC 签名验票是什么？</h3>
          <p>后端验证 ticket 时，前端使用 App Secret 对请求进行 HMAC-SHA256 签名，服务端校验签名后才能通过。这确保了 ticket 无法被伪造或重放。</p>
        </div>
        <div class="faq-item">
          <h3>如何接入？</h3>
          <p>只需两步：1. 前端引入 SDK，调用 SliderCaptcha 的 verify 方法获取 ticket；2. 后端使用 App Key 和 App Secret 对请求签名，调用验票接口验证 ticket 有效性。</p>
        </div>
        <div class="faq-item">
          <h3>支持多租户吗？</h3>
          <p>支持。每个应用独立分配 appId、appKey 和 appSecret，数据完全隔离，互不影响。</p>
        </div>
        <div class="faq-item">
          <h3>能防止哪些攻击？</h3>
          <p>可防止自动化脚本批量注册、暴力破解登录、恶意抢购薅羊毛、爬虫恶意抓取等。同时通过行为轨迹分析（滑动速度、加速度）精准区分人机。</p>
        </div>
        <div class="faq-item">
          <h3>是开源的吗？</h3>
          <p>是。Waterproof Wall 是开源滑块验证码平台，前端 Vue 3 + Vite，后端 Go，支持自部署。</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const domain = import.meta.env.VITE_APP_DOMAIN || 'https://007.hallo.run'
const { isLoggedIn } = useAuth()
const router = useRouter()

const codeExampleFront = computed(() => {
  return `<span class="comment">// 1. 引入 SDK</span>
&lt;script src="${domain}/sdk/captcha.js"&gt;&lt;/script&gt;

<span class="comment">// 2. 初始化验证码</span>
const captcha = new SliderCaptcha({
  appId: 'your-app-id',
  endpoint: '${domain}'
});
captcha.verify().then(result => {
  <span class="comment">// 3. 将 ticket 发送到后端验票</span>
  fetch('/api/v1/ticket/check', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ticket: result.ticket })
  });
});`
})

function onStart() {
  if (!isLoggedIn.value) {
    window.dispatchEvent(new Event('show-auth'))
  } else {
    router.push('/docs')
  }
}
</script>