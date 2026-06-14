<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">巡检模板</span>
      <el-button type="primary" @click="createTemplate">快速创建机房日巡检模板</el-button>
    </div>
    <el-table :data="templates" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="模板名称" />
      <el-table-column prop="applicable_roles" label="适用角色" />
      <el-table-column prop="enabled" label="启用" width="100" />
      <el-table-column prop="updated_at" label="更新时间" />
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type Template = { id: number; name: string; applicable_roles: string; enabled: boolean; updated_at: string }
const templates = ref<Template[]>([])

// load 查询巡检模板列表并刷新表格。
async function load() {
  templates.value = await apiGet<Template[]>('/api/admin/templates')
}

// createTemplate 创建一套演示巡检模板，便于快速验证任务计划流程。
async function createTemplate() {
  await apiPost('/api/admin/templates', {
    name: '机房日巡检',
    description: '包含到场、设备面板、温度指示灯和异常确认节点',
    applicable_roles: '巡检员,班组长',
    enabled: true,
    nodes: [
      { name: '到达 A 区设备柜', node_type: 'checkin', min_photos: 0, require_text: false, allow_abnormal: false, require_live_capture: false },
      { name: '拍摄设备面板状态', node_type: 'photo', min_photos: 1, require_text: false, allow_abnormal: true, require_live_capture: true },
      { name: '记录温度与指示灯', node_type: 'text', min_photos: 1, require_text: true, allow_abnormal: true, require_live_capture: true },
      { name: '确认现场恢复正常', node_type: 'confirm', min_photos: 1, require_text: false, allow_abnormal: true, require_live_capture: true }
    ]
  })
  ElMessage.success('模板已创建')
  await load()
}

onMounted(load)
</script>
