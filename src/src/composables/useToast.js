import { reactive } from 'vue'

const state = reactive({ message: '', type: 'success', showing: false })
let timer = null

export function showToast(msg, t = 'success') {
  state.message = msg
  state.type = t
  state.showing = true
  clearTimeout(timer)
  timer = setTimeout(() => { state.showing = false }, 3000)
}

export { state as toastState }