import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

const apiTarget = process.env.VITE_PROXY_TARGET ?? 'http://localhost:8089'

export default defineConfig({
    plugins: [
        vue(),
        tailwindcss(),
        AutoImport({
            resolvers: [ElementPlusResolver()],
        }),
        Components({
            resolvers: [ElementPlusResolver()],
        }),
    ],
    css: {
        preprocessorOptions: {
            scss: { api: 'modern-compiler' },
        },
    },
    // Pre-bundle Element Plus component styles so dev doesn't reload on first visit to each page
    optimizeDeps: {
        include: [
            'element-plus/es',
            'element-plus/es/components/*/style/css',
            '@element-plus/icons-vue',
            'vue-i18n',
        ],
    },
    server: {
        port: 5175,
        proxy: {
            '/api': apiTarget,
            '/uploads': apiTarget,
        },
    },
})
