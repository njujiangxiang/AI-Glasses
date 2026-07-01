<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">任务计划</span>
      <div>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px" @change="load">
          <el-option label="已启用" value="enabled" />
          <el-option label="已停用" value="disabled" />
        </el-select>
        <el-input v-model="filters.keyword" placeholder="搜索计划" clearable style="width: 200px; margin-left: 8px" @keyup.enter="load" />
        <el-button @click="load" style="margin-left: 8px">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增计划</el-button>
      </div>
    </div>
    <el-table :data="items" stripe row-key="id" v-loading="loading">
      <el-table-column type="index" label="序号" width="70" align="center" :index="indexMethod" />
      <el-table-column prop="name" label="计划名称" min-width="150" />
      <el-table-column label="模板" width="120">
        <template #default="scope">{{ templateMap[scope.row.template_id] || scope.row.template_id }}</template>
      </el-table-column>
      <el-table-column prop="plan_type" label="计划类型" width="100" />
      <el-table-column label="调度时间" width="160">
        <template #default="scope">
          <span>{{ formatSchedule(scope.row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="巡检点位" width="140">
        <template #default="scope">
          <el-tag v-if="scope.row.point_name" size="small">{{ scope.row.point_name }}</el-tag>
          <span v-else style="color: #999">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="equipment_name" label="设备" width="120" />
      <el-table-column label="指派给" width="120">
        <template #default="scope">{{ userMap[scope.row.assignee_id] || scope.row.assignee_id || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? '已启用' : '已停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button link :type="scope.row.enabled ? 'warning' : 'success'" @click="toggleEnable(scope.row)">
            {{ scope.row.enabled ? '停用' : '启用' }}
          </el-button>
          <el-button link type="primary" @click="generateNow(scope.row)">立即生成</el-button>
          <el-button link type="danger" @click="remove(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager" v-if="total > filters.page_size">
      <el-pagination v-model:current-page="filters.page" :page-size="filters.page_size" :total="total" layout="total, prev, pager, next" @current-change="load" />
    </div>
  </el-card>

  <!-- 计划编辑弹窗 -->
  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑计划' : '新增计划'" width="720px" top="3vh">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-divider content-position="left">基本信息</el-divider>
      <el-form-item label="计划名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入计划名称" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="巡检模板" prop="template_id">
            <el-select v-model="form.template_id" placeholder="选择模板" filterable>
              <el-option v-for="t in templateOptions" :key="t.id" :label="t.name" :value="t.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="计划类型">
            <el-select v-model="form.plan_type" placeholder="选择类型" clearable>
              <el-option label="日常例行" value="日常例行" />
              <el-option label="专项防雷" value="专项防雷" />
              <el-option label="缺陷复查" value="缺陷复查" />
              <el-option label="保电特巡" value="保电特巡" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">调度配置</el-divider>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="执行频率" prop="schedule_frequency">
            <el-select v-model="form.schedule_frequency" placeholder="选择频率" style="width: 100%">
              <el-option label="每天" value="daily" />
              <el-option label="每周" value="weekly" />
              <el-option label="每月" value="monthly" />
              <el-option label="自定义Cron" value="custom" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12" v-if="form.schedule_frequency === 'weekly'">
          <el-form-item label="星期">
            <el-select v-model="form.schedule_day_of_week" placeholder="选择星期" style="width: 100%">
              <el-option label="周一" :value="1" />
              <el-option label="周二" :value="2" />
              <el-option label="周三" :value="3" />
              <el-option label="周四" :value="4" />
              <el-option label="周五" :value="5" />
              <el-option label="周六" :value="6" />
              <el-option label="周日" :value="0" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12" v-if="form.schedule_frequency === 'monthly'">
          <el-form-item label="日期">
            <el-input-number v-model="form.schedule_day_of_month" :min="1" :max="31" style="width: 100%" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16" v-if="form.schedule_frequency !== 'custom'">
        <el-col :span="12">
          <el-form-item label="执行时间" prop="schedule_time">
            <el-time-picker v-model="form.schedule_time" placeholder="选择执行时间" format="HH:mm" value-format="HH:mm" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="时区" prop="timezone">
            <el-input v-model="form.timezone" placeholder="Asia/Shanghai" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16" v-if="form.schedule_frequency === 'custom'">
        <el-col :span="12">
          <el-form-item label="Cron表达式" prop="cron_expr">
            <el-input v-model="form.cron_expr" placeholder="如 0 8 * * * 表示每天8点" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="时区" prop="timezone">
            <el-input v-model="form.timezone" placeholder="Asia/Shanghai" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="开始时间" prop="start_at">
            <el-date-picker v-model="form.start_at" type="datetime" placeholder="选择开始时间" value-format="YYYY-MM-DDTHH:mm:ssZ" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="完成时限(分)" prop="due_duration_minutes">
            <el-input-number v-model="form.due_duration_minutes" :min="1" :max="1440" style="width: 100%" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">指派配置</el-divider>
      <el-row :gutter="16">
        <el-col :span="24">
          <el-form-item label="指派给" prop="assignee_id">
            <el-select v-model="form.assignee_id" placeholder="选择用户" filterable style="width: 100%">
              <el-option v-for="u in userOptions" :key="u.id" :label="u.display_name || u.name || u.username" :value="u.id" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">巡检信息</el-divider>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="巡检点位" prop="point_name">
            <el-select v-model="form.point_name" placeholder="选择巡检点位" filterable clearable style="width: 100%" @change="onPointChange">
              <el-option v-for="p in pointOptions" :key="p.id" :label="p.name" :value="p.name">
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <span>{{ p.name }}</span>
                  <span style="color: #999; font-size: 12px">{{ p.area }}{{ p.substation ? ' · ' + p.substation : '' }}</span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="设备名称">
            <el-input v-model="form.equipment_name" placeholder="巡检设备名称" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="作业区域">
            <el-input v-model="form.inspect_area" placeholder="作业区域" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="变电站">
            <el-input v-model="form.substation_name" placeholder="变电站名称" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">人员配置</el-divider>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="计划负责人">
            <el-input v-model="form.plan_principal" placeholder="计划负责人" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="作业人员">
            <el-input v-model="form.operator_user" placeholder="作业人员" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="监护人">
            <el-input v-model="form.guardian" placeholder="现场监护人" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="所属单位">
            <el-input v-model="form.belong_unit" placeholder="计划所属单位" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="工作概述">
        <el-input v-model="form.plan_desc" type="textarea" :rows="2" placeholder="工作内容概述" />
      </el-form-item>
      <el-form-item label="创建人">
        <el-input v-model="form.creator" placeholder="创建人" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submit" :loading="submitting">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type Plan = {
  id: number
  template_id: number
  name: string
  cron_expr: string
  timezone: string
  start_at: string
  due_duration_minutes: number
  assignee_type: string
  assignee_id: number
  point_name: string
  equipment_name: string
  plan_type: string
  enabled: boolean
  inspect_area: string
  substation_name: string
  belong_unit: string
  operator_unit: string
  plan_principal: string
  operator_user: string
  guardian: string
  plan_desc: string
  creator: string
}

type TemplateOption = { id: number; name: string }
type UserOption = { id: number; username: string; name: string; display_name: string }
type PointOption = { id: number; name: string; equipment_name: string; location: string; area: string; substation: string }

const items = ref<Plan[]>([])
const total = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const filters = reactive({ keyword: '', status: '', page: 1, page_size: 20 })
function indexMethod(rowIndex: number) { return (filters.page - 1) * filters.page_size + rowIndex + 1 }
const templateOptions = ref<TemplateOption[]>([])
const templateMap = ref<Record<number, string>>({})
const userOptions = ref<UserOption[]>([])
const userMap = ref<Record<number, string>>({})
const pointOptions = ref<PointOption[]>([])

const form = reactive({
  name: '',
  template_id: undefined as number | undefined,
  cron_expr: '',
  timezone: 'Asia/Shanghai',
  start_at: '',
  due_duration_minutes: 90,
  assignee_type: 'user',
  assignee_id: undefined as number | undefined,
  point_name: '',
  equipment_name: '',
  plan_type: '',
  inspect_area: '',
  substation_name: '',
  belong_unit: '',
  operator_unit: '',
  plan_principal: '',
  operator_user: '',
  guardian: '',
  plan_desc: '',
  creator: '',
  schedule_frequency: 'daily' as 'daily' | 'weekly' | 'monthly' | 'custom',
  schedule_time: '08:00' as string,
  schedule_day_of_week: 1,
  schedule_day_of_month: 1,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入计划名称', trigger: 'blur' }],
  template_id: [{ required: true, message: '请选择巡检模板', trigger: 'change' }],
  start_at: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  due_duration_minutes: [{ required: true, message: '请输入完成时限', trigger: 'blur' }],
  assignee_id: [{ required: true, message: '请选择指派用户', trigger: 'change' }],
  point_name: [{ required: true, message: '请选择巡检点位', trigger: 'change' }],
  schedule_frequency: [{ required: true, message: '请选择执行频率', trigger: 'change' }],
  schedule_time: [{ required: true, message: '请选择执行时间', trigger: 'change' }],
  cron_expr: [
    {
      validator: (_rule: any, value: string, callback: any) => {
        // 非自定义模式下 cron 由频率+时间自动生成，无需手动填写
        if (form.schedule_frequency === 'custom' && !value) {
          callback(new Error('请输入Cron表达式'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

// 将调度选择转换为 Cron 表达式
function buildCronExpr(): string {
  if (form.schedule_frequency === 'custom') {
    return form.cron_expr
  }
  const time = form.schedule_time || '08:00'
  const parts = time.split(':')
  const hour = parts[0] || '08'
  const minute = parts[1] || '00'
  switch (form.schedule_frequency) {
    case 'daily':
      return `${minute} ${hour} * * *`
    case 'weekly':
      return `${minute} ${hour} * * ${form.schedule_day_of_week}`
    case 'monthly':
      return `${minute} ${hour} ${form.schedule_day_of_month} * *`
    default:
      return `${minute} ${hour} * * *`
  }
}

// 从 Cron 表达式反向解析调度配置
function parseCronToSchedule(cronExpr: string) {
  const parts = cronExpr.trim().split(/\s+/)
  if (parts.length < 5) {
    form.schedule_frequency = 'custom'
    form.cron_expr = cronExpr
    return
  }
  const [min, hour, dom, mon, dow] = parts
  form.schedule_time = `${hour.padStart(2, '0')}:${min.padStart(2, '0')}`
  if (dom !== '*' && mon === '*') {
    form.schedule_frequency = 'monthly'
    form.schedule_day_of_month = parseInt(dom) || 1
  } else if (dow !== '*') {
    form.schedule_frequency = 'weekly'
    form.schedule_day_of_week = parseInt(dow) || 1
  } else {
    form.schedule_frequency = 'daily'
  }
}

// 格式化调度时间的显示
function formatSchedule(plan: Plan): string {
  const cronExpr = plan.cron_expr
  const parts = cronExpr.trim().split(/\s+/)
  if (parts.length < 5) return cronExpr
  const [min, hour, dom, mon, dow] = parts
  const time = `${hour.padStart(2, '0')}:${min.padStart(2, '0')}`
  if (dom !== '*' && mon === '*') return `每月${dom}日 ${time}`
  if (dow !== '*') {
    const weekLabels: Record<string, string> = { '0': '周日', '1': '周一', '2': '周二', '3': '周三', '4': '周四', '5': '周五', '6': '周六' }
    return `${weekLabels[dow] || '周' + dow} ${time}`
  }
  return `每天 ${time}`
}

// 选择巡检点位时自动填充关联信息
function onPointChange(pointName: string) {
  const point = pointOptions.value.find(p => p.name === pointName)
  if (point) {
    form.equipment_name = point.equipment_name || form.equipment_name
    form.inspect_area = point.area || form.inspect_area
    form.substation_name = point.substation || form.substation_name
  }
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.keyword) params.set('keyword', filters.keyword)
    if (filters.status) params.set('status', filters.status)
    params.set('page', String(filters.page))
    params.set('page_size', String(filters.page_size))
    const result = await apiGet<{ items: Plan[]; total: number }>(`/api/admin/plans?${params.toString()}`)
    items.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

async function loadTemplates() {
  templateOptions.value = await apiGet<TemplateOption[]>('/api/admin/templates/all')
  templateMap.value = Object.fromEntries(templateOptions.value.map(t => [t.id, t.name]))
}

async function loadUsers() {
  const users = await apiGet<UserOption[]>('/api/admin/users/all')
  userOptions.value = users
  userMap.value = Object.fromEntries(users.map(u => [u.id, u.display_name || u.name || u.username]))
}

async function loadPoints() {
  pointOptions.value = await apiGet<PointOption[]>('/api/admin/inspection-points/all')
}

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    name: '', template_id: undefined, cron_expr: '', timezone: 'Asia/Shanghai',
    start_at: '', due_duration_minutes: 90, assignee_type: 'user', assignee_id: undefined,
    point_name: '', equipment_name: '', plan_type: '', inspect_area: '',
    substation_name: '', belong_unit: '', operator_unit: '',
    plan_principal: '', operator_user: '', guardian: '', plan_desc: '', creator: '',
    schedule_frequency: 'daily', schedule_time: '08:00', schedule_day_of_week: 1, schedule_day_of_month: 1,
  })
  loadTemplates()
  loadPoints()
  dialogVisible.value = true
}

async function openEdit(row: Plan) {
  editingId.value = row.id
  // 从 Cron 表达式解析调度配置
  parseCronToSchedule(row.cron_expr)
  Object.assign(form, {
    name: row.name,
    template_id: row.template_id,
    cron_expr: row.cron_expr,
    timezone: row.timezone,
    start_at: row.start_at,
    due_duration_minutes: row.due_duration_minutes,
    assignee_type: row.assignee_type,
    assignee_id: row.assignee_id,
    point_name: row.point_name,
    equipment_name: row.equipment_name,
    plan_type: row.plan_type,
    inspect_area: row.inspect_area,
    substation_name: row.substation_name,
    belong_unit: row.belong_unit,
    operator_unit: row.operator_unit,
    plan_principal: row.plan_principal,
    operator_user: row.operator_user,
    guardian: row.guardian,
    plan_desc: row.plan_desc,
    creator: row.creator
  })
  loadTemplates()
  loadPoints()
  dialogVisible.value = true
}

async function submit() {
  // 先生成 cron 表达式写入 form，再校验，避免隐藏的 cron_expr 字段因空值拦截提交
  form.cron_expr = buildCronExpr()
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...form }
    if (editingId.value) {
      await apiPost(`/api/admin/plans/${editingId.value}/update`, payload)
    } else {
      await apiPost('/api/admin/plans', payload)
    }
    ElMessage.success('计划已保存')
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function toggleEnable(row: Plan) {
  const action = row.enabled ? 'disable' : 'enable'
  await apiPost(`/api/admin/plans/${row.id}/${action}`)
  ElMessage.success(row.enabled ? '计划已停用' : '计划已启用')
  await load()
}

async function generateNow(row: Plan) {
  await ElMessageBox.confirm('确定立即生成一次巡检任务吗？', '提示', { type: 'info' })
  await apiPost(`/api/admin/plans/${row.id}/generate-now`)
  ElMessage.success('任务已生成')
}

async function remove(row: Plan) {
  await ElMessageBox.confirm(`确定删除计划"${row.name}"吗？`, '提示', { type: 'warning' })
  await apiPost(`/api/admin/plans/${row.id}/delete`)
  ElMessage.success('计划已删除')
  await load()
}

onMounted(() => { load(); loadTemplates(); loadUsers(); loadPoints() })
</script>
