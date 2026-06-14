<!--
  巡检管理后台应用壳层。该组件负责区分公开页面和已登录工作区，渲染 likeadmin 风格的
  左侧可收缩导航、多 Tab 页面导航，并提供右上角用户菜单。
-->
<template>
  <router-view v-if="route.meta.public" />
  <div v-else class="layout-default" :class="{ 'is-collapse': isCollapse }">
    <aside class="app-aside">
      <div class="app-logo">
        <el-icon><Monitor /></el-icon>
        <span v-show="!isCollapse">智能眼镜巡检</span>
      </div>
      <el-menu router :collapse="isCollapse" :default-active="route.path" class="app-menu">
        <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
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
            admin
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
          <el-tab-pane v-for="tab in tabs" :key="tab.path" :label="tab.title" :name="tab.path">
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
  DataAnalysis,
  Document,
  Expand,
  Fold,
  Monitor,
  Setting,
  Tickets
} from '@element-plus/icons-vue'
import { clearToken } from '@/api/client'

// route 和 router 用于同步当前路由、菜单选中态和多 Tab 导航。
const route = useRoute()
const router = useRouter()
// isCollapse 控制左侧菜单是否收缩为仅图标模式。
const isCollapse = ref(false)
// activeTab 保存当前选中的页面 Tab 路径。
const activeTab = ref('')
// tabs 保存用户已经打开的业务页面。
const tabs = ref<{ path: string; title: string }[]>([])
// menuItems 是侧栏菜单和可打开 Tab 的唯一来源。
const menuItems = [
  { path: '/workbench', title: '工作台', icon: DataAnalysis },
  { path: '/templates', title: '巡检模板', icon: Document },
  { path: '/plans', title: '任务计划', icon: Calendar },
  { path: '/tasks', title: '任务管理', icon: Tickets },
  { path: '/defects', title: '缺陷管理', icon: Bell },
  { path: '/devices', title: '设备管理', icon: Setting }
]

// currentTitle 根据路由元信息生成面包屑标题。
const currentTitle = computed(() => String(route.meta.title || '工作台'))

// 监听路由变化，进入业务页面时自动创建或激活对应 Tab。
watch(
  () => route.fullPath,
  () => {
    if (route.meta.public) return
    const item = menuItems.find((menu) => menu.path === route.path)
    if (!item) return
    activeTab.value = item.path
    if (!tabs.value.some((tab) => tab.path === item.path)) {
      tabs.value.push({ path: item.path, title: item.title })
    }
  },
  { immediate: true, flush: 'post' }
)

// changeTab 在用户点击 Tab 时切换到对应路由。
function changeTab(name: string | number) {
  if (String(name) !== route.path) router.push(String(name))
}

// removeTab 关闭指定 Tab，并在关闭当前页时切换到相邻页面。
function removeTab(name: string | number) {
  const path = String(name)
  if (tabs.value.length === 1) return
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
