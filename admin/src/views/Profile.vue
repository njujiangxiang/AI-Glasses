<template>
  <el-card shadow="never">
    <template #header>
      <span class="card-title">个人中心</span>
    </template>

    <div class="profile-container">
      <!-- 头像区域 -->
      <div class="avatar-section">
        <div class="avatar-wrapper">
          <el-avatar v-if="avatarBlobUrl" :size="120" :src="avatarBlobUrl" />
          <el-avatar v-else :size="120">{{ userInitial }}</el-avatar>
          <div v-if="avatarBlobUrl" class="avatar-status">✓ 已设置头像</div>
        </div>
        <div class="avatar-actions">
          <el-upload :http-request="uploadAvatar" :show-file-list="false" accept="image/png,image/jpeg,image/webp">
            <el-button type="primary">{{ avatarBlobUrl ? '更换头像' : '上传头像' }}</el-button>
          </el-upload>
        </div>
      </div>

      <!-- 个人信息表单 -->
      <div class="form-section">
        <el-form ref="formRef" :model="form" :rules="rules" label-width="110px" style="max-width: 500px;">
          <el-form-item label="用户名">
            <el-input v-model="form.username" disabled />
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
          <el-form-item label="所属单位">
            <el-input v-model="orgName" disabled />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="submit">保存修改</el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import { ElMessage } from 'element-plus'
import { apiGet, apiPost, apiUpload, getCurrentUser } from '@/api/client'

type User = { id: number; username: string; name: string; gender: string; birth_year: number; birth_month: number; id_card_no: string; org_code: string; has_avatar: boolean; updated_at: string }

const user = ref<User>({ id: 0, username: '', name: '', gender: 'unknown', birth_year: 0, birth_month: 0, id_card_no: '', org_code: '', has_avatar: false, updated_at: '' })
const orgName = ref('')
const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ username: '', name: '', gender: 'unknown', birth_year: 0, birth_month: 0, id_card_no: '' })

const userInitial = computed(() => user.value.name?.slice(0, 1) || user.value.username?.slice(0, 1) || 'U')
const avatarBlobUrl = ref('')

// 使用 fetch + Blob URL 方式加载头像（支持 Bearer Token 认证）
async function loadCurrentUserAvatar() {
  if (!user.value.id || !user.value.has_avatar) {
    avatarBlobUrl.value = ''
    return
  }
  try {
    const token = localStorage.getItem('admin_token')
    const res = await fetch(`/api/admin/users/${user.value.id}/avatar`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {}
    })
    if (res.ok) {
      const blob = await res.blob()
      avatarBlobUrl.value = URL.createObjectURL(blob)
    }
  } catch (e) {
    console.warn('Failed to load avatar in profile:', e)
  }
}

const rules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }]
}

async function load() {
  const current = getCurrentUser()
  if (!current) return
  const data = await apiGet<User>(`/api/admin/users/${current.id}`)
  user.value = data
  Object.assign(form, {
    username: data.username,
    name: data.name,
    gender: data.gender,
    birth_year: data.birth_year,
    birth_month: data.birth_month,
    id_card_no: data.id_card_no
  })
  // 加载单位名称
  const orgs = await apiGet<{ code: string; name: string }[]>('/api/admin/organizations')
  const org = orgs.find(o => o.code === data.org_code)
  orgName.value = org ? `${org.name}（${org.code}）` : data.org_code
  // 加载头像
  await loadCurrentUserAvatar()
}

async function submit() {
  await formRef.value?.validate()
  saving.value = true
  try {
    await apiPost(`/api/admin/users/${user.value.id}/update`, form)
    ElMessage.success('个人信息已更新')
    await load()
  } finally {
    saving.value = false
  }
}

async function uploadAvatar(options: UploadRequestOptions) {
  const fd = new FormData()
  fd.append('avatar', options.file)
  await apiUpload(`/api/admin/users/${user.value.id}/avatar`, fd)
  ElMessage.success('头像已更新')
  // 清除旧的 blob URL 并重新加载
  if (avatarBlobUrl.value) URL.revokeObjectURL(avatarBlobUrl.value)
  await load()
  // 更新 localStorage 中的用户头像状态，供右上角显示
  const stored = localStorage.getItem('admin_user')
  if (stored) {
    const userData = JSON.parse(stored)
    userData.has_avatar = true
    localStorage.setItem('admin_user', JSON.stringify(userData))
  }
}

onMounted(load)
</script>

<style scoped>
.profile-container {
  display: flex;
  gap: 60px;
  padding: 20px;
}

.avatar-section {
  text-align: center;
  padding: 20px 40px;
  border-right: 1px solid #ebeef5;
}

.avatar-wrapper {
  position: relative;
  display: inline-block;
  margin-bottom: 20px;
}

.avatar-status {
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  background: #67c23a;
  color: #fff;
  font-size: 13px;
  padding: 4px 12px;
  border-radius: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.avatar-actions {
  margin-top: 30px;
}

.form-section {
  flex: 1;
  padding-top: 10px;
}
</style>
