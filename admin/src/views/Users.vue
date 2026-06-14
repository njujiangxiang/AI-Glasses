<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">用户管理</span>
      <div class="filters">
        <el-input v-model="filters.keyword" clearable placeholder="用户名/姓名/身份证" style="width: 180px" @keyup.enter="load" />
        <el-select v-model="filters.org_code" clearable filterable placeholder="所属单位" style="width: 180px">
          <el-option v-for="org in organizations" :key="org.code" :label="`${org.name}（${org.code}）`" :value="org.code" />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="状态" style="width: 120px">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="disabled" />
        </el-select>
        <el-button @click="load">查询</el-button>
        <el-button type="primary" @click="openCreate">新增用户</el-button>
      </div>
    </div>
    <el-table :data="users" stripe>
      <el-table-column label="头像" width="80">
        <template #default="scope">
          <el-avatar v-if="scope.row.has_avatar" :src="avatarUrl(scope.row)" />
          <el-avatar v-else>{{ scope.row.name?.slice(0, 1) || scope.row.username.slice(0, 1) }}</el-avatar>
        </template>
      </el-table-column>
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="name" label="姓名" />
      <el-table-column label="性别" width="90">
        <template #default="scope">{{ genderLabel(scope.row.gender) }}</template>
      </el-table-column>
      <el-table-column label="出生年月" width="120">
        <template #default="scope">{{ birthText(scope.row) }}</template>
      </el-table-column>
      <el-table-column label="身份证号" width="180">
        <template #default="scope">{{ maskID(scope.row.id_card_no) }}</template>
      </el-table-column>
      <el-table-column prop="org_code" label="所属单位" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column label="操作" width="220">
        <template #default="scope">
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button v-if="scope.row.status !== 'active'" link type="success" @click="enable(scope.row.id)">启用</el-button>
          <el-button v-else link type="warning" @click="disable(scope.row.id)">停用</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager">
      <el-pagination v-model:current-page="filters.page" v-model:page-size="filters.page_size" layout="total, sizes, prev, pager, next" :total="total" @current-change="load" @size-change="load" />
    </div>
  </el-card>

  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑用户' : '新增用户'" width="620px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="用户名" prop="username">
        <el-input v-model="form.username" :disabled="!!editingId" placeholder="请输入用户名" />
      </el-form-item>
      <el-form-item label="姓名" prop="name">
        <el-input v-model="form.name" placeholder="请输入姓名" />
      </el-form-item>
      <el-form-item label="性别" prop="gender">
        <el-select v-model="form.gender">
          <el-option label="男" value="male" />
          <el-option label="女" value="female" />
          <el-option label="未知" value="unknown" />
        </el-select>
      </el-form-item>
      <el-form-item label="出生年月">
        <el-input-number v-model="form.birth_year" :min="0" :max="2100" placeholder="年" />
        <el-input-number v-model="form.birth_month" :min="0" :max="12" placeholder="月" style="margin-left: 12px" />
      </el-form-item>
      <el-form-item label="身份证号" prop="id_card_no">
        <el-input v-model="form.id_card_no" placeholder="请输入身份证号码" />
      </el-form-item>
      <el-form-item label="所属单位" prop="org_code">
        <el-select v-model="form.org_code" clearable filterable placeholder="请选择单位">
          <el-option v-for="org in organizations" :key="org.code" :label="`${org.name}（${org.code}）`" :value="org.code" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="disabled" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="editingId" label="头像">
        <el-upload :http-request="uploadAvatar" :show-file-list="false" accept="image/png,image/jpeg,image/webp">
          <el-button>上传头像</el-button>
        </el-upload>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import { ElMessage } from 'element-plus'
import { apiGet, apiPost, apiUpload } from '@/api/client'

type Organization = { id: number; code: string; name: string }
type User = { id: number; username: string; name: string; gender: string; birth_year: number; birth_month: number; id_card_no: string; org_code: string; status: string; has_avatar: boolean; UpdatedAt: string }
type UserList = { items: User[]; total: number }

const users = ref<User[]>([])
const total = ref(0)
const organizations = ref<Organization[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const filters = reactive({ keyword: '', org_code: '', status: '', page: 1, page_size: 20 })
const form = reactive({ username: '', name: '', gender: 'unknown', birth_year: 0, birth_month: 0, id_card_no: '', org_code: '', status: 'active' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }]
}

// load 查询用户列表。
async function load() {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([key, value]) => { if (value) params.set(key, String(value)) })
  const result = await apiGet<UserList>(`/api/admin/users?${params.toString()}`)
  users.value = result.items
  total.value = result.total
}
// loadOrganizations 查询单位选项。
async function loadOrganizations() { organizations.value = await apiGet<Organization[]>('/api/admin/organizations') }
// openCreate 打开新增用户弹窗。
function openCreate() {
  editingId.value = null
  Object.assign(form, { username: '', name: '', gender: 'unknown', birth_year: 0, birth_month: 0, id_card_no: '', org_code: '', status: 'active' })
  dialogVisible.value = true
}
// openEdit 打开编辑用户弹窗。
function openEdit(row: User) {
  editingId.value = row.id
  Object.assign(form, { username: row.username, name: row.name, gender: row.gender, birth_year: row.birth_year, birth_month: row.birth_month, id_card_no: row.id_card_no, org_code: row.org_code, status: row.status })
  dialogVisible.value = true
}
// submit 保存用户基础资料。
async function submit() {
  await formRef.value?.validate()
  if (editingId.value) await apiPost(`/api/admin/users/${editingId.value}/update`, form)
  else await apiPost('/api/admin/users', form)
  ElMessage.success('用户已保存')
  dialogVisible.value = false
  await load()
}
// uploadAvatar 上传头像到数据库。
async function uploadAvatar(options: UploadRequestOptions) {
  if (!editingId.value) return
  const fd = new FormData()
  fd.append('avatar', options.file)
  await apiUpload(`/api/admin/users/${editingId.value}/avatar`, fd)
  ElMessage.success('头像已上传')
  await load()
}
// enable 启用用户。
async function enable(id: number) {
  await apiPost(`/api/admin/users/${id}/enable`)
  ElMessage.success('用户已启用')
  await load()
}
// disable 停用用户。
async function disable(id: number) {
  await apiPost(`/api/admin/users/${id}/disable`)
  ElMessage.success('用户已停用')
  await load()
}
function avatarUrl(row: User) { return `/api/admin/users/${row.id}/avatar?v=${encodeURIComponent(row.UpdatedAt || '')}` }
function genderLabel(gender: string) { return gender === 'male' ? '男' : gender === 'female' ? '女' : '未知' }
function birthText(row: User) { return row.birth_year ? `${row.birth_year}-${String(row.birth_month).padStart(2, '0')}` : '-' }
function maskID(value: string) { return value ? `${value.slice(0, 6)}********${value.slice(-4)}` : '-' }
onMounted(async () => { await Promise.all([loadOrganizations(), load()]) })
</script>

<style scoped>
.filters { display: flex; gap: 8px; align-items: center; }
.pager { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
