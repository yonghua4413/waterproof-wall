import { ref, computed } from 'vue'

const token = ref(localStorage.getItem('captcha_token') || '')
const userStatus = ref(localStorage.getItem('captcha_user_status') || '')
const isLoggedIn = computed(() => !!token.value)

function setAuth(data) {
  if (data.token) {
    token.value = data.token
    localStorage.setItem('captcha_token', data.token)
  }
  if (data.user && data.user.status) {
    userStatus.value = data.user.status
    localStorage.setItem('captcha_user_status', data.user.status)
  }
}

function logout() {
  token.value = ''
  userStatus.value = ''
  localStorage.removeItem('captcha_token')
  localStorage.removeItem('captcha_user_status')
}

export function useAuth() {
  return { token, isLoggedIn, userStatus, setAuth, logout }
}