<template>
  <div class="auth-page">
    <div class="auth-container">
      <!-- Left Decorative Panel -->
      <div class="auth-decor">
        <div class="decor-content">
          <h2>欢迎回来</h2>
          <p>登录你的账号，继续探索精彩内容</p>
          <div class="decor-features">
            <div class="feature-item">
              <el-icon><Document /></el-icon>
              <span>阅读优质文章</span>
            </div>
            <div class="feature-item">
              <el-icon><ChatDotRound /></el-icon>
              <span>参与实时讨论</span>
            </div>
            <div class="feature-item">
              <el-icon><User /></el-icon>
              <span>关注感兴趣的人</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Login Form -->
      <div class="auth-form-panel">
        <div class="form-wrapper">
          <div class="form-header">
            <h1 class="form-title">登录</h1>
            <p class="form-subtitle">请输入你的账号信息</p>
          </div>

          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            class="auth-form"
            @submit.prevent="handleLogin"
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
                placeholder="请输入密码"
                size="large"
                type="password"
                show-password
                @keyup.enter="handleLogin"
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
                @click="handleLogin"
              >
                登录
              </el-button>
            </el-form-item>
          </el-form>

          <div class="form-footer">
            <span>还没有账号？</span>
            <router-link to="/register" class="link">立即注册</router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  Message,
  Lock,
  Document,
  ChatDotRound,
  User,
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = ref({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 个字符', trigger: 'blur' },
  ],
}

const handleLogin = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await authStore.login(form.value.username, form.value.password)
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/'
      router.push(redirect)
    } catch (error: any) {
      ElMessage.error(error?.response?.data?.error || error?.response?.data?.message || '登录失败，请检查账号密码')
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  padding: 48px;

  .form-wrapper {
    width: 100%;
    max-width: 360px;
  }

  .form-header {
    margin-bottom: 32px;

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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
    color: #667eea;
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
    padding: 32px 24px;
  }

  .auth-container {
    border-radius: 16px;
  }
}
</style>