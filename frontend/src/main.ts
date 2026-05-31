import { createApp } from 'vue'

import App from './App.vue'
import { pinia } from './app/pinia'
import { setUnauthorizedHandler } from './lib/api'
import { queryClient, VueQueryPlugin } from './lib/query-client'
import router from './router'
import { useAuthStore } from './stores/auth'
import './style.css'

const app = createApp(App)

app.use(pinia)
app.use(VueQueryPlugin, { queryClient })
app.use(router)

const authStore = useAuthStore(pinia)
setUnauthorizedHandler(async () => {
  authStore.reset()
  const current = router.currentRoute.value
  if (current.name !== 'login') {
    await router.push({
      name: 'login',
      query: current.fullPath ? { redirect: current.fullPath } : undefined,
    })
  }
})

app.mount('#app')
