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
          <el-avatar v-if="avatarBlobUrls[scope.row.id]" :src="avatarBlobUrls[scope.row.id]" />
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
      <el-table-column label="所属单位" width="200">
        <template #default="scope">
          <div>{{ orgNameMap[scope.row.org_code] || '-' }}</div>
          <div style="font-size: 12px; color: #909399">{{ scope.row.org_code || '-' }}</div>
        </template>
      </el-table-column>
      <el-table-column label="角色" width="120">
        <template #default="scope">{{ getRoleName(scope.row.role_id) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
            {{ scope.row.status === 'active' ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
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
      <el-form-item label="角色" prop="role_id">
        <el-select v-model="form.role_id" placeholder="请选择角色">
          <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="disabled" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="dialogVisible" label="头像">
        <div style="display: flex; align-items: center; gap: 16px">
          <div class="avatar-preview">
            <el-avatar v-if="editingAvatarBlobUrl" :size="64" :src="editingAvatarBlobUrl" />
            <el-avatar v-else :size="64">{{ form.name?.slice(0, 1) || form.username?.slice(0, 1) }}</el-avatar>
            <div v-if="editingAvatarBlobUrl" class="avatar-badge">✓ 已上传</div>
          </div>
          <el-upload :http-request="editingId ? uploadAvatar : uploadAvatarForCreate" :show-file-list="false" accept="image/png,image/jpeg,image/webp">
            <el-button>{{ editingAvatarBlobUrl ? '更换头像' : '上传头像' }}</el-button>
          </el-upload>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="closeDialog">取消</el-button>
      <el-button type="primary" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import { ElMessage } from 'element-plus'
import { apiGet, apiPost, apiUpload } from '@/api/client'

type Organization = { id: number; code: string; name: string }
type Role = { id: number; name: string }
type User = { id: number; username: string; name: string; gender: string; birth_year: number; birth_month: number; id_card_no: string; org_code: string; status: string; has_avatar: boolean; role_id: number; UpdatedAt: string }
type UserList = { items: User[]; total: number }

const users = ref<User[]>([])
const total = ref(0)
const organizations = ref<Organization[]>([])
const roles = ref<Role[]>([])
// 头像 Blob URL 缓存
const avatarBlobUrls = ref<Record<number, string>>({})

// 使用 fetch + Blob URL 方式加载头像（支持 Bearer Token 认证）
async function loadAvatarForUser(user: User) {
  if (!user.has_avatar || avatarBlobUrls.value[user.id]) return
  try {
    const token = localStorage.getItem('admin_token')
    const res = await fetch(`/api/admin/users/${user.id}/avatar`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {}
    })
    if (res.ok) {
      const blob = await res.blob()
      avatarBlobUrls.value[user.id] = URL.createObjectURL(blob)
    }
  } catch (e) {
    console.warn(`Failed to load avatar for user ${user.id}:`, e)
  }
}

// 组织编码到名称的映射，用于列表显示
const orgNameMap = computed(() => {
  const map: Record<string, string> = {}
  organizations.value.forEach(org => { map[org.code] = org.name })
  return map
})
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const editingAvatarBlobUrl = ref('')
const pendingAvatarFile = ref<File | null>(null) // 新增用户时暂存的头像文件
const formRef = ref<FormInstance>()
const filters = reactive({ keyword: '', org_code: '', status: '', page: 1, page_size: 20 })
const form = reactive({ username: '', name: '', gender: 'unknown', birth_year: 0, birth_month: 0, id_card_no: '', org_code: '', status: 'active', role_id: 0 })
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
  // 异步加载每个用户的头像
  users.value.forEach(user => { if (user.has_avatar) loadAvatarForUser(user) })
}
// loadOrganizations 查询单位选项。
async function loadOrganizations() { organizations.value = await apiGet<Organization[]>('/api/admin/organizations') }

// loadRoles 查询角色列表
async function loadRoles() {
  const res = await apiGet<Role[]>('/api/admin/roles/all')
  roles.value = res
}

// getRoleName 获取角色名称
function getRoleName(roleId: string | number) {
  const role = roles.value.find(r => r.id === Number(roleId))
  return role?.name || '-'
}
// openCreate 打开新增用户弹窗。
function openCreate() {
  editingId.value = null
  pendingAvatarFile.value = null
  editingAvatarBlobUrl.value = ''
  Object.assign(form, { username: '', name: '', gender: 'unknown', birth_year: 0, birth_month: 0, id_card_no: '', org_code: '', status: 'active', role_id: 0 })
  dialogVisible.value = true
}

// uploadAvatarForCreate 新增用户时预览头像（暂存文件，保存时再上传）。
function uploadAvatarForCreate(options: UploadRequestOptions) {
  pendingAvatarFile.value = options.file
  editingAvatarBlobUrl.value = URL.createObjectURL(options.file)
  ElMessage.success('头像已选择，保存用户时自动上传')
}

// closeDialog 关闭弹窗并清理 blob URL 避免内存泄漏
function closeDialog() {
  if (editingAvatarBlobUrl.value) {
    URL.revokeObjectURL(editingAvatarBlobUrl.value)
    editingAvatarBlobUrl.value = ''
  }
  pendingAvatarFile.value = null
  dialogVisible.value = false
}
// openEdit 打开编辑用户弹窗。
async function openEdit(row: User) {
  editingId.value = row.id
  Object.assign(form, { username: row.username, name: row.name, gender: row.gender, birth_year: row.birth_year, birth_month: row.birth_month, id_card_no: row.id_card_no, org_code: row.org_code, status: row.status, role_id: row.role_id || 0 })
  // 使用 Blob URL 方式加载编辑用户的头像
  editingAvatarBlobUrl.value = ''
  if (row.has_avatar) {
    try {
      const token = localStorage.getItem('admin_token')
      const res = await fetch(`/api/admin/users/${row.id}/avatar`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      })
      if (res.ok) {
        const blob = await res.blob()
        editingAvatarBlobUrl.value = URL.createObjectURL(blob)
      }
    } catch (e) {
      console.warn('Failed to load editing avatar:', e)
    }
  }
  dialogVisible.value = true
}
// submit 保存用户基础资料。
async function submit() {
  await formRef.value?.validate()
  let userId: number
  if (editingId.value) {
    await apiPost(`/api/admin/users/${editingId.value}/update`, form)
    userId = editingId.value
  } else {
    const result = await apiPost<{ id: number }>('/api/admin/users', form)
    userId = result.id
    // 如果有暂存的头像，创建用户后上传
    if (pendingAvatarFile.value) {
      const fd = new FormData()
      fd.append('avatar', pendingAvatarFile.value)
      await apiUpload(`/api/admin/users/${userId}/avatar`, fd)
      pendingAvatarFile.value = null
    }
  }
  ElMessage.success('用户已保存')
  closeDialog()
  await load()
}
// uploadAvatar 上传头像到数据库。
async function uploadAvatar(options: UploadRequestOptions) {
  if (!editingId.value) return
  const fd = new FormData()
  fd.append('avatar', options.file)
  await apiUpload(`/api/admin/users/${editingId.value}/avatar`, fd)
  ElMessage.success('头像已上传')
  // 清除缓存并重新加载列表和编辑弹窗的头像
  if (avatarBlobUrls.value[editingId.value]) {
    URL.revokeObjectURL(avatarBlobUrls.value[editingId.value])
    delete avatarBlobUrls.value[editingId.value]
  }
  await load()
  // 重新加载编辑弹窗中的头像
  const updated = users.value.find(u => u.id === editingId.value)
  if (updated?.has_avatar) {
    const token = localStorage.getItem('admin_token')
    const res = await fetch(`/api/admin/users/${updated.id}/avatar`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {}
    })
    if (res.ok) {
      const blob = await res.blob()
      editingAvatarBlobUrl.value = URL.createObjectURL(blob)
    }
  }
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
function genderLabel(gender: string) { return gender === 'male' ? '男' : gender === 'female' ? '女' : '未知' }
function birthText(row: User) { return row.birth_year ? `${row.birth_year}-${String(row.birth_month).padStart(2, '0')}` : '-' }
function maskID(value: string) { return value ? `${value.slice(0, 6)}********${value.slice(-4)}` : '-' }
onMounted(async () => { await Promise.all([loadOrganizations(), loadRoles(), load()]) })
</script>

<style scoped>
.filters { display: flex; gap: 8px; align-items: center; }
.pager { display: flex; justify-content: flex-end; margin-top: 16px; }

.avatar-preview { position: relative; display: inline-block; }
.avatar-badge {
  position: absolute;
  bottom: -8px;
  right: -8px;
  background: #67c23a;
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}
</style>
