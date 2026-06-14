// 后台管理端 Vue Router 配置。路由 meta.title 会驱动 App.vue 中的侧栏、面包屑和多 Tab
// 标题；meta.public 用于让登录页脱离已登录后的管理端布局单独渲染。
import { createRouter, createWebHistory } from 'vue-router'
import Workbench from '@/views/Workbench.vue'
import Templates from '@/views/Templates.vue'
import Plans from '@/views/Plans.vue'
import Tasks from '@/views/Tasks.vue'
import Defects from '@/views/Defects.vue'
import Devices from '@/views/Devices.vue'
import Organizations from '@/views/Organizations.vue'
import Users from '@/views/Users.vue'
import Login from '@/views/Login.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/workbench' },
    { path: '/login', component: Login, meta: { title: '登录', public: true } },
    { path: '/workbench', component: Workbench, meta: { title: '工作台' } },
    { path: '/templates', component: Templates, meta: { title: '巡检模板' } },
    { path: '/plans', component: Plans, meta: { title: '任务计划' } },
    { path: '/tasks', component: Tasks, meta: { title: '任务管理' } },
    { path: '/defects', component: Defects, meta: { title: '缺陷管理' } },
    { path: '/devices', component: Devices, meta: { title: '设备管理' } },
    { path: '/organizations', component: Organizations, meta: { title: '组织管理' } },
    { path: '/users', component: Users, meta: { title: '用户管理' } }
  ]
})
