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
            <el-avatar v-if="currentUserAvatarBlobUrl" :size="32" :src="currentUserAvatarBlobUrl" />
            <el-avatar v-else :size="32">{{ currentUserInitial }}</el-avatar>
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
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowDown,
  Bell,
  Calendar,
  Collection,
  DataAnalysis,
  Document,
  Expand,
  Fold,
  Key,
  Monitor,
  OfficeBuilding,
  Operation,
  Setting,
  Tickets,
  User,
  UserFilled,
  Menu
} from '@element-plus/icons-vue'
import { clearToken, getCurrentUser, apiGet } from '@/api/client'
import { getIconComponent } from '@/iconCatalog'

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
// menuItems 是侧栏菜单和可打开 Tab 的唯一来源 - 从后端动态加载
const menuItems = ref<MenuItem[]>([])

// loadMenus 从后端加载当前用户的权限菜单
async function loadMenus() {
  try {
    console.log('开始加载用户菜单...')
    const menus = await apiGet<any[]>('/api/admin/menus/mine')
    console.log('从后端获取的菜单:', menus)
    // 如果后端没有返回菜单（空表），使用默认菜单
    if (!menus || menus.length === 0) {
      console.warn('后端未返回菜单，使用默认菜单')
      menuItems.value = [
        { path: '/workbench', title: '工作台', icon: DataAnalysis },
        { path: '/templates', title: '巡检模板', icon: Document },
        { path: '/workflows', title: '工作流管理', icon: Operation },
        { path: '/plans', title: '任务计划', icon: Calendar },
        { path: '/tasks', title: '任务管理', icon: Tickets },
        { path: '/tasksheets', title: '作业任务单', icon: Document },
        { path: '/defects', title: '缺陷管理', icon: Bell },
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
            { path: '/roles', title: '角色管理', icon: UserFilled },
            { path: '/menus', title: '菜单权限', icon: Menu },
            { path: '/business-codes', title: '业务编码配置', icon: Key },
            { path: '/monitoring/logs', title: '实时监控', icon: Monitor }
          ]
        }
      ]
      return
    }
    // 转换后端菜单格式为前端格式
    menuItems.value = menus.map(menu => transformMenu(menu))
    console.log('菜单加载完成，共加载了', menuItems.value.length, '个顶级菜单')
  } catch (e) {
    console.warn('加载菜单失败，使用默认菜单:', e)
    // 加载失败时使用默认菜单
    menuItems.value = [
      { path: '/workbench', title: '工作台', icon: DataAnalysis },
      { path: '/templates', title: '巡检模板', icon: Document },
      { path: '/workflows', title: '工作流管理', icon: Operation },
      { path: '/plans', title: '任务计划', icon: Calendar },
      { path: '/tasks', title: '任务管理', icon: Tickets },
      { path: '/tasksheets', title: '作业任务单', icon: Document },
      { path: '/defects', title: '缺陷管理', icon: Bell },
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
          { path: '/roles', title: '角色管理', icon: UserFilled },
          { path: '/menus', title: '菜单权限', icon: Menu },
          { path: '/business-codes', title: '业务编码配置', icon: Key }
        ]
      }
    ]
  }
}

function transformMenu(menu: any): MenuItem {
  const path = menu.path || `/${menu.id}`
  const item: MenuItem = {
    path,
    title: menu.name,
    icon: path === '/roles' ? UserFilled : path === '/menus' ? Menu : getIconComponent(menu.icon)
  }
  if (menu.children && menu.children.length > 0) {
    // 过滤掉按钮级权限，只保留目录和菜单
    const validChildren = menu.children.filter((c: any) => c.type !== 'A')
    if (validChildren.length > 0) {
      item.children = validChildren.map(transformMenu)
    }
  }
  return item
}
const leafMenuItems = computed(() => menuItems.value.flatMap((item) => item.children || [item]))
const currentUser = ref(getCurrentUser())

function refreshCurrentUser() {
  currentUser.value = getCurrentUser()
}

// 无论是否在布局中，都监听用户更新事件（登录成功后通知壳层刷新）
onMounted(() => window.addEventListener('admin-user-updated', refreshCurrentUser))
onUnmounted(() => window.removeEventListener('admin-user-updated', refreshCurrentUser))

// 路由变化时同步刷新 currentUser，确保头像加载时数据已就绪
watch(
  () => route.fullPath,
  () => {
    const cached = getCurrentUser()
    if (cached && (!currentUser.value || currentUser.value.id !== cached.id)) {
      currentUser.value = cached
    }
  }
)

function shouldLoadWorkspaceData() {
  return route.matched.length > 0 && !route.meta.public && !!localStorage.getItem('admin_token')
}

// 监听路由变化 - 每次进入非公开页面时检查并加载菜单
watch(
  () => route.fullPath,
  () => {
    if (shouldLoadWorkspaceData() && menuItems.value.length === 0) {
      loadMenus()
    }
  },
  { immediate: true }
)

// 监听令牌变化，登录成功后重新加载菜单
watch(
  () => localStorage.getItem('admin_token'),
  () => {
    if (shouldLoadWorkspaceData()) {
      loadMenus()
    }
  }
)

// 页面加载时获取菜单
onMounted(() => {
  if (shouldLoadWorkspaceData()) {
    loadMenus()
  }
})
const currentUserName = computed(() => currentUser.value?.name || currentUser.value?.display_name || currentUser.value?.username || 'admin')
const currentUserRole = computed(() => currentUser.value?.org_name || currentUser.value?.company_name || currentUser.value?.org_code || '未设置单位')
const currentUserInitial = computed(() => (currentUserName.value || 'A').slice(0, 1).toUpperCase())
const activeMenu = computed(() => route.path.startsWith('/tasksheets') ? '/tasksheets' : route.path)
// 右上角用户头像 - 与个人中心保持一致，支持显示上传的头像图片
const currentUserAvatarBlobUrl = ref('')

async function loadCurrentUserAvatar() {
  // 检查是否在工作区（有 token 且非公开页面）
  const token = localStorage.getItem('admin_token')
  if (!token || !currentUser.value?.id) {
    currentUserAvatarBlobUrl.value = ''
    return
  }
  // 检查路由是否匹配（排除公开页面）
  if (route.meta.public) {
    currentUserAvatarBlobUrl.value = ''
    return
  }
  try {
    const res = await fetch(`/api/admin/users/${currentUser.value.id}/avatar`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {}
    })
    if (res.ok) {
      const blob = await res.blob()
      // 确保返回的是有效图片（不是空响应）
      if (blob.size > 0) {
        currentUserAvatarBlobUrl.value = URL.createObjectURL(blob)
      } else {
        currentUserAvatarBlobUrl.value = ''
      }
    } else {
      currentUserAvatarBlobUrl.value = ''
    }
  } catch (e) {
    console.warn('Failed to load header avatar:', e)
    currentUserAvatarBlobUrl.value = ''
  }
}

// 用户变化时重新加载头像
watch(currentUser, () => loadCurrentUserAvatar(), { immediate: true })

// currentTitle 根据路由元信息生成面包屑标题。
const currentTitle = computed(() => String(route.meta.title || '工作台'))

// 监听路由变化，进入业务页面时自动创建或激活对应 Tab，并刷新头像
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
    // 路由变化时刷新头像（从个人中心返回后更新）
    loadCurrentUserAvatar()
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
    await ElMessageBox.confirm('确定退出当前账号吗？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })
    clearToken()
    await router.push('/login')
    ElMessage.success('已退出登录')
    return
  }
  if (command === 'profile') {
    // 打开个人中心Tab
    const tabPath = '/profile'
    if (!tabs.value.some(tab => tab.path === tabPath)) {
      tabs.value.push({ path: tabPath, title: '个人中心' })
    }
    activeTab.value = tabPath
    await router.push(tabPath)
    return
  }
}
</script>
