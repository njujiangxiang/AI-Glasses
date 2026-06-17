<template>
  <div class="workflows-page">
    <div class="page-header">
      <span class="page-title">工作流管理</span>
      <el-button type="primary" @click="createWorkflow">+ 新建工作流</el-button>
    </div>

    <el-card shadow="never">
      <template #header>
        <div class="filter-bar">
          <el-input
            v-model="keyword"
            placeholder="搜索工作流名称"
            style="width: 280px"
            clearable
            @input="load"
          >
            <template #prefix>🔍</template>
          </el-input>
          <el-select
            v-model="statusFilter"
            placeholder="状态筛选"
            style="width: 140px"
            clearable
            @change="load"
          >
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
            <el-option label="已归档" value="archived" />
          </el-select>
        </div>
      </template>

      <el-empty v-if="workflows.length === 0 && !loading" description="暂无工作流，点击右上角创建第一个工作流" />

      <el-table v-else :data="workflows" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="工作流名称" min-width="200">
          <template #default="{ row }">
            <div style="font-weight: 500; cursor: pointer; color: #1890ff" @click="editWorkflow(row.id)">
              {{ row.name }}
            </div>
            <div style="font-size: 12px; color: #8c8c8c; margin-top: 4px">{{ row.description }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_by" label="创建人" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column prop="updated_at" label="最近更新" width="160" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editWorkflow(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" link type="success" size="small" @click="publish(row.id)">发布</el-button>
            <el-button v-else link size="small" @click="unpublish(row.id)">取消发布</el-button>
            <el-button link type="danger" size="small" @click="deleteWorkflow(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

const router = useRouter()
const keyword = ref('')
const statusFilter = ref('')
const loading = ref(false)

type Workflow = {
  id: number
  name: string
  description: string
  status: string
  created_by: number
  created_at: string
  updated_at: string
}

const workflows = ref<Workflow[]>([])

async function load() {
  loading.value = true
  try {
    let url = '/api/admin/workflows'
    const params = new URLSearchParams()
    if (keyword.value) params.append('keyword', keyword.value)
    if (statusFilter.value) params.append('status', statusFilter.value)
    if (params.toString()) url += '?' + params.toString()
    workflows.value = await apiGet<Workflow[]>(url)
  } finally {
    loading.value = false
  }
}

function createWorkflow() {
  router.push('/workflows/new')
}

function editWorkflow(id: number) {
  router.push(`/workflows/${id}`)
}

async function publish(id: number) {
  try {
    await ElMessageBox.confirm('发布后工作流将可以被巡检任务使用，确认发布吗？', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await apiPost(`/api/admin/workflows/${id}/publish`)
    ElMessage.success('发布成功')
    await load()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(String(err))
  }
}

async function unpublish(id: number) {
  try {
    await ElMessageBox.confirm('取消发布后工作流将不能被新任务使用，确认取消吗？', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await apiPost(`/api/admin/workflows/${id}/unpublish`)
    ElMessage.success('已取消发布')
    await load()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(String(err))
  }
}

async function deleteWorkflow(id: number) {
  try {
    await ElMessageBox.confirm('删除后无法恢复，确认删除吗？', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await apiPost(`/api/admin/workflows/${id}/delete`)
    ElMessage.success('删除成功')
    await load()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(String(err))
  }
}

function statusLabel(status: string) {
  const map: Record<string, string> = { draft: '草稿', published: '已发布', archived: '已归档' }
  return map[status] || status
}

function statusTagType(status: string) {
  const map: Record<string, string> = { draft: 'info', published: 'success', archived: 'warning' }
  return map[status] || ''
}

onMounted(load)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #262626;
}

.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
}
</style>
