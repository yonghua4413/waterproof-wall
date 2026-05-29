<template>
  <div class="demo-root">
    <a class="back-link" href="/">&larr; 返回首页</a>

    <header class="demo-header">
      <h1>Waterproof Wall</h1>
      <p class="subtitle">开源滑块验证码平台 · 完整接入流程演示</p>
    </header>

    <div class="demo-grid">
      <section class="demo-card steps-card">
        <h2>接入流程</h2>
        <ol class="steps">
          <li class="step" :class="{ active: activeStep === 1, done: activeStep > 1 }">
            <span class="step-num">{{ activeStep > 1 ? '✓' : '1' }}</span>
            <div>
              <strong>加载 SDK</strong>
              <code>&lt;script src="/sdk/captcha.js"&gt;&lt;/script&gt;</code>
            </div>
          </li>
          <li class="step" :class="{ active: activeStep === 2, done: activeStep > 2 }">
            <span class="step-num">{{ activeStep > 2 ? '✓' : '2' }}</span>
            <div>
              <strong>初始化 &amp; 验证</strong>
              <code>new SliderCaptcha(&#123; appId, endpoint &#125;)</code>
            </div>
          </li>
          <li class="step" :class="{ active: activeStep === 3, done: activeStep > 3 }">
            <span class="step-num">{{ activeStep > 3 ? '✓' : '3' }}</span>
            <div>
              <strong>获取 ticket</strong>
              <code>captcha.verify() &rarr; &#123; ticket &#125;</code>
            </div>
          </li>
          <li class="step" :class="{ active: activeStep === 4, done: activeStep > 4 }">
            <span class="step-num">{{ activeStep > 4 ? '✓' : '4' }}</span>
            <div>
              <strong>发送到后端验票</strong>
              <code>POST /api/demo/verify-captcha</code>
            </div>
          </li>
          <li class="step" :class="{ active: activeStep === 5, done: activeStep > 5 }">
            <span class="step-num">{{ activeStep > 5 ? '✓' : '5' }}</span>
            <div>
              <strong>后端验票结果</strong>
              <code>&#123; success: true, appId: ..., scene: ... &#125;</code>
            </div>
          </li>
        </ol>
      </section>

      <section class="demo-card demo-form-card">
        <h2>登录演示</h2>
        <div class="demo-form">
          <div class="form-group">
            <label>用户名</label>
            <input type="text" v-model="username" placeholder="请输入用户名" />
          </div>
          <div class="form-group">
            <label>密码</label>
            <input type="password" v-model="password" placeholder="请输入密码" />
          </div>
          <div class="form-group">
            <button class="btn btn-primary captcha-btn" :disabled="loading" @click="startVerify">
              {{ loading ? '验证中...' : '验证并登录' }}
            </button>
          </div>
          <div class="captcha-container"></div>
        </div>
      </section>
    </div>

    <section class="demo-card debug-card">
      <div class="debug-header">
        <h2>调试日志</h2>
        <button class="btn btn-sm" @click="logs = []">清空</button>
      </div>
      <div class="debug-log" ref="logRef">
        <div v-for="(log, i) in logs" :key="i" class="log-entry" :class="log.type">{{ log.text }}</div>
        <div v-if="!logs.length" class="log-entry info">等待操作...</div>
      </div>
    </section>

    <Teleport to="body">
      <div class="modal-overlay" :class="{ show: modal.show }" @click.self="modal.show = false">
        <div class="modal-dialog">
          <div class="modal-icon" :class="modal.success ? 'success' : 'error'">{{ modal.success ? '✓' : '✗' }}</div>
          <h3>{{ modal.title }}</h3>
          <div class="modal-body" v-html="modal.body"></div>
          <button class="btn btn-primary modal-close-btn" @click="modal.show = false">关闭</button>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'

const endpoint = import.meta.env.VITE_APP_API_BASE || ''
const demoAppId = ref('')

const username = ref('demo')
const password = ref('123456')
const loading = ref(false)
const activeStep = ref(0)
const logs = ref([])
const logRef = ref(null)

const modal = reactive({
  show: false,
  success: true,
  title: '',
  body: '',
})

let configLoaded = false
let captchaInstance = null

function addLog(message, type = 'info') {
  const time = new Date().toLocaleTimeString()
  logs.value.push({ text: `[${time}] ${message}`, type })
  nextTick(() => {
    if (logRef.value) logRef.value.scrollTop = logRef.value.scrollHeight
  })
}

function setActiveStep(step) {
  activeStep.value = step
}

function showModal(success, title, bodyHTML) {
  modal.success = success
  modal.title = title
  modal.body = bodyHTML
  modal.show = true
}

async function loadConfig() {
  addLog('请求平台配置...', 'info')
  const resp = await fetch('/api/demo/config')
  const config = await resp.json()
  demoAppId.value = config.appId
  configLoaded = true
  addLog(`获取到 App ID: ${config.appId}, endpoint: ${config.endpoint || '(默认)'}`, 'success')
  setActiveStep(1)
}

function initCaptcha() {
  if (typeof window.SliderCaptcha === 'undefined') {
    addLog('等待 SDK 加载...', 'warn')
    return new Promise((resolve) => {
      const check = () => {
        if (typeof window.SliderCaptcha !== 'undefined') {
          addLog('SDK 加载完成', 'success')
          resolve()
        } else {
          setTimeout(check, 100)
        }
      }
      check()
    })
  }
}

async function startVerify() {
  if (!captchaInstance) {
    captchaInstance = new window.SliderCaptcha({
      appId: demoAppId.value,
      endpoint,
      scene: 'default',
    })
    addLog('SliderCaptcha 实例已创建', 'success')
  }

  loading.value = true
  setActiveStep(2)

  try {
    addLog('弹出滑块验证...', 'info')

    const result = await captchaInstance.verify()

    if (!result.success) {
      throw new Error('验证未通过，请重试')
    }

    addLog(`验证通过！ticket: ${result.ticket.substring(0, 40)}...`, 'success')
    addLog(`评分: ${result.score || 'N/A'}`, 'info')
    setActiveStep(3)

    addLog('发送 ticket 到后端验票 (POST /api/demo/verify-captcha)...', 'info')
    setActiveStep(4)

    const verifyResp = await fetch('/api/demo/verify-captcha', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        appId: demoAppId.value,
        ticket: result.ticket,
        scene: 'default',
        bizId: '',
      }),
    })

    const verifyResult = await verifyResp.json()

    if (verifyResult.success) {
      addLog('后端验票通过！', 'success')
      setActiveStep(5)

      const time = new Date().toLocaleTimeString()
      showModal(true, '验证成功！登录完成', `
        <p>滑块验证码验证通过，后端验票成功。</p>
        <pre>验票时间: ${time}
评分: ${result.score || 'N/A'}
有效期: ${result.expiresIn || 180}秒
验票结果: ${JSON.stringify(verifyResult, null, 2)}</pre>
      `)
    } else {
      throw new Error(verifyResult.error || '验票失败')
    }
  } catch (err) {
    if (err.message === 'captcha cancelled') {
      addLog('用户取消了验证', 'warn')
    } else {
      addLog(`失败: ${err.message}`, 'error')
      showModal(false, '验证失败', `<p>${err.message}</p>`)
    }
    setActiveStep(0)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (!window.SliderCaptcha) {
    const s = document.createElement('script')
    s.src = '/sdk/captcha.js'
    document.head.appendChild(s)
  }
  await loadConfig()
  await initCaptcha()
})
</script>

<style scoped>
.demo-root {
  max-width: 960px;
  margin: 0 auto;
  padding: 20px 20px 40px;
}

.back-link {
  display: inline-block;
  margin-bottom: 12px;
  color: var(--primary);
  font-size: 14px;
}

/* ===== Header ===== */
.demo-header {
  text-align: center;
  margin-bottom: 32px;
}

.demo-header h1 {
  font-size: 2rem;
  font-weight: 800;
  background: linear-gradient(135deg, #60a5fa, #a78bfa);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.demo-header .subtitle {
  color: var(--text2);
  font-size: 0.95rem;
  margin-top: 4px;
}

/* ===== Card ===== */
.demo-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
}

.demo-card h2 {
  font-size: 1.1rem;
  font-weight: 700;
  margin-bottom: 16px;
  color: var(--text);
}

/* ===== Grid ===== */
.demo-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}

@media (max-width: 768px) {
  .demo-grid {
    grid-template-columns: 1fr;
  }
}

/* ===== Steps ===== */
.steps {
  list-style: none;
}

.step {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid var(--border);
  opacity: 0.45;
  transition: opacity 0.3s;
}

.step:last-child {
  border-bottom: none;
}

.step.active {
  opacity: 1;
}

.step.done {
  opacity: 0.7;
}

.step.done .step-num {
  background: #22c55e;
}

.step-num {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--input-border);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 700;
}

.step strong {
  display: block;
  font-size: 0.9rem;
  color: var(--text);
  margin-bottom: 2px;
}

.step code {
  font-size: 0.75rem;
  color: var(--text2);
  background: var(--bg);
  padding: 2px 6px;
  border-radius: 4px;
  word-break: break-all;
}

/* ===== Form ===== */
.demo-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text2);
}

.form-group input {
  padding: 10px 14px;
  border: 1px solid var(--input-border);
  border-radius: 8px;
  background: var(--bg);
  color: var(--text);
  font-size: 0.9rem;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s;
}

.form-group input:focus {
  border-color: var(--primary);
}

.captcha-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  border: none;
}

.captcha-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.captcha-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.captcha-container {
  margin-top: 4px;
}

/* ===== Debug ===== */
.debug-card {
  margin-bottom: 20px;
}

.debug-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.debug-header h2 {
  margin-bottom: 0;
}

.debug-header .btn-sm {
  background: var(--input-border);
  color: var(--text2);
  border: none;
  font-size: 0.8rem;
  cursor: pointer;
}

.debug-header .btn-sm:hover {
  background: #475569;
  color: var(--text);
}

.debug-log {
  background: var(--bg);
  border-radius: 8px;
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
  font-family: ui-monospace, SFMono-Regular, "Consolas", monospace;
  scrollbar-width: none;
  -ms-overflow-style: none;
  font-size: 0.78rem;
  line-height: 1.6;
}

.debug-log::-webkit-scrollbar {
  display: none;
}

.log-entry {
  padding: 3px 0;
  border-bottom: 1px solid var(--border);
  word-break: break-all;
}

.log-entry:last-child {
  border-bottom: none;
}

.log-entry.info {
  color: #60a5fa;
}

.log-entry.success {
  color: #22c55e;
}

.log-entry.error {
  color: #ef4444;
}

.log-entry.warn {
  color: #f59e0b;
}

/* ===== Modal ===== */
.modal-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-overlay.show {
  display: flex;
}

.modal-dialog {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 32px;
  max-width: 440px;
  width: 90%;
  text-align: center;
  animation: demoModalIn 0.3s ease;
}

@keyframes demoModalIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.modal-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  margin: 0 auto 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.6rem;
  font-weight: 700;
}

.modal-icon.success {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.modal-icon.error {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.modal-dialog h3 {
  font-size: 1.2rem;
  margin-bottom: 12px;
}

.modal-body {
  color: var(--text2);
  font-size: 0.85rem;
  margin-bottom: 20px;
  text-align: left;
}

.modal-body pre {
  background: var(--bg);
  border-radius: 6px;
  padding: 10px;
  font-size: 0.75rem;
  overflow-x: auto;
  margin-top: 8px;
  line-height: 1.6;
}

.modal-close-btn {
  width: 100%;
}
</style>
