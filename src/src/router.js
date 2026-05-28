import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import Dashboard from './views/Dashboard.vue'
import Example from './views/Example.vue'
import Docs from './views/Docs.vue'
import Source from './views/Source.vue'
import { useAuth } from './composables/useAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Home },
    { path: '/dashboard', component: Dashboard, meta: { auth: true } },
    { path: '/example', component: Example },
    { path: '/docs', component: Docs },
    { path: '/source', component: Source },
  ],
})

router.beforeEach((to) => {
  if (to.meta.auth) {
    const { isLoggedIn } = useAuth()
    if (!isLoggedIn.value) return '/'
  }
})

export default router