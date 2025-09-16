import { createApp } from 'vue'
import App from './App.vue'
import './style.css';
import router from './router/index.js';
import { i18n } from './i18n/i18n.js';

const app = createApp(App)
app.use(router)
app.use(i18n)
app.mount('#app')
