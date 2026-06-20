<template>
  <el-card shadow="never">
    <template #header>
      <div class="card-header">
        <span>角色管理</span>
        <el-button type="primary" @click="openCreate">新增角色</el-button>
      </div>
    </template>

    <el-table :data="roles" stripe>
      <el-table-column prop="name" label="角色名称" min-width="150" />
      <el-table-column prop="code" label="角色编码" width="150" />
      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="member_count" label="成员数" width="100" align="center" />
      <el-table-column prop="data_scope" label="数据范围" width="150" align="center">
        <template #default="scope">
          <el-tag :type="getDataScopeTagType(scope.row.data_scope)" size="small">
            {{ getDataScopeText(scope.row.data_scope) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="100" align="center" />
      <el-table-column label="状态" width="100" align="center">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'disabled' ? 'info' : 'success'">
            {{ scope.row.status === 'disabled' ? '禁用' : '启用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="scope">{{ formatDate(scope.row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button link type="success" @click="openPermission(scope.row)">分配权限</el-button>
          <el-button link type="danger" @click="handleDelete(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <!-- 新增/编辑角色弹窗 -->
  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑角色' : '新增角色'" width="600px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="角色名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入角色名称" maxlength="30" />
      </el-form-item>
      <el-form-item label="角色编码" prop="code">
        <el-input v-model="form.code" placeholder="请输入角色编码" maxlength="30" />
      </el-form-item>
      <el-form-item label="排序" prop="sort">
        <el-input-number v-model="form.sort" :min="0" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio value="active">启用</el-radio>
          <el-radio value="disabled">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="数据范围" prop="data_scope">
        <el-select v-model="form.data_scope" placeholder="请选择数据范围">
          <el-option label="全部数据" value="all">
            <div class="flex flex-col">
              <span>全部数据</span>
              <span style="font-size: 12px; color: #909399">可查看所有组织和所有用户的数据</span>
            </div>
          </el-option>
          <el-option label="本组织及下级" value="org_and_sub">
            <div class="flex flex-col">
              <span>本组织及下级</span>
              <span style="font-size: 12px; color: #909399">可查看本组织及所有下级组织的数据</span>
            </div>
          </el-option>
          <el-option label="仅本组织" value="org_only">
            <div class="flex flex-col">
              <span>仅本组织</span>
              <span style="font-size: 12px; color: #909399">只能查看本组织内部的数据</span>
            </div>
          </el-option>
          <el-option label="仅自己" value="self_only">
            <div class="flex flex-col">
              <span>仅自己</span>
              <span style="font-size: 12px; color: #909399">只能查看自己创建或分配给自己的数据</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="描述" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" maxlength="200" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确定</el-button>
    </template>
  </el-dialog>

  <!-- 分配权限弹窗 -->
  <el-dialog v-model="permissionDialogVisible" title="分配权限" width="700px">
    <el-tree
      ref="treeRef"
      :data="menuTree"
      :props="{ label: 'name', children: 'children' }"
      show-checkbox
      node-key="id"
      :default-checked-keys="checkedMenuIds"
      default-expand-all
    />
    <template #footer>
      <el-button @click="permissionDialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submitPermission">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

interface Role {
  id: number
  name: string
  code: string
  description: string
  data_scope: string
  member_count: number
  sort: number
  status: string
  created_at: string
  updated_at: string
  menus: number[]
}

interface Menu {
  id: number
  pid: number
  type: string
  name: string
  icon: string
  sort: number
  perms: string
  path: string
  component: string
  visible: boolean
  status: string
  children?: Menu[]
}

const formRef = ref<FormInstance>()
const treeRef = ref() // 树形选择器的引用
const roles = ref<Role[]>([])
const menuTree = ref<Menu[]>([])
const checkedMenuIds = ref<number[]>([])
const dialogVisible = ref(false)
const permissionDialogVisible = ref(false)
const editingId = ref<number | null>(null)
const editingRoleId = ref<number | null>(null)
const submitting = ref(false)

const form = reactive({
  name: '',
  code: '',
  description: '',
  data_scope: 'org_only',
  sort: 0,
  status: 'active'
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
}

async function loadRoles() {
  const res = await apiGet<{ items: Role[]; total: number }>('/api/admin/roles')
  roles.value = res.items
}

async function loadMenuTree() {
  const res = await apiGet<Menu[]>('/api/admin/menus/tree')
  menuTree.value = res
}

function openCreate() {
  editingId.value = null
  form.name = ''
  form.code = ''
  form.description = ''
  form.data_scope = 'org_only'
  form.sort = 0
  form.status = 'active'
  dialogVisible.value = true
}

async function openEdit(row: Role) {
  editingId.value = row.id
  form.name = row.name
  form.code = row.code
  form.description = row.description
  form.data_scope = row.data_scope || 'org_only'
  form.sort = row.sort
  form.status = row.status
  dialogVisible.value = true
}

async function openPermission(row: Role) {
  editingRoleId.value = row.id
  checkedMenuIds.value = []
  const res = await apiGet<Role>(`/api/admin/roles/${row.id}`)
  checkedMenuIds.value = res.menus || []
  permissionDialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (editingId.value) {
      await apiPost(`/api/admin/roles/${editingId.value}/update`, form)
      ElMessage.success('编辑成功')
    } else {
      await apiPost('/api/admin/roles', form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadRoles()
  } finally {
    submitting.value = false
  }
}

async function submitPermission() {
  submitting.value = true
  try {
    const checkedKeys = treeRef.value?.getCheckedKeys() || []

    await apiPost(`/api/admin/roles/${editingRoleId.value}/menus`, {
      menu_ids: checkedKeys.join(',')
    })
    ElMessage.success('权限分配成功')
    permissionDialogVisible.value = false
    loadRoles()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: Role) {
  try {
    await ElMessageBox.confirm(`确定删除角色【${row.name}】吗？`, '提示', {
      type: 'warning'
    })
    await apiPost(`/api/admin/roles/${row.id}/delete`)
    ElMessage.success('删除成功')
    loadRoles()
  } catch {
    // 用户取消
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

function getDataScopeText(scope: string): string {
  const map: Record<string, string> = {
    all: '全部数据',
    org_and_sub: '本组织及下级',
    org_only: '仅本组织',
    self_only: '仅自己'
  }
  return map[scope] || scope
}

function getDataScopeTagType(scope: string): string {
  const map: Record<string, string> = {
    all: 'danger',
    org_and_sub: 'warning',
    org_only: 'success',
    self_only: 'info'
  }
  return map[scope] || 'info'
}

onMounted(() => {
  loadRoles()
  loadMenuTree()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
