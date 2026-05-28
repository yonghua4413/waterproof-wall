<template>
  <div class="auth-mask" @click.self="$emit('close')">
    <div class="auth-card">
      <div class="auth-brand">
        <svg viewBox="0 0 24 24" fill="none" stroke="#3b82f6" stroke-width="2" width="28" height="28"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
        <span>Waterproof Wall</span>
      </div>
      <div class="auth-switch">
        <button :class="{ active: tab === 'login' }" @click="tab = 'login'">登录</button>
        <button :class="{ active: tab === 'register' }" @click="tab = 'register'">注册</button>
      </div>
      <form v-if="tab === 'login'" @submit.prevent="doLogin" class="auth-form">
        <div class="auth-field">
          <label>邮箱</label>
          <input type="email" v-model="loginEmail" placeholder="you@example.com" autocomplete="email">
        </div>
        <div class="auth-field">
          <label>密码</label>
          <input type="password" v-model="loginPw" placeholder="请输入密码" autocomplete="current-password">
        </div>
        <button type="submit" class="auth-submit" :disabled="loginning">{{ loginning ? '登录中...' : '登录' }}</button>
        <p class="auth-alt">还没有账号？<a href="#" @click.prevent="tab = 'register'">立即注册</a></p>
      </form>
      <form v-else @submit.prevent="doRegister" class="auth-form">
        <div class="auth-field">
          <label>邮箱</label>
          <input type="email" v-model="regEmail" placeholder="you@example.com" autocomplete="email">
        </div>
        <div class="auth-field">
          <label>密码</label>
          <input type="password" v-model="regPw" placeholder="至少 8 位" autocomplete="new-password">
        </div>
        <div class="auth-field">
          <label>确认密码</label>
          <input type="password" v-model="regPw2" placeholder="再次输入密码" autocomplete="new-password">
        </div>
        <button type="submit" class="auth-submit" :disabled="registering">{{ registering ? '注册中...' : '注册' }}</button>
        <p class="auth-alt">已有账号？<a href="#" @click.prevent="tab = 'login'">去登录</a></p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useApi } from '../composables/useApi'
import { useAuth } from '../composables/useAuth'
import { showToast } from '../composables/useToast'

const emit = defineEmits(['close', 'done'])
const { api } = useApi()
const { setAuth } = useAuth()

const tab = ref('login')
const loginning = ref(false)
const registering = ref(false)
const loginEmail = ref('')
const loginPw = ref('')
const regEmail = ref('')
const regPw = ref('')
const regPw2 = ref('')

async function doLogin() {
  if (loginning.value) return
  loginning.value = true
  try {
    const res = await api('POST', '/api/v1/users/login', { email: loginEmail.value.trim(), password: loginPw.value })
    showToast('登录成功')
    setAuth(res)
    emit('done')
  } catch (e) { showToast(e.message, 'error') }
  finally { loginning.value = false }
}

async function doRegister() {
  if (registering.value) return
  if (regPw.value !== regPw2.value) return showToast('两次密码不一致', 'error')
  if (regPw.value.length < 8) return showToast('密码至少8位', 'error')
  registering.value = true
  try {
    const res = await api('POST', '/api/v1/users/register', { email: regEmail.value.trim(), password: regPw.value })
    if (res.user && res.user.status !== 'active') {
      showToast('注册成功，请等待管理员激活', 'error')
      setAuth(res)
      emit('done')
      return
    }
    showToast('注册成功')
    setAuth(res)
    emit('done')
  } catch (e) { showToast(e.message, 'error') }
  finally { registering.value = false }
}
</script>

<style scoped>
.auth-mask {
  position: fixed; inset: 0; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; z-index: 200;
}
.auth-card {
  background: #fff; border-radius: 12px; padding: 32px 28px; width: min(400px, 92vw); color: #1e293b;
  box-shadow: 0 20px 60px rgba(0,0,0,.15);
}
.auth-brand {
  display: flex; align-items: center; justify-content: center; gap: 10px; margin-bottom: 24px; font-size: 18px; font-weight: 700; color: #0f172a;
}
.auth-switch {
  display: flex; background: #f1f5f9; border-radius: 8px; padding: 3px; margin-bottom: 24px;
}
.auth-switch button {
  flex: 1; height: 36px; border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;
  background: transparent; color: #64748b; transition: all .2s;
}
.auth-switch button.active {
  background: #fff; color: #0f172a; box-shadow: 0 1px 3px rgba(0,0,0,.1);
}

.auth-form { display: flex; flex-direction: column; }
.auth-field { margin-bottom: 16px; }
.auth-field label { display: block; font-size: 13px; font-weight: 600; color: #475569; margin-bottom: 6px; }
.auth-field input {
  width: 100%; height: 40px; border: 1px solid #cbd5e1; border-radius: 6px; padding: 0 12px; font-size: 14px;
  color: #0f172a; background: #fff; transition: border-color .15s;
}
.auth-field input:focus { outline: none; border-color: #3b82f6; box-shadow: 0 0 0 2px rgba(59,130,246,.15); }
.auth-field input::placeholder { color: #94a3b8; }

.auth-submit {
  width: 100%; height: 42px; border: none; border-radius: 6px; font-size: 15px; font-weight: 600;
  cursor: pointer; background: #3b82f6; color: #fff; transition: background .15s; margin-top: 4px;
}
.auth-submit:hover { background: #2563eb; }
.auth-submit:disabled { background: #94a3b8; cursor: not-allowed; }

.auth-alt {
  text-align: center; font-size: 13px; color: #94a3b8; margin-top: 16px; margin-bottom: 0;
}
.auth-alt a { color: #3b82f6; text-decoration: none; font-weight: 500; }
.auth-alt a:hover { text-decoration: underline; }
</style>