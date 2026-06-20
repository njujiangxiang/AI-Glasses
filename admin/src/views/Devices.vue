<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">设备管理</span>
      <div>
        <el-button @click="refresh">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增设备</el-button>
      </div>
    </div>

    <el-table :data="devices" stripe v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="serial_no" label="序列号" width="200" />
      <el-table-column prop="name" label="名称" width="200" />
      <el-table-column label="组织机构" width="200">
        <template #default="scope">
          <span v-if="scope.row.org_code">{{ orgLabel(scope.row.org_code) }}</span>
          <span v-else class="muted-text">—</span>
        </template>
      </el-table-column>
      <el-table-column label="绑定用户" width="150">
        <template #default="scope">
          <span v-if="scope.row.bound_user_name">{{ scope.row.bound_user_name }}</span>
          <span v-else class="muted-text">未绑定</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="120" align="center">
        <template #default="scope">
          <el-tag :type="statusType(scope.row.status)">
            {{ statusLabel(scope.row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="320" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button link type="success" @click="doEnable(scope.row)" v-if="scope.row.status === 'disabled'">启用</el-button>
          <el-button link type="warning" @click="doDisable(scope.row)" v-else-if="scope.row.status === 'active'">停用</el-button>
          <el-button link type="danger" @click="doDelete(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑设备弹窗 -->
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑设备' : '新增设备'" width="560px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="设备名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入设备名称" />
        </el-form-item>

        <el-form-item label="组织机构" prop="org_code">
          <el-select v-model="form.org_code" filterable placeholder="请选择组织机构" style="width: 100%;">
            <el-option v-for="org in organizations" :key="org.code" :label="org.name" :value="org.code" />
          </el-select>
        </el-form-item>

        <el-form-item label="绑定用户">
          <div style="display: flex; gap: 10px; align-items: center;">
            <el-input v-model="selectedUserName" placeholder="请选择用户" readonly @click="openUserLookup" style="flex: 1;">
              <template #suffix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button v-if="form.bound_user_id" type="info" @click="clearUser">清除</el-button>
          </div>
        </el-form-item>

        <el-form-item label="序列号">
          <el-input v-model="generatedSerial" disabled placeholder="自动生成" style="background: #f5f7fa;">
            <template #prefix>
              <el-icon><InfoFilled /></el-icon>
            </template>
          </el-input>
          <div class="form-hint">序列号由"AI 眼镜设备编码"（GLASS）规则自动生成，用户不可修改</div>
        </el-form-item>

        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" placeholder="请选择状态" style="width: 100%;">
            <el-option label="待绑定" value="pending" />
            <el-option label="启用" value="active" />
            <el-option label="停用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 用户选择弹窗 -->
    <el-dialog v-model="userLookupVisible" title="选择用户" width="700px" @open="loadUserPage">
      <div class="lookup-toolbar">
        <el-input v-model="userQuery.keyword" placeholder="搜索用户名/姓名" clearable style="width: 200px;" @input="onUserSearch">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="userQuery.status" placeholder="状态" clearable style="width: 120px;" @change="loadUserPage">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="disabled" />
        </el-select>
        <el-select v-model="userQuery.org_code" placeholder="组织" clearable style="width: 160px;" @change="loadUserPage">
          <el-option v-for="org in organizations" :key="org.code" :label="org.name" :value="org.code" />
        </el-select>
      </div>
      <el-table :data="users" stripe v-loading="userLoading" @row-click="selectUser" style="max-height: 360px;">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="name" label="姓名" width="150" />
        <el-table-column label="所属组织" width="200">
          <template #default="scope">
            <span v-if="scope.row.org_name">{{ scope.row.org_name }}</span>
            <span v-else class="muted-text">—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'" size="small">
              {{ scope.row.status === 'active' ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div class="lookup-pagination">
        <el-pagination v-model:current-page="userQuery.page" :page-size="userQuery.page_size" :total="userTotal" layout="prev, pager, next, jumper, total" @current-change="loadUserPage" />
      </div>
      <template #footer>
        <el-button @click="userLookupVisible = false">取消</el-button>
        <el-button type="primary" @click="userLookupVisible = false">确定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'
import { InfoFilled, Search } from '@element-plus/icons-vue'

type Device = {
  id: number
  serial_no: string
  name: string
  org_code: string
  status: string
  bound_user_id?: number
  bound_user_name?: string
  created_at: string
  CreatedAt: string
}

type Organization = { code: string; name: string }

type User = {
  id: number
  username: string
  name: string
  org_name?: string
  role_name?: string
  status: string
}

const devices = ref<Device[]>([])
const organizations = ref<Organization[]>([])
const users = ref<User[]>([])
const loading = ref(false)
const userLoading = ref(false)
const dialogVisible = ref(false)
const userLookupVisible = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const generatedSerial = ref('')
const selectedUserName = ref('')
const userTotal = ref(0)
const userQuery = reactive({
  keyword: '',
  status: '',
  org_code: '',
  page: 1,
  page_size: 10
})
let userSearchTimer: ReturnType<typeof setTimeout> | null = null

const form = reactive({
  name: '',
  org_code: '',
  bound_user_id: undefined as number | undefined,
  status: 'pending'
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  org_code: [{ required: true, message: '请选择组织机构', trigger: 'change' }]
}

// 加载设备列表
async function refresh() {
  loading.value = true
  try {
    devices.value = await apiGet<Device[]>('/api/admin/devices')
    // 格式化处理：创建时间 + 绑定用户名称
    devices.value.forEach(d => {
      if (d.CreatedAt) {
        const dt = new Date(d.CreatedAt)
        d.created_at = dt.toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
      }
    })
  } finally {
    loading.value = false
  }
}

// 加载组织机构列表
async function loadOrganizations() {
  organizations.value = await apiGet<Organization[]>('/api/admin/organizations')
}

// 加载用户列表（分页 + 搜索）
async function loadUserPage() {
  userLoading.value = true
  try {
    const params = new URLSearchParams({
      page: String(userQuery.page),
      page_size: String(userQuery.page_size)
    })
    if (userQuery.keyword) params.append('keyword', userQuery.keyword)
    if (userQuery.status) params.append('status', userQuery.status)
    if (userQuery.org_code) params.append('org_code', userQuery.org_code)

    const data = await apiGet<{ items: User[]; total: number }>(`/api/admin/users?${params}`)
    users.value = data.items || []
    userTotal.value = data.total || 0
  } finally {
    userLoading.value = false
  }
}

// 搜索防抖
function onUserSearch() {
  if (userSearchTimer) clearTimeout(userSearchTimer)
  userSearchTimer = setTimeout(() => {
    userQuery.page = 1
    loadUserPage()
  }, 300)
}

// 组织机构标签
const orgLabelMap = computed(() => {
  const map = new Map<string, string>()
  organizations.value.forEach(org => map.set(org.code, org.name))
  return map
})

function orgLabel(code: string): string {
  return orgLabelMap.value.get(code) || code
}

// 状态标签
function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: '待绑定',
    active: '启用',
    revoked: '撤销',
    lost_disabled: '丢失禁用',
    disabled: '停用'
  }
  return labels[status] || status
}

function statusType(status: string): string {
  const types: Record<string, string> = {
    pending: 'info',
    active: 'success',
    revoked: 'danger',
    lost_disabled: 'danger',
    disabled: 'info'
  }
  return types[status] || ''
}

// 打开新增弹窗
function openCreate() {
  editingId.value = null
  generatedSerial.value = ''
  selectedUserName.value = ''
  form.bound_user_id = undefined
  // 生成序列号：调用业务编码生成接口
  generateSerial()
  dialogVisible.value = true
}

// 打开编辑弹窗
async function openEdit(row: Device) {
  editingId.value = row.id
  form.name = row.name
  form.org_code = row.org_code || ''
  form.bound_user_id = row.bound_user_id
  form.status = row.status
  generatedSerial.value = row.serial_no
  selectedUserName.value = row.bound_user_name || ''
  dialogVisible.value = true
}

// 生成序列号（调用业务编码接口）
async function generateSerial() {
  try {
    const result = await apiPost<{ code: string }>('/api/admin/business-codes/generate', {
      code: 'GLASS'
    })
    generatedSerial.value = result.code
  } catch (err: any) {
    ElMessage.error('生成序列号失败: ' + (err?.message || '未知错误'))
    generatedSerial.value = ''
  }
}

// 打开用户选择弹窗
function openUserLookup() {
  userQuery.page = 1
  userQuery.keyword = ''
  userQuery.status = ''
  userQuery.org_code = ''
  userLookupVisible.value = true
}

// 选择用户
function selectUser(user: User) {
  form.bound_user_id = user.id
  selectedUserName.value = user.name || user.username
  userLookupVisible.value = false
}

// 清除用户
function clearUser() {
  form.bound_user_id = undefined
  selectedUserName.value = ''
}

// 重置表单
function resetForm() {
  formRef.value?.resetFields()
  form.name = ''
  form.org_code = ''
  form.bound_user_id = undefined
  form.status = 'pending'
  generatedSerial.value = ''
  selectedUserName.value = ''
  editingId.value = null
}

// 提交表单
async function submit() {
  await formRef.value?.validate()
  if (!generatedSerial.value) {
    ElMessage.error('序列号尚未生成，请检查业务编码配置')
    return
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await apiPost(`/api/admin/devices/${editingId.value}/update`, {
        name: form.name,
        org_code: form.org_code,
        status: form.status,
        bound_user_id: form.bound_user_id
      })
      ElMessage.success('设备已更新')
    } else {
      await apiPost('/api/admin/devices', {
        serial_no: generatedSerial.value,
        name: form.name,
        org_code: form.org_code,
        status: form.status,
        bound_user_id: form.bound_user_id
      })
      ElMessage.success('设备已添加')
    }
    dialogVisible.value = false
    await refresh()
  } catch (err: any) {
    ElMessage.error(err?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

// 删除设备
async function doDelete(row: Device) {
  try {
    await ElMessageBox.confirm(`确定删除设备【${row.name}】吗？`, '提示', {
      type: 'warning'
    })
    await apiPost(`/api/admin/devices/${row.id}/delete`)
    ElMessage.success('删除成功')
    await refresh()
  } catch {
    // 用户取消
  }
}

// 启用设备
async function doEnable(row: Device) {
  await apiPost(`/api/admin/devices/${row.id}/enable`)
  ElMessage.success('设备已启用')
  await refresh()
}

// 停用设备
async function doDisable(row: Device) {
  await apiPost(`/api/admin/devices/${row.id}/disable`)
  ElMessage.success('设备已停用')
  await refresh()
}

onMounted(() => {
  refresh()
  loadOrganizations()
})
</script>

<style scoped>
.page-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-right: 16px;
}

.muted-text {
  color: #909399;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.lookup-toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  align-items: center;
}

.lookup-pagination {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}
</style>
