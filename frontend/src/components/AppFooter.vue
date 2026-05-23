<template>
  <footer class="app-footer">
    <div class="footer-container">
      <div class="footer-content">
        <div class="footer-brand">
          <p class="footer-desc">分享你的想法，探索无限可能</p>
        </div>
        <div class="footer-links">
          <router-link to="/" class="footer-link">首页</router-link>
          <router-link to="/feed" class="footer-link">关注</router-link>
          <router-link to="/chat" class="footer-link">聊天室</router-link>
        </div>
      </div>
      <div class="footer-bottom">
        <span>© {{ currentYear }} All rights reserved.</span>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
const currentYear = computed(() => new Date().getFullYear())
</script>

<style scoped lang="scss">
.app-footer {
  background: #1a1a2e;
  color: #aaa;
  margin-top: auto;
}

.footer-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32px 0 24px;
  flex-wrap: wrap;
  gap: 24px;
}

.footer-brand {
  .footer-desc {
    font-size: 13px;
    margin: 6px 0 0;
    color: #777;
  }
}

.footer-links {
  display: flex;
  gap: 24px;

  .footer-link {
    font-size: 14px;
    color: #888;
    text-decoration: none;
    transition: color 0.2s;

    &:hover {
      color: #667eea;
    }
  }
}

.footer-bottom {
  border-top: 1px solid #2a2a3e;
  padding: 16px 0;
  font-size: 13px;
  color: #555;
  text-align: center;
}
</style>