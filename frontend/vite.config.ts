import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss()
  ],
  css: { devSourcemap: true },
  define: { __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: false }
})
