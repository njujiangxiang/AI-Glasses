<template>
  <div style="min-height: 100vh; display: grid; place-items: center; background: var(--el-bg-color-page);">
    <el-card shadow="never" style="width: 380px;">
      <template #header><span class="card-title">后台登录</span></template>
      <el-form label-position="top">
        <el-form-item label="用户名">
          <el-input v-model="username" placeholder="输入 active 用户名" />
        </el-form-item>
        <el-button type="primary" style="width: 100%;" @click="login">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { apiPost, setToken } from '@/api/client'

const username = ref('admin')
const router = useRouter()
// login 调用后台登录接口，保存 token 后进入工作台。
async function login() {
  const data = await apiPost<{ access_token: string }>('/api/admin/auth/login', { username: username.value })
  setToken(data.access_token)
  ElMessage.success('登录成功')
  await router.push('/workbench')
}
</script>
