import { ref } from 'vue'

const DEFAULT_BASE = import.meta.env.VITE_APP_API_BASE || ''

export function useApi() {
  const apiBase = ref(DEFAULT_BASE)

  function api(method, path, body, token) {
    const base = apiBase.value
    const opts = { method, headers: { 'Content-Type': 'application/json' } }
    if (token) opts.headers['Authorization'] = 'Bearer ' + token
    if (body) opts.body = JSON.stringify(body)
    return fetch(base + path, opts).then(async r => {
      const data = await r.json()
      if (!r.ok) throw new Error(data.error || r.statusText)
      return data
    })
  }

  return { apiBase, api, DEFAULT_BASE }
}

export { DEFAULT_BASE }