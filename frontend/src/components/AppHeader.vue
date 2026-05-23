<template>
  <header class="app-header">
    <div class="header-container">
      <!-- Logo -->
      <router-link to="/" class="logo">
        <span class="logo-text">BlogHub</span>
      </router-link>

      <!-- Desktop Navigation -->
      <nav class="nav-links">
        <router-link to="/" class="nav-link" :class="{ active: $route.path === '/' }">
          <el-icon><House /></el-icon>
          <span>首页</span>
        </router-link>
        <router-link to="/feed" class="nav-link" :class="{ active: $route.path === '/feed' }">
          <el-icon><Compass /></el-icon>
          <span>关注</span>
        </router-link>
        <router-link to="/chat" class="nav-link" :class="{ active: $route.path === '/chat' }">
          <el-icon><ChatDotRound /></el-icon>
          <span>聊天室</span>
        </router-link>
      </nav>

      <!-- Right Section -->
      <div class="header-right">
        <template v-if="isLoggedIn">
          <!-- Write Button -->
          <el-button
            type="primary"
            class="write-btn"
            @click="$router.push('/editor')"
            round
          >
            <el-icon><EditPen /></el-icon>
            <span class="write-text">写文章</span>
          </el-button>

          <!-- User Menu -->
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-menu">
              <el-avatar
                :size="32"
                :src="getImageUrl(currentUser?.image || '')"
                class="menu-avatar"
              >
                <span class="avatar-fallback">{{ currentUser?.username?.charAt(0)?.toUpperCase() }}</span>
              </el-avatar>
              <span class="username-text">{{ currentUser?.username }}</span>
              <el-icon class="arrow-icon"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  个人主页
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>
                  设置
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>

        <template v-else>
          <el-button @click="$router.push('/login')" round>
            登录
          </el-button>
          <el-button type="primary" @click="$router.push('/register')" round>
            注册
          </el-button>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import {
  House,
  Compass,
  ChatDotRound,
  EditPen,
  User,
  Setting,
  SwitchButton,
  ArrowDown,
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const isLoggedIn = computed(() => authStore.isLoggedIn)
const currentUser = computed(() => authStore.user)

/** 拼接图片完整 URL */
const getImageUrl = (path: string): string => {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `http://localhost:3000${path}`
}

const handleCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push(`/profile/${(currentUser.value as any)?.ID ?? currentUser.value?.id}`)
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      authStore.logout()
      ElMessage.success('已退出登录')
      router.push('/')
      break
  }
}
</script>

<style scoped lang="scss">
.app-header {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid #f0f0f0;
}

.header-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.logo {
  text-decoration: none;
  flex-shrink: 0;

  .logo-text {
    font-size: 22px;
    font-weight: 700;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
  justify-content: center;

  .nav-link {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    border-radius: 8px;
    color: #666;
    text-decoration: none;
    font-size: 14px;
    font-weight: 500;
    transition: all 0.2s ease;

    &:hover {
      color: #333;
      background: #f5f7fa;
    }

    &.active {
      color: #667eea;
      background: #f0f2ff;
    }
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;

  .write-btn {
    .write-text {
      margin-left: 4px;
    }
  }
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.2s;

  &:hover {
    background: #f5f7fa;
  }

  .menu-avatar {
    flex-shrink: 0;

    .avatar-fallback {
      font-size: 14px;
      font-weight: 600;
      color: #667eea;
      background: linear-gradient(135deg, #e0e7ff, #f0f5ff);
      width: 100%;
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  .username-text {
    font-size: 14px;
    font-weight: 500;
    color: #333;
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .arrow-icon {
    font-size: 12px;
    color: #999;
  }
}

@media (max-width: 768px) {
  .header-container {
    padding: 0 16px;
  }

  .nav-links .nav-link span {
    display: none;
  }

  .nav-links .nav-link {
    padding: 8px 10px;
  }

  .username-text {
    display: none;
  }

  .write-btn .write-text {
    display: none;
  }

  .write-btn {
    padding: 8px 12px;
  }
}
</style>