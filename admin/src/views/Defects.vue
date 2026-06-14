<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">缺陷管理</span>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-table :data="defects" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="status" label="状态" width="140" />
      <el-table-column prop="task_id" label="任务" width="100" />
      <el-table-column prop="node_id" label="节点" width="100" />
      <el-table-column prop="description" label="异常描述" />
      <el-table-column label="操作" width="180">
        <template #default="scope">
          <el-button link type="primary" @click="confirm(scope.row.id)">确认</el-button>
          <el-button link type="danger" @click="close(scope.row.id)">关闭</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type Defect = { id: number; status: string; task_id: number; node_id: number; description: string }
const defects = ref<Defect[]>([])
// load 查询缺陷列表并刷新表格。
async function load() { defects.value = await apiGet<Defect[]>('/api/admin/defects') }
// confirm 将指定缺陷从待确认状态切换为已确认。
async function confirm(id: number) {
  await apiPost(`/api/admin/defects/${id}/confirm`)
  ElMessage.success('缺陷已确认')
  await load()
}
// close 关闭指定缺陷并写入默认关闭原因。
async function close(id: number) {
  await apiPost(`/api/admin/defects/${id}/close`, { reason: '后台关闭' })
  ElMessage.success('缺陷已关闭')
  await load()
}
onMounted(load)
</script>
