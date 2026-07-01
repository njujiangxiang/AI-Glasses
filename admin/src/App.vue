<!--
  巡检管理后台应用壳层。该组件负责区分公开页面和已登录工作区，渲染 likeadmin 风格的
  左侧可收缩导航、多 Tab 页面导航，并提供右上角用户菜单。
-->
<template>
  <router-view v-if="route.meta.public" />
  <div v-else class="layout-default" :class="{ 'is-collapse': isCollapse }">
    <aside class="app-aside">
      <div class="app-logo">
        <svg class="app-logo-icon" viewBox="0 0 32 18" aria-hidden="true">
          <path d="M3 9.5c0-3.6 2.4-6.5 6-6.5h3.6c1.4 0 2.6.8 3.4 2 .8-1.2 2-2 3.4-2H23c3.6 0 6 2.9 6 6.5S26.6 16 23 16h-3.2c-2.3 0-4-1.8-4-4h-.6c0 2.2-1.7 4-4 4H9c-3.6 0-6-2.9-6-6.5Z" />
          <path d="M15.2 8.2h1.6" />
        </svg>
        <span v-show="!isCollapse">智镜任务中台</span>
      </div>
      <el-menu router :collapse="isCollapse" :default-active="activeMenu" class="app-menu">
        <template v-for="item in menuItems" :key="item.path">
          <el-sub-menu v-if="item.children" :index="item.path">
            <template #title>
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.title }}</span>
            </template>
            <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
              <el-icon><component :is="child.icon" /></el-icon>
              <template #title>{{ child.title }}</template>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </aside>
    <section class="app-body">
      <header class="app-header">
        <div class="header-left">
          <el-button text class="collapse-button" @click="isCollapse = !isCollapse">
            <el-icon><Fold v-if="!isCollapse" /><Expand v-else /></el-icon>
          </el-button>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <el-dropdown trigger="click" @command="handleUserCommand">
          <span class="user-dropdown">
            <el-avatar :size="32" :src="currentUserAvatarUrl">{{ currentUserInitial }}</el-avatar>
            <span class="user-meta">
              <span class="user-name">{{ currentUserName }}</span>
              <span class="user-role">{{ currentUserRole }}</span>
            </span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人中心</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>
      <div class="app-tabs">
        <el-tabs v-model="activeTab" type="card" closable @tab-change="changeTab" @tab-remove="removeTab">
          <el-tab-pane v-for="tab in tabs" :key="tab.path" :label="tab.title" :name="tab.path" :closable="tab.path !== '/workbench'">
            <template #label>
              <span class="tab-label">{{ tab.title }}</span>
            </template>
          </el-tab-pane>
        </el-tabs>
      </div>
      <main class="app-main">
        <router-view />
      </main>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowDown,
  Bell,
  Calendar,
  Collection,
  DataAnalysis,
  DataBoard,
  Document,
  Expand,
  Fold,
  Key,
  MapLocation,
  Monitor,
  OfficeBuilding,
  Operation,
  Setting,
  User
} from '@element-plus/icons-vue'
import { clearToken, getCurrentUser } from '@/api/client'

// route 和 router 用于同步当前路由、菜单选中态和多 Tab 导航。
const route = useRoute()
const router = useRouter()
// isCollapse 控制左侧菜单是否收缩为仅图标模式。
const isCollapse = ref(false)
// activeTab 保存当前选中的页面 Tab 路径。
const activeTab = ref('')
// tabs 保存用户已经打开的业务页面，工作台固定在首位且不可关闭。
const tabs = ref<{ path: string; title: string }[]>([{ path: '/workbench', title: '工作台' }])
type MenuItem = { path: string; title: string; icon: unknown; children?: MenuItem[] }
// menuItems 是侧栏菜单和可打开 Tab 的唯一来源。
const menuItems: MenuItem[] = [
  { path: '/workbench', title: '工作台', icon: DataAnalysis },
  { path: '/inspection-points', title: '点位管理', icon: MapLocation },
  { path: '/templates', title: '任务模板', icon: Document },
  { path: '/plans', title: '任务计划', icon: Calendar },
  { path: '/tasks', title: '任务管理', icon: Monitor },
  { path: '/defects', title: '缺陷管理', icon: Bell },
  { path: '/reports', title: '巡检报告', icon: DataBoard },
  {
    path: '/master-data',
    title: '台账和主数据管理',
    icon: Collection,
    children: [
      { path: '/devices', title: '设备管理', icon: Monitor }
    ]
  },
  {
    path: '/system',
    title: '系统管理',
    icon: Setting,
    children: [
      { path: '/organizations', title: '组织管理', icon: OfficeBuilding },
      { path: '/users', title: '用户管理', icon: User },
      { path: '/business-codes', title: '业务编码配置', icon: Key }
    ]
  }
]
const leafMenuItems = computed(() => menuItems.flatMap((item) => item.children || [item]))
const currentUser = computed(() => getCurrentUser())
const currentUserName = computed(() => currentUser.value?.name || currentUser.value?.display_name || currentUser.value?.username || 'admin')
const currentUserRole = computed(() => currentUser.value?.company_name || '未设置公司')
const currentUserInitial = computed(() => currentUserName.value.slice(0, 1))
const currentUserAvatarUrl = computed(() => currentUser.value?.avatar_size ? `/api/admin/users/${currentUser.value.id}/avatar` : '')
const activeMenu = computed(() => route.path.startsWith('/tasksheets') ? '/tasksheets' : route.path)

// currentTitle 根据路由元信息生成面包屑标题。
const currentTitle = computed(() => String(route.meta.title || '工作台'))

// 监听路由变化，进入业务页面时自动创建或激活对应 Tab。
watch(
  () => route.fullPath,
  () => {
    if (route.meta.public) return
    const tab = tabForRoute()
    if (!tab) return
    activeTab.value = tab.path
    if (!tabs.value.some((item) => item.path === tab.path)) {
      tabs.value.push(tab)
    }
  },
  { immediate: true, flush: 'post' }
)

// tabForRoute 根据当前路由生成可打开的业务 Tab，动态表单路由按单据模式命名。
function tabForRoute() {
  if (route.path.startsWith('/tasksheets/') && route.path !== '/tasksheets') {
    const mode = String(route.params.mode || 'create')
    const prefix = mode === 'view' ? '查看' : mode === 'edit' ? '编辑' : '新增'
    const code = route.params.id ? `TASK-${String(route.params.id).padStart(3, '0')}` : '任务单'
    return { path: route.fullPath, title: mode === 'create' ? '新增任务单' : `${prefix}：${code}` }
  }
  const item = leafMenuItems.value.find((menu) => menu.path === route.path)
  return item ? { path: item.path, title: item.title } : null
}

// changeTab 在用户点击 Tab 时切换到对应路由。
function changeTab(name: string | number) {
  if (String(name) !== route.path) router.push(String(name))
}

// removeTab 关闭指定 Tab，并在关闭当前页时切换到相邻页面。
function removeTab(name: string | number) {
  const path = String(name)
  if (path === '/workbench' || tabs.value.length === 1) return
  const index = tabs.value.findIndex((tab) => tab.path === path)
  tabs.value = tabs.value.filter((tab) => tab.path !== path)
  if (path === route.path) {
    const next = tabs.value[index] || tabs.value[index - 1] || tabs.value[0]
    router.push(next.path)
  }
}

// handleUserCommand 处理右上角用户菜单命令，包括退出登录。
async function handleUserCommand(command: string) {
  if (command === 'logout') {
    await ElMessageBox.confirm('确定退出当前账号吗？', '提示', { type: 'warning' })
    clearToken()
    await router.push('/login')
    ElMessage.success('已退出登录')
    return
  }
  ElMessage.info('功能建设中')
}
</script>
