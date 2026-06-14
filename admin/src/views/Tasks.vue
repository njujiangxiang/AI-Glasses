<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">任务管理</span>
      <el-space>
        <el-select v-model="status" placeholder="状态" clearable style="width: 180px" @change="load">
          <el-option label="待领取" value="pending" />
          <el-option label="已分配" value="assigned" />
          <el-option label="执行中" value="in_progress" />
          <el-option label="已提交" value="submitted" />
          <el-option label="已完成" value="completed" />
          <el-option label="已逾期" value="overdue" />
        </el-select>
        <el-button @click="load">刷新</el-button>
      </el-space>
    </div>
    <el-table :data="tasks" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="status" label="状态" width="120" />
      <el-table-column prop="point_name" label="点位" />
      <el-table-column prop="equipment_name" label="设备" />
      <el-table-column prop="due_at" label="截止时间" />
      <el-table-column prop="executor_id" label="执行人" width="100" />
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiGet } from '@/api/client'

type Task = { id: number; status: string; point_name: string; equipment_name: string; due_at: string; executor_id?: number }
const tasks = ref<Task[]>([])
const status = ref('')
// load 按当前状态筛选条件查询任务列表。
async function load() {
  const query = status.value ? `?status=${status.value}` : ''
  tasks.value = await apiGet<Task[]>(`/api/admin/tasks${query}`)
}
onMounted(load)
</script>
