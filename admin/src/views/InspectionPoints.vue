<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">点位管理</span>
      <div>
        <el-input v-model="filters.keyword" placeholder="搜索点位" clearable style="width: 200px" @keyup.enter="load" />
        <el-button @click="load" style="margin-left: 8px">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增点位</el-button>
      </div>
    </div>
    <el-table :data="items" stripe row-key="id" v-loading="loading">
      <el-table-column type="index" label="序号" width="70" align="center" :index="indexMethod" />
      <el-table-column prop="name" label="点位名称" min-width="150" />
      <el-table-column prop="equipment_name" label="关联设备" width="150" />
      <el-table-column prop="area" label="所属区域" width="120" />
      <el-table-column prop="substation" label="变电站" width="140" />
      <el-table-column prop="location" label="位置描述" min-width="150" />
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? '已启用' : '已停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="scope">
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

  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑点位' : '新增点位'" width="640px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="点位名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入巡检点位名称" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="关联设备">
            <el-input v-model="form.equipment_name" placeholder="关联设备名称" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="位置描述">
            <el-input v-model="form.location" placeholder="具体位置描述" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="所属区域">
            <el-input v-model="form.area" placeholder="所属区域" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="变电站">
            <el-input v-model="form.substation" placeholder="所属变电站" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="点位说明">
        <el-input v-model="form.description" type="textarea" :rows="3" placeholder="点位详细说明" />
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

type Point = {
  id: number
  name: string
  equipment_name: string
  location: string
  area: string
  substation: string
  description: string
  enabled: boolean
}

const items = ref<Point[]>([])
const total = ref(0)
const loading = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const filters = reactive({ keyword: '', page: 1, page_size: 20 })
function indexMethod(rowIndex: number) { return (filters.page - 1) * filters.page_size + rowIndex + 1 }

const form = reactive({
  name: '',
  equipment_name: '',
  location: '',
  area: '',
  substation: '',
  description: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入点位名称', trigger: 'blur' }],
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.keyword) params.set('keyword', filters.keyword)
    params.set('page', String(filters.page))
    params.set('page_size', String(filters.page_size))
    const result = await apiGet<{ items: Point[]; total: number }>(`/api/admin/inspection-points?${params.toString()}`)
    items.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, { name: '', equipment_name: '', location: '', area: '', substation: '', description: '' })
  dialogVisible.value = true
}

function openEdit(row: Point) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    equipment_name: row.equipment_name,
    location: row.location,
    area: row.area,
    substation: row.substation,
    description: row.description,
  })
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (editingId.value) {
      await apiPost(`/api/admin/inspection-points/${editingId.value}/update`, form)
    } else {
      await apiPost('/api/admin/inspection-points', form)
    }
    ElMessage.success('点位已保存')
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function toggleEnable(row: Point) {
  const action = row.enabled ? 'disable' : 'enable'
  await apiPost(`/api/admin/inspection-points/${row.id}/${action}`)
  ElMessage.success(row.enabled ? '点位已停用' : '点位已启用')
  await load()
}

async function remove(row: Point) {
  await ElMessageBox.confirm(`确定删除点位"${row.name}"吗？`, '提示', { type: 'warning' })
  await apiPost(`/api/admin/inspection-points/${row.id}/delete`)
  ElMessage.success('点位已删除')
  await load()
}

onMounted(load)
</script>
