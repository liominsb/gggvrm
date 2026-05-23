<template>
  <div class="settings-page">
    <div class="main-container">
      <div class="settings-header">
        <h1 class="page-title">
          <el-icon><Setting /></el-icon>
          个人设置
        </h1>
      </div>

      <!-- 个人资料 -->
      <el-card class="settings-card" shadow="never">
        <template #header>
          <span class="card-title">个人资料</span>
        </template>
        <el-form
          ref="profileFormRef"
          :model="profileForm"
          :rules="profileRules"
          label-position="top"
          class="settings-form"
        >
          <!-- Avatar Upload -->
          <div class="avatar-section">
            <el-upload
              class="avatar-uploader"
              :show-file-list="false"
              :before-upload="beforeAvatarUpload"
              :http-request="handleAvatarUpload"
              accept="image/*"
            >
              <el-avatar
                :size="96"
                :src="getImageUrl(profileForm.avatar)"
                class="current-avatar"
              >
                <span class="avatar-fallback">{{ profileForm.username?.charAt(0)?.toUpperCase() }}</span>
              </el-avatar>
              <div class="avatar-overlay">
                <el-icon size="24"><Camera /></el-icon>
                <span>更换头像</span>
              </div>
            </el-upload>
            <div class="avatar-info">
              <h3>{{ profileForm.username }}</h3>
              <p>点击头像更换</p>
            </div>
          </div>

          <!-- Username -->
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="profileForm.username"
              placeholder="请输入用户名"
              maxlength="30"
              show-word-limit
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item class="form-actions">
            <el-button
              type="primary"
              size="large"
              :loading="profileSaving"
              @click="handleProfileSubmit"
            >
              <el-icon><Check /></el-icon>
              保存资料
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 修改密码 -->
      <el-card class="settings-card" shadow="never">
        <template #header>
          <span class="card-title">修改密码</span>
        </template>
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-position="top"
          class="settings-form"
        >
          <el-form-item label="旧密码" prop="old_password">
            <el-input
              v-model="passwordForm.old_password"
              placeholder="请输入旧密码"
              type="password"
              show-password
            >
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="新密码" prop="new_password">
            <el-input
              v-model="passwordForm.new_password"
              placeholder="请输入新密码"
              type="password"
              show-password
            >
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item class="form-actions">
            <el-button
              type="primary"
              size="large"
              :loading="passwordSaving"
              @click="handlePasswordSubmit"
            >
              <el-icon><Check /></el-icon>
              修改密码
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- Danger Zone -->
      <el-card class="danger-card" shadow="never">
        <template #header>
          <div class="danger-header">
            <el-icon color="#f56c6c"><WarningFilled /></el-icon>
            <span>危险操作</span>
          </div>
        </template>
        <div class="danger-content">
          <div class="danger-item">
            <div class="danger-info">
              <h4>退出登录</h4>
              <p>退出当前账号，需要重新登录</p>
            </div>
            <el-button type="danger" plain @click="handleLogout">
              退出登录
            </el-button>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { uploadApi } from '@/api/upload'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import {
  Setting,
  User,
  Lock,
  Check,
  Camera,
  WarningFilled,
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const profileFormRef = ref<FormInstance>()
const passwordFormRef = ref<FormInstance>()
const profileSaving = ref(false)
const passwordSaving = ref(false)

const profileForm = ref({
  username: '',
  avatar: '',
})

const passwordForm = ref({
  old_password: '',
  new_password: '',
})

const profileRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 30, message: '用户名长度为 2-30 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, message: '用户名只能包含字母、数字、下划线和中文', trigger: 'blur' },
  ],
}

const passwordRules: FormRules = {
  old_password: [
    { required: true, message: '请输入旧密码', trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 50, message: '密码长度为 6-50 个字符', trigger: 'blur' },
  ],
}

/** 拼接图片完整 URL */
const getImageUrl = (path: string): string => {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `http://localhost:3000${path}`
}

/** 初始化表单 */
const initForm = () => {
  const user = authStore.user
  if (user) {
    profileForm.value.username = user.username || ''
    profileForm.value.avatar = (user as any).image || ''
  }
}

/** 头像上传前校验 */
const beforeAvatarUpload = (file: File) => {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('头像大小不能超过 2MB')
    return false
  }
  return true
}

/** 上传头像 */
const handleAvatarUpload = async (options: UploadRequestOptions) => {
  try {
    const urls = await uploadApi.uploadImages([options.file])
    profileForm.value.avatar = urls[0] || ''
    ElMessage.success('头像上传成功')
  } catch (error) {
    ElMessage.error('头像上传失败')
  }
}

/** 提交个人资料 */
const handleProfileSubmit = async () => {
  if (!profileFormRef.value) return
  await profileFormRef.value.validate(async (valid) => {
    if (!valid) return
    profileSaving.value = true
    try {
      await authStore.updateProfile({
        username: profileForm.value.username,
        image: profileForm.value.avatar,
      })
      ElMessage.success('资料保存成功')
    } catch (error: any) {
      ElMessage.error(error?.response?.data?.error || '保存失败')
    } finally {
      profileSaving.value = false
    }
  })
}

/** 提交密码修改 */
const handlePasswordSubmit = async () => {
  if (!passwordFormRef.value) return
  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return
    passwordSaving.value = true
    try {
      await authStore.changePassword({
        old_password: passwordForm.value.old_password,
        new_password: passwordForm.value.new_password,
      })
      ElMessage.success('密码修改成功')
      passwordForm.value.old_password = ''
      passwordForm.value.new_password = ''
      passwordFormRef.value?.resetFields()
    } catch (error: any) {
      ElMessage.error(error?.response?.data?.error || '密码修改失败')
    } finally {
      passwordSaving.value = false
    }
  })
}

/** 退出登录 */
const handleLogout = () => {
  authStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(() => {
  initForm()
})
</script>

<style scoped lang="scss">
.main-container {
  max-width: 680px;
  margin: 0 auto;
  padding: 32px 24px 48px;
}

.settings-header {
  margin-bottom: 28px;

  .page-title {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 28px;
    font-weight: 700;
    color: var(--el-text-color-primary);
  }
}

.settings-card {
  margin-bottom: 24px;
  border-radius: 12px;

  .card-title {
    font-weight: 600;
    font-size: 16px;
  }
}

.avatar-section {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 24px;

  .avatar-uploader {
    position: relative;
    cursor: pointer;

    .current-avatar {
      border: 3px solid var(--el-border-color-lighter);
      transition: border-color 0.2s;
    }

    .avatar-overlay {
      position: absolute;
      inset: 0;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      background: rgba(0, 0, 0, 0.5);
      border-radius: 50%;
      color: white;
      opacity: 0;
      transition: opacity 0.2s;
      font-size: 12px;
      gap: 4px;
    }

    &:hover .avatar-overlay {
      opacity: 1;
    }
  }

  .avatar-info {
    h3 {
      margin: 0 0 4px;
      font-size: 18px;
    }
    p {
      margin: 0;
      color: var(--el-text-color-secondary);
      font-size: 13px;
    }
  }
}

.form-actions {
  margin-top: 8px;
}

.danger-card {
  border-radius: 12px;
  border-color: var(--el-color-danger-light-7);

  .danger-header {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
  }

  .danger-item {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .danger-info {
      h4 {
        margin: 0 0 4px;
        font-size: 15px;
      }
      p {
        margin: 0;
        color: var(--el-text-color-secondary);
        font-size: 13px;
      }
    }
  }
}
</style>
