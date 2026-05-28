<template>
  <div id="app-root">
    <nav class="nav">
      <router-link to="/" class="nav-logo">
        <svg viewBox="0 0 24 24" fill="none" stroke="var(--primary)" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
        Waterproof Wall
      </router-link>
      <div class="nav-right">
        <router-link to="/">首页</router-link>
        <router-link to="/docs">文档</router-link>
        <router-link to="/example">体验</router-link>
        <router-link to="/source">源码</router-link>
        <button v-if="isLoggedIn" class="btn btn-white btn-sm" @click="router.push('/dashboard')">控制台</button>
        <button v-else class="btn btn-white btn-sm" @click="showAuth = true">登录</button>
      </div>
    </nav>

    <router-view />

    <AuthModal v-if="showAuth" @close="showAuth = false" @done="onAuthDone" />
    <Toast />

    <footer>
      <p>Waterproof Wall &mdash; 开源滑块验证码平台</p>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from './composables/useAuth'
import { useApi } from './composables/useApi'
import AuthModal from './components/AuthModal.vue'
import Toast from './components/Toast.vue'

const router = useRouter()
const { isLoggedIn, token, setAuth, logout } = useAuth()
const { api } = useApi()
const showAuth = ref(false)

function onAuthDone() {
  showAuth.value = false
  router.push('/dashboard')
}

function onShowAuth() { showAuth.value = true }

onMounted(() => {
  window.addEventListener('show-auth', onShowAuth)
  if (isLoggedIn.value) {
    api('GET', '/api/v1/users/me', null, token.value).then(res => {
      setAuth({ user: res.user })
    }).catch(err => {
      if (err.message === 'unauthorized') logout()
    })
  }
})
onUnmounted(() => window.removeEventListener('show-auth', onShowAuth))
</script>