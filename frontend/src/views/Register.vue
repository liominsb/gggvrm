<template>
  <div class="auth-page">
    <div class="auth-container">
      <!-- 左侧装饰 -->
      <div class="auth-decor">
        <div class="decor-inner">
          <DecorPlant variant="sprout" />
          <h2 class="decor-title">加入我们</h2>
          <p class="decor-desc">创建账号，开启你的创作之旅</p>
          <div class="decor-features">
            <div class="decor-feature">
              <FreshIcon name="edit" :size="18" color="pink" />
              <span>发表你的观点</span>
            </div>
            <div class="decor-feature">
              <FreshIcon name="compass" :size="18" color="pink" />
              <span>找到志同道合的人</span>
            </div>
            <div class="decor-feature">
              <FreshIcon name="sparkle" :size="18" color="pink" />
              <span>建立你的影响力</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧表单 -->
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
                placeholder="用户名"
                size="large"
                clearable
                class="fresh-form-input"
              >
                <template #prefix>
                  <FreshIcon name="user" :size="16" color="pink" />
                </template>
              </el-input>
            </el-form-item>

            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                placeholder="密码（至少 6 位）"
                size="large"
                type="password"
                show-password
                class="fresh-form-input"
              >
                <template #prefix>
                  <FreshIcon name="lock" :size="16" color="pink" />
                </template>
              </el-input>
            </el-form-item>

            <el-form-item prop="confirmPassword">
              <el-input
                v-model="form.confirmPassword"
                placeholder="确认密码"
                size="large"
                type="password"
                show-password
                class="fresh-form-input"
                @keyup.enter="handleRegister"
              >
                <template #prefix>
                  <FreshIcon name="lock" :size="16" color="pink" />
                </template>
              </el-input>
            </el-form-item>

            <el-form-item>
              <FreshButton
                variant="pink"
                size="lg"
                block
                :disabled="loading"
                native-type="submit"
                @click="handleRegister"
              >
                {{ loading ? '注册中...' : '注册' }}
              </FreshButton>
            </el-form-item>
          </el-form>

          <div class="form-footer">
            <span>已有账号？</span>
            <router-link to="/login" class="footer-link">立即登录</router-link>
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
import FreshIcon from '@/components/fresh/FreshIcon.vue'
import FreshButton from '@/components/fresh/FreshButton.vue'
import DecorPlant from '@/components/fresh/DecorPlant.vue'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = ref({
  username: '',
  password: '',
  confirmPassword: '',
})

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
    { pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, message: '只能包含字母、数字、下划线和中文', trigger: 'blur' },
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
      await authStore.register(form.value.username, form.value.password)
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
  min-height: calc(100dvh - 120px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--fresh-space-xl) var(--fresh-space-lg);
}

.auth-container {
  display: flex;
  width: 100%;
  max-width: 880px;
  border-radius: var(--fresh-radius-xl);
  overflow: hidden;
  box-shadow: var(--fresh-shadow-lg);
}

/* 左侧装饰 */
.auth-decor {
  flex: 1;
  background: linear-gradient(160deg, var(--fresh-pink-light) 0%, #fdf5f0 50%, var(--fresh-bg-surface) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--fresh-space-2xl);
}

.decor-inner {
  text-align: center;
  max-width: 280px;
}

.decor-title {
  font-size: var(--fresh-text-2xl);
  font-weight: 700;
  color: var(--fresh-text-primary);
  margin: var(--fresh-space-lg) 0 var(--fresh-space-sm);
}

.decor-desc {
  font-size: var(--fresh-text-sm);
  color: var(--fresh-text-secondary);
  margin: 0 0 var(--fresh-space-xl);
  line-height: 1.6;
}

.decor-features {
  display: flex;
  flex-direction: column;
  gap: var(--fresh-space-md);
}

.decor-feature {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-md);
  padding: var(--fresh-space-sm) var(--fresh-space-md);
  background: rgba(255, 255, 255, 0.7);
  border-radius: var(--fresh-radius-sm);
  font-size: 14px;
  color: var(--fresh-text-secondary);
  backdrop-filter: blur(4px);
}

/* 右侧表单 */
.auth-form-panel {
  flex: 1;
  background: var(--fresh-bg-surface);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--fresh-space-2xl);
}

.form-wrapper {
  width: 100%;
  max-width: 340px;
}

.form-header {
  margin-bottom: var(--fresh-space-xl);
}

.form-title {
  font-size: var(--fresh-text-3xl);
  font-weight: 700;
  color: var(--fresh-text-primary);
  margin: 0 0 var(--fresh-space-xs);
}

.form-subtitle {
  font-size: var(--fresh-text-sm);
  color: var(--fresh-text-muted);
  margin: 0;
}

.auth-form {
  :deep(.el-form-item) {
    margin-bottom: var(--fresh-space-lg);
  }

  :deep(.el-form-item__error) {
    font-size: 12px;
    color: var(--fresh-error);
    padding-top: 4px;
  }
}

.fresh-form-input {
  :deep(.el-input__wrapper) {
    background: var(--fresh-bg-page);
    border: 1.5px solid var(--fresh-border-light);
    border-radius: var(--fresh-radius-sm);
    box-shadow: none;
    padding: 4px 12px;
    transition: border-color var(--fresh-transition-fast);

    &:hover { border-color: var(--fresh-border-default); }

    &.is-focus {
      border-color: var(--fresh-pink);
      box-shadow: 0 0 0 3px rgba(232, 160, 180, 0.15);
    }
  }

  :deep(.el-input__inner) {
    font-size: 15px;
    color: var(--fresh-text-primary);

    &::placeholder { color: var(--fresh-text-muted); }
  }

  :deep(.el-input__prefix) {
    color: var(--fresh-pink);
  }
}

.form-footer {
  text-align: center;
  margin-top: var(--fresh-space-lg);
  font-size: 14px;
  color: var(--fresh-text-secondary);

  .footer-link {
    color: var(--fresh-pink);
    font-weight: 600;
    margin-left: 4px;

    &:hover { color: var(--fresh-pink-hover); }
  }
}

@media (max-width: 768px) {
  .auth-decor { display: none; }
  .auth-form-panel { padding: var(--fresh-space-xl) var(--fresh-space-lg); }
  .form-wrapper { max-width: 100%; }
}
</style>
