<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">节点管理</span>
      <div>
        <el-select v-model="filters.node_type" placeholder="节点类型" clearable style="width: 140px" @change="load">
          <el-option label="文本" value="text" />
          <el-option label="读取" value="read" />
          <el-option label="检查" value="check" />
          <el-option label="拍照" value="photo" />
          <el-option label="录像" value="video" />
          <el-option label="录音" value="audio" />
        </el-select>
        <el-input v-model="filters.keyword" placeholder="搜索节点" clearable style="width: 200px; margin-left: 8px" @keyup.enter="load" />
        <el-button @click="load" style="margin-left: 8px">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增节点</el-button>
      </div>
    </div>
    <el-table :data="items" stripe row-key="id" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="节点名称" min-width="150" />
      <el-table-column prop="node_type" label="节点类型" width="100">
        <template #default="scope">
          <el-tag>{{ nodeTypeLabel(scope.row.node_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="min_photos" label="最少照片" width="100" align="center" />
      <el-table-column label="要求" width="160">
        <template #default="scope">
          <el-tag v-if="scope.row.is_required === '1'" type="success" size="small">必做</el-tag>
          <el-tag v-if="scope.row.is_mandatory === '1'" type="warning" size="small">强制</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="timeout_second" label="超时(秒)" width="100" align="center" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button link type="danger" @click="remove(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager" v-if="total > filters.page_size">
      <el-pagination v-model:current-page="filters.page" :page-size="filters.page_size" :total="total" layout="total, prev, pager, next" @current-change="load" />
    </div>
  </el-card>

  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑节点' : '新增节点'" width="640px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item label="节点名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入节点名称" />
      </el-form-item>
      <el-form-item label="节点类型" prop="node_type">
        <el-select v-model="form.node_type" placeholder="选择节点类型">
          <el-option label="文本" value="text" />
          <el-option label="读取" value="read" />
          <el-option label="检查" value="check" />
          <el-option label="拍照" value="photo" />
          <el-option label="录像" value="video" />
          <el-option label="录音" value="audio" />
        </el-select>
      </el-form-item>
      <el-form-item label="节点说明" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="节点详细说明" />
      </el-form-item>
      <el-form-item label="简短提示" prop="node_desc">
        <el-input v-model="form.node_desc" placeholder="AR眼镜端展示的简短提示" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="最少照片数" prop="min_photos">
            <el-input-number v-model="form.min_photos" :min="0" :max="10" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="超时时间(秒)" prop="timeout_second">
            <el-input-number v-model="form.timeout_second" :min="0" :max="3600" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="8">
          <el-form-item label="要求文本">
            <el-switch v-model="form.require_text" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="允许异常">
            <el-switch v-model="form.allow_abnormal" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="实时拍摄">
            <el-switch v-model="form.require_live_capture" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="是否必做">
            <el-select v-model="form.is_required">
              <el-option label="是" value="1" />
              <el-option label="否" value="0" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="是否强制">
            <el-select v-model="form.is_mandatory">
              <el-option label="是" value="1" />
              <el-option label="否" value="0" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="备注" prop="remark">
        <el-input v-model="form.remark" placeholder="备注信息" />
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

type Node = {
  id: number
  template_id: number | null
  name: string
  node_type: string
  description: string
  node_desc: string
  min_photos: number
  require_text: boolean
  allow_abnormal: boolean
  require_live_capture: boolean
  is_mandatory: string
  is_required: string
  timeout_second: number
  remark: string
}

const items = ref<Node[]>([])
const total = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const filters = reactive({ keyword: '', node_type: '', assigned: '', page: 1, page_size: 20 })

const form = reactive({
  name: '',
  node_type: 'text',
  description: '',
  node_desc: '',
  min_photos: 0,
  require_text: false,
  allow_abnormal: false,
  require_live_capture: true,
  is_required: '1',
  is_mandatory: '1',
  timeout_second: 0,
  remark: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  node_type: [{ required: true, message: '请选择节点类型', trigger: 'change' }]
}

function nodeTypeLabel(type: string) {
  const map: Record<string, string> = { text: '文本', read: '读取', check: '检查', photo: '拍照', video: '录像', audio: '录音' }
  return map[type] || type
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.keyword) params.set('keyword', filters.keyword)
    if (filters.node_type) params.set('node_type', filters.node_type)
    if (filters.assigned) params.set('assigned', filters.assigned)
    params.set('page', String(filters.page))
    params.set('page_size', String(filters.page_size))
    const result = await apiGet<{ items: Node[]; total: number }>(`/api/admin/nodes?${params.toString()}`)
    items.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    name: '', node_type: 'text', description: '', node_desc: '',
    min_photos: 0, require_text: false, allow_abnormal: false,
    require_live_capture: true, is_required: '1', is_mandatory: '1',
    timeout_second: 0, remark: ''
  })
  dialogVisible.value = true
}

function openEdit(row: Node) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name, node_type: row.node_type, description: row.description,
    node_desc: row.node_desc, min_photos: row.min_photos,
    require_text: row.require_text, allow_abnormal: row.allow_abnormal,
    require_live_capture: row.require_live_capture,
    is_required: row.is_required, is_mandatory: row.is_mandatory,
    timeout_second: row.timeout_second, remark: row.remark
  })
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (editingId.value) {
      await apiPost(`/api/admin/nodes/${editingId.value}/update`, form)
    } else {
      await apiPost('/api/admin/nodes', form)
    }
    ElMessage.success('节点已保存')
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function remove(row: Node) {
  await ElMessageBox.confirm(`确定删除节点"${row.name}"吗？`, '提示', { type: 'warning' })
  await apiPost(`/api/admin/nodes/${row.id}/delete`)
  ElMessage.success('节点已删除')
  await load()
}

onMounted(load)
</script>
