<template>
  <div class="workflow-editor-page">
    <div class="page-header">
      <div style="display: flex; align-items: center; gap: 16px">
        <el-button link @click="goBack" style="font-size: 18px">←</el-button>
        <span class="page-title">{{ workflow.id ? '编辑工作流' : '新建工作流' }}</span>
        <el-tag v-if="workflow.status" :type="workflow.status === 'published' ? 'success' : 'info'" size="small">
          {{ statusLabel(workflow.status) }}
        </el-tag>
      </div>
      <div style="display: flex; gap: 12px">
        <el-button size="large" @click="goBack">取消</el-button>
        <el-button size="large" type="primary" :loading="saving" @click="saveDraft">保存草稿</el-button>
        <el-button v-if="workflow.status !== 'published'" size="large" type="success" :loading="publishing" @click="publish">发布工作流</el-button>
      </div>
    </div>

    <div class="tab-bar">
      <div class="tab-item active">📝 步骤配置</div>
      <div class="tab-item" @click="showBasicInfo = true">⚙️ 基本信息</div>
    </div>

    <div class="editor-layout">
      <!-- Steps Panel -->
      <div class="steps-panel">
        <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center">
          <span style="font-weight: 500; font-size: 15px">检查步骤 ({{ steps.length }})</span>
        </div>

        <draggable
          v-model="steps"
          item-key="id"
          handle=".drag-handle"
          animation="200"
          ghost-class="ghost"
          @end="onDragEnd"
        >
          <template #item="{ element: step, index }">
            <div
              class="step-card"
              :class="{ selected: selectedStepId === step.id }"
              @click="selectStep(step.id)"
            >
              <div class="step-header">
                <span class="drag-handle">⠿</span>
                <div class="step-number">{{ index + 1 }}</div>
                <div class="step-icon" :class="step.type">
                  <span>{{ stepIcon(step.type) }}</span>
                </div>
                <div class="step-info">
                  <div class="step-name">{{ step.name || '未命名步骤' }}</div>
                  <div class="step-type">{{ stepTypeLabel(step.type) }}</div>
                </div>
                <el-tag v-if="step.required" size="small" type="danger" effect="light">必填</el-tag>
                <div class="step-actions">
                  <el-button size="small" link @click.stop="duplicateStep(step.id)">复制</el-button>
                  <el-button size="small" link type="danger" @click.stop="deleteStep(step.id)">删除</el-button>
                </div>
              </div>
            </div>
          </template>
        </draggable>

        <button class="add-step-btn" @click="showStepTypePicker = true">
          <span>➕</span>
          添加检查步骤
        </button>

        <!-- Step Type Picker -->
        <div v-if="showStepTypePicker" class="step-type-picker-modal">
          <div style="font-weight: 500; margin-bottom: 16px">选择步骤类型</div>
          <div class="step-type-grid">
            <div
              v-for="type in stepTypes"
              :key="type.value"
              class="type-option"
              @click="addStep(type.value)"
            >
              <div class="icon">{{ type.icon }}</div>
              <div class="label">{{ type.label }}</div>
            </div>
          </div>
          <el-button link @click="showStepTypePicker = false">取消</el-button>
        </div>
      </div>

      <!-- Config Panel -->
      <div class="config-panel">
        <div class="config-title">⚙️ 步骤配置</div>

        <div v-if="selectedStep">
          <div class="config-section">
            <label class="config-label">步骤名称</label>
            <el-input v-model="selectedStep.name" placeholder="输入步骤名称" />
          </div>

          <div class="config-section">
            <label class="config-label">步骤描述</label>
            <el-input v-model="selectedStep.description" type="textarea" :rows="2" placeholder="输入步骤描述（可选）" />
          </div>

          <div class="config-section">
            <label class="config-label">是否必填</label>
            <el-switch v-model="selectedStep.required" active-text="是" inactive-text="否" />
          </div>

          <div class="config-section" v-if="selectedStep.type === 'select'">
            <label class="config-label">选择项配置</label>
            <div v-for="(opt, i) in selectedStepOptions" :key="i" style="margin-bottom: 8px; display: flex; gap: 8px">
              <el-input v-model="opt.label" placeholder="选项名称" size="small" />
              <el-button size="small" type="danger" link @click="selectedStepOptions.splice(i, 1)">删除</el-button>
            </div>
            <el-button size="small" link @click="selectedStepOptions.push({ label: '', value: '' })">➕ 添加选项</el-button>
          </div>

          <div class="config-section" style="padding-top: 16px; border-top: 1px solid #f0f0f0">
            <label class="config-label">🚨 异常触发配置</label>
            <div style="margin-top: 12px">
              <el-checkbox v-model="selectedStep.abnormal_enabled">启用异常触发</el-checkbox>
              <div style="margin: 12px 0 0 24px">
                <el-checkbox v-model="selectedStep.abnormal_require_photo">📷 必须拍照</el-checkbox><br>
                <el-checkbox v-model="selectedStep.abnormal_require_video" style="margin-top: 8px">🎥 必须录像</el-checkbox><br>
                <el-checkbox v-model="selectedStep.abnormal_require_note" style="margin-top: 8px">📝 必须填写备注</el-checkbox><br>
                <el-checkbox v-model="selectedStep.abnormal_require_signature" style="margin-top: 8px">✍️ 必须签字确认</el-checkbox><br>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="empty-config">
          <div style="font-size: 48px; margin-bottom: 16px">👆</div>
          选择一个步骤进行配置
        </div>
      </div>
    </div>

    <!-- Basic Info Dialog -->
    <el-dialog v-model="showBasicInfo" title="基本信息" width="500px">
      <el-form label-width="80px">
        <el-form-item label="工作流名称">
          <el-input v-model="workflow.name" placeholder="输入工作流名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="workflow.description" type="textarea" :rows="3" placeholder="输入工作流描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBasicInfo = false">取消</el-button>
        <el-button type="primary" @click="showBasicInfo = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import draggable from 'vuedraggable'
import { apiGet, apiPost } from '@/api/client'

const route = useRoute()
const router = useRouter()

const workflowId = computed(() => route.params.id as string)
const isNew = computed(() => workflowId.value === 'new')

const saving = ref(false)
const publishing = ref(false)
const showBasicInfo = ref(false)
const showStepTypePicker = ref(false)
const selectedStepId = ref<number | null>(null)

type Workflow = {
  id?: number
  name: string
  description: string
  status: string
}

type WorkflowStep = {
  id: number
  workflow_id: number
  name: string
  description: string
  type: string
  required: boolean
  options_json?: string
  abnormal_enabled: boolean
  abnormal_require_photo: boolean
  abnormal_require_video: boolean
  abnormal_require_note: boolean
  abnormal_require_signature: boolean
  sort_order: number
}

type SelectOption = { label: string; value: string }

const workflow = ref<Workflow>({ name: '', description: '', status: 'draft' })
const steps = ref<WorkflowStep[]>([])

const stepTypes = [
  { value: 'text', label: '文本输入', icon: '📝' },
  { value: 'number', label: '数值输入', icon: '🔢' },
  { value: 'select', label: '选择清单', icon: '📋' },
  { value: 'photo', label: '拍照', icon: '📷' },
  { value: 'video', label: '录像', icon: '🎥' },
  { value: 'audio', label: '录音', icon: '🎤' },
]

const selectedStep = computed(() => steps.value.find(s => s.id === selectedStepId.value))

const selectedStepOptions = computed({
  get: (): SelectOption[] => {
    if (!selectedStep.value?.options_json) return []
    try { return JSON.parse(selectedStep.value.options_json) } catch { return [] }
  },
  set: (val: SelectOption[]) => {
    if (selectedStep.value) {
      selectedStep.value.options_json = JSON.stringify(val)
    }
  }
})

async function load() {
  if (isNew.value) return
  try {
    const data = await apiGet<{ workflow: Workflow; steps: WorkflowStep[] }>(`/api/admin/workflows/${workflowId.value}`)
    workflow.value = data.workflow
    steps.value = data.steps.sort((a, b) => a.sort_order - b.sort_order)
  } catch (err) {
    ElMessage.error(String(err))
  }
}

async function ensureWorkflowCreated(): Promise<number> {
  if (workflow.value.id) return workflow.value.id
  const result = await apiPost<Workflow>('/api/admin/workflows', {
    name: workflow.value.name || '新建工作流',
    description: workflow.value.description
  })
  workflow.value.id = result.id
  workflow.value.status = result.status
  return result.id!
}

async function saveDraft() {
  saving.value = true
  try {
    const id = await ensureWorkflowCreated()
    await apiPost(`/api/admin/workflows/${id}`, {
      name: workflow.value.name,
      description: workflow.value.description
    })
    ElMessage.success('已保存')
  } catch (err) {
    ElMessage.error(String(err))
  } finally {
    saving.value = false
  }
}

async function publish() {
  if (steps.value.length === 0) {
    ElMessage.warning('请先添加至少一个步骤')
    return
  }
  publishing.value = true
  try {
    const id = await ensureWorkflowCreated()
    await apiPost(`/api/admin/workflows/${id}/publish`)
    workflow.value.status = 'published'
    ElMessage.success('发布成功')
  } catch (err) {
    ElMessage.error(String(err))
  } finally {
    publishing.value = false
  }
}

async function addStep(type: string) {
  try {
    const id = await ensureWorkflowCreated()
    const step = await apiPost<WorkflowStep>(`/api/admin/workflows/${id}/steps`, {
      name: '新' + stepTypeLabel(type) + '步骤',
      description: '',
      type: type,
      required: true,
      options: type === 'select' ? [{ label: '正常', value: 'normal' }, { label: '异常', value: 'abnormal' }] : [],
      abnormal_enabled: false,
      abnormal_require_photo: true,
      abnormal_require_video: false,
      abnormal_require_note: true,
      abnormal_require_signature: false
    })
    steps.value.push(step)
    selectedStepId.value = step.id
    showStepTypePicker.value = false
  } catch (err) {
    ElMessage.error(String(err))
  }
}

async function duplicateStep(id: number) {
  try {
    const wfId = await ensureWorkflowCreated()
    const step = await apiPost<WorkflowStep>(`/api/admin/workflows/${wfId}/steps/${id}/duplicate`)
    steps.value.push(step)
  } catch (err) {
    ElMessage.error(String(err))
  }
}

async function deleteStep(id: number) {
  try {
    await ElMessageBox.confirm('确认删除此步骤？', '提示', { type: 'warning' })
    const wfId = await ensureWorkflowCreated()
    await apiPost(`/api/admin/workflows/${wfId}/steps/${id}/delete`)
    const idx = steps.value.findIndex(s => s.id === id)
    if (idx >= 0) steps.value.splice(idx, 1)
    if (selectedStepId.value === id) selectedStepId.value = null
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(String(err))
  }
}

function selectStep(id: number) {
  selectedStepId.value = id
}

async function onDragEnd() {
  if (!workflow.value.id) return
  try {
    await apiPost(`/api/admin/workflows/${workflow.value.id}/steps/reorder`, {
      step_ids: steps.value.map(s => s.id)
    })
  } catch (err) {
    console.error('reorder failed', err)
  }
}

function goBack() {
  router.push('/workflows')
}

function statusLabel(status: string) {
  const map: Record<string, string> = { draft: '草稿', published: '已发布', archived: '已归档' }
  return map[status] || status
}

function stepIcon(type: string) {
  const t = stepTypes.find(x => x.value === type)
  return t?.icon || '❓'
}

function stepTypeLabel(type: string) {
  const t = stepTypes.find(x => x.value === type)
  return t?.label || type
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

.tab-bar {
  display: flex;
  gap: 0;
  margin-bottom: 20px;
  border-bottom: 1px solid #e8e8e8;
}

.tab-item {
  padding: 12px 24px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  color: #595959;
}

.tab-item.active {
  color: #1890ff;
  border-bottom-color: #1890ff;
  font-weight: 500;
}

.editor-layout {
  display: flex;
  gap: 20px;
  min-height: calc(100vh - 220px);
}

.steps-panel {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  overflow-y: auto;
}

.config-panel {
  width: 320px;
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  border-left: 1px solid #f0f0f0;
  overflow-y: auto;
}

.config-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.config-section {
  margin-bottom: 24px;
}

.config-label {
  font-size: 14px;
  font-weight: 500;
  color: #262626;
  margin-bottom: 8px;
  display: block;
}

.step-card {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  margin-bottom: 12px;
  transition: all 0.2s;
}

.step-card:hover {
  border-color: #1890ff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.1);
}

.step-card.selected {
  border-color: #1890ff;
  box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}

.step-header {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.step-number {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 13px;
  flex-shrink: 0;
}

.step-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: #f0f5ff;
  color: #2f54eb;
}

.step-icon.text { background: #f0f5ff; color: #2f54eb; }
.step-icon.number { background: #e6f7ff; color: #1890ff; }
.step-icon.select { background: #fff7e6; color: #fa8c16; }
.step-icon.photo { background: #f6ffed; color: #52c41a; }
.step-icon.video { background: #f9f0ff; color: #722ed1; }
.step-icon.audio { background: #fff1f0; color: #f5222d; }

.step-info {
  flex: 1;
  min-width: 0;
}

.step-name {
  font-weight: 500;
  color: #262626;
  margin-bottom: 2px;
}

.step-type {
  font-size: 12px;
  color: #8c8c8c;
}

.step-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.step-card:hover .step-actions {
  opacity: 1;
}

.drag-handle {
  color: #bfbfbf;
  cursor: grab;
  padding: 4px;
}

.drag-handle:active {
  cursor: grabbing;
}

.ghost {
  opacity: 0.5;
  background: #e6f7ff;
}

.add-step-btn {
  width: 100%;
  padding: 16px;
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  background: none;
  cursor: pointer;
  color: #8c8c8c;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.add-step-btn:hover {
  border-color: #1890ff;
  color: #1890ff;
  background: #f0f9ff;
}

.step-type-picker-modal {
  margin-top: 16px;
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.step-type-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}

.type-option {
  padding: 16px 12px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}

.type-option:hover {
  border-color: #1890ff;
  background: #f0f9ff;
}

.type-option .icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.type-option .label {
  font-size: 13px;
  color: #262626;
}

.empty-config {
  text-align: center;
  padding: 60px 20px;
  color: #bfbfbf;
}
</style>
