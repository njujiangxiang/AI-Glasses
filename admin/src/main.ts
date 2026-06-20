// 后台前端入口文件。它挂载 Vue 3 应用，并注册 Element Plus、全局项目样式和管理端路由。
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import './styles/index.css'
import App from './App.vue'
import router from './router'

createApp(App)
  .use(router)
  .use(ElementPlus, {
    locale: zhCn,
  })
  .mount('#app')
