<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">任务模板</span>
      <div>
        <el-input v-model="filters.keyword" placeholder="搜索模板" clearable style="width: 200px" @keyup.enter="load" />
        <el-button @click="load" style="margin-left: 8px">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增模板</el-button>
      </div>
    </div>
    <el-table :data="items" stripe row-key="id" v-loading="loading">
      <el-table-column type="index" label="序号" width="70" align="center" :index="indexMethod" />
      <el-table-column prop="name" label="模板名称" min-width="150" />
      <el-table-column prop="type" label="类型" width="120" />
      <el-table-column prop="scene" label="场景" width="120" />
      <el-table-column prop="applicable_roles" label="适用角色" width="150" />
      <el-table-column prop="version" label="版本" width="80" />
      <el-table-column label="节点数" width="80" align="center">
        <template #default="scope">
          <el-tag>{{ nodeCountMap[scope.row.id] || '-' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? '已启用' : '已停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="340" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="viewDetail(scope.row)">详情</el-button>
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button link type="primary" @click="copyTemplate(scope.row)">复制模板</el-button>
          <el-button link :type="scope.row.enabled ? 'warning' : 'success'" @click="toggleEnable(scope.row)">
            {{ scope.row.enabled ? '停用' : '启用' }}
          </el-button>
          <el-button link type="danger" @click="remove(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager" v-if="total > filters.page_size">
      <el-pagination v-model:current-page="filters.page" :page-size="filters.page_size" :total="total" layout="total, prev, pager, next" @current-change="load" />
    </div>
  </el-card>

  <!-- 模板编辑弹窗 -->
  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑模板' : '新增模板'" width="720px" top="5vh">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="模板名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入模板名称" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="类型" prop="type">
            <el-select v-model="form.type" placeholder="选择类型" clearable>
              <el-option label="操作票执行" value="操作票执行" />
              <el-option label="设备巡检" value="设备巡检" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="场景" prop="scene">
            <el-select v-model="form.scene" placeholder="选择场景" clearable>
              <el-option label="变电巡视" value="变电巡视" />
              <el-option label="配电巡检" value="配电巡检" />
              <el-option label="输电线路" value="输电线路" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="版本号">
            <el-input v-model="form.version" placeholder="如 v1.0" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="适用角色">
            <el-input v-model="form.applicable_roles" placeholder="如 巡检员,班组长" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="模板说明">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="模板说明" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" placeholder="备注" />
      </el-form-item>
      <el-divider content-position="left">选择巡检节点</el-divider>
      <el-form-item label="已选节点">
        <div style="width: 100%">
          <div style="margin-bottom: 8px; display: flex; gap: 8px; align-items: center">
            <el-button size="small" type="primary" @click="openNodeCreate">新增节点</el-button>
            <span style="color: #909399; font-size: 12px">共 {{ selectedNodes.length }} 个节点，可拖拽排序</span>
          </div>
          <draggable
            v-model="selectedNodes"
            item-key="id"
            handle=".drag-handle"
            @end="onDragEnd"
          >
            <template #item="{ element, index }">
              <div class="node-row" :class="{ 'dragging': dragIndex === index }">
                <span class="drag-handle" :title="'拖拽排序'">☰</span>
                <span class="node-index">{{ index + 1 }}</span>
                <span class="node-name">{{ element.name }}</span>
                <el-tag size="small">{{ nodeTypeLabel(element.node_type) }}</el-tag>
                <span class="node-actions">
                  <el-button link type="primary" size="small" @click="openNodeEdit(element)">编辑</el-button>
                  <el-button link type="danger" size="small" @click="removeNode(index)">删除</el-button>
                </span>
              </div>
            </template>
          </draggable>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submit" :loading="submitting">保存</el-button>
    </template>
  </el-dialog>

  <!-- 节点新增/编辑弹窗 -->
  <el-dialog v-model="nodeDialogVisible" :title="nodeEditingId ? '编辑节点' : '新增节点'" width="640px" append-to-body>
    <el-form ref="nodeFormRef" :model="nodeForm" :rules="nodeRules" label-width="120px">
      <el-form-item label="节点名称" prop="name">
        <el-input v-model="nodeForm.name" placeholder="请输入节点名称" />
      </el-form-item>
      <el-form-item label="节点类型" prop="node_type">
        <el-select v-model="nodeForm.node_type" placeholder="选择节点类型">
          <el-option label="文本" value="text" />
          <el-option label="读取" value="read" />
          <el-option label="检查" value="check" />
          <el-option label="拍照" value="photo" />
          <el-option label="录像" value="video" />
          <el-option label="录音" value="audio" />
        </el-select>
      </el-form-item>
      <el-form-item label="节点说明" prop="description">
        <el-input v-model="nodeForm.description" type="textarea" :rows="2" placeholder="节点详细说明" />
      </el-form-item>
      <el-form-item label="简短提示" prop="node_desc">
        <el-input v-model="nodeForm.node_desc" placeholder="AR眼镜端展示的简短提示" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="最少照片数">
            <el-input-number v-model="nodeForm.min_photos" :min="0" :max="10" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="超时时间(秒)">
            <el-input-number v-model="nodeForm.timeout_second" :min="0" :max="3600" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="8">
          <el-form-item label="要求文本">
            <el-switch v-model="nodeForm.require_text" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="允许异常">
            <el-switch v-model="nodeForm.allow_abnormal" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="实时拍摄">
            <el-switch v-model="nodeForm.require_live_capture" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="是否必做">
            <el-select v-model="nodeForm.is_required">
              <el-option label="是" value="1" />
              <el-option label="否" value="0" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="是否强制">
            <el-select v-model="nodeForm.is_mandatory">
              <el-option label="是" value="1" />
              <el-option label="否" value="0" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="备注">
        <el-input v-model="nodeForm.remark" placeholder="备注信息" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="nodeDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submitNode" :loading="nodeSubmitting">保存</el-button>
    </template>
  </el-dialog>

  <!-- 详情抽屉 -->
  <el-drawer v-model="detailVisible" title="模板详情" size="500px">
    <div v-if="detailData">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="模板名称">{{ detailData.template.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ detailData.template.type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="场景">{{ detailData.template.scene || '-' }}</el-descriptions-item>
        <el-descriptions-item label="版本">{{ detailData.template.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="适用角色">{{ detailData.template.applicable_roles || '-' }}</el-descriptions-item>
        <el-descriptions-item label="说明">{{ detailData.template.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detailData.template.enabled ? 'success' : 'info'">{{ detailData.template.enabled ? '已启用' : '已停用' }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <h4 style="margin: 16px 0 8px">巡检节点 ({{ detailData.nodes.length }})</h4>
      <el-table :data="detailData.nodes" stripe size="small">
        <el-table-column prop="sort_order" label="序号" width="60" />
        <el-table-column prop="name" label="节点名称" />
        <el-table-column prop="node_type" label="类型" width="80">
          <template #default="scope">{{ nodeTypeLabel(scope.row.node_type) }}</template>
        </el-table-column>
        <el-table-column prop="min_photos" label="照片" width="60" align="center" />
        <el-table-column label="必做" width="60" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.is_required === '1'" type="success" size="small">是</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import draggable from 'vuedraggable'
import { apiGet, apiPost } from '@/api/client'

type Template = {
  id: number
  name: string
  description: string
  type: string
  scene: string
  version: string
  applicable_roles: string
  enabled: boolean
  remark: string
}

type TemplateNode = {
  id: number
  template_id: number | null
  name: string
  node_type: string
  sort_order: number
  min_photos: number
  is_required: string
  description: string
  node_desc: string
  require_text: boolean
  allow_abnormal: boolean
  require_live_capture: boolean
  is_mandatory: string
  timeout_second: number
  remark: string
  created_at?: string
}

const items = ref<Template[]>([])
const total = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const filters = reactive({ keyword: '', page: 1, page_size: 20 })
function indexMethod(rowIndex: number) { return (filters.page - 1) * filters.page_size + rowIndex + 1 }
const nodeCountMap = ref<Record<number, number>>({})

const form = reactive({
  name: '',
  description: '',
  type: '',
  scene: '',
  version: '',
  applicable_roles: '',
  remark: '',
  node_ids: [] as number[]
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }]
}

// 节点选择
const templateNodes = ref<TemplateNode[]>([])

// 新增/编辑节点弹窗
const nodeDialogVisible = ref(false)
const nodeEditingId = ref<number | null>(null)
const nodeFormRef = ref<FormInstance>()
const nodeSubmitting = ref(false)
const nodeForm = reactive({
  name: '',
  node_type: 'text',
  description: '',
  node_desc: '',
  min_photos: 0,
  require_text: false,
  allow_abnormal: false,
  require_live_capture: false,
  is_required: '1',
  is_mandatory: '1',
  timeout_second: 0,
  remark: ''
})
const nodeRules: FormRules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  node_type: [{ required: true, message: '请选择节点类型', trigger: 'change' }]
}

// 拖拽排序
const dragIndex = ref(-1)

// 详情
const detailVisible = ref(false)
const detailData = ref<{ template: Template; nodes: TemplateNode[] } | null>(null)

const selectedNodes = computed({
  get: () => {
    return form.node_ids.map((id, index) => {
      const node = templateNodes.value.find(n => n.id === id)
                    || { id, name: `节点#${id}`, node_type: '', sort_order: index + 1 }
      return { ...node, sort_order: index + 1 }
    })
  },
  set: (newNodes: TemplateNode[]) => {
    form.node_ids = newNodes.map(n => n.id)
  }
})

function nodeTypeLabel(type: string) {
  const map: Record<string, string> = { text: '文本', read: '读取', check: '检查', photo: '拍照', video: '录像', audio: '录音' }
  return map[type] || type
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.keyword) params.set('keyword', filters.keyword)
    params.set('page', String(filters.page))
    params.set('page_size', String(filters.page_size))
    const result = await apiGet<{ items: Template[]; total: number }>(`/api/admin/templates?${params.toString()}`)
    items.value = result.items
    total.value = result.total
    // 加载每个模板的节点数
    for (const t of items.value) {
      const detail = await apiGet<{ nodes: TemplateNode[] }>(`/api/admin/templates/${t.id}`)
      nodeCountMap.value[t.id] = detail.nodes.length
    }
  } finally {
    loading.value = false
  }
}

function removeNode(index: number) {
  form.node_ids.splice(index, 1)
}

function onDragEnd() {
  dragIndex.value = -1
}

// 新增节点
function openNodeCreate() {
  nodeEditingId.value = null
  Object.assign(nodeForm, {
    name: '', node_type: 'text', description: '', node_desc: '',
    min_photos: 0, require_text: false, allow_abnormal: false,
    require_live_capture: false, is_required: '1', is_mandatory: '1',
    timeout_second: 0, remark: ''
  })
  nodeDialogVisible.value = true
}

// 编辑已有节点
function openNodeEdit(node: TemplateNode) {
  nodeEditingId.value = node.id
  const toStr = (v: unknown) => (v === true || v === '1' || v === 1 ? '1' : '0')
  Object.assign(nodeForm, {
    name: node.name,
    node_type: node.node_type,
    description: node.description || '',
    node_desc: node.node_desc || '',
    min_photos: node.min_photos || 0,
    require_text: node.require_text ?? false,
    allow_abnormal: node.allow_abnormal ?? false,
    require_live_capture: node.require_live_capture ?? false,
    is_required: toStr(node.is_required),
    is_mandatory: toStr(node.is_mandatory),
    timeout_second: node.timeout_second || 0,
    remark: node.remark || ''
  })
  nodeDialogVisible.value = true
}

// 保存节点（新增或编辑）
async function submitNode() {
  await nodeFormRef.value?.validate()
  nodeSubmitting.value = true
  try {
    if (nodeEditingId.value) {
      await apiPost(`/api/admin/nodes/${nodeEditingId.value}/update`, nodeForm)
      // 更新本地节点数据
      const idx = templateNodes.value.findIndex(n => n.id === nodeEditingId.value)
      if (idx !== -1) {
        templateNodes.value[idx] = { ...templateNodes.value[idx], ...nodeForm }
      }
      ElMessage.success('节点已更新')
    } else {
      const newNode = await apiPost<TemplateNode>('/api/admin/nodes', nodeForm)
      // 将新节点添加到已选节点列表末尾
      templateNodes.value.push(newNode)
      form.node_ids.push(newNode.id)
      ElMessage.success('节点已创建并添加')
    }
    nodeDialogVisible.value = false
  } finally {
    nodeSubmitting.value = false
  }
}

function openCreate() {
  editingId.value = null
  templateNodes.value = []
  Object.assign(form, { name: '', description: '', type: '', scene: '', version: '', applicable_roles: '', remark: '', node_ids: [] })
  dialogVisible.value = true
}

async function openEdit(row: Template) {
  editingId.value = row.id
  const detail = await apiGet<{ template: Template; nodes: TemplateNode[] }>(`/api/admin/templates/${row.id}`)
  templateNodes.value = detail.nodes
  Object.assign(form, {
    name: detail.template.name,
    description: detail.template.description,
    type: detail.template.type,
    scene: detail.template.scene,
    version: detail.template.version,
    applicable_roles: detail.template.applicable_roles,
    remark: detail.template.remark,
    node_ids: detail.nodes.map(n => n.id)
  })
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  if (form.node_ids.length === 0) {
    ElMessage.warning('请至少选择一个节点')
    return
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await apiPost(`/api/admin/templates/${editingId.value}/update`, form)
    } else {
      await apiPost('/api/admin/templates', form)
    }
    ElMessage.success('模板已保存')
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function toggleEnable(row: Template) {
  const action = row.enabled ? 'disable' : 'enable'
  await apiPost(`/api/admin/templates/${row.id}/${action}`)
  ElMessage.success(row.enabled ? '模板已停用' : '模板已启用')
  await load()
}

async function viewDetail(row: Template) {
  detailData.value = await apiGet(`/api/admin/templates/${row.id}`)
  detailVisible.value = true
}

async function remove(row: Template) {
  await ElMessageBox.confirm(`确定删除模板"${row.name}"吗？关联的节点将被释放。`, '提示', { type: 'warning' })
  await apiPost(`/api/admin/templates/${row.id}/delete`)
  ElMessage.success('模板已删除')
  await load()
}

async function copyTemplate(row: Template) {
  await ElMessageBox.confirm(`确定复制模板"${row.name}"吗？将创建一个名称前缀为"（复制）"的新模板，并同步复制所有节点。`, '提示', { type: 'info' })
  await apiPost(`/api/admin/templates/${row.id}/copy`)
  ElMessage.success('模板已复制')
  await load()
}

onMounted(load)
</script>

<style scoped>
.node-row {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
  transition: background 0.2s;
}
.node-row:hover {
  background: #f5f7fa;
}
.node-row.dragging {
  background: #ecf5ff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
.drag-handle {
  cursor: grab;
  margin-right: 8px;
  font-size: 16px;
  color: #909399;
  user-select: none;
}
.drag-handle:active {
  cursor: grabbing;
}
.node-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  font-size: 12px;
  margin-right: 12px;
  flex-shrink: 0;
}
.node-name {
  flex: 1;
  font-size: 14px;
  margin-right: 8px;
}
.node-actions {
  display: flex;
  gap: 4px;
  margin-left: auto;
}
</style>
