<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">业务编码配置</span>
      <div>
        <el-button @click="load">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增编码规则</el-button>
      </div>
    </div>
    <el-table :data="businessCodes" stripe row-key="id">
      <el-table-column prop="name" label="编码名称" />
      <el-table-column prop="code" label="代码" width="120" />
      <el-table-column label="日期" width="120">
        <template #default="scope">
          {{ scope.row.use_date ? scope.row.date_format : '不添加' }}
        </template>
      </el-table-column>
      <el-table-column prop="seq_padding" label="流水号位数" width="120" />
      <el-table-column label="分隔符" width="120">
        <template #default="scope">
          {{ scope.row.use_separator ? scope.row.separator : '无' }}
        </template>
      </el-table-column>
      <el-table-column label="组织编码" width="200">
        <template #default="scope">
          <template v-if="scope.row.use_org_code">
            <span v-if="scope.row.org_source === 'current'">当前用户机构</span>
            <span v-else>{{ orgLabel(scope.row.org_code) }}</span>
          </template>
          <span v-else class="muted-text">未启用</span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="120">
        <template #default="scope">
          <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
            {{ scope.row.status === 'active' ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="CreatedAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="320">
        <template #default="scope">
          <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
          <el-button link type="success" @click="testGenerate(scope.row)">试生成</el-button>
          <el-button v-if="scope.row.status !== 'active'" link type="success" @click="enable(scope.row.id)">启用</el-button>
          <el-button v-else link type="warning" @click="disable(scope.row.id)">停用</el-button>
          <el-button link type="danger" @click="remove(scope.row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑编码规则' : '新增编码规则'" width="600px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item label="编码名称" prop="name">
        <el-input v-model="form.name" placeholder="如 任务单编号" />
      </el-form-item>
      <el-form-item label="代码" prop="code">
        <el-input v-model="form.code" placeholder="如 TK，保存时自动转大写" />
      </el-form-item>
      <el-form-item label="添加日期" prop="use_date">
        <el-switch v-model="form.use_date" />
      </el-form-item>
      <el-form-item v-if="form.use_date" label="日期格式" prop="date_format">
        <el-select v-model="form.date_format">
          <el-option label="yyyyMMdd（20260616）" value="yyyyMMdd" />
          <el-option label="yyMMdd（260616）" value="yyMMdd" />
        </el-select>
      </el-form-item>
      <el-form-item label="流水号位数" prop="seq_padding">
        <el-input-number v-model="form.seq_padding" :min="1" :max="12" />
      </el-form-item>
      <el-form-item label="使用分隔符" prop="use_separator">
        <el-switch v-model="form.use_separator" />
      </el-form-item>
      <el-form-item v-if="form.use_separator" label="分隔符" prop="separator">
        <el-input v-model="form.separator" placeholder="如 - 或 /" maxlength="8" />
      </el-form-item>
      <el-form-item label="添加组织编码" prop="use_org_code">
        <el-switch v-model="form.use_org_code" />
      </el-form-item>
      <el-form-item v-if="form.use_org_code" label="组织编码来源" prop="org_source">
        <el-radio-group v-model="form.org_source">
          <el-radio label="fixed">固定组织编码</el-radio>
          <el-radio label="current">当前登录人所在机构</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.use_org_code && form.org_source === 'fixed'" label="组织机构编码" prop="org_code">
        <el-select v-model="form.org_code" clearable filterable placeholder="请选择公司/组织编码">
          <el-option v-for="org in organizations" :key="org.code" :label="`${org.name}（${org.code}）`" :value="org.code" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status">
          <el-option label="启用" value="active" />
          <el-option label="停用" value="disabled" />
        </el-select>
      </el-form-item>
      <el-form-item label="预览">
        <div class="preview-text">{{ previewCode }}</div>
        <div class="preview-hint">实际生成时使用 Asia/Shanghai 时区</div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submit">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="generateDialogVisible" title="试生成结果" width="400px">
    <div v-if="generatedCode">
      <p>生成的编号：</p>
      <div class="generated-code">{{ generatedCode }}</div>
      <p class="generate-hint">该操作已消耗一个真实流水号</p>
    </div>
    <div v-else-if="generateError">
      <el-alert type="error" :title="generateError" show-icon />
    </div>
    <template #footer>
      <el-button @click="generateDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGet, apiPost } from '@/api/client'

type BusinessCode = {
  id: number
  name: string
  code: string
  use_date: boolean
  date_format: string
  seq_padding: number
  separator: string
  use_separator: boolean
  use_org_code: boolean
  org_source: string
  org_code: string
  status: string
  CreatedAt: string
}

type Organization = { id: number; code: string; name: string }

const businessCodes = ref<BusinessCode[]>([])
const organizations = ref<Organization[]>([])
const dialogVisible = ref(false)
const generateDialogVisible = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const generatedCode = ref<string>('')
const generateError = ref<string>('')

const form = reactive({
  name: '',
  code: '',
  use_date: true,
  date_format: 'yyyyMMdd',
  seq_padding: 4,
  separator: '',
  use_separator: false,
  use_org_code: false,
  org_source: 'fixed',
  org_code: '',
  status: 'active'
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入编码名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入代码', trigger: 'blur' },
    { pattern: /^[A-Za-z0-9_-]{1,64}$/, message: '只能包含字母、数字、下划线和中划线', trigger: 'blur' }
  ],
  date_format: [{ required: true, message: '请选择日期格式', trigger: 'change' }],
  seq_padding: [{ required: true, message: '请输入流水号位数', trigger: 'blur' }],
  org_code: [
    {
      validator: (_rule, value, callback) => {
        if (form.use_org_code && !value) callback(new Error('请选择组织机构编码'))
        else callback()
      },
      trigger: 'change'
    }
  ]
}

// 使用 Asia/Shanghai 时区生成预览代码
const previewCode = computed(() => {
  const now = new Date()
  const shanghaiTime = new Date(now.toLocaleString('en-US', { timeZone: 'Asia/Shanghai' }))
  const year = shanghaiTime.getFullYear()
  const month = String(shanghaiTime.getMonth() + 1).padStart(2, '0')
  const day = String(shanghaiTime.getDate()).padStart(2, '0')
  const dateStr = `${year}${month}${day}`
  const seqStr = '1'.padStart(form.seq_padding, '0')
  const code = form.code.toUpperCase()
  const parts = [code]
  if (form.use_org_code) {
    if (form.org_source === 'current') {
      parts.push('[当前用户组织]')
    } else if (form.org_code) {
      parts.push(form.org_code)
    }
  }
  if (form.use_date) parts.push(dateStr)
  parts.push(seqStr)

  if (form.use_separator && form.separator) {
    return parts.join(form.separator)
  }
  return parts.join('')
})

const orgNameMap = computed(() => {
  const map = new Map<string, string>()
  organizations.value.forEach((org) => map.set(org.code, org.name))
  return map
})

// load 查询业务编码配置列表并刷新表格。
async function load() {
  businessCodes.value = await apiGet<BusinessCode[]>('/api/admin/business-codes')
}

async function loadOrganizations() {
  organizations.value = await apiGet<Organization[]>('/api/admin/organizations')
}

function orgLabel(code: string) {
  if (!code) return '未选择'
  const name = orgNameMap.value.get(code)
  return name ? `${name}（${code}）` : code
}

// openCreate 打开新增编码规则弹窗。
function openCreate() {
  editingId.value = null
  Object.assign(form, {
    name: '',
    code: '',
    use_date: true,
    date_format: 'yyyyMMdd',
    seq_padding: 4,
    separator: '',
    use_separator: false,
    use_org_code: false,
    org_source: 'fixed',
    org_code: '',
    status: 'active'
  })
  dialogVisible.value = true
}

// openEdit 打开编辑编码规则弹窗。
function openEdit(row: BusinessCode) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    code: row.code,
    use_date: row.use_date ?? true,
    date_format: row.date_format,
    seq_padding: row.seq_padding,
    separator: row.separator,
    use_separator: row.use_separator,
    use_org_code: row.use_org_code,
    org_source: row.org_source || 'fixed',
    org_code: row.org_code || '',
    status: row.status
  })
  dialogVisible.value = true
}

// submit 保存编码规则，校验通过后调用接口，成功提示并关闭弹窗。
async function submit() {
  try {
    await formRef.value?.validate()
    const payload = {
      ...form,
      org_code: form.use_org_code ? form.org_code : '',
      org_source: form.use_org_code ? form.org_source : 'fixed'
    }
    if (editingId.value) {
      await apiPost(`/api/admin/business-codes/${editingId.value}/update`, payload)
    } else {
      await apiPost('/api/admin/business-codes', payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await load()
  } catch (err: any) {
    if (err?.message) ElMessage.error(err.message)
  }
}

// enable 启用编码规则。
async function enable(id: number) {
  await apiPost(`/api/admin/business-codes/${id}/enable`)
  ElMessage.success('编码规则已启用')
  await load()
}

// disable 停用编码规则。
async function disable(id: number) {
  await apiPost(`/api/admin/business-codes/${id}/disable`)
  ElMessage.success('编码规则已停用')
  await load()
}

// remove 删除编码规则。
async function remove(id: number) {
  await ElMessageBox.confirm('确定删除该编码规则吗？', '提示', { type: 'warning' })
  await apiPost(`/api/admin/business-codes/${id}/delete`)
  ElMessage.success('编码规则已删除')
  await load()
}

// testGenerate 调用生成接口测试编号生成。
async function testGenerate(row: BusinessCode) {
  generatedCode.value = ''
  generateError.value = ''
  generateDialogVisible.value = true

  try {
    const result = await apiPost<{ code: string }>('/api/admin/business-codes/generate', {
      code: row.code
    })
    generatedCode.value = result.code
  } catch (err: any) {
    generateError.value = err.message || '生成失败'
  }
}

onMounted(() => {
  load()
  loadOrganizations()
})
</script>

<style scoped>
.muted-text {
  color: #909399;
}

.preview-text {
  font-family: monospace;
  font-size: 18px;
  font-weight: bold;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}
.preview-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.generated-code {
  font-family: monospace;
  font-size: 24px;
  font-weight: bold;
  padding: 16px;
  background: #f0f9ff;
  border: 2px solid #409eff;
  border-radius: 8px;
  text-align: center;
  margin: 16px 0;
}
.generate-hint {
  font-size: 12px;
  color: #e6a23c;
  text-align: center;
}
</style>
