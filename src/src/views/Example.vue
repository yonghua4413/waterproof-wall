<template>
  <div style="min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center">
    <div class="card">
      <a class="back" href="/">&larr; 返回首页</a>
      <h2>SDK 演示</h2>
      <div class="field"><label>App ID</label><input v-model="appId" placeholder="从控制台获取"></div>
      <button class="btn" @click="startCaptcha">开始验证</button>
      <div v-if="result" class="result">{{ result }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const endpoint = import.meta.env.VITE_APP_API_BASE || ''
const appId = ref('app_Ovx-y8QZNRKZK5cSbOTo1mP6A4Gp5BNi')
const result = ref('')

onMounted(() => {
  if (!window.SliderCaptcha) {
    const s = document.createElement('script')
    s.src = '/sdk/captcha.js'
    document.head.appendChild(s)
  }
})

function startCaptcha() {
  if (!appId.value.trim()) { alert('请输入 App ID'); return }
  const c = new window.SliderCaptcha({ endpoint, appId: appId.value })
  c.verify().then(r => {
    result.value = '验证成功！Ticket: ' + JSON.stringify(r, null, 2)
  }).catch(err => {
    result.value = '验证失败: ' + err.message
  })
}
</script>

<style scoped>
.card { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 24px; width: min(460px, 90vw); }
.card h2 { font-size: 18px; margin-bottom: 16px; }
.field { margin-bottom: 12px; }
.field label { display: block; font-size: 13px; font-weight: 600; margin-bottom: 4px; color: var(--text2); }
.field input { width: 100%; height: 38px; border: 1px solid var(--input-border); border-radius: 6px; padding: 0 12px; font: inherit; color: var(--text); background: var(--input-bg); }
.btn { display: inline-flex; align-items: center; justify-content: center; height: 40px; padding: 0 20px; border: 0; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer; background: var(--primary); color: #fff; width: 100%; margin-top: 8px; }
.result { margin-top: 12px; padding: 10px; background: var(--input-bg); border: 1px solid var(--border); border-radius: 6px; font-size: 13px; color: var(--text2); word-break: break-all; max-height: 120px; overflow-y: auto; }
.back { display: inline-block; margin-bottom: 16px; color: var(--primary); font-size: 14px; }
</style>