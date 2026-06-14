<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">任务计划</span>
      <el-button type="primary" @click="createPlan">创建每日 08:00 计划</el-button>
    </div>
    <el-table :data="plans" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="计划名称" />
      <el-table-column prop="cron_expr" label="Cron" />
      <el-table-column prop="timezone" label="时区" />
      <el-table-column prop="point_name" label="点位" />
      <el-table-column prop="equipment_name" label="设备" />
      <el-table-column label="操作" width="120">
        <template #default="scope">
          <el-button link type="primary" @click="enable(scope.row.id)">启用</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type Plan = { id: number; name: string; cron_expr: string; timezone: string; point_name: string; equipment_name: string }
const plans = ref<Plan[]>([])

// load 查询任务计划列表并刷新表格。
async function load() { plans.value = await apiGet<Plan[]>('/api/admin/plans') }
// createPlan 基于演示模板创建每日巡检计划。
async function createPlan() {
  await apiPost('/api/admin/plans', {
    template_id: 1,
    name: '机房日巡检计划',
    cron_expr: '0 8 * * *',
    timezone: 'Asia/Shanghai',
    start_at: new Date().toISOString(),
    due_duration_minutes: 90,
    assignee_type: 'team',
    assignee_id: 1,
    point_name: 'A 区设备柜',
    equipment_name: '核心交换机柜 A-102'
  })
  ElMessage.success('计划已创建')
  await load()
}
// enable 启用指定任务计划，让调度器可以生成巡检任务。
async function enable(id: number) {
  await apiPost(`/api/admin/plans/${id}/enable`)
  ElMessage.success('计划已启用')
  await load()
}

onMounted(load)
</script>
