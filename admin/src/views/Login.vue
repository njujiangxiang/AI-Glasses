<template>
  <div class="login-container">
    <!-- 左侧品牌区 -->
    <div class="login-brand">
      <div class="brand-bg">
        <div class="brand-pattern"></div>
        <div class="brand-glow"></div>
      </div>
      <div class="brand-content">
        <div class="brand-icon">
          <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="32" cy="32" r="28" stroke="currentColor" stroke-width="2" />
            <circle cx="32" cy="32" r="18" stroke="currentColor" stroke-width="2" />
            <circle cx="24" cy="28" r="6" fill="currentColor" />
            <circle cx="40" cy="28" r="6" fill="currentColor" />
            <path d="M20 44 Q32 52 44 44" stroke="currentColor" stroke-width="3" stroke-linecap="round" fill="none" />
          </svg>
        </div>
        <h1 class="brand-title">智镜巡检</h1>
        <div class="brand-divider"></div>
        <p class="brand-subtitle">智能眼镜巡检管理系统</p>
        <div class="brand-features">
          <div class="feature-item">
            <span class="feature-icon">✓</span>
            <span>任务数字化管理</span>
          </div>
          <div class="feature-item">
            <span class="feature-icon">✓</span>
            <span>操作可视化指引</span>
          </div>
          <div class="feature-item">
            <span class="feature-icon">✓</span>
            <span>证据可追溯记录</span>
          </div>
        </div>
      </div>
      <div class="brand-footer">
        <p>© 2025 智镜巡检 版权所有</p>
      </div>
    </div>

    <!-- 右侧登录区 - 融合设计 -->
    <div class="login-form-section">
      <div class="login-card">
        <div class="login-header">
          <div class="login-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <h2>欢迎登录</h2>
          <p>请输入您的账号密码</p>
        </div>

        <el-form ref="formRef" :model="form" :rules="rules" class="login-form" @submit.prevent="login">
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              size="large"
              prefix-icon="User"
              class="login-input"
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              prefix-icon="Lock"
              show-password
              class="login-input"
              @keyup.enter="login"
            />
          </el-form-item>

          <div class="form-actions">
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

        <div class="login-tip">
          <p>演示账号：admin / admin123</p>
        </div>
      </div>
    </div>
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
  min-height: 100vh;
  display: flex;
  position: relative;
  overflow: hidden;
}

/* ========== 左侧品牌区 - 柔和渐变 ========== */
.login-brand {
  flex: 1;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: white;
}

.brand-bg {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #007a3d 0%, #00572b 50%, #004d26 100%);
  z-index: 0;
}

/* 装饰图案 - 增加层次感 */
.brand-pattern {
  position: absolute;
  top: -50%;
  right: -50%;
  width: 200%;
  height: 200%;
  background:
    radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.08) 0%, transparent 50%),
    radial-gradient(circle at 70% 70%, rgba(255, 255, 255, 0.05) 0%, transparent 40%);
  animation: patternFloat 60s ease-in-out infinite;
}

@keyframes patternFloat {
  0%, 100% { transform: translate(0, 0) rotate(0deg); }
  50% { transform: translate(-5%, 5%) rotate(2deg); }
}

/* 发光效果 - 向右延伸，连接右侧 */
.brand-glow {
  position: absolute;
  top: 0;
  right: -200px;
  width: 400px;
  height: 100%;
  background: radial-gradient(ellipse at center, rgba(0, 122, 61, 0.15) 0%, transparent 70%);
  z-index: 1;
}

.brand-content {
  position: relative;
  z-index: 2;
  text-align: center;
  max-width: 400px;
}

.brand-icon {
  width: 100px;
  height: 100px;
  margin: 0 auto 32px;
  color: rgba(255, 255, 255, 0.95);
  filter: drop-shadow(0 4px 16px rgba(0, 0, 0, 0.2));
}

.brand-title {
  font-size: 42px;
  font-weight: 600;
  margin-bottom: 16px;
  letter-spacing: 3px;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.brand-divider {
  width: 80px;
  height: 3px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.6), transparent);
  margin: 24px auto;
  border-radius: 2px;
}

.brand-subtitle {
  font-size: 18px;
  font-weight: 300;
  opacity: 0.9;
  letter-spacing: 1px;
  margin-bottom: 40px;
}

.brand-features {
  text-align: left;
  display: inline-block;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  font-size: 15px;
  opacity: 0.85;
  transition: all 0.3s ease;
}

.feature-item:hover {
  opacity: 1;
  transform: translateX(4px);
}

.feature-icon {
  width: 24px;
  height: 24px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.brand-footer {
  position: absolute;
  bottom: 30px;
  left: 0;
  right: 0;
  text-align: center;
  font-size: 13px;
  opacity: 0.6;
  z-index: 2;
}

/* ========== 右侧登录区 - 融合设计 ========== */
.login-form-section {
  flex: 0 0 520px;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background: linear-gradient(135deg, #f8faf9 0%, #f0f5f2 100%);
}

/* 左侧连接条 - 视觉融合关键 */
.login-form-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 6px;
  height: 100%;
  background: linear-gradient(180deg, #007a3d, #00572b);
  box-shadow: 0 0 30px rgba(0, 122, 61, 0.3);
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 20px;
  padding: 48px 40px;
  box-shadow:
    0 4px 24px rgba(0, 122, 61, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.05),
    0 0 0 1px rgba(0, 122, 61, 0.06);
  position: relative;
  overflow: hidden;
}

/* 卡片顶部绿色装饰条 - 与左侧呼应 */
.login-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #007a3d, #00a854, #007a3d);
}

.login-header {
  text-align: center;
  margin-bottom: 36px;
}

.login-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 20px;
  background: linear-gradient(135deg, #007a3d, #00a854);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 4px 12px rgba(0, 122, 61, 0.3);
}

.login-header h2 {
  font-size: 26px;
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
  border-radius: 12px;
  padding: 8px 16px;
  box-shadow: 0 0 0 1px #e5e7eb inset;
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

.login-form :deep(.el-input__prefix) {
  color: #9ca3af;
}

.form-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
  font-size: 14px;
}

.form-actions :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: #007a3d;
  border-color: #007a3d;
}

.form-actions :deep(.el-checkbox__input.is-focus .el-checkbox__inner) {
  border-color: #007a3d;
}

.login-btn {
  width: 100%;
  height: 50px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 12px;
  background: linear-gradient(135deg, #007a3d, #00572b);
  border: none;
  box-shadow: 0 4px 12px rgba(0, 122, 61, 0.25);
  transition: all 0.2s ease;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 122, 61, 0.35);
}

.login-btn:active {
  transform: translateY(0);
}

.login-tip {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #f3f4f6;
  text-align: center;
  font-size: 13px;
  color: #9ca3af;
}

/* ========== 响应式适配 ========== */
@media (max-width: 968px) {
  .login-container {
    flex-direction: column;
  }

  .login-brand {
    flex: none;
    min-height: 280px;
    padding: 40px 30px;
  }

  .brand-icon {
    width: 70px;
    height: 70px;
    margin-bottom: 20px;
  }

  .brand-title {
    font-size: 30px;
  }

  .brand-subtitle {
    font-size: 15px;
    margin-bottom: 24px;
  }

  .brand-features {
    display: none;
  }

  .brand-footer {
    display: none;
  }

  .login-form-section {
    flex: 1;
    padding: 30px 20px;
  }

  .login-form-section::before {
    width: 100%;
    height: 4px;
    top: 0;
    bottom: auto;
    background: linear-gradient(90deg, #007a3d, #00a854, #007a3d);
  }

  .login-card {
    padding: 36px 28px;
  }
}
</style>
