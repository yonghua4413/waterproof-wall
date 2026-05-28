<template>
  <div class="dash">
    <div v-if="showInactiveModal" class="modal-mask" @click.self="showInactiveModal = false">
      <div class="modal-card">
        <h3 style="margin:0 0 12px;font-size:18px">账号未激活</h3>
        <p style="color:var(--text2);font-size:14px;margin:0 0 16px">请添加微信联系管理员激活账号后再使用</p>
        <img src="/me.jpg" alt="微信二维码" style="width:200px;height:200px;border-radius:8px;border:1px solid var(--border)">
        <button class="btn" style="margin-top:16px;width:100%" @click="showInactiveModal = false">知道了</button>
      </div>
    </div>
    <div class="dash-header">
      <h1>控制台</h1>
      <div style="display:flex;align-items:center;gap:10px">
        <button class="btn btn-ghost btn-sm" @click="logout">退出</button>
      </div>
    </div>
    <div style="display:flex;gap:8px;margin-bottom:16px;align-items:center">
      <button class="btn btn-primary" @click="userStatus === 'active' ? showCreate = true : showInactiveModal = true">+ 创建应用</button>
      <div style="flex:1"></div>
      <label style="font-size:13px;color:var(--text2)">日期</label>
      <input type="date" v-model="usageDate" style="height:32px;border:1px solid var(--input-border);border-radius:4px;padding:0 8px;font:inherit;font-size:13px;color:var(--text);background:var(--input-bg)">
      <button class="btn btn-ghost btn-sm" @click="loadUsage">刷新</button>
    </div>
    <div>
      <div v-if="!apps.length" class="app-item" style="text-align:center;color:#8896a7">暂无应用，点击上方按钮创建</div>
      <div v-for="app in apps" :key="app.appId" class="app-item">
        <div class="app-head">
          <div><span class="app-name">{{ app.name }}</span> <span class="app-status">{{ app.status }}</span></div>
          <div class="app-actions">
            <button class="btn btn-ghost btn-sm" @click="editApp = app">编辑域名</button>
            <button class="btn btn-ghost btn-sm" @click="confirmRotate(app.appId)">重置 Secret</button>
          </div>
        </div>
        <div class="field-row"><label>App ID</label><div class="field-val">{{ app.appId }}</div></div>
        <div class="field-row"><label>App Key</label><div class="field-val">{{ app.appKey }}</div></div>
        <div class="field-row"><label>域名</label><div class="field-val">{{ app.domains && app.domains.length ? app.domains.join(', ') : '*' }}</div></div>
        <div class="field-row"><label>创建时间</label><div class="field-val">{{ fmtTime(app.createdAt) }}</div></div>
        <div v-if="usageMap[app.appId]" class="stats-grid">
          <div class="stat-card"><div class="stat-num">{{ usageMap[app.appId].Challenges }}</div><div class="stat-label">请求</div></div>
          <div class="stat-card"><div class="stat-num">{{ usageMap[app.appId].Passes }}</div><div class="stat-label">通过</div></div>
          <div class="stat-card"><div class="stat-num">{{ usageMap[app.appId].Fails }}</div><div class="stat-label">失败</div></div>
          <div class="stat-card"><div class="stat-num">{{ usageMap[app.appId].Tickets }}</div><div class="stat-label">验票</div></div>
          <div class="stat-card"><div class="stat-num">{{ usageMap[app.appId].Passes > 0 ? Math.round(usageMap[app.appId].Passes / (usageMap[app.appId].Passes + usageMap[app.appId].Fails) * 100) + '%' : '0%' }}</div><div class="stat-label">通过率</div></div>
        </div>
      </div>
    </div>

    <CreateAppModal v-if="showCreate" @close="showCreate = false" @created="onCreateApp" />
    <SecretModal v-if="secretApp" :app="secretApp" @close="secretApp = null" />
    <EditDomainsModal v-if="editApp" :initial="editApp.domains ? editApp.domains.join('\n') : ''" @close="editApp = null" @save="onSaveDomains" />
    <ConfirmModal v-if="confirmAction" :title="confirmAction.title" :message="confirmAction.message" @close="confirmAction = null" @confirm="runConfirm" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'
import { useAuth } from '../composables/useAuth'
import { showToast } from '../composables/useToast'
import CreateAppModal from '../components/CreateAppModal.vue'
import SecretModal from '../components/SecretModal.vue'
import EditDomainsModal from '../components/EditDomainsModal.vue'
import ConfirmModal from '../components/ConfirmModal.vue'

const { token, userStatus, logout } = useAuth()
const { api } = useApi()

const apps = ref([])
const usageMap = ref({})
const usageDate = ref(new Date().toISOString().slice(0, 10))
const showCreate = ref(false)
const secretApp = ref(null)
const editApp = ref(null)
const confirmAction = ref(null)
const showInactiveModal = ref(userStatus.value && userStatus.value !== 'active')

async function loadApps() {
  try {
    const res = await api('GET', '/api/v1/apps', null, token.value)
    apps.value = res.apps || []
  } catch (e) { showToast(e.message, 'error') }
}

async function loadUsage() {
  if (!apps.value.length) return
  const date = usageDate.value || new Date().toISOString().slice(0, 10)
  try {
    const res = await api('GET', '/api/v1/apps/usage?date=' + date, null, token.value)
    const m = {}
    if (res.apps) res.apps.forEach(a => { m[a.app.ID] = a.daily })
    usageMap.value = m
  } catch (e) { showToast(e.message, 'error') }
}

async function onCreateApp(data) {
  try {
    const res = await api('POST', '/api/v1/apps', data, token.value)
    showCreate.value = false
    if (res.app && res.app.appSecret) secretApp.value = res.app
    await loadApps()
    await loadUsage()
  } catch (e) { showToast(e.message, 'error') }
}

function confirmRotate(appId) {
  confirmAction.value = {
    title: '重置 App Secret',
    message: '重置后旧 Secret 立即失效。确定？',
    action: async () => {
      try {
        const res = await api('POST', '/api/v1/apps/rotate-secret', { appId }, token.value)
        secretApp.value = res.app
        await loadApps()
      } catch (e) { showToast(e.message, 'error') }
    }
  }
}

function runConfirm() {
  if (confirmAction.value && confirmAction.value.action) confirmAction.value.action()
  confirmAction.value = null
}

async function onSaveDomains(domains) {
  try {
    await api('POST', '/api/v1/apps/update-domains', { appId: editApp.value.appId, domains }, token.value)
    editApp.value = null
    showToast('域名已更新')
    await loadApps()
    await loadUsage()
  } catch (e) { showToast(e.message, 'error') }
}

function fmtTime(s) {
  if (!s) return '-'
  return s.slice(0, 19).replace('T', ' ')
}

loadApps().then(() => loadUsage())
</script>

<style scoped>
.modal-mask {
  position: fixed; inset: 0; background: rgba(0,0,0,.45);
  display: flex; align-items: center; justify-content: center; z-index: 200;
}
.modal-card {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: 12px; padding: 28px; width: min(360px, 90vw);
  text-align: center; box-shadow: 0 20px 60px rgba(0,0,0,.3);
}
</style>