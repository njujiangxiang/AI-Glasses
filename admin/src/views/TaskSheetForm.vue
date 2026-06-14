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
          <el-col :span="6"><el-form-item label="所属单位" prop="org_code"><el-select v-model="master.org_code"><el-option label="默认单位" value="ROOT" /></el-select></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="计划日期" prop="plan_date"><el-date-picker v-model="master.plan_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="负责人" prop="owner_name"><el-select v-model="master.owner_name"><el-option label="巡检班组长" value="巡检班组长" /><el-option label="巡检员" value="巡检员" /></el-select></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="预计完成时长" prop="estimated_hours"><el-input-number v-model="master.estimated_hours" :min="0" :precision="1" style="width: 100%" /></el-form-item></el-col>
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
      <el-table v-else :data="details" stripe row-key="local_id" :row-class-name="rowClassName">
        <el-table-column type="index" label="序号" width="70" />
        <el-table-column label="作业点位" min-width="160">
          <template #default="scope">
            <el-input v-if="!isView" v-model="scope.row.point_name" placeholder="请输入点位" />
            <span v-else>{{ scope.row.point_name }}</span>
            <div v-if="!scope.row.point_name" class="row-error">缺少作业点位</div>
          </template>
        </el-table-column>
        <el-table-column label="设备名称" min-width="160">
          <template #default="scope"><el-input v-if="!isView" v-model="scope.row.device_name" /><span v-else>{{ scope.row.device_name }}</span></template>
        </el-table-column>
        <el-table-column label="作业内容" min-width="220">
          <template #default="scope"><el-input v-if="!isView" v-model="scope.row.work_content" /><span v-else>{{ scope.row.work_content }}</span></template>
        </el-table-column>
        <el-table-column label="标准工时" width="130">
          <template #default="scope"><el-input-number v-if="!isView" v-model="scope.row.standard_hours" :min="0" :precision="1" style="width: 100%" /><span v-else>{{ scope.row.standard_hours }}</span></template>
        </el-table-column>
        <el-table-column label="风险等级" width="130">
          <template #default="scope">
            <el-select v-if="!isView" v-model="scope.row.risk_level"><el-option label="低风险" value="low" /><el-option label="中风险" value="medium" /><el-option label="高风险" value="high" /></el-select>
            <el-tag v-else :type="scope.row.risk_level === 'high' ? 'danger' : scope.row.risk_level === 'medium' ? 'warning' : 'success'">{{ riskLabel(scope.row.risk_level) }}</el-tag>
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
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'

type Detail = { local_id: number; point_name: string; device_name: string; work_content: string; standard_hours: number; risk_level: 'low' | 'medium' | 'high' }
const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const isView = computed(() => route.params.mode === 'view')
const isCreate = computed(() => route.params.mode === 'create')
const copyFrom = computed(() => String(route.query.copy_from || ''))
const master = reactive({ code: isCreate.value ? '' : 'TASK-20260614-001', name: copyFrom.value ? 'A区变电站 AI眼镜巡检作业（复制）' : 'A区变电站 AI眼镜巡检作业', org_code: 'ROOT', plan_date: '2026-06-15', owner_name: '巡检班组长', priority: 'normal', estimated_hours: 2, remark: '明细中高风险点位提交前需要班组长复核。' })
const details = ref<Detail[]>([
  { local_id: 1, point_name: 'A区一号柜', device_name: '开关柜 KYN28', work_content: '红外测温与外观检查', standard_hours: 0.5, risk_level: 'medium' },
  { local_id: 2, point_name: '', device_name: '主变压器', work_content: '油温与声音巡检', standard_hours: 1, risk_level: 'high' },
  { local_id: 3, point_name: '电缆夹层', device_name: '电缆桥架', work_content: '积水与异物检查', standard_hours: 0.5, risk_level: 'medium' }
])
const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  org_code: [{ required: true, message: '请选择所属单位', trigger: 'change' }],
  plan_date: [{ required: true, message: '请选择计划日期', trigger: 'change' }],
  owner_name: [{ required: true, message: '请选择负责人', trigger: 'change' }],
  estimated_hours: [{ required: true, message: '请输入预计完成时长', trigger: 'blur' }]
}
const pageTitle = computed(() => isView.value ? `查看：${master.code}` : isCreate.value ? '新增任务单' : `编辑：${master.code}`)
const subtitle = computed(() => isView.value ? '只读查看任务单主表和明细' : '先维护主表信息，再录入作业明细')
const invalidRows = computed(() => details.value.filter((row) => !row.point_name || !row.device_name || !row.work_content))
const summaryText = computed(() => isView.value ? '查看模式不可编辑，可复制为新任务单。' : invalidRows.value.length ? `明细表有 ${invalidRows.value.length} 行待完善，保存草稿不受影响，提交前需修正。` : '主表和明细已完整，可提交任务单。')

// addDetail 新增一行作业明细。
function addDetail() { details.value.push({ local_id: Date.now(), point_name: '', device_name: '', work_content: '', standard_hours: 0, risk_level: 'low' }) }
// copyLastDetail 复制上一行明细。
function copyLastDetail() { if (details.value.length) copyDetail(details.value.length - 1); else addDetail() }
// copyDetail 复制指定明细行。
function copyDetail(index: number) { details.value.splice(index + 1, 0, { ...details.value[index], local_id: Date.now() }) }
// removeDetail 删除指定明细行。
function removeDetail(index: number) { details.value.splice(index, 1) }
// saveDraft 保存草稿，新增页保存后切换为编辑页。
async function saveDraft() {
  ElMessage.success('草稿已保存')
  if (isCreate.value) await router.replace('/tasksheets/1/edit')
}
// submitSheet 提交任务单并返回列表。
async function submitSheet() {
  await formRef.value?.validate()
  if (!details.value.length || invalidRows.value.length) {
    ElMessage.error('请先完善作业明细')
    return
  }
  ElMessage.success('任务单已提交')
  await router.push('/tasksheets?status=submitted')
}
// backToList 返回列表，存在未保存变更时给出确认。
async function backToList() {
  if (!isView.value) await ElMessageBox.confirm('当前内容可能未保存，确定返回列表吗？', '提示', { type: 'warning' })
  await router.push('/tasksheets')
}
function copyCurrent() { router.push(`/tasksheets/create?copy_from=${route.params.id || 1}`) }
function rowClassName({ row }: { row: Detail }) { return !row.point_name || !row.device_name || !row.work_content ? 'warning-row' : '' }
function riskLabel(value: Detail['risk_level']) { return value === 'high' ? '高风险' : value === 'medium' ? '中风险' : '低风险' }
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
