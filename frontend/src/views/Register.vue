<template>
  <div class="auth-page">
    <div class="auth-container">
      <!-- Left Decorative Panel -->
      <div class="auth-decor">
        <div class="decor-content">
          <h2>加入我们</h2>
          <p>创建账号，开启你的创作之旅</p>
          <div class="decor-features">
            <div class="feature-item">
              <el-icon><EditPen /></el-icon>
              <span>发表你的观点和想法</span>
            </div>
            <div class="feature-item">
              <el-icon><Connection /></el-icon>
              <span>与志同道合的人交流</span>
            </div>
            <div class="feature-item">
              <el-icon><TrendCharts /></el-icon>
              <span>建立你的影响力</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Register Form -->
      <div class="auth-form-panel">
        <div class="form-wrapper">
          <div class="form-header">
            <h1 class="form-title">注册</h1>
            <p class="form-subtitle">创建一个新账号</p>
          </div>

          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            class="auth-form"
            @submit.prevent="handleRegister"
          >
            <el-form-item prop="username">
              <el-input
                v-model="form.username"
                placeholder="请输入用户名"
                size="large"
                clearable
              >
                <template #prefix>
                  <el-icon><User /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                placeholder="请输入密码（至少 6 位）"
                size="large"
                type="password"
                show-password
              >
                <template #prefix>
                  <el-icon><Lock /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item prop="confirmPassword">
              <el-input
                v-model="form.confirmPassword"
                placeholder="请确认密码"
                size="large"
                type="password"
                show-password
                @keyup.enter="handleRegister"
              >
                <template #prefix>
                  <el-icon><Lock /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                size="large"
                class="submit-btn"
                :loading="loading"
                @click="handleRegister"
              >
                注册
              </el-button>
            </el-form-item>
          </el-form>

          <div class="form-footer">
            <span>已有账号？</span>
            <router-link to="/login" class="link">立即登录</router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  User,
  Lock,
  EditPen,
  Connection,
  TrendCharts,
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = ref({
  username: '',
  password: '',
  confirmPassword: '',
})

/** 确认密码校验 */
const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== form.value.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 30, message: '用户名长度为 2-30 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, message: '用户名只能包含字母、数字、下划线和中文', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 50, message: '密码长度为 6-50 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

const handleRegister = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await authStore.register(
        form.value.username,
        form.value.password
      )
      ElMessage.success('注册成功，已自动登录')
      router.push('/')
    } catch (error: any) {
      ElMessage.error(error?.response?.data?.error || error?.response?.data?.message || '注册失败，请稍后重试')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped lang="scss">
.auth-page {
  min-height: calc(100vh - 120px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 24px;
  background: #f5f7fa;
}

.auth-container {
  display: flex;
  width: 100%;
  max-width: 900px;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.08);
}

.auth-decor {
  flex: 1;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  padding: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;

  .decor-content {
    h2 {
      font-size: 32px;
      font-weight: 700;
      margin: 0 0 12px;
    }

    p {
      font-size: 16px;
      opacity: 0.85;
      margin: 0 0 32px;
    }

    .decor-features {
      display: flex;
      flex-direction: column;
      gap: 16px;

      .feature-item {
        display: flex;
        align-items: center;
        gap: 12px;
        font-size: 15px;
        opacity: 0.9;
      }
    }
  }
}

.auth-form-panel {
  flex: 1;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;

  .form-wrapper {
    width: 100%;
    max-width: 360px;
  }

  .form-header {
    margin-bottom: 28px;

    .form-title {
      font-size: 28px;
      font-weight: 700;
      color: #1a1a1a;
      margin: 0 0 8px;
    }

    .form-subtitle {
      font-size: 15px;
      color: #999;
      margin: 0;
    }
  }
}

.auth-form {
  :deep(.el-input__wrapper) {
    border-radius: 10px;
    padding: 4px 12px;
  }
}

.submit-btn {
  width: 100%;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  height: 48px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  border: none;

  &:hover {
    opacity: 0.9;
  }
}

.form-footer {
  text-align: center;
  margin-top: 24px;
  font-size: 14px;
  color: #666;

  .link {
    color: #f5576c;
    text-decoration: none;
    font-weight: 500;
    margin-left: 4px;

    &:hover {
      text-decoration: underline;
    }
  }
}

@media (max-width: 768px) {
  .auth-decor {
    display: none;
  }

  .auth-form-panel {
    padding: 24px 20px;
  }

  .auth-container {
    border-radius: 16px;
  }
}
</style>