<template>
  <footer class="app-footer">
    <div class="footer-container">
      <div class="footer-content">
        <div class="footer-left">
          <span class="footer-brand">博客空间</span>
          <span class="footer-desc">分享想法，发现美好</span>
        </div>
        <div class="footer-links">
          <router-link to="/" class="footer-link">首页</router-link>
          <router-link to="/feed" class="footer-link">关注</router-link>
          <router-link to="/chat" class="footer-link">聊天室</router-link>
        </div>
        <div class="footer-right">
          <span class="footer-copy">&copy; {{ currentYear }} 博客空间</span>
        </div>
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
  margin-top: auto;
  background: var(--fresh-bg-surface);
  border-top: 1px solid var(--fresh-border-light);
}

.footer-container {
  max-width: var(--fresh-max-width);
  margin: 0 auto;
  padding: 0 var(--fresh-space-lg);
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--fresh-space-lg) 0;
  flex-wrap: wrap;
  gap: var(--fresh-space-md);
}

.footer-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.footer-brand {
  font-size: 15px;
  font-weight: 600;
  color: var(--fresh-text-primary);
}

.footer-desc {
  font-size: 13px;
  color: var(--fresh-text-muted);
}

.footer-links {
  display: flex;
  gap: var(--fresh-space-lg);
}

.footer-link {
  font-size: 14px;
  color: var(--fresh-text-secondary);
  text-decoration: none;
  transition: color var(--fresh-transition-fast);

  &:hover {
    color: var(--fresh-mint-hover);
  }
}

.footer-right {
  font-size: 13px;
  color: var(--fresh-text-muted);
}

@media (max-width: 768px) {
  .footer-content {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }
}
</style>
