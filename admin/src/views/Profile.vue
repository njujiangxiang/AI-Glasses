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
import { apiGet, apiPost, apiUpload, getCurrentUser, setCurrentUser } from '@/api/client'

type User = { id: number; username: string; display_name: string; name: string; gender: string; birth_year: number; birth_month: number; id_card_no: string; org_code: string; org_name?: string; company_name?: string; role_id: number; avatar_size: number; has_avatar: boolean; updated_at: string }
type CurrentUserResponse = { user: User; org_name: string; company_name: string }

const user = ref<User>({ id: 0, username: '', display_name: '', name: '', gender: 'unknown', birth_year: 0, birth_month: 0, id_card_no: '', org_code: '', role_id: 0, avatar_size: 0, has_avatar: false, updated_at: '' })
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

function applyCurrentUser(current: User) {
  user.value = current
  setCurrentUser(current)
  Object.assign(form, {
    username: current.username,
    name: current.name,
    gender: current.gender,
    birth_year: current.birth_year,
    birth_month: current.birth_month,
    id_card_no: current.id_card_no
  })
  orgName.value = current.company_name ? `${current.company_name}（${current.org_code}）` : current.org_code
}

async function load() {
  try {
    const data = await apiGet<CurrentUserResponse>('/api/admin/users/me')
    applyCurrentUser({ ...data.user, org_name: data.org_name || '', company_name: data.company_name || data.org_name || '' })
  } catch (e) {
    const cached = getCurrentUser()
    if (!cached?.id) throw e
    console.warn('加载当前用户接口失败，降级读取用户详情:', e)
    const data = await apiGet<User>(`/api/admin/users/${cached.id}`)
    applyCurrentUser({ ...data, org_name: cached.org_name || '', company_name: cached.company_name || cached.org_name || '' })
  }
  await loadCurrentUserAvatar()
}

async function submit() {
  await formRef.value?.validate()
  saving.value = true
  try {
    const data = await apiPost<CurrentUserResponse>('/api/admin/users/me/update', form)
    applyCurrentUser({ ...data.user, org_name: data.org_name || '', company_name: data.company_name || data.org_name || '' })
    ElMessage.success('个人信息已更新')
    await loadCurrentUserAvatar()
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
