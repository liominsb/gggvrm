<template>
  <div class="not-found-page">
    <div class="not-found-container">
      <div class="illustration">
        <div class="error-code">404</div>
        <div class="error-emoji">🔍</div>
      </div>
      <h1 class="error-title">页面未找到</h1>
      <p class="error-desc">
        抱歉，你访问的页面不存在或已被移除
      </p>
      <div class="error-actions">
        <el-button type="primary" size="large" @click="goHome" round>
          <el-icon><HomeFilled /></el-icon>
          返回首页
        </el-button>
        <el-button size="large" @click="goBack" round>
          <el-icon><Back /></el-icon>
          返回上一页
        </el-button>
      </div>

      <!-- Suggested Links -->
      <div class="suggestions">
        <h3>你可以试试：</h3>
        <div class="suggestion-links">
          <router-link to="/" class="suggestion-item">
            <el-icon><House /></el-icon>
            <span>浏览文章</span>
          </router-link>
          <router-link to="/feed" class="suggestion-item">
            <el-icon><Compass /></el-icon>
            <span>我的关注</span>
          </router-link>
          <router-link to="/chat" class="suggestion-item">
            <el-icon><ChatDotRound /></el-icon>
            <span>聊天室</span>
          </router-link>
          <router-link to="/editor" class="suggestion-item">
            <el-icon><EditPen /></el-icon>
            <span>写文章</span>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import {
  HomeFilled,
  Back,
  House,
  Compass,
  ChatDotRound,
  EditPen,
} from '@element-plus/icons-vue'

const router = useRouter()

const goHome = () => {
  router.push('/')
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}
</script>

<style scoped lang="scss">
.not-found-page {
  min-height: calc(100vh - 120px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  background: #f5f7fa;
}

.not-found-container {
  text-align: center;
  max-width: 500px;
}

.illustration {
  position: relative;
  margin-bottom: 24px;

  .error-code {
    font-size: 120px;
    font-weight: 800;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    line-height: 1;
    letter-spacing: -4px;
    opacity: 0.3;
  }

  .error-emoji {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 64px;
    animation: float 3s ease-in-out infinite;
  }
}

@keyframes float {
  0%, 100% { transform: translate(-50%, -50%) translateY(0); }
  50% { transform: translate(-50%, -50%) translateY(-10px); }
}

.error-title {
  font-size: 28px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0 0 12px;
}

.error-desc {
  font-size: 16px;
  color: #666;
  margin: 0 0 32px;
  line-height: 1.6;
}

.error-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-bottom: 48px;
}

.suggestions {
  h3 {
    font-size: 15px;
    color: #999;
    font-weight: 500;
    margin: 0 0 16px;
  }

  .suggestion-links {
    display: flex;
    justify-content: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .suggestion-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
    background: #fff;
    border-radius: 10px;
    color: #333;
    text-decoration: none;
    font-size: 14px;
    border: 1px solid #eee;
    transition: all 0.3s ease;

    &:hover {
      border-color: #667eea;
      color: #667eea;
      box-shadow: 0 2px 12px rgba(102, 126, 234, 0.15);
      transform: translateY(-2px);
    }
  }
}

@media (max-width: 768px) {
  .illustration .error-code {
    font-size: 80px;
  }

  .illustration .error-emoji {
    font-size: 48px;
  }

  .error-title {
    font-size: 22px;
  }

  .error-actions {
    flex-direction: column;
    align-items: center;
  }

  .suggestion-links {
    gap: 8px;

    .suggestion-item {
      padding: 8px 14px;
      font-size: 13px;
    }
  }
}
</style>