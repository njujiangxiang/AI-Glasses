// 后台管理端 Vue Router 配置。路由 meta.title 会驱动 App.vue 中的侧栏、面包屑和多 Tab
// 标题；meta.public 用于让登录页脱离已登录后的管理端布局单独渲染。
import { createRouter, createWebHistory } from 'vue-router'
import Workbench from '@/views/Workbench.vue'
import Templates from '@/views/Templates.vue'
import Plans from '@/views/Plans.vue'
import Tasks from '@/views/Tasks.vue'
import TaskSheets from '@/views/TaskSheets.vue'
import TaskSheetForm from '@/views/TaskSheetForm.vue'
import Defects from '@/views/Defects.vue'
import Devices from '@/views/Devices.vue'
import Organizations from '@/views/Organizations.vue'
import Users from '@/views/Users.vue'
import BusinessCodes from '@/views/BusinessCodes.vue'
import Login from '@/views/Login.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/workbench' },
    { path: '/login', component: Login, meta: { title: '登录', public: true } },
    { path: '/workbench', component: Workbench, meta: { title: '工作台' } },
    { path: '/templates', component: Templates, meta: { title: '巡检模板' } },
    { path: '/plans', component: Plans, meta: { title: '任务计划' } },
    { path: '/tasks', component: Tasks, meta: { title: '任务管理' } },
    { path: '/tasksheets', component: TaskSheets, meta: { title: '作业任务单' } },
    { path: '/tasksheets/create', component: TaskSheetForm, meta: { title: '新增任务单' }, props: { mode: 'create' } },
    { path: '/tasksheets/:id/:mode(edit|view)', component: TaskSheetForm, meta: { title: '任务单详情' } },
    { path: '/defects', component: Defects, meta: { title: '缺陷管理' } },
    { path: '/devices', component: Devices, meta: { title: '设备管理' } },
    { path: '/organizations', component: Organizations, meta: { title: '组织管理' } },
    { path: '/users', component: Users, meta: { title: '用户管理' } },
    { path: '/business-codes', component: BusinessCodes, meta: { title: '业务编码配置' } }
  ]
})

// 路由守卫：未登录用户重定向到登录页，已登录用户访问登录页重定向到工作台。
router.beforeEach((to) => {
  const token = localStorage.getItem('admin_token')
  if (!token && !to.meta.public) {
    return { path: '/login' }
  }
  if (token && to.path === '/login') {
    return { path: '/workbench' }
  }
})

export default router
