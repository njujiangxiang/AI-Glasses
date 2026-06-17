<template>
  <div class="login-container">
    <!-- 左侧品牌区 -->
    <div class="login-brand">
      <div class="brand-overlay"></div>
      <div class="brand-content">
        <div class="brand-icon">
          <svg viewBox="0 0 120 80" fill="none" xmlns="http://www.w3.org/2000/svg">
            <!-- 左镜片 -->
            <ellipse cx="35" cy="40" rx="28" ry="24" stroke="currentColor" stroke-width="3" fill="rgba(255,255,255,0.1)"/>
            <!-- 右镜片 -->
            <ellipse cx="85" cy="40" rx="28" ry="24" stroke="currentColor" stroke-width="3" fill="rgba(255,255,255,0.1)"/>
            <!-- 鼻梁架 -->
            <path d="M42 40 Q50 32 60 32 Q70 32 78 40" stroke="currentColor" stroke-width="3" fill="none"/>
            <!-- 左镜腿 -->
            <path d="M7 40 Q0 30 5 20 L8 18" stroke="currentColor" stroke-width="3" stroke-linecap="round" fill="none"/>
            <!-- 右镜腿 -->
            <path d="M113 40 Q120 30 115 20 L112 18" stroke="currentColor" stroke-width="3" stroke-linecap="round" fill="none"/>
            <!-- 镜片高光 -->
            <ellipse cx="25" cy="32" rx="8" ry="5" fill="rgba(255,255,255,0.2)" transform="rotate(-20 25 32)"/>
            <ellipse cx="75" cy="32" rx="8" ry="5" fill="rgba(255,255,255,0.2)" transform="rotate(-20 75 32)"/>
          </svg>
        </div>
        <h1 class="brand-title">智镜巡检</h1>
        <div class="brand-divider"></div>
        <p class="brand-subtitle">智能眼镜巡检管理系统</p>
        <p class="brand-desc">
          基于智能眼镜的现场作业巡检解决方案，实现任务数字化、操作可视化、证据可追溯
        </p>
      </div>
    </div>

    <!-- 右侧登录区 -->
    <div class="login-form-wrapper">
      <div class="login-card">
        <div class="login-header">
          <h2>欢迎登录</h2>
          <p>请输入您的账号密码</p>
        </div>

        <el-form ref="formRef" :model="form" :rules="rules" class="login-form" @submit.prevent="login">
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              size="large"
              class="login-input"
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              show-password
              class="login-input"
              @keyup.enter="login"
            >
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <div class="form-row">
            <el-checkbox v-model="rememberMe">记住我</el-checkbox>
          </div>

          <el-button
            type="primary"
            class="login-btn"
            :loading="loading"
            @click="login"
            size="large"
          >
            登录系统
          </el-button>
        </el-form>
      </div>

      <div class="login-footer">
        <p>© 2025 智镜巡检 版权所有</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import { apiPost, setCurrentUser, setToken } from '@/api/client'

const formRef = ref<FormInstance>()
const form = reactive({ username: 'admin', password: '' })
const loading = ref(false)
const rememberMe = ref(false)
const router = useRouter()

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

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
    setCurrentUser({ ...data.user, company_name: data.company_name || '' })
    ElMessage.success('登录成功')
    await router.push('/workbench')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 100vh;
}

/* ========== 左侧品牌区 ========== */
.login-brand {
  position: relative;
  background-image: url('https://images.unsplash.com/photo-1581093458791-91186078f78a?ixlib=rb-4.0.3&auto=format&fit=crop&w=1920&q=80');
  background-size: cover;
  background-position: center;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 渐变遮罩 - 国网绿色调 */
.brand-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(0, 122, 61, 0.92) 0%, rgba(0, 87, 43, 0.88) 100%);
}

.brand-content {
  position: relative;
  z-index: 1;
  text-align: center;
  color: white;
  padding: 60px;
}

.brand-icon {
  width: 120px;
  height: 80px;
  margin: 0 auto 24px;
  color: rgba(255, 255, 255, 0.95);
}

.brand-title {
  font-size: 42px;
  font-weight: 600;
  margin-bottom: 12px;
  letter-spacing: 2px;
}

.brand-divider {
  width: 60px;
  height: 3px;
  background: rgba(255, 255, 255, 0.5);
  margin: 32px auto;
  border-radius: 2px;
}

.brand-subtitle {
  font-size: 18px;
  opacity: 0.9;
  font-weight: 300;
  letter-spacing: 1px;
}

.brand-desc {
  font-size: 14px;
  opacity: 0.75;
  max-width: 320px;
  margin: 30px auto 0;
  line-height: 1.8;
}

/* ========== 右侧登录区 ========== */
.login-form-wrapper {
  background: #f0f5f2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: white;
  border-radius: 16px;
  padding: 48px 40px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06), 0 1px 3px rgba(0, 0, 0, 0.04);
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.login-header h2 {
  font-size: 28px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 8px;
}

.login-header p {
  color: #6b7280;
  font-size: 14px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 24px;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 8px 16px;
  box-shadow: 0 0 0 1px #d1d5db inset;
  transition: all 0.2s ease;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #9ca3af inset;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #007a3d inset, 0 0 0 4px rgba(0, 122, 61, 0.1);
}

.login-form :deep(.el-input__inner) {
  font-size: 15px;
}

.form-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  font-size: 14px;
}

.form-row :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: #007a3d;
  border-color: #007a3d;
}

.form-row :deep(.el-checkbox__input.is-focus .el-checkbox__inner) {
  border-color: #007a3d;
}

.login-btn {
  width: 100%;
  height: 48px;
  background: #007a3d;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0, 122, 61, 0.25);
  transition: all 0.2s ease;
}

.login-btn:hover {
  background: #006b35;
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(0, 122, 61, 0.35);
}

.login-btn:active {
  transform: translateY(0);
}

.login-footer {
  margin-top: 32px;
  text-align: center;
  color: #9ca3af;
  font-size: 13px;
}

/* ========== 响应式适配 ========== */
@media (max-width: 768px) {
  .login-container {
    grid-template-columns: 1fr;
  }

  .login-brand {
    min-height: 240px;
  }

  .brand-content {
    padding: 30px 20px;
  }

  .brand-icon {
    width: 60px;
    height: 60px;
    margin-bottom: 16px;
  }

  .brand-title {
    font-size: 28px;
  }

  .brand-subtitle {
    font-size: 15px;
  }

  .brand-divider {
    margin: 20px auto;
  }

  .brand-desc {
    display: none;
  }

  .login-form-wrapper {
    padding: 30px 20px;
  }

  .login-card {
    padding: 32px 24px;
  }

  .login-header h2 {
    font-size: 22px;
  }
}
</style>
