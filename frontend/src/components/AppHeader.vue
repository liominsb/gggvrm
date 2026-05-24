<template>
  <header class="app-header">
    <div class="header-container">
      <!-- Mobile Toggle -->
      <button class="mobile-toggle" @click="mobileOpen = !mobileOpen" aria-label="菜单">
        <FreshIcon name="menu" :size="20" />
      </button>

      <!-- Navigation -->
      <nav class="nav-links" :class="{ 'nav-links--open': mobileOpen }">
        <router-link to="/" class="nav-link" :class="{ active: $route.path === '/' }">
          <FreshIcon name="home" :size="16" />
          <span>首页</span>
        </router-link>
        <router-link to="/feed" class="nav-link" :class="{ active: $route.path === '/feed' }">
          <FreshIcon name="compass" :size="16" />
          <span>关注</span>
        </router-link>
        <router-link to="/chat" class="nav-link" :class="{ active: $route.path === '/chat' }">
          <FreshIcon name="chat" :size="16" />
          <span>聊天室</span>
        </router-link>
      </nav>

      <!-- Brand -->
      <div class="header-brand">
        <span class="brand-text">博客空间</span>
      </div>

      <!-- Right -->
      <div class="header-right">
        <template v-if="isLoggedIn">
          <FreshButton
            variant="mint"
            size="sm"
            icon="edit"
            @click="$router.push('/editor')"
          >
            写文章
          </FreshButton>

          <el-dropdown trigger="click" @command="handleCommand" popper-class="fresh-dropdown">
            <div class="user-menu">
              <div class="user-avatar">
                {{ currentUser?.username?.charAt(0)?.toUpperCase() }}
              </div>
              <span class="username-text">{{ currentUser?.username }}</span>
              <FreshIcon name="arrow-down" :size="12" />
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <FreshIcon name="user" :size="14" />
                  <span>个人主页</span>
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <FreshIcon name="settings" :size="14" />
                  <span>设置</span>
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <FreshIcon name="logout" :size="14" />
                  <span>退出登录</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>

        <template v-else>
          <FreshButton variant="ghost" size="sm" @click="$router.push('/login')">
            登录
          </FreshButton>
          <FreshButton variant="mint" size="sm" @click="$router.push('/register')">
            注册
          </FreshButton>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import FreshIcon from '@/components/fresh/FreshIcon.vue'
import FreshButton from '@/components/fresh/FreshButton.vue'

const router = useRouter()
const authStore = useAuthStore()
const mobileOpen = ref(false)

const isLoggedIn = computed(() => authStore.isLoggedIn)
const currentUser = computed(() => authStore.user)

const handleCommand = (command: string) => {
  mobileOpen.value = false
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
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: var(--z-header);
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--fresh-border-light);
  height: var(--fresh-header-height);
}

.header-container {
  max-width: var(--fresh-max-width);
  margin: 0 auto;
  padding: 0 var(--fresh-space-lg);
  height: 100%;
  display: flex;
  align-items: center;
  gap: var(--fresh-space-lg);
}

.mobile-toggle {
  display: none;
  background: none;
  border: none;
  padding: 6px;
  cursor: pointer;
  color: var(--fresh-text-secondary);
  border-radius: var(--fresh-radius-sm);

  &:hover {
    background: var(--fresh-bg-hover);
  }
}

/* Navigation */
.nav-links {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;

  .nav-link {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    border-radius: var(--fresh-radius-sm);
    color: var(--fresh-text-secondary);
    text-decoration: none;
    font-size: 14px;
    font-weight: 500;
    transition: all var(--fresh-transition-fast);

    &:hover {
      color: var(--fresh-text-primary);
      background: var(--fresh-bg-hover);
    }

    &.active {
      color: var(--fresh-mint-hover);
      background: var(--fresh-mint-light);
      font-weight: 600;
    }
  }
}

/* Brand */
.header-brand {
  flex: 1;
  display: flex;
  justify-content: center;
}

.brand-text {
  font-size: 18px;
  font-weight: 700;
  color: var(--fresh-text-primary);
  letter-spacing: 0.05em;
}

/* Right */
.header-right {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  flex-shrink: 0;
}

/* User Menu */
.user-menu {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: var(--fresh-radius-sm);
  transition: background var(--fresh-transition-fast);

  &:hover {
    background: var(--fresh-bg-hover);
  }
}

.user-avatar {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--fresh-mint-light);
  color: var(--fresh-mint-hover);
  font-size: 13px;
  font-weight: 600;
  border-radius: 50%;
  flex-shrink: 0;
}

.username-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--fresh-text-primary);
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Responsive */
@media (max-width: 768px) {
  .mobile-toggle {
    display: flex;
  }

  .nav-links {
    position: fixed;
    top: var(--fresh-header-height);
    left: 0;
    right: 0;
    background: rgba(255, 255, 255, 0.98);
    backdrop-filter: blur(12px);
    flex-direction: column;
    padding: var(--fresh-space-md);
    border-bottom: 1px solid var(--fresh-border-light);
    transform: translateY(-100%);
    opacity: 0;
    pointer-events: none;
    transition: all 0.25s var(--fresh-ease-out);
    z-index: 999;

    &--open {
      transform: translateY(0);
      opacity: 1;
      pointer-events: auto;
    }

    .nav-link {
      width: 100%;
      padding: 12px 16px;
      font-size: 15px;
    }
  }

  .header-brand {
    justify-content: flex-start;
  }

  .username-text {
    display: none;
  }
}
</style>

<!-- Global dropdown -->
<style lang="scss">
.fresh-dropdown {
  background: var(--fresh-bg-surface) !important;
  border: 1px solid var(--fresh-border-light) !important;
  border-radius: var(--fresh-radius-md) !important;
  box-shadow: var(--fresh-shadow-md) !important;
  padding: 6px !important;

  .el-dropdown-menu__item {
    display: flex !important;
    align-items: center;
    gap: 8px;
    font-family: var(--fresh-font-display) !important;
    font-size: 14px !important;
    color: var(--fresh-text-primary) !important;
    padding: 9px 14px !important;
    border-radius: var(--fresh-radius-xs) !important;
    font-weight: 500;

    &:hover {
      background: var(--fresh-bg-hover) !important;
      color: var(--fresh-mint-hover) !important;
    }
  }

  .el-dropdown-menu__item--divided {
    border-top: 1px solid var(--fresh-border-light) !important;
  }
}
</style>
