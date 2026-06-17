<template>
  <div style="min-height: 100vh; display: grid; place-items: center; background: var(--el-bg-color-page);">
    <el-card shadow="never" style="width: 380px;">
      <template #header><span class="card-title">后台登录</span></template>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="login">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-button type="primary" style="width: 100%;" :loading="loading" @click="login">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { apiPost, setCurrentUser, setToken } from '@/api/client'

const formRef = ref<FormInstance>()
const form = reactive({ username: 'admin', password: '' })
const loading = ref(false)
const router = useRouter()

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

// login 调用后台登录接口，保存 token 后进入工作台。
async function login() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const data = await apiPost<{ access_token: string; user: any; company_name: string }>('/api/admin/auth/login', {
      username: form.username,
      password: form.password
    })
    setToken(data.access_token)
    console.log('Login user data:', data.user)
    console.log('avatar_size value:', data.user.avatar_size)
    setCurrentUser({ ...data.user, company_name: data.company_name || '' })
    ElMessage.success('登录成功')
    await router.push('/workbench')
  } finally {
    loading.value = false
  }
}
</script>
