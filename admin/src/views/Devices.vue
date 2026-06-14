<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">设备管理</span>
      <el-button type="primary" @click="register">登记测试眼镜</el-button>
    </div>
    <el-table :data="devices" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="serial_no" label="序列号" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="bound_user_id" label="绑定用户" />
      <el-table-column label="操作" width="220">
        <template #default="scope">
          <el-button link type="warning" @click="revoke(scope.row.id)">撤销</el-button>
          <el-button link type="danger" @click="disableLost(scope.row.id)">标记丢失</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type Device = { id: number; serial_no: string; name: string; status: string; bound_user_id?: number }
const devices = ref<Device[]>([])
// load 查询设备列表并刷新表格。
async function load() { devices.value = await apiGet<Device[]>('/api/admin/devices') }
// register 登记一台测试智能眼镜设备。
async function register() {
  await apiPost('/api/admin/devices', { serial_no: `GLASS-${Date.now()}`, name: '测试智能眼镜' })
  ElMessage.success('设备已登记')
  await load()
}
// revoke 撤销指定设备的访问权限。
async function revoke(id: number) {
  await apiPost(`/api/admin/devices/${id}/revoke`)
  ElMessage.success('设备已撤销')
  await load()
}
// disableLost 将指定设备标记为丢失禁用。
async function disableLost(id: number) {
  await apiPost(`/api/admin/devices/${id}/disable-lost`)
  ElMessage.success('设备已标记丢失')
  await load()
}
onMounted(load)
</script>
