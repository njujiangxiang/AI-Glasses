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
        <el-select v-model="filters.org_code" clearable placeholder="全部单位" style="width: 160px">
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
        <el-select v-model="filters.owner" clearable placeholder="全部人员" style="width: 160px">
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
          <el-button link type="primary" @click="viewSheet(scope.row)">{{ scope.row.code }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="任务名称" min-width="220" />
      <el-table-column prop="org_name" label="所属单位" width="130" />
      <el-table-column prop="plan_date" label="计划日期" width="120" />
      <el-table-column prop="owner_name" label="负责人" width="120" />
      <el-table-column prop="detail_count" label="明细数" width="90" />
      <el-table-column label="状态" width="110">
        <template #default="scope"><el-tag :type="statusType(scope.row.status)">{{ statusLabel(scope.row.status) }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="160" />
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

type TaskSheet = {
  id: number
  code: string
  name: string
  org_code: string
  org_name: string
  plan_date: string
  owner_name: string
  detail_count: number
  status: 'draft' | 'submitted' | 'completed' | 'voided'
  updated_at: string
}

const router = useRouter()
const filters = reactive({ keyword: '', org_code: '', status: '', owner: '' })
const sheets = reactive<TaskSheet[]>([
  { id: 1, code: 'TASK-20260614-001', name: 'A区变电站 AI眼镜巡检作业', org_code: 'ROOT', org_name: '默认单位', plan_date: '2026-06-15', owner_name: '巡检班组长', detail_count: 3, status: 'draft', updated_at: '2026-06-14 14:32' },
  { id: 2, code: 'TASK-20260613-009', name: 'B区电缆夹层复检', org_code: 'ROOT', org_name: '默认单位', plan_date: '2026-06-14', owner_name: '巡检员', detail_count: 5, status: 'submitted', updated_at: '2026-06-13 18:20' },
  { id: 3, code: 'TASK-20260612-006', name: '主变区域常规巡视', org_code: 'ROOT', org_name: '默认单位', plan_date: '2026-06-12', owner_name: '巡检班组长', detail_count: 4, status: 'completed', updated_at: '2026-06-12 17:10' }
])
const filteredSheets = computed(() => sheets.filter((sheet) => {
  const keywordHit = !filters.keyword || sheet.code.includes(filters.keyword) || sheet.name.includes(filters.keyword)
  const orgHit = !filters.org_code || sheet.org_code === filters.org_code
  const statusHit = !filters.status || sheet.status === filters.status
  const ownerHit = !filters.owner || sheet.owner_name === filters.owner
  return keywordHit && orgHit && statusHit && ownerHit
}))

// load 按当前筛选条件刷新列表。示例页使用本地数据模拟。
function load() { ElMessage.success('查询完成') }
// reset 重置筛选条件。
function reset() {
  Object.assign(filters, { keyword: '', org_code: '', status: '', owner: '' })
  load()
}
// createSheet 打开新增任务单 Tab。
function createSheet() { router.push('/tasksheets/create') }
// editSheet 打开指定任务单编辑页。
function editSheet(row: TaskSheet) { router.push(`/tasksheets/${row.id}/edit`) }
// viewSheet 打开指定任务单查看页。
function viewSheet(row: TaskSheet) { router.push(`/tasksheets/${row.id}/view`) }
// copySheet 从当前任务单复制生成新草稿。
function copySheet(row: TaskSheet) { router.push(`/tasksheets/create?copy_from=${row.id}`) }
// removeSheet 删除草稿任务单。
async function removeSheet(row: TaskSheet) {
  await ElMessageBox.confirm(`确定删除 ${row.code} 吗？`, '提示', { type: 'warning' })
  ElMessage.success('任务单已删除')
}
// withdrawSheet 撤回已提交任务单。
async function withdrawSheet(row: TaskSheet) {
  await ElMessageBox.confirm(`确定撤回 ${row.code} 吗？`, '提示', { type: 'warning' })
  ElMessage.success('任务单已撤回')
}
function statusLabel(status: TaskSheet['status']) { return ({ draft: '草稿', submitted: '已提交', completed: '已完成', voided: '已作废' }[status]) }
function statusType(status: TaskSheet['status']) { return status === 'submitted' ? 'success' : status === 'completed' ? 'info' : status === 'voided' ? 'danger' : 'warning' }
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 12px 0; background: #f8fbf9; border-radius: 4px; }
.pager { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
