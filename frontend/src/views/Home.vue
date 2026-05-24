<template>
  <div class="home-page">
    <div class="main-container">
      <!-- ===== 欢迎横幅 ===== -->
      <div class="welcome-banner">
        <div class="banner-left">
          <h1 class="banner-title">欢迎回来</h1>
          <p class="banner-desc">探索精彩文章，发现更多有趣的想法</p>
          <div class="banner-stats">
            <div class="stat-item">
              <span class="stat-value">{{ total }}</span>
              <span class="stat-label">篇文章</span>
            </div>
            <div class="stat-dot"></div>
            <div class="stat-item">
              <span class="stat-value">{{ tagsCount }}</span>
              <span class="stat-label">个标签</span>
            </div>
            <div class="stat-dot"></div>
            <div class="stat-item">
              <span class="stat-value">{{ categoriesCount }}</span>
              <span class="stat-label">个分类</span>
            </div>
          </div>
        </div>
        <div class="banner-right">
          <DecorPlant variant="cloud" />
        </div>
      </div>

      <div class="dashboard-grid">
        <!-- 主内容区 -->
        <div class="dashboard-main">
          <!-- 筛选标签 -->
          <div v-if="selectedTag || selectedCategoryId" class="active-filter">
            <span class="filter-label">当前筛选：</span>
            <span v-if="selectedTag" class="filter-chip" @click="clearTagFilter">
              # {{ selectedTag }}
              <span class="filter-close">×</span>
            </span>
            <span v-if="selectedCategoryId" class="filter-chip filter-chip--apricot" @click="clearCategoryFilter">
              {{ selectedCategoryName }}
              <span class="filter-close">×</span>
            </span>
          </div>

          <!-- 文章列表 -->
          <div v-loading="loading" class="article-list" element-loading-background="rgba(250,249,246,0.6)">
            <transition-group name="fade-list" tag="div">
              <article
                v-for="(article, idx) in articles"
                :key="article.id"
                class="article-card card-soft accent-border-left"
                :style="{ animationDelay: idx * 0.04 + 's' }"
                @click="goToArticle(article.id)"
              >
                <div class="article-card-inner">
                  <div class="article-info">
                    <div class="article-meta">
                      <span class="article-avatar">{{ (article.username || 'U').charAt(0) }}</span>
                      <span class="article-author">{{ article.username || `用户 #${article.user_id}` }}</span>
                      <span class="article-date">{{ formatDate(article.created_at) }}</span>
                    </div>

                    <h2 class="article-title">{{ article.title }}</h2>
                    <p class="article-preview">{{ article.preview }}</p>

                    <div class="article-footer-row">
                      <div class="article-tags">
                        <span
                          v-for="tag in article.tags"
                          :key="tag"
                          class="tag-soft"
                          @click.stop="filterByTag(tag)"
                        >{{ tag }}</span>
                      </div>
                      <span class="read-more">阅读全文 →</span>
                    </div>
                  </div>

                  <div v-if="(article as any).cover_img" class="article-cover">
                    <el-image
                      :src="getImageUrl((article as any).cover_img)"
                      fit="cover"
                      class="cover-img"
                      lazy
                    >
                      <template #error>
                        <div class="cover-placeholder">
                          <FreshIcon name="article" :size="24" color="mint" />
                        </div>
                      </template>
                    </el-image>
                  </div>
                </div>
              </article>
            </transition-group>

            <!-- 空状态 -->
            <div v-if="!loading && articles.length === 0" class="empty-state card-soft">
              <DecorPlant variant="sprout" />
              <p class="empty-title">还没有文章</p>
              <p class="empty-desc">暂时没有找到相关内容，换个筛选试试吧</p>
            </div>
          </div>

          <!-- 分页 -->
          <div v-if="total > pageSize" class="pagination-wrapper">
            <button
              class="page-btn"
              :disabled="currentPage <= 1"
              @click="handlePageChange(currentPage - 1)"
            >上一页</button>
            <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
            <button
              class="page-btn"
              :disabled="currentPage >= totalPages"
              @click="handlePageChange(currentPage + 1)"
            >下一页</button>
          </div>
        </div>

        <!-- 侧边栏 -->
        <aside class="dashboard-sidebar">
          <!-- 标签云 -->
          <FreshCard title="热门标签" accent="mint">
            <template #title-icon>
              <FreshIcon name="tag" :size="16" color="mint" />
            </template>
            <div v-loading="tagsLoading" class="tags-cloud" element-loading-background="rgba(250,249,246,0.6)">
              <span
                v-for="tag in tags"
                :key="tag.id"
                class="tag-soft"
                :class="{ 'tag-soft--active': selectedTag === tag.name }"
                @click="filterByTag(tag.name)"
              >{{ tag.name }}</span>
            </div>
            <div v-if="!tagsLoading && tags.length === 0" class="sidebar-empty text-muted">
              暂无标签
            </div>
          </FreshCard>

          <!-- 分类 -->
          <FreshCard title="文章分类" accent="apricot">
            <template #title-icon>
              <FreshIcon name="folder" :size="16" color="apricot" />
            </template>
            <div v-loading="categoriesLoading" class="categories-list" element-loading-background="rgba(250,249,246,0.6)">
              <div
                v-for="cat in categories"
                :key="cat.id"
                class="category-item"
                :class="{ 'category-item--active': selectedCategoryId === cat.id }"
                @click="filterByCategory(cat)"
              >
                <span class="cat-dot"></span>
                <span class="cat-name">{{ cat.name }}</span>
              </div>
            </div>
            <div v-if="!categoriesLoading && categories.length === 0" class="sidebar-empty text-muted">
              暂无分类
            </div>
          </FreshCard>

          <!-- 快捷操作 -->
          <FreshCard title="快捷操作" accent="pink">
            <template #title-icon>
              <FreshIcon name="sparkle" :size="16" color="pink" />
            </template>
            <div class="quick-actions">
              <router-link to="/feed" class="quick-action">
                <FreshIcon name="compass" :size="16" color="mint" />
                <span>关注动态</span>
                <FreshIcon name="arrow-right" :size="14" />
              </router-link>
              <router-link to="/editor" class="quick-action">
                <FreshIcon name="edit" :size="16" color="pink" />
                <span>写文章</span>
                <FreshIcon name="arrow-right" :size="14" />
              </router-link>
              <router-link to="/chat" class="quick-action">
                <FreshIcon name="chat" :size="16" color="apricot" />
                <span>聊天室</span>
                <FreshIcon name="arrow-right" :size="14" />
              </router-link>
            </div>
          </FreshCard>

          <!-- 装饰 -->
          <div class="sidebar-decor">
            <DecorPlant variant="geometric" />
          </div>
        </aside>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useArticleStore } from '@/stores/article'
import { tagsApi } from '@/api/tags'
import type { ArticleListItem } from '@/types/article'
import type { Tag, Category } from '@/types/common'
import FreshCard from '@/components/fresh/FreshCard.vue'
import FreshIcon from '@/components/fresh/FreshIcon.vue'
import DecorPlant from '@/components/fresh/DecorPlant.vue'

const router = useRouter()
const route = useRoute()
const store = useArticleStore()

const loading = ref(false)
const tagsLoading = ref(false)
const categoriesLoading = ref(false)
const articles = ref<ArticleListItem[]>([])
const tags = ref<Tag[]>([])
const categories = ref<Category[]>([])
const selectedTag = ref('')
const selectedCategoryId = ref<number | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const selectedCategoryName = computed(() => {
  if (!selectedCategoryId.value) return ''
  return categories.value.find(c => c.id === selectedCategoryId.value)?.name || ''
})

const selectedTagId = computed(() => {
  if (!selectedTag.value) return undefined
  return tags.value.find(t => t.name === selectedTag.value)?.id
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const tagsCount = computed(() => tags.value.length)
const categoriesCount = computed(() => categories.value.length)

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

const fetchArticles = async () => {
  loading.value = true
  try {
    await store.fetchArticles({
      tag_id: selectedTagId.value,
      category_id: selectedCategoryId.value || undefined,
      page: currentPage.value,
      page_size: pageSize.value,
    })
    articles.value = store.articles
    total.value = store.articlesCount
  } catch (error) {
    console.error('获取文章失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchTags = async () => {
  tagsLoading.value = true
  try { tags.value = await tagsApi.getTags() }
  catch (error) { console.error('获取标签失败:', error) }
  finally { tagsLoading.value = false }
}

const fetchCategories = async () => {
  categoriesLoading.value = true
  try { categories.value = await tagsApi.getCategories() }
  catch (error) { console.error('获取分类失败:', error) }
  finally { categoriesLoading.value = false }
}

const goToArticle = (id: number) => router.push(`/article/${id}`)

const filterByTag = (tag: string) => {
  selectedTag.value = tag
  currentPage.value = 1
  fetchArticles()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const clearTagFilter = () => {
  selectedTag.value = ''
  currentPage.value = 1
  fetchArticles()
}

const filterByCategory = (cat: { id: number; name: string; slug: string }) => {
  selectedCategoryId.value = selectedCategoryId.value === cat.id ? null : cat.id
  currentPage.value = 1
  fetchArticles()
}

const clearCategoryFilter = () => {
  selectedCategoryId.value = null
  currentPage.value = 1
  fetchArticles()
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchArticles()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

watch(
  () => route.query.tag,
  (newTag) => {
    if (newTag && typeof newTag === 'string') selectedTag.value = newTag
    fetchArticles()
  },
  { immediate: false }
)

onMounted(() => {
  if (route.query.tag && typeof route.query.tag === 'string') selectedTag.value = route.query.tag
  fetchArticles()
  fetchTags()
  fetchCategories()
})
</script>

<style scoped lang="scss">
.main-container {
  max-width: var(--fresh-max-width);
  margin: 0 auto;
  padding: var(--fresh-space-xl) var(--fresh-space-lg) var(--fresh-space-3xl);
}

/* ===== 欢迎横幅 ===== */
.welcome-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--fresh-space-xl) var(--fresh-space-2xl);
  margin-bottom: var(--fresh-space-xl);
  background: var(--fresh-bg-surface);
  border-radius: var(--fresh-radius-xl);
  box-shadow: var(--fresh-shadow-sm);
  gap: var(--fresh-space-lg);
  overflow: hidden;
}

.banner-title {
  font-size: var(--fresh-text-3xl);
  font-weight: 700;
  color: var(--fresh-text-primary);
  margin: 0 0 var(--fresh-space-sm);
  letter-spacing: 0.02em;
}

.banner-desc {
  font-size: var(--fresh-text-base);
  color: var(--fresh-text-secondary);
  margin: 0 0 var(--fresh-space-xl);
  max-width: 380px;
}

.banner-stats {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-md);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-value {
  font-size: var(--fresh-text-2xl);
  font-weight: 700;
  color: var(--fresh-mint-hover);
  line-height: 1.2;
}

.stat-label {
  font-size: var(--fresh-text-xs);
  color: var(--fresh-text-muted);
}

.stat-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--fresh-border-default);
}

.banner-right {
  flex-shrink: 0;
}

/* ===== 布局网格 ===== */
.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: var(--fresh-space-lg);
}

.dashboard-main {
  min-width: 0;
}

.dashboard-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--fresh-space-md);
}

/* ===== 活跃筛选 ===== */
.active-filter {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  margin-bottom: var(--fresh-space-md);
  padding: var(--fresh-space-sm) var(--fresh-space-md);
  background: var(--fresh-mint-light);
  border-radius: var(--fresh-radius-sm);
  font-size: 14px;
}

.filter-label {
  color: var(--fresh-text-secondary);
  font-size: 13px;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  background: var(--fresh-bg-surface);
  border-radius: var(--fresh-radius-xs);
  font-size: 13px;
  font-weight: 500;
  color: var(--fresh-mint-hover);
  cursor: pointer;
  transition: all var(--fresh-transition-fast);

  &:hover { background: var(--fresh-bg-hover); }

  &--apricot {
    color: var(--fresh-apricot-hover);
  }
}

.filter-close {
  font-size: 16px;
  opacity: 0.5;
}

/* ===== 文章卡片 ===== */
.article-list {
  min-height: 300px;
}

.article-card {
  padding: var(--fresh-space-lg);
  margin-bottom: var(--fresh-space-md);
  padding-left: calc(var(--fresh-space-lg) + 3px);
}

.article-card-inner {
  display: flex;
  gap: var(--fresh-space-lg);
}

.article-info {
  flex: 1;
  min-width: 0;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  margin-bottom: var(--fresh-space-sm);
}

.article-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--fresh-mint-light);
  color: var(--fresh-mint-hover);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.article-author {
  font-size: 13px;
  font-weight: 500;
  color: var(--fresh-mint-hover);
}

.article-date {
  font-size: 13px;
  color: var(--fresh-text-muted);
  margin-left: auto;
}

.article-title {
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

.article-preview {
  font-size: var(--fresh-text-sm);
  color: var(--fresh-text-secondary);
  line-height: 1.6;
  margin: 0 0 var(--fresh-space-md);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.article-footer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--fresh-space-sm);
}

.article-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.read-more {
  font-size: 13px;
  font-weight: 500;
  color: var(--fresh-mint);
  white-space: nowrap;
}

/* 封面图 */
.article-cover {
  flex-shrink: 0;
  width: 150px;
  height: 110px;
  border-radius: var(--fresh-radius-sm);
  overflow: hidden;

  .cover-img {
    width: 100%;
    height: 100%;
  }
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--fresh-bg-hover);
}

/* ===== 空状态 ===== */
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
}

/* ===== 分页 ===== */
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

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
}

.page-info {
  font-size: 14px;
  color: var(--fresh-text-secondary);
}

/* ===== 侧边栏 ===== */
.tags-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.categories-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.category-item {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  padding: 9px 10px;
  cursor: pointer;
  font-size: 14px;
  color: var(--fresh-text-secondary);
  border-radius: var(--fresh-radius-xs);
  transition: all var(--fresh-transition-fast);

  &:hover {
    background: var(--fresh-bg-hover);
    color: var(--fresh-text-primary);
  }

  &--active {
    background: var(--fresh-apricot-light);
    color: var(--fresh-apricot-hover);
    font-weight: 500;
  }
}

.cat-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--fresh-apricot);
  flex-shrink: 0;
}

/* 快捷操作 */
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.quick-action {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  padding: 10px;
  text-decoration: none;
  color: var(--fresh-text-secondary);
  font-size: 14px;
  font-weight: 500;
  border-radius: var(--fresh-radius-xs);
  transition: all var(--fresh-transition-fast);

  &:hover {
    background: var(--fresh-bg-hover);
    color: var(--fresh-text-primary);
  }
}

.sidebar-empty {
  text-align: center;
  padding: var(--fresh-space-md);
  font-size: 13px;
}

.sidebar-decor {
  display: flex;
  justify-content: center;
  padding: var(--fresh-space-sm) 0;
  opacity: 0.5;
}

/* ===== 过渡动画 ===== */
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
.fade-list-leave-to {
  opacity: 0;
}

/* ===== 响应式 ===== */
@media (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .dashboard-sidebar {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  }

  .sidebar-decor {
    display: none;
  }
}

@media (max-width: 768px) {
  .welcome-banner {
    flex-direction: column;
    align-items: flex-start;
    padding: var(--fresh-space-lg);
    gap: 0;
  }

  .banner-title {
    font-size: var(--fresh-text-2xl);
  }

  .banner-right {
    display: none;
  }

  .article-card-inner {
    flex-direction: column-reverse;
  }

  .article-cover {
    width: 100%;
    height: 160px;
  }

  .article-title {
    font-size: var(--fresh-text-lg);
  }

  .dashboard-sidebar {
    grid-template-columns: 1fr;
  }

  .pagination-wrapper {
    flex-wrap: wrap;
  }
}
</style>
