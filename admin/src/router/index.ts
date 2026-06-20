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
import Roles from '@/views/Roles.vue'
import Menus from '@/views/Menus.vue'
import BusinessCodes from '@/views/BusinessCodes.vue'
import RealtimeMonitor from '@/views/RealtimeMonitor.vue'
import Workflows from '@/views/Workflows.vue'
import WorkflowEditor from '@/views/WorkflowEditor.vue'
import Profile from '@/views/Profile.vue'
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
    { path: '/roles', component: Roles, meta: { title: '角色管理' } },
    { path: '/menus', component: Menus, meta: { title: '菜单权限' } },
    { path: '/business-codes', component: BusinessCodes, meta: { title: '业务编码配置' } },
    { path: '/monitoring/logs', component: RealtimeMonitor, meta: { title: '实时监控' } },
    { path: '/workflows', component: Workflows, meta: { title: '工作流管理' } },
    { path: '/workflows/:id', component: WorkflowEditor, meta: { title: '编辑工作流' } },
    { path: '/profile', component: Profile, meta: { title: '个人中心' } }
  ]
})

// 路由守卫：未登录用户重定向到登录页。不要仅凭本地 token 把登录页重定向到工作台，
// 否则残留的失效 token 会先进入业务页，再被 401 拉回登录页，造成刷新/跳转循环。
router.beforeEach((to) => {
  const token = localStorage.getItem('admin_token')
  if (!token && !to.meta.public) {
    return { path: '/login' }
  }
})

export default router
