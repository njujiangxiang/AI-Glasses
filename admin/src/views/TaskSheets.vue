<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">作业任务单</span>
      <el-button type="primary" @click="createSheet">新增任务单</el-button>
    </div>

    <el-form class="filter-form" :inline="true" :model="filters">
      <el-form-item label="关键字">
        <el-input v-model="filters.keyword" clearable placeholder="编号/名称" style="width: 180px" @keyup.enter="load" />
      </el-form-item>
      <el-form-item label="所属单位">
        <el-select v-model="filters.orgCode" clearable placeholder="全部单位" style="width: 160px">
          <el-option label="默认单位" value="ROOT" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="filters.status" clearable placeholder="全部状态" style="width: 140px">
          <el-option label="草稿" value="draft" />
          <el-option label="已提交" value="submitted" />
          <el-option label="已完成" value="completed" />
          <el-option label="已作废" value="voided" />
        </el-select>
      </el-form-item>
      <el-form-item label="负责人">
        <el-select v-model="filters.ownerName" clearable placeholder="全部人员" style="width: 160px">
          <el-option label="巡检班组长" value="巡检班组长" />
          <el-option label="巡检员" value="巡检员" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="filteredSheets" stripe>
      <el-table-column prop="code" label="任务编号" min-width="160">
        <template #default="scope">
          <el-button link type="primary" @click="viewSheet(scope.row)">{{ scope.row.code || '-' }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="任务名称" min-width="220" />
      <el-table-column prop="orgName" label="所属单位" width="130" />
      <el-table-column prop="planDate" label="计划日期" width="120" />
      <el-table-column prop="ownerName" label="负责人" width="120" />
      <el-table-column prop="detailCount" label="明细数" width="90" />
      <el-table-column label="状态" width="110">
        <template #default="scope">
          <el-tag :type="statusType(scope.row.status)">{{ statusLabel(scope.row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="updatedAt" label="更新时间" width="160" />
      <el-table-column label="操作" width="250" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="viewSheet(scope.row)">查看</el-button>
          <el-button v-if="scope.row.status === 'draft'" link type="primary" @click="editSheet(scope.row)">编辑</el-button>
          <el-button link type="primary" @click="copySheet(scope.row)">复制</el-button>
          <el-button v-if="scope.row.status === 'draft'" link type="danger" @click="removeSheet(scope.row)">删除</el-button>
          <el-button v-if="scope.row.status === 'submitted'" link type="warning" @click="withdrawSheet(scope.row)">撤回</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination layout="total, prev, pager, next" :total="filteredSheets.length" :page-size="20" />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { taskStore, type TaskSheet } from '@/stores/taskStore'

const router = useRouter()

const filters = reactive({
  keyword: '',
  orgCode: '',
  status: '',
  ownerName: ''
})

const sheets = computed(() => taskStore.getAll())

const filteredSheets = computed(() => sheets.value.filter((sheet) => {
  const keywordHit = !filters.keyword ||
    (sheet.code && sheet.code.includes(filters.keyword)) ||
    (sheet.name && sheet.name.includes(filters.keyword))
  const orgHit = !filters.orgCode || sheet.orgCode === filters.orgCode
  const statusHit = !filters.status || sheet.status === filters.status
  const ownerHit = !filters.ownerName || sheet.ownerName === filters.ownerName
  return keywordHit && orgHit && statusHit && ownerHit
}))

function load() {
  ElMessage.success('查询完成')
}

function reset() {
  Object.assign(filters, { keyword: '', orgCode: '', status: '', ownerName: '' })
  load()
}

function createSheet() {
  router.push('/tasksheets/create')
}

function editSheet(row: TaskSheet) {
  router.push(`/tasksheets/${row.id}/edit`)
}

function viewSheet(row: TaskSheet) {
  router.push(`/tasksheets/${row.id}/view`)
}

function copySheet(row: TaskSheet) {
  router.push(`/tasksheets/create?copy_from=${row.id}`)
}

async function removeSheet(row: TaskSheet) {
  await ElMessageBox.confirm(`确定删除 ${row.code || '此任务单'} 吗？`, '提示', { type: 'warning' })
  taskStore.delete(row.id)
  ElMessage.success('任务单已删除')
}

async function withdrawSheet(row: TaskSheet) {
  await ElMessageBox.confirm(`确定撤回 ${row.code} 吗？`, '提示', { type: 'warning' })
  taskStore.withdraw(row.id)
  ElMessage.success('任务单已撤回')
}

function statusLabel(status: TaskSheet['status']) {
  return { draft: '草稿', submitted: '已提交', completed: '已完成', voided: '已作废' }[status]
}

function statusType(status: TaskSheet['status']) {
  if (status === 'submitted') return 'success'
  if (status === 'completed') return 'info'
  if (status === 'voided') return 'danger'
  return 'warning'
}
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 12px 0; background: #f8fbf9; border-radius: 4px; }
.pager { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
