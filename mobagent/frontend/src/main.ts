import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import './styles/main.css'

const tg = window.Telegram?.WebApp
tg?.ready()
tg?.expand()
tg?.setHeaderColor('#0a0e17')
tg?.setBackgroundColor('#0a0e17')

createApp(App).use(createPinia()).use(router).mount('#app')

declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        ready: () => void
        expand: () => void
        setHeaderColor: (c: string) => void
        setBackgroundColor: (c: string) => void
        HapticFeedback?: { impactOccurred: (s: string) => void }
        showAlert?: (msg: string) => void
      }
    }
  }
}
