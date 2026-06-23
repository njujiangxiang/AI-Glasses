<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">巡检模板</span>
      <div>
        <el-input v-model="filters.keyword" placeholder="搜索模板" clearable style="width: 200px" @keyup.enter="load" />
        <el-button @click="load" style="margin-left: 8px">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增模板</el-button>
      </div>
    </div>
    <el-table :data="items" stripe row-key="id" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
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
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="viewDetail(scope.row)">详情</el-button>
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
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
              <el-option label="设备巡检" value="设备巡检" />
              <el-option label="缺陷复查" value="缺陷复查" />
              <el-option label="安全交底" value="安全交底" />
              <el-option label="保电特巡" value="保电特巡" />
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
            <el-button size="small" @click="showNodeSelector = true">添加节点</el-button>
            <span style="color: #909399; font-size: 12px">已选择 {{ form.node_ids.length }} 个节点</span>
          </div>
          <el-table :data="selectedNodes" stripe size="small" max-height="200">
            <el-table-column prop="sort_order" label="序号" width="60" />
            <el-table-column prop="name" label="节点名称" />
            <el-table-column prop="node_type" label="类型" width="80">
              <template #default="scope">{{ nodeTypeLabel(scope.row.node_type) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="scope">
                <el-button link type="danger" size="small" @click="removeNode(scope.$index)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submit" :loading="submitting">保存</el-button>
    </template>
  </el-dialog>

  <!-- 节点选择弹窗 -->
  <el-dialog v-model="showNodeSelector" title="选择节点" width="600px" append-to-body>
    <div style="margin-bottom: 8px; display: flex; gap: 8px">
      <el-select v-model="nodeFilterType" placeholder="节点类型" clearable style="width: 140px" @change="loadUnassignedNodes">
        <el-option label="文本" value="text" />
        <el-option label="读取" value="read" />
        <el-option label="检查" value="check" />
        <el-option label="拍照" value="photo" />
        <el-option label="录像" value="video" />
        <el-option label="录音" value="audio" />
      </el-select>
      <el-button @click="loadUnassignedNodes">刷新</el-button>
    </div>
    <el-table :data="unassignedNodes" stripe size="small" max-height="360" @selection-change="onNodeSelect">
      <el-table-column type="selection" width="40" />
      <el-table-column prop="name" label="节点名称" />
      <el-table-column prop="node_type" label="类型" width="80">
        <template #default="scope">{{ nodeTypeLabel(scope.row.node_type) }}</template>
      </el-table-column>
      <el-table-column prop="min_photos" label="最少照片" width="80" align="center" />
      <el-table-column label="必做" width="60" align="center">
        <template #default="scope">
          <el-tag v-if="scope.row.is_required === '1'" type="success" size="small">是</el-tag>
        </template>
      </el-table-column>
    </el-table>
    <template #footer>
      <el-button @click="showNodeSelector = false">取消</el-button>
      <el-button type="primary" @click="confirmAddNodes">确定添加</el-button>
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
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
}

const items = ref<Template[]>([])
const total = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const filters = reactive({ keyword: '', page: 1, page_size: 20 })
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
const showNodeSelector = ref(false)
const unassignedNodes = ref<TemplateNode[]>([])
const nodeFilterType = ref('')
const tempSelectedNodes = ref<TemplateNode[]>([])

// 详情
const detailVisible = ref(false)
const detailData = ref<{ template: Template; nodes: TemplateNode[] } | null>(null)

const selectedNodes = computed(() => {
  return form.node_ids.map((id, index) => {
    const node = unassignedNodes.value.find(n => n.id === id) || { id, name: `节点#${id}`, node_type: '', sort_order: index + 1 }
    return { ...node, sort_order: index + 1 }
  })
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

async function loadUnassignedNodes() {
  const params = new URLSearchParams()
  if (nodeFilterType.value) params.set('node_type', nodeFilterType.value)
  unassignedNodes.value = await apiGet<TemplateNode[]>(`/api/admin/nodes/unassigned?${params.toString()}`)
}

function onNodeSelect(nodes: TemplateNode[]) {
  tempSelectedNodes.value = nodes
}

function confirmAddNodes() {
  for (const node of tempSelectedNodes.value) {
    if (!form.node_ids.includes(node.id)) {
      form.node_ids.push(node.id)
    }
  }
  showNodeSelector.value = false
}

function removeNode(index: number) {
  form.node_ids.splice(index, 1)
}

function openCreate() {
  editingId.value = null
  Object.assign(form, { name: '', description: '', type: '', scene: '', version: '', applicable_roles: '', remark: '', node_ids: [] })
  loadUnassignedNodes()
  dialogVisible.value = true
}

async function openEdit(row: Template) {
  editingId.value = row.id
  const detail = await apiGet<{ template: Template; nodes: TemplateNode[] }>(`/api/admin/templates/${row.id}`)
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
  loadUnassignedNodes()
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

watch(showNodeSelector, (v) => { if (v) loadUnassignedNodes() })

onMounted(load)
</script>
