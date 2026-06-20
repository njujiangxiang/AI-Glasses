<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">组织管理</span>
      <div>
        <el-button @click="loadTree">刷新</el-button>
        <el-button type="primary" @click="openCreate('')">新增单位</el-button>
      </div>
    </div>

    <el-table :data="orgTree" row-key="id" :tree-props="{ children: 'children' }" stripe default-expand-all>
      <el-table-column prop="code" label="单位编码" min-width="150" />
      <el-table-column prop="name" label="单位名称" min-width="200" />
      <el-table-column label="完整层级路径" min-width="300" show-overflow-tooltip>
        <template #default="scope">
          <template v-for="(item, idx) in scope.row._path" :key="idx">
            <span v-if="idx === 0" class="path-root">{{ item.name }}</span>
            <span v-else-if="idx === scope.row._path.length - 1" class="path-current">{{ item.name }}</span>
            <span v-else class="path-item">{{ item.name }}</span>
            <span v-if="idx < scope.row._path.length - 1" class="path-separator"> > </span>
          </template>
        </template>
      </el-table-column>
      <el-table-column label="上级单位" min-width="180">
        <template #default="scope">
          {{ scope.row._parentName || '—' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="120" align="center">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
            {{ scope.row.status === 'active' ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="层级" width="80" align="center">
        <template #default="scope">
          <el-tag size="small" type="info">{{ scope.row._depth }}层</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="CreatedAt" label="创建时间" width="180" />
      <el-table-column prop="UpdatedAt" label="更新时间" width="180" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="openCreate(scope.row.code)">新增子单位</el-button>
          <el-button link type="success" @click="openEdit(scope.row)">编辑</el-button>
          <el-button v-if="scope.row.status !== 'active'" link type="warning" @click="enable(scope.row.id)">启用</el-button>
          <el-button v-else link type="warning" @click="disable(scope.row.id)">停用</el-button>
          <el-button link type="danger" @click="handleDelete(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <!-- 新增/编辑单位弹窗 -->
  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑单位' : '新增单位'" width="520px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="单位编码" prop="code">
        <el-input v-model="form.code" placeholder="如 ROOT 或 SG-001" />
      </el-form-item>
      <el-form-item label="单位名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入单位名称" />
      </el-form-item>
      <el-form-item label="上级单位" prop="parent_code">
        <el-tree-select
          v-model="form.parent_code"
          :data="parentOptions"
          :props="{ label: 'name', value: 'code' }"
          check-strictly
          placeholder="请选择上级单位"
          clearable
          :disabled="editingId !== null"
        />
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
      <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

interface Organization {
  id: number
  code: string
  name: string
  parent_code: string
  status: string
  CreatedAt: string
  children?: Organization[]
  // 平铺时附加的展示字段
  _depth?: number
  _path?: Array<{ code: string; name: string }>
  _parentName?: string
}

const formRef = ref<FormInstance>()
const orgTree = ref<Organization[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)

const form = reactive({
  code: '',
  name: '',
  parent_code: '',
  status: 'active'
})

const rules: FormRules = {
  code: [{ required: true, message: '请输入单位编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入单位名称', trigger: 'blur' }]
}

// 从组织树中平铺所有可作父级的单位（排除当前编辑的单位）
const parentOptions = computed(() => {
  const options: Organization[] = []
  const flatten = (nodes: Organization[]) => {
    nodes.forEach(node => {
      if (node.id !== editingId.value) {
        options.push({ ...node, children: undefined })
        if (node.children) flatten(node.children)
      }
    })
  }
  flatten(orgTree.value)
  // 添加"顶级单位"选项（code 为空字符串，与后端 parent_code='' 对应）
  return [{ id: 0, name: '顶级单位', code: '', parent_code: '', status: 'active', CreatedAt: '' } as Organization, ...options]
})

async function loadTree() {
  const res = await apiGet<Organization[]>('/api/admin/organizations/tree')
  // 递归附加展示字段（深度、路径、父级名称）到每个节点
  attachDisplayFields(res, 1, [])
  orgTree.value = res
}

// 递归附加展示字段到树形结构的每个节点
function attachDisplayFields(nodes: Organization[], depth: number, path: Array<{ code: string; name: string }>) {
  for (const node of nodes) {
    const currentPath = [...path, { code: node.code, name: node.name }]
    node._depth = depth
    node._path = currentPath
    node._parentName = path.length > 0 ? path[path.length - 1].name : ''
    if (node.children && node.children.length > 0) {
      attachDisplayFields(node.children, depth + 1, currentPath)
    }
  }
}

function openCreate(parentCode: string) {
  editingId.value = null
  form.code = ''
  form.name = ''
  form.parent_code = parentCode  // '' 表示顶级单位，否则为父级单位的 code
  form.status = 'active'
  dialogVisible.value = true
}

async function openEdit(row: Organization) {
  editingId.value = row.id
  form.code = row.code
  form.name = row.name
  form.parent_code = row.parent_code
  form.status = row.status
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (editingId.value) {
      await apiPost(`/api/admin/organizations/${editingId.value}/update`, form)
      ElMessage.success('编辑成功')
    } else {
      await apiPost('/api/admin/organizations', form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadTree()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: Organization) {
  if (row.children && row.children.length > 0) {
    ElMessage.warning('该单位下有子单位，请先删除子单位')
    return
  }
  try {
    await ElMessageBox.confirm(`确定删除单位【${row.name}】吗？`, '提示', {
      type: 'warning'
    })
    await apiPost(`/api/admin/organizations/${row.id}/delete`)
    ElMessage.success('删除成功')
    loadTree()
  } catch {
    // 用户取消
  }
}

async function enable(id: number) {
  await apiPost(`/api/admin/organizations/${id}/enable`)
  ElMessage.success('单位已启用')
  await loadTree()
}

async function disable(id: number) {
  await apiPost(`/api/admin/organizations/${id}/disable`)
  ElMessage.success('单位已停用')
  await loadTree()
}

onMounted(() => {
  loadTree()
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

.path-breadcrumb {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.path-root {
  color: #606266;
  font-weight: 600;
}

.path-item {
  color: #606266;
}

.path-current {
  color: #409eff;
  font-weight: 600;
}

.path-separator {
  color: #c0c4cc;
}
</style>
