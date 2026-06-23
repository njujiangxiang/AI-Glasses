<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">巡检任务</span>
      <div>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px" @change="load">
          <el-option label="待领取" value="pending" />
          <el-option label="已分配" value="assigned" />
          <el-option label="执行中" value="in_progress" />
          <el-option label="已提交" value="submitted" />
          <el-option label="已完成" value="completed" />
          <el-option label="已逾期" value="overdue" />
          <el-option label="已取消" value="cancelled" />
        </el-select>
        <el-select v-model="filters.template_id" placeholder="模板" clearable style="width: 140px; margin-left: 8px" @change="load">
          <el-option v-for="t in templateOptions" :key="t.id" :label="t.name" :value="t.id" />
        </el-select>
        <el-input v-model="filters.keyword" placeholder="搜索任务" clearable style="width: 180px; margin-left: 8px" @keyup.enter="load" />
        <el-button @click="load" style="margin-left: 8px">刷新</el-button>
        <el-button type="primary" @click="openCreate">手动创建任务</el-button>
      </div>
    </div>
    <el-table :data="items" stripe row-key="id" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="task_name" label="任务名称" min-width="150" />
      <el-table-column label="模板" width="120">
        <template #default="scope">{{ scope.row.template_name || '-' }}</template>
      </el-table-column>
      <el-table-column prop="point_name" label="巡检点位" width="120" />
      <el-table-column prop="equipment_name" label="设备" width="120" />
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="statusType(scope.row.status)">{{ statusLabel(scope.row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="指派给" width="120">
        <template #default="scope">{{ userMap[scope.row.assignee_id] || scope.row.assignee_id || '-' }}</template>
      </el-table-column>
      <el-table-column prop="due_at" label="截止时间" width="160">
        <template #default="scope">{{ formatTime(scope.row.due_at) }}</template>
      </el-table-column>
      <el-table-column label="执行人" width="100">
        <template #default="scope">{{ scope.row.executor_id || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="viewDetail(scope.row)">详情</el-button>
          <el-button link type="primary" @click="viewResults(scope.row)">结果</el-button>
          <el-button v-if="canCancel(scope.row)" link type="warning" @click="cancelTask(scope.row)">取消</el-button>
          <el-button v-if="scope.row.status === 'submitted'" link type="success" @click="completeTask(scope.row)">完成</el-button>
          <el-button link type="danger" @click="removeTask(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager" v-if="total > filters.page_size">
      <el-pagination v-model:current-page="filters.page" :page-size="filters.page_size" :total="total" layout="total, prev, pager, next" @current-change="load" />
    </div>
  </el-card>

  <!-- 手动创建任务弹窗 -->
  <el-dialog v-model="dialogVisible" title="手动创建任务" width="640px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="巡检模板" prop="template_id">
        <el-select v-model="form.template_id" placeholder="选择模板" filterable>
          <el-option v-for="t in templateOptions" :key="t.id" :label="t.name" :value="t.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="任务名称" prop="task_name">
        <el-input v-model="form.task_name" placeholder="请输入任务名称" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="点位名称">
            <el-input v-model="form.point_name" placeholder="巡检点位名称" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="设备名称">
            <el-input v-model="form.equipment_name" placeholder="巡检设备名称" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="作业区域">
        <el-input v-model="form.inspect_area" placeholder="作业区域" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="24">
          <el-form-item label="指派给" prop="assignee_id">
            <el-select v-model="form.assignee_id" placeholder="选择用户" filterable style="width: 100%">
              <el-option v-for="u in userOptions" :key="u.id" :label="u.display_name || u.name || u.username" :value="u.id" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="截止时间" prop="due_at">
        <el-date-picker v-model="form.due_at" type="datetime" placeholder="选择截止时间" value-format="YYYY-MM-DDTHH:mm:ssZ" style="width: 100%" />
      </el-form-item>
      <el-form-item label="眼镜编号">
        <el-input v-model="form.glasses_sn" placeholder="指定AR眼镜编号（可选）" />
      </el-form-item>
      <el-form-item label="下发人">
        <el-input v-model="form.assign_user" placeholder="任务下发人" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submit" :loading="submitting">创建</el-button>
    </template>
  </el-dialog>

  <!-- 任务详情抽屉 -->
  <el-drawer v-model="detailVisible" title="任务详情" size="550px">
    <div v-if="detailData">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="任务名称">{{ detailData.task.task_name }}</el-descriptions-item>
        <el-descriptions-item label="巡检点位">{{ detailData.task.point_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="设备名称">{{ detailData.task.equipment_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="作业区域">{{ detailData.task.inspect_area || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType(detailData.task.status)">{{ statusLabel(detailData.task.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="指派给">{{ userMap[detailData.task.assignee_id] || detailData.task.assignee_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="执行人">{{ detailData.task.executor_id ? (userMap[detailData.task.executor_id] || detailData.task.executor_id) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="截止时间">{{ formatTime(detailData.task.due_at) }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ detailData.task.started_at ? formatTime(detailData.task.started_at) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ detailData.task.submitted_at ? formatTime(detailData.task.submitted_at) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ detailData.task.completed_at ? formatTime(detailData.task.completed_at) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="眼镜编号">{{ detailData.task.glasses_sn || '-' }}</el-descriptions-item>
        <el-descriptions-item label="下发人">{{ detailData.task.assign_user || '-' }}</el-descriptions-item>
      </el-descriptions>

      <h4 style="margin: 20px 0 10px">巡检节点 ({{ detailData.nodes.length }})</h4>
      <el-table :data="detailData.nodes" stripe size="small">
        <el-table-column prop="sort_order" label="序号" width="60" />
        <el-table-column prop="name" label="节点名称" />
        <el-table-column prop="node_type" label="类型" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'completed' ? 'success' : scope.row.status === 'abnormal' ? 'danger' : 'info'" size="small">
              {{ scope.row.status === 'completed' ? '已完成' : scope.row.status === 'abnormal' ? '异常' : '待执行' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </el-drawer>

  <!-- 执行结果抽屉 -->
  <el-drawer v-model="resultsVisible" title="执行结果" size="600px">
    <div v-if="resultsData">
      <h4 style="margin: 0 0 10px">节点执行结果</h4>
      <el-table :data="resultsData.nodes" stripe size="small">
        <el-table-column prop="sort_order" label="序号" width="60" />
        <el-table-column prop="name" label="节点名称" width="150" />
        <el-table-column prop="node_type" label="类型" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'completed' ? 'success' : scope.row.status === 'abnormal' ? 'danger' : 'info'" size="small">
              {{ scope.row.status === 'completed' ? '已完成' : scope.row.status === 'abnormal' ? '异常' : '待执行' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="结果" min-width="200">
          <template #default="scope">
            <template v-if="getResult(scope.row.id)">
              <div v-if="getResult(scope.row.id)?.feedback_content"><strong>反馈:</strong> {{ getResult(scope.row.id)?.feedback_content }}</div>
              <div v-if="getResult(scope.row.id)?.text_note"><strong>备注:</strong> {{ getResult(scope.row.id)?.text_note }}</div>
              <div v-if="getResult(scope.row.id)?.is_abnormal"><strong style="color: #f56c6c">异常:</strong> {{ getResult(scope.row.id)?.abnormal_desc }}</div>
            </template>
            <span v-else style="color: #909399">未提交</span>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="resultsData.defects.length > 0">
        <h4 style="margin: 20px 0 10px">缺陷记录 ({{ resultsData.defects.length }})</h4>
        <el-table :data="resultsData.defects" stripe size="small">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="description" label="缺陷描述" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag size="small">{{ scope.row.status === 'reported' ? '已上报' : scope.row.status === 'confirmed' ? '已确认' : '已关闭' }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type Task = {
  id: number
  plan_id: number | null
  template_id: number
  template_name: string
  task_name: string
  point_name: string
  equipment_name: string
  inspect_area: string
  status: string
  assignee_type: string
  assignee_id: number
  executor_id: number | null
  due_at: string
  started_at: string | null
  submitted_at: string | null
  completed_at: string | null
  glasses_sn: string
  assign_user: string
}

type TaskNode = {
  id: number
  sort_order: number
  name: string
  node_type: string
  status: string
}

type NodeResult = {
  id: number
  node_id: number
  feedback_content: string
  text_note: string
  is_abnormal: boolean
  abnormal_desc: string
}

type TemplateOption = { id: number; name: string }
type UserOption = { id: number; username: string; name: string; display_name: string }

const items = ref<Task[]>([])
const total = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const filters = reactive({ keyword: '', status: '', template_id: undefined as number | undefined, page: 1, page_size: 20 })
const templateOptions = ref<TemplateOption[]>([])
const userOptions = ref<UserOption[]>([])
const userMap = ref<Record<number, string>>({})

const form = reactive({
  template_id: undefined as number | undefined,
  task_name: '',
  point_name: '',
  equipment_name: '',
  inspect_area: '',
  assignee_type: 'user',
  assignee_id: undefined as number | undefined,
  due_at: '',
  glasses_sn: '',
  assign_user: ''
})

const rules: FormRules = {
  template_id: [{ required: true, message: '请选择巡检模板', trigger: 'change' }],
  task_name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  assignee_id: [{ required: true, message: '请选择指派用户', trigger: 'change' }],
  due_at: [{ required: true, message: '请选择截止时间', trigger: 'change' }]
}

// 详情
const detailVisible = ref(false)
const detailData = ref<{ task: Task; nodes: TaskNode[]; results: NodeResult[]; attachments: any[]; defects: any[] } | null>(null)

// 结果
const resultsVisible = ref(false)
const resultsData = ref<{ nodes: TaskNode[]; results: NodeResult[]; attachments: any[]; defects: any[] } | null>(null)

function statusLabel(status: string) {
  const map: Record<string, string> = {
    pending: '待领取', assigned: '已分配', in_progress: '执行中',
    submitted: '已提交', completed: '已完成', overdue: '已逾期', cancelled: '已取消'
  }
  return map[status] || status
}

function statusType(status: string) {
  const map: Record<string, string> = {
    pending: 'warning', assigned: '', in_progress: '',
    submitted: 'info', completed: 'success', overdue: 'danger', cancelled: 'info'
  }
  return map[status] || ''
}

function canCancel(task: Task) {
  return ['pending', 'assigned', 'in_progress'].includes(task.status)
}

function formatTime(time: string | null) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

function getResult(nodeId: number): NodeResult | undefined {
  return resultsData.value?.results.find(r => r.node_id === nodeId)
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.keyword) params.set('keyword', filters.keyword)
    if (filters.status) params.set('status', filters.status)
    if (filters.template_id) params.set('template_id', String(filters.template_id))
    params.set('page', String(filters.page))
    params.set('page_size', String(filters.page_size))
    const result = await apiGet<{ items: Task[]; total: number }>(`/api/admin/tasks?${params.toString()}`)
    items.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

async function loadTemplates() {
  templateOptions.value = await apiGet<TemplateOption[]>('/api/admin/templates/all')
}

async function loadUsers() {
  const users = await apiGet<UserOption[]>('/api/admin/users/all')
  userOptions.value = users
  userMap.value = Object.fromEntries(users.map(u => [u.id, u.display_name || u.name || u.username]))
}

function openCreate() {
  Object.assign(form, {
    template_id: undefined, task_name: '', point_name: '', equipment_name: '',
    inspect_area: '', assignee_type: 'user', assignee_id: undefined,
    due_at: '', glasses_sn: '', assign_user: ''
  })
  loadTemplates()
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    await apiPost('/api/admin/tasks', form)
    ElMessage.success('任务已创建')
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function viewDetail(row: Task) {
  detailData.value = await apiGet(`/api/admin/tasks/${row.id}`)
  detailVisible.value = true
}

async function viewResults(row: Task) {
  resultsData.value = await apiGet(`/api/admin/tasks/${row.id}/results`)
  resultsVisible.value = true
}

async function cancelTask(row: Task) {
  await ElMessageBox.confirm(`确定取消任务"${row.task_name}"吗？`, '提示', { type: 'warning' })
  await apiPost(`/api/admin/tasks/${row.id}/cancel`)
  ElMessage.success('任务已取消')
  await load()
}

async function completeTask(row: Task) {
  await ElMessageBox.confirm(`确定完成任务"${row.task_name}"吗？`, '提示', { type: 'info' })
  await apiPost(`/api/admin/tasks/${row.id}/complete`)
  ElMessage.success('任务已完成')
  await load()
}

async function removeTask(row: Task) {
  await ElMessageBox.confirm(`确定删除任务"${row.task_name}"吗？删除后不可恢复。`, '提示', { type: 'warning' })
  await apiPost(`/api/admin/tasks/${row.id}/delete`)
  ElMessage.success('任务已删除')
  await load()
}

onMounted(() => { load(); loadTemplates(); loadUsers() })
</script>
