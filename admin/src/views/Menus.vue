<template>
  <el-card shadow="never">
    <template #header>
      <div class="card-header">
        <span>菜单权限管理</span>
        <el-button type="primary" @click="openCreate(0)">新增菜单</el-button>
      </div>
    </template>

    <el-table :data="menuTree" row-key="id" :tree-props="{ children: 'children' }" stripe default-expand-all>
      <el-table-column prop="name" label="菜单名称" min-width="200">
        <template #default="scope">
          <el-icon v-if="scope.row.icon" style="margin-right: 8px; color: #606266">
            <component :is="getIconComponent(scope.row.icon)" />
          </el-icon>
          {{ scope.row.name }}
          <el-tag v-if="scope.row.type === 'M'" type="info" size="small" style="margin-left: 8px">目录</el-tag>
          <el-tag v-else-if="scope.row.type === 'C'" type="success" size="small" style="margin-left: 8px">菜单</el-tag>
          <el-tag v-else type="warning" size="small" style="margin-left: 8px">按钮</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="100" align="center" />
      <el-table-column prop="perms" label="权限标识" min-width="180" />
      <el-table-column prop="path" label="路由地址" min-width="180" />
      <el-table-column prop="component" label="前端组件" min-width="200" show-overflow-tooltip />
      <el-table-column label="状态" width="100" align="center">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'disabled' ? 'info' : 'success'">
            {{ scope.row.status === 'disabled' ? '禁用' : '启用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="显示" width="100" align="center">
        <template #default="scope">
          <el-tag :type="scope.row.visible ? 'success' : 'info'">
            {{ scope.row.visible ? '显示' : '隐藏' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="openCreate(scope.row.id)">新增子菜单</el-button>
          <el-button link type="success" @click="openEdit(scope.row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <!-- 新增/编辑菜单弹窗 -->
  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑菜单' : '新增菜单'" width="700px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="上级菜单" prop="pid">
            <el-tree-select
              v-model="form.pid"
              :data="parentOptions"
              :props="{ label: 'name', value: 'id' }"
              check-strictly
              placeholder="请选择上级菜单"
              clearable
              :disabled="editingId !== null"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="菜单类型" prop="type">
            <el-select v-model="form.type" placeholder="请选择菜单类型">
              <el-option label="目录" value="M" />
              <el-option label="菜单" value="C" />
              <el-option label="按钮" value="A" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="菜单名称" prop="name">
            <el-input v-model="form.name" placeholder="请输入菜单名称" maxlength="30" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="菜单编码" prop="code">
            <el-input v-model="form.code" placeholder="请输入菜单编码" maxlength="30" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="图标" prop="icon">
            <el-select v-model="form.icon" placeholder="请选择图标" clearable filterable class="icon-select">
              <el-option v-for="item in iconOptions" :key="item.value" :label="`${item.label} ${item.value}`" :value="item.value">
                <span class="icon-option">
                  <el-icon><component :is="getIconComponent(item.value)" /></el-icon>
                  <span class="icon-option-name">{{ item.label }}</span>
                  <span class="icon-option-value">{{ item.value }}</span>
                </span>
              </el-option>
            </el-select>
            <div v-if="form.icon" class="selected-icon-preview">
              <el-icon><component :is="getIconComponent(form.icon)" /></el-icon>
              <span>{{ form.icon }}</span>
            </div>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="排序" prop="sort">
            <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="权限标识" prop="perms">
            <el-input v-model="form.perms" placeholder="如: system:role:list" maxlength="100" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="路由地址" prop="path" v-if="form.type !== 'A'">
        <el-input v-model="form.path" placeholder="请输入路由地址，如: /system/roles" maxlength="200" />
      </el-form-item>

      <el-form-item label="前端组件" prop="component" v-if="form.type === 'C'">
        <el-input v-model="form.component" placeholder="请输入组件路径，如: Roles" maxlength="200" />
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item label="是否显示" prop="visible">
            <el-radio-group v-model="form.visible">
              <el-radio :value="true">显示</el-radio>
              <el-radio :value="false">隐藏</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="是否禁用" prop="status">
            <el-radio-group v-model="form.status">
              <el-radio value="active">启用</el-radio>
              <el-radio value="disabled">禁用</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="是否缓存" prop="is_cache" v-if="form.type === 'C'">
            <el-radio-group v-model="form.is_cache">
              <el-radio :value="false">不缓存</el-radio>
              <el-radio :value="true">缓存</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'
import { getIconComponent, iconOptions } from '@/iconCatalog'

interface Menu {
  id: number
  pid: number
  type: string
  name: string
  code: string
  icon: string
  sort: number
  perms: string
  path: string
  component: string
  is_cache: boolean
  visible: boolean
  status: string
  created_at: string
  updated_at: string
  children?: Menu[]
}

const formRef = ref<FormInstance>()
const menuTree = ref<Menu[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)

const form = reactive({
  pid: 0,
  type: 'C',
  name: '',
  code: '',
  icon: '',
  sort: 0,
  perms: '',
  path: '',
  component: '',
  is_cache: false,
  visible: true,
  status: 'active'
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择菜单类型', trigger: 'change' }]
}

const parentOptions = computed(() => {
  const options: Menu[] = []
  const flatten = (menus: Menu[]) => {
    menus.forEach(menu => {
      if (menu.type === 'M' || menu.type === 'C') {
        options.push(menu)
        if (menu.children) flatten(menu.children)
      }
    })
  }
  flatten(menuTree.value)
  return [{ id: 0, name: '顶级菜单' } as Menu, ...options]
})

async function loadMenuTree() {
  const res = await apiGet<Menu[]>('/api/admin/menus/tree')
  menuTree.value = res
}

function openCreate(pid: number) {
  editingId.value = null
  form.pid = pid
  form.type = pid === 0 ? 'M' : 'C'
  form.name = ''
  form.code = ''
  form.icon = ''
  form.sort = 0
  form.perms = ''
  form.path = ''
  form.component = ''
  form.is_cache = false
  form.visible = true
  form.status = 'active'
  dialogVisible.value = true
}

async function openEdit(row: Menu) {
  editingId.value = row.id
  form.pid = row.pid
  form.type = row.type
  form.name = row.name
  form.code = row.code
  form.icon = row.icon
  form.sort = row.sort
  form.perms = row.perms
  form.path = row.path
  form.component = row.component
  form.is_cache = row.is_cache
  form.visible = row.visible
  form.status = row.status
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (editingId.value) {
      await apiPost(`/api/admin/menus/${editingId.value}/update`, form)
      ElMessage.success('编辑成功')
    } else {
      await apiPost('/api/admin/menus', form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    loadMenuTree()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: Menu) {
  if (row.children && row.children.length > 0) {
    ElMessage.warning('该菜单下有子菜单，请先删除子菜单')
    return
  }
  try {
    await ElMessageBox.confirm(`确定删除菜单【${row.name}】吗？`, '提示', {
      type: 'warning'
    })
    await apiPost(`/api/admin/menus/${row.id}/delete`)
    ElMessage.success('删除成功')
    loadMenuTree()
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  loadMenuTree()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.icon-select {
  width: 100%;
}

.icon-option {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.icon-option-name {
  min-width: 72px;
  color: #303133;
}

.icon-option-value {
  color: #909399;
  font-size: 12px;
}

.selected-icon-preview {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  color: #606266;
  font-size: 13px;
}
</style>
