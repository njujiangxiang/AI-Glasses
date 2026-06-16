<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">组织管理</span>
      <div>
        <el-button @click="load">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增单位</el-button>
      </div>
    </div>
    <el-table :data="organizations" stripe row-key="id">
      <el-table-column prop="code" label="单位编码" />
      <el-table-column prop="name" label="单位名称" />
      <el-table-column prop="parent_code" label="上级单位编码" />
      <el-table-column label="状态" width="120">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
            {{ scope.row.status === 'active' ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="CreatedAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="260">
        <template #default="scope">
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button v-if="scope.row.status !== 'active'" link type="success" @click="enable(scope.row.id)">启用</el-button>
          <el-button v-else link type="warning" @click="disable(scope.row.id)">停用</el-button>
          <el-button link type="danger" @click="remove(scope.row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑单位' : '新增单位'" width="520px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="单位编码" prop="code">
        <el-input v-model="form.code" placeholder="如 ROOT 或 SG-001" />
      </el-form-item>
      <el-form-item label="单位名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入单位名称" />
      </el-form-item>
      <el-form-item label="上级单位" prop="parent_code">
        <el-select v-model="form.parent_code" clearable filterable placeholder="顶级单位">
          <el-option v-for="org in parentOptions" :key="org.code" :label="`${org.name}（${org.code}）`" :value="org.code" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="disabled" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type Organization = { id: number; code: string; name: string; parent_code: string; status: string; CreatedAt: string }
const organizations = ref<Organization[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const form = reactive({ code: '', name: '', parent_code: '', status: 'active' })
const rules: FormRules = {
  code: [{ required: true, message: '请输入单位编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入单位名称', trigger: 'blur' }]
}
const parentOptions = computed(() => organizations.value.filter((org) => org.id !== editingId.value))

// load 查询单位组织列表并刷新表格。
async function load() { organizations.value = await apiGet<Organization[]>('/api/admin/organizations') }
// openCreate 打开新增单位弹窗。
function openCreate() {
  editingId.value = null
  Object.assign(form, { code: '', name: '', parent_code: '', status: 'active' })
  dialogVisible.value = true
}
// openEdit 打开编辑单位弹窗。
function openEdit(row: Organization) {
  editingId.value = row.id
  Object.assign(form, { code: row.code, name: row.name, parent_code: row.parent_code, status: row.status })
  dialogVisible.value = true
}
// submit 保存单位组织。
async function submit() {
  await formRef.value?.validate()
  if (editingId.value) await apiPost(`/api/admin/organizations/${editingId.value}/update`, form)
  else await apiPost('/api/admin/organizations', form)
  ElMessage.success('单位已保存')
  dialogVisible.value = false
  await load()
}
// enable 启用单位组织。
async function enable(id: number) {
  await apiPost(`/api/admin/organizations/${id}/enable`)
  ElMessage.success('单位已启用')
  await load()
}
// disable 停用单位组织。
async function disable(id: number) {
  await apiPost(`/api/admin/organizations/${id}/disable`)
  ElMessage.success('单位已停用')
  await load()
}
// remove 删除单位组织。
async function remove(id: number) {
  await ElMessageBox.confirm('确定删除该单位吗？存在下级单位或用户时将无法删除。', '提示', { type: 'warning' })
  await apiPost(`/api/admin/organizations/${id}/delete`)
  ElMessage.success('单位已删除')
  await load()
}
onMounted(load)
</script>
