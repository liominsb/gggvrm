<template>
  <div class="feed-page">
    <div class="main-container">
      <!-- 未登录 -->
      <div v-if="!isLoggedIn" class="empty-state card-soft">
        <DecorPlant variant="potted" />
        <h2 class="empty-title">登录后查看关注动态</h2>
        <p class="empty-desc">登录你的账号，关注感兴趣的作者，获取最新内容推送</p>
        <div class="empty-actions">
          <FreshButton variant="mint" size="md" @click="$router.push('/login')">
            立即登录
          </FreshButton>
          <FreshButton variant="outline" size="md" @click="$router.push('/register')">
            注册账号
          </FreshButton>
        </div>
      </div>

      <!-- 已登录 -->
      <template v-else>
        <div class="feed-header">
          <h1 class="feed-title">我的关注</h1>
          <span class="feed-count">{{ total }} 条动态</span>
        </div>

        <div v-loading="loading" class="feed-list" element-loading-background="rgba(250,249,246,0.6)">
          <transition-group name="fade-list" tag="div">
            <article
              v-for="(article, idx) in feedArticles"
              :key="article.id"
              class="feed-card card-soft accent-border-left"
              :style="{ animationDelay: idx * 0.04 + 's' }"
              @click="goToArticle(article.id)"
            >
              <div class="feed-card-header">
                <div class="author-info">
                  <span class="author-avatar">{{ (article.username || 'U').charAt(0) }}</span>
                  <span class="author-name">{{ article.username || `用户 #${article.user_id}` }}</span>
                  <span class="feed-time">{{ formatDate(article.created_at) }}</span>
                </div>
              </div>

              <div class="feed-card-body">
                <div class="feed-text">
                  <h2 class="feed-article-title">{{ article.title }}</h2>
                  <p class="feed-preview">{{ article.preview }}</p>
                  <div class="feed-tags">
                    <span
                      v-for="tag in article.tags"
                      :key="tag"
                      class="tag-soft"
                      @click.stop="goToTag(tag)"
                    >{{ tag }}</span>
                  </div>
                </div>
                <div v-if="(article as any).cover_img" class="feed-cover">
                  <el-image :src="getImageUrl((article as any).cover_img)" fit="cover" class="cover-img" lazy>
                    <template #error>
                      <div class="cover-placeholder">
                        <FreshIcon name="article" :size="20" color="mint" />
                      </div>
                    </template>
                  </el-image>
                </div>
              </div>

              <div class="feed-card-footer">
                <span class="footer-stat">
                  <FreshIcon name="heart" :size="12" color="pink" />
                  <span>{{ article.likes }} 赞</span>
                </span>
                <span class="read-more-link">阅读全文 →</span>
              </div>
            </article>
          </transition-group>

          <!-- 空 -->
          <div v-if="!loading && feedArticles.length === 0" class="empty-state card-soft">
            <DecorPlant variant="sprout" />
            <p class="empty-title">暂无关注动态</p>
            <p class="empty-desc">你关注的作者还没有发布文章，去发现更多有趣的人吧</p>
            <FreshButton variant="mint" size="sm" @click="$router.push('/')">
              浏览推荐
            </FreshButton>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="total > pageSize" class="pagination-wrapper">
          <button class="page-btn" :disabled="currentPage <= 1" @click="handlePageChange(currentPage - 1)">
            上一页
          </button>
          <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
          <button class="page-btn" :disabled="currentPage >= totalPages" @click="handlePageChange(currentPage + 1)">
            下一页
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useArticleStore } from '@/stores/article'
import { useAuthStore } from '@/stores/auth'
import type { ArticleListItem } from '@/types/article'
import FreshIcon from '@/components/fresh/FreshIcon.vue'
import FreshButton from '@/components/fresh/FreshButton.vue'
import DecorPlant from '@/components/fresh/DecorPlant.vue'

const router = useRouter()
const store = useArticleStore()
const authStore = useAuthStore()

const loading = ref(false)
const feedArticles = ref<ArticleListItem[]>([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const isLoggedIn = computed(() => authStore.isLoggedIn)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const getImageUrl = (path: string): string => {
  if (!path || path.startsWith('http')) return path
  return path
}

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / 86400000)

  if (days > 30) {
    return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
  }
  if (days > 0) return `${days} 天前`
  const hours = Math.floor(diff / 3600000)
  if (hours > 0) return `${hours} 小时前`
  return '刚刚'
}

const fetchFeed = async () => {
  if (!isLoggedIn.value) return
  loading.value = true
  try {
    await store.fetchFeed({ limit: pageSize.value, offset: (currentPage.value - 1) * pageSize.value })
    feedArticles.value = store.feedArticles
    total.value = store.articlesCount
  } catch (error) {
    console.error('获取动态失败:', error)
  } finally {
    loading.value = false
  }
}

const goToArticle = (id: number) => router.push(`/article/${id}`)
const goToTag = (tag: string) => router.push({ path: '/', query: { tag } })

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchFeed()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => { fetchFeed() })
</script>

<style scoped lang="scss">
.main-container {
  max-width: 800px;
  margin: 0 auto;
  padding: var(--fresh-space-xl) var(--fresh-space-lg) var(--fresh-space-3xl);
}

.feed-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: var(--fresh-space-lg);
}

.feed-title {
  font-size: var(--fresh-text-2xl);
  font-weight: 700;
  color: var(--fresh-text-primary);
  margin: 0;
}

.feed-count {
  font-size: 14px;
  color: var(--fresh-text-muted);
}

/* Cards */
.feed-card {
  padding: var(--fresh-space-lg);
  margin-bottom: var(--fresh-space-md);
  padding-left: calc(var(--fresh-space-lg) + 3px);
}

.feed-card-header {
  margin-bottom: var(--fresh-space-md);
}

.author-info {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  flex-wrap: wrap;
}

.author-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--fresh-pink-light);
  color: var(--fresh-pink-hover);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.author-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--fresh-mint-hover);
}

.feed-time {
  font-size: 13px;
  color: var(--fresh-text-muted);
}

.feed-card-body {
  display: flex;
  gap: var(--fresh-space-lg);
  margin-bottom: var(--fresh-space-md);
}

.feed-text {
  flex: 1;
  min-width: 0;
}

.feed-article-title {
  font-size: var(--fresh-text-xl);
  font-weight: 600;
  color: var(--fresh-text-primary);
  margin: 0 0 var(--fresh-space-sm);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.feed-preview {
  font-size: var(--fresh-text-sm);
  color: var(--fresh-text-secondary);
  line-height: 1.6;
  margin: 0 0 var(--fresh-space-sm);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.feed-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.feed-cover {
  flex-shrink: 0;
  width: 150px;
  height: 110px;
  border-radius: var(--fresh-radius-sm);
  overflow: hidden;

  .cover-img { width: 100%; height: 100%; }
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--fresh-bg-hover);
}

.feed-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: var(--fresh-space-sm);
  border-top: 1px solid var(--fresh-border-light);
}

.footer-stat {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--fresh-text-muted);
}

.read-more-link {
  font-size: 13px;
  font-weight: 500;
  color: var(--fresh-mint);
}

/* Empty */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--fresh-space-3xl) var(--fresh-space-md);
  text-align: center;
  gap: var(--fresh-space-sm);
}

.empty-title {
  font-size: var(--fresh-text-lg);
  font-weight: 600;
  color: var(--fresh-text-primary);
  margin: 0;
}

.empty-desc {
  font-size: var(--fresh-text-sm);
  color: var(--fresh-text-muted);
  margin: 0;
  max-width: 320px;
}

.empty-actions {
  display: flex;
  gap: var(--fresh-space-sm);
  margin-top: var(--fresh-space-sm);
}

/* Pagination */
.pagination-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--fresh-space-md);
  padding: var(--fresh-space-xl) 0;
}

.page-btn {
  font-size: 14px;
  font-weight: 500;
  color: var(--fresh-mint);
  background: var(--fresh-bg-surface);
  border: 1px solid var(--fresh-border-light);
  padding: 8px 20px;
  border-radius: var(--fresh-radius-sm);
  cursor: pointer;
  transition: all var(--fresh-transition-fast);

  &:hover:not(:disabled) {
    border-color: var(--fresh-mint);
    background: var(--fresh-mint-light);
  }

  &:disabled { opacity: 0.4; cursor: not-allowed; }
}

.page-info {
  font-size: 14px;
  color: var(--fresh-text-secondary);
}

/* Transitions */
.fade-list-enter-active {
  transition: opacity 0.3s var(--fresh-ease-out), transform 0.3s var(--fresh-ease-out);
}
.fade-list-leave-active {
  transition: opacity 0.15s var(--fresh-ease-out);
}
.fade-list-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.fade-list-leave-to { opacity: 0; }

@media (max-width: 768px) {
  .feed-card-body { flex-direction: column-reverse; }
  .feed-cover { width: 100%; height: 160px; }
  .feed-article-title { font-size: var(--fresh-text-lg); }
}
</style>
