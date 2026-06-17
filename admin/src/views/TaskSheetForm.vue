<template>
  <div>
    <div class="page-toolbar">
      <div>
        <span class="card-title">{{ pageTitle }}</span>
        <div class="page-subtitle">{{ subtitle }}</div>
      </div>
      <el-space>
        <el-button @click="backToList">返回列表</el-button>
        <el-button v-if="isView" type="primary" @click="copyCurrent">复制新建</el-button>
        <template v-else>
          <el-button @click="saveDraft">保存草稿</el-button>
          <el-button type="primary" @click="submitSheet">提交任务单</el-button>
        </template>
      </el-space>
    </div>

    <el-alert v-if="copyFrom" title="已从历史任务单复制，请确认计划日期和负责人后再提交。" type="warning" show-icon :closable="false" class="copy-alert" />

    <el-card shadow="never" class="form-card">
      <template #header><span class="card-title">主表信息</span></template>
      <el-form ref="formRef" :model="master" :rules="rules" label-width="120px" :disabled="isView">
        <el-row :gutter="16">
          <el-col :span="6"><el-form-item label="任务编号"><el-input v-model="master.code" disabled placeholder="保存后自动生成" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="任务名称" prop="name"><el-input v-model="master.name" placeholder="请输入任务名称" /></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="优先级" prop="priority"><el-select v-model="master.priority"><el-option label="普通" value="normal" /><el-option label="紧急" value="urgent" /></el-select></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="所属单位" prop="orgCode"><el-select v-model="master.orgCode"><el-option label="默认单位" value="ROOT" /></el-select></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="计划日期" prop="planDate"><el-date-picker v-model="master.planDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="负责人" prop="ownerName"><el-select v-model="master.ownerName"><el-option label="巡检班组长" value="巡检班组长" /><el-option label="巡检员" value="巡检员" /></el-select></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="预计完成时长" prop="estimatedHours"><el-input-number v-model="master.estimatedHours" :min="0" :precision="1" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="备注"><el-input v-model="master.remark" type="textarea" :rows="3" placeholder="请输入备注" /></el-form-item></el-col>
        </el-row>
      </el-form>
    </el-card>

    <el-card shadow="never" class="form-card">
      <template #header>
        <div class="card-header-row">
          <span class="card-title">作业明细</span>
          <span class="page-subtitle">共 {{ details.length }} 条，{{ invalidRows.length }} 条待完善</span>
        </div>
      </template>
      <div v-if="!isView" class="detail-toolbar">
        <el-button type="primary" @click="addDetail">新增明细</el-button>
        <el-button @click="copyLastDetail">复制上一行</el-button>
        <el-button>批量导入</el-button>
      </div>
      <el-empty v-if="details.length === 0" description="还没有作业明细，至少添加一条明细后才能提交任务单。">
        <el-button v-if="!isView" type="primary" @click="addDetail">新增第一条明细</el-button>
      </el-empty>
      <el-table v-else :data="details" stripe row-key="localId" :row-class-name="rowClassName">
        <el-table-column type="index" label="序号" width="70" />
        <el-table-column label="作业点位" min-width="160">
          <template #default="scope">
            <el-input v-if="!isView" v-model="scope.row.pointName" placeholder="请输入点位" />
            <span v-else>{{ scope.row.pointName }}</span>
            <div v-if="!scope.row.pointName" class="row-error">缺少作业点位</div>
          </template>
        </el-table-column>
        <el-table-column label="设备名称" min-width="160">
          <template #default="scope"><el-input v-if="!isView" v-model="scope.row.deviceName" /><span v-else>{{ scope.row.deviceName }}</span></template>
        </el-table-column>
        <el-table-column label="作业内容" min-width="220">
          <template #default="scope"><el-input v-if="!isView" v-model="scope.row.workContent" /><span v-else>{{ scope.row.workContent }}</span></template>
        </el-table-column>
        <el-table-column label="标准工时" width="130">
          <template #default="scope"><el-input-number v-if="!isView" v-model="scope.row.standardHours" :min="0" :precision="1" style="width: 100%" /><span v-else>{{ scope.row.standardHours }}</span></template>
        </el-table-column>
        <el-table-column label="风险等级" width="130">
          <template #default="scope">
            <el-select v-if="!isView" v-model="scope.row.riskLevel"><el-option label="低风险" value="low" /><el-option label="中风险" value="medium" /><el-option label="高风险" value="high" /></el-select>
            <el-tag v-else :type="scope.row.riskLevel === 'high' ? 'danger' : scope.row.riskLevel === 'medium' ? 'warning' : 'success'">{{ riskLabel(scope.row.riskLevel) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="!isView" label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button link type="primary" @click="copyDetail(scope.$index)">复制</el-button>
            <el-button link type="danger" @click="removeDetail(scope.$index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="sticky-actions">
      <span :class="invalidRows.length ? 'danger-text' : 'page-subtitle'">{{ summaryText }}</span>
      <el-space>
        <el-button @click="backToList">返回列表</el-button>
        <el-button v-if="isView" type="primary" @click="copyCurrent">复制新建</el-button>
        <template v-else>
          <el-button @click="saveDraft">保存草稿</el-button>
          <el-button type="primary" @click="submitSheet">提交任务单</el-button>
        </template>
      </el-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { taskStore, type TaskSheet, type TaskDetail } from '@/stores/taskStore'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()

const isView = computed(() => route.params.mode === 'view')
const isCreate = computed(() => route.params.mode === 'create')
const isEdit = computed(() => route.params.mode === 'edit')
const copyFrom = computed(() => String(route.query.copy_from || ''))

// Current task being edited - use reactive and update in place with Object.assign
const currentTask = reactive<TaskSheet>(taskStore.create())

// Master form fields (reactive view of currentTask)
const master = reactive({
  code: '',
  name: '',
  orgCode: 'ROOT',
  orgName: '默认单位',
  planDate: '',
  ownerName: '',
  priority: 'normal' as const,
  estimatedHours: 2,
  remark: ''
})

// Details array
const details = ref<TaskDetail[]>([])

// Form validation rules
const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  orgCode: [{ required: true, message: '请选择所属单位', trigger: 'change' }],
  planDate: [{ required: true, message: '请选择计划日期', trigger: 'change' }],
  ownerName: [{ required: true, message: '请选择负责人', trigger: 'change' }],
  estimatedHours: [{ required: true, message: '请输入预计完成时长', trigger: 'blur' }]
}

// Computed properties
const pageTitle = computed(() => isView.value ? `查看：${master.code || '任务单'}` : isCreate.value ? '新增任务单' : `编辑：${master.code || '任务单'}`)
const subtitle = computed(() => isView.value ? '只读查看任务单主表和明细' : '先维护主表信息，再录入作业明细')
const invalidRows = computed(() => details.value.filter((row) => !row.pointName || !row.deviceName || !row.workContent))
const summaryText = computed(() => isView.value ? '查看模式不可编辑，可复制为新任务单。' : invalidRows.value.length ? `明细表有 ${invalidRows.value.length} 行待完善，保存草稿不受影响，提交前需修正。` : '主表和明细已完整，可提交任务单。')

// Load data on mount
onMounted(() => {
  loadTaskData()
})

// Load task data based on route params
function loadTaskData() {
  const id = Number(route.params.id)
  let newTask: TaskSheet

  if (isCreate.value) {
    if (copyFrom.value) {
      // Copy from existing
      try {
        const sourceId = Number(copyFrom.value)
        newTask = taskStore.copy(sourceId)
      } catch (e) {
        ElMessage.error('复制失败，原任务单不存在')
        newTask = taskStore.create()
      }
    } else {
      // Create new
      newTask = taskStore.create()
    }
  } else if (id) {
    // View or Edit existing
    const existing = taskStore.getById(id)
    if (existing) {
      newTask = existing
    } else {
      ElMessage.error('任务单不存在')
      router.push('/tasksheets')
      return
    }
  } else {
    newTask = taskStore.create()
  }

  // Update currentTask in place
  Object.assign(currentTask, newTask)
  // Sync master and details from currentTask
  syncFromTask()
}

// Sync currentTask data to form
function syncFromTask() {
  master.code = currentTask.code || ''
  master.name = currentTask.name || ''
  master.orgCode = currentTask.orgCode
  master.orgName = currentTask.orgName
  master.planDate = currentTask.planDate
  master.ownerName = currentTask.ownerName
  master.priority = currentTask.priority
  master.estimatedHours = currentTask.estimatedHours
  master.remark = currentTask.remark || ''
  details.value = [...currentTask.details]
}

// Sync form changes back to currentTask
function syncToTask() {
  currentTask.code = master.code
  currentTask.name = master.name
  currentTask.orgCode = master.orgCode
  currentTask.orgName = master.orgName
  currentTask.planDate = master.planDate
  currentTask.ownerName = master.ownerName
  currentTask.priority = master.priority
  currentTask.estimatedHours = master.estimatedHours
  currentTask.remark = master.remark
  currentTask.details = details.value
  currentTask.detailCount = details.value.length
}

// Add new detail row
function addDetail() {
  details.value.push({ localId: Date.now(), pointName: '', deviceName: '', workContent: '', standardHours: 0, riskLevel: 'low' })
}

// Copy last detail row
function copyLastDetail() {
  if (details.value.length) {
    copyDetail(details.value.length - 1)
  } else {
    addDetail()
  }
}

// Copy specific detail row
function copyDetail(index: number) {
  details.value.splice(index + 1, 0, { ...details.value[index], localId: Date.now() })
}

// Remove detail row
function removeDetail(index: number) {
  details.value.splice(index, 1)
}

// Save draft
async function saveDraft() {
  await formRef.value?.validate()
  syncToTask()
  taskStore.save(currentTask)
  ElMessage.success('草稿已保存')
  // If was create, switch to edit mode
  if (isCreate.value) {
    await router.replace(`/tasksheets/${currentTask.id}/edit`)
  }
}

// Submit task sheet
async function submitSheet() {
  await formRef.value?.validate()
  if (!details.value.length || invalidRows.value.length) {
    ElMessage.error('请先完善作业明细')
    return
  }
  syncToTask()
  taskStore.save(currentTask)
  taskStore.submit(currentTask.id)
  ElMessage.success('任务单已提交')
  await router.push('/tasksheets?status=submitted')
}

// Back to list
async function backToList() {
  if (!isView.value) {
    // Check for unsaved changes
    const original = taskStore.getById(currentTask.id)
    let hasChanges = false
    if (original) {
      // Simple comparison
      hasChanges = JSON.stringify(original) !== JSON.stringify({ ...currentTask, details: details.value })
    } else if (master.name || details.value.length) {
      hasChanges = true
    }
    if (hasChanges) {
      try {
        await ElMessageBox.confirm('当前内容可能未保存，确定返回列表吗？', '提示', { type: 'warning' })
      } catch {
        return
      }
    }
  }
  await router.push('/tasksheets')
}

// Copy current task as new
function copyCurrent() {
  router.push(`/tasksheets/create?copy_from=${route.params.id || currentTask.id}`)
}

// Table row class
function rowClassName({ row }: { row: TaskDetail }) {
  return !row.pointName || !row.deviceName || !row.workContent ? 'warning-row' : ''
}

// Risk level label
function riskLabel(value: TaskDetail['riskLevel']) {
  return value === 'high' ? '高风险' : value === 'medium' ? '中风险' : '低风险'
}
</script>

<style scoped>
.page-subtitle { margin-top: 4px; color: var(--el-text-color-secondary); font-size: 13px; }
.copy-alert { margin-bottom: 12px; }
.form-card { margin-bottom: 12px; }
.card-header-row { display: flex; justify-content: space-between; align-items: center; width: 100%; }
.detail-toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.row-error { margin-top: 4px; color: var(--el-color-danger); font-size: 12px; }
.danger-text { color: var(--el-color-danger); }
.sticky-actions { position: sticky; bottom: 0; z-index: 2; display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: rgba(255, 255, 255, .96); border: 1px solid var(--el-border-color-lighter); border-radius: 4px; box-shadow: 0 -4px 14px rgba(0, 0, 0, .04); }
:deep(.warning-row) { --el-table-tr-bg-color: #fff8f6; }
</style>
