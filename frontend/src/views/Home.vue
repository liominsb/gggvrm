<template>
  <div class="home-page">
    <!-- Hero Banner -->
    <div class="hero-banner">
      <div class="hero-content">
        <h1 class="hero-title">BlogHub</h1>
        <p class="hero-subtitle">分享你的想法，探索无限可能</p>
      </div>
    </div>

    <div class="main-container">
      <el-row :gutter="24">
        <!-- Main Content -->
        <el-col :xs="24" :sm="24" :md="17" :lg="17" :xl="17">
          <!-- Active Filters -->
          <div v-if="selectedTag || selectedCategoryId" class="active-filter">
            <el-tag v-if="selectedTag" closable size="large" type="primary" @close="clearTagFilter">
              <el-icon><PriceTag /></el-icon>
              {{ selectedTag }}
            </el-tag>
            <el-tag v-if="selectedCategoryId" closable size="large" type="success" @close="clearCategoryFilter">
              <el-icon><Folder /></el-icon>
              {{ selectedCategoryName }}
            </el-tag>
            <span class="filter-hint">筛选结果</span>
          </div>

          <!-- Article List -->
          <div v-loading="loading" class="article-list">
            <transition-group name="fade" tag="div">
              <div
                v-for="article in articles"
                :key="article.id"
                class="article-card"
                @click="goToArticle(article.id)"
              >
                <div class="article-card-body">
                  <div class="article-info">
                    <div class="article-meta-top">
                      <span class="article-author">{{ article.username || `用户 #${article.user_id}` }}</span>
                      <span class="article-date">{{ formatDate(article.created_at) }}</span>
                    </div>

                    <h2 class="article-title">{{ article.title }}</h2>
                    <p class="article-preview">{{ article.preview }}</p>

                    <div class="article-footer">
                      <div class="article-tags">
                        <el-tag
                          v-for="tag in article.tags"
                          :key="tag"
                          size="small"
                          type="info"
                          class="tag-item"
                          @click.stop="filterByTag(tag)"
                        >
                          {{ tag }}
                        </el-tag>
                      </div>
                      <div class="article-actions">
                        <span class="read-more">阅读全文 →</span>
                      </div>
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
                        <div class="image-placeholder">
                          <el-icon size="24"><Picture /></el-icon>
                        </div>
                      </template>
                    </el-image>
                  </div>
                </div>
              </div>
            </transition-group>

            <!-- Empty State -->
            <el-empty
              v-if="!loading && articles.length === 0"
              description="暂无文章"
              :image-size="120"
            />
          </div>

          <!-- Pagination -->
          <div v-if="total > pageSize" class="pagination-wrapper">
            <el-pagination
              v-model:current-page="currentPage"
              :page-size="pageSize"
              :total="total"
              layout="prev, pager, next"
              background
              @current-change="handlePageChange"
            />
          </div>
        </el-col>

        <!-- Sidebar -->
        <el-col :xs="24" :sm="24" :md="7" :lg="7" :xl="7">
          <aside class="sidebar">
            <!-- Popular Tags -->
            <div class="sidebar-card">
              <h3 class="sidebar-title">
                <el-icon><PriceTag /></el-icon>
                热门标签
              </h3>
              <div v-loading="tagsLoading" class="tags-cloud">
                <el-tag
                  v-for="tag in tags"
                  :key="tag.id"
                  :type="selectedTag === tag.name ? '' : 'info'"
                  :effect="selectedTag === tag.name ? 'dark' : 'plain'"
                  class="sidebar-tag"
                  @click="filterByTag(tag.name)"
                >
                  {{ tag.name }}
                </el-tag>
              </div>
              <el-empty
                v-if="!tagsLoading && tags.length === 0"
                description="暂无标签"
                :image-size="60"
              />
            </div>

            <!-- Categories -->
            <div class="sidebar-card">
              <h3 class="sidebar-title">
                <el-icon><Folder /></el-icon>
                分类
              </h3>
              <div v-loading="categoriesLoading" class="categories-list">
                <div
                  v-for="cat in categories"
                  :key="cat.id"
                  class="category-item"
                  :class="{ active: selectedCategoryId === cat.id }"
                  @click="filterByCategory(cat)"
                >
                  {{ cat.name }}
                </div>
              </div>
              <el-empty
                v-if="!categoriesLoading && categories.length === 0"
                description="暂无分类"
                :image-size="60"
              />
            </div>

            <!-- Quick Links -->
            <div class="sidebar-card">
              <h3 class="sidebar-title">
                <el-icon><Link /></el-icon>
                快速导航
              </h3>
              <div class="quick-links">
                <router-link to="/feed" class="quick-link-item">
                  <el-icon><Compass /></el-icon>
                  <span>我的关注</span>
                  <el-icon class="arrow"><ArrowRight /></el-icon>
                </router-link>
                <router-link to="/editor" class="quick-link-item">
                  <el-icon><EditPen /></el-icon>
                  <span>写文章</span>
                  <el-icon class="arrow"><ArrowRight /></el-icon>
                </router-link>
                <router-link to="/chat" class="quick-link-item">
                  <el-icon><ChatDotRound /></el-icon>
                  <span>聊天室</span>
                  <el-icon class="arrow"><ArrowRight /></el-icon>
                </router-link>
              </div>
            </div>
          </aside>
        </el-col>
      </el-row>
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
import {
  Star,
  Picture,
  PriceTag,
  Folder,
  Link,
  Compass,
  ArrowRight,
  EditPen,
  ChatDotRound,
} from '@element-plus/icons-vue'

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

/** 拼接图片完整 URL */
const getImageUrl = (path: string): string => {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `http://localhost:3000${path}`
}

/** 格式化日期 */
const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 30) {
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }
  if (days > 0) return `${days} 天前`
  if (hours > 0) return `${hours} 小时前`
  if (minutes > 0) return `${minutes} 分钟前`
  return '刚刚'
}

/** 获取文章列表 */
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
    console.error('获取文章列表失败:', error)
  } finally {
    loading.value = false
  }
}

/** 获取标签列表 */
const fetchTags = async () => {
  tagsLoading.value = true
  try {
    tags.value = await tagsApi.getTags()
  } catch (error) {
    console.error('获取标签失败:', error)
  } finally {
    tagsLoading.value = false
  }
}

const fetchCategories = async () => {
  categoriesLoading.value = true
  try {
    categories.value = await tagsApi.getCategories()
  } catch (error) {
    console.error('获取分类失败:', error)
  } finally {
    categoriesLoading.value = false
  }
}

/** 跳转文章详情 */
const goToArticle = (id: number) => {
  router.push(`/article/${id}`)
}

/** 按标签筛选 */
const filterByTag = (tag: string) => {
  selectedTag.value = tag
  currentPage.value = 1
  fetchArticles()
}

/** 清除标签筛选 */
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

/** 分页 */
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchArticles()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

// 监听路由 query 中的 tag 参数
watch(
  () => route.query.tag,
  (newTag) => {
    if (newTag && typeof newTag === 'string') {
      selectedTag.value = newTag
    }
    fetchArticles()
  },
  { immediate: false }
)

onMounted(() => {
  if (route.query.tag && typeof route.query.tag === 'string') {
    selectedTag.value = route.query.tag
  }
  fetchArticles()
  fetchTags()
  fetchCategories()
})
</script>

<style scoped lang="scss">
.hero-banner {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 48px 0;
  text-align: center;
  margin-bottom: 32px;

  .hero-content {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
  }

  .hero-title {
    font-size: 42px;
    font-weight: 700;
    color: #fff;
    margin: 0 0 12px;
    letter-spacing: 2px;
  }

  .hero-subtitle {
    font-size: 18px;
    color: rgba(255, 255, 255, 0.85);
    margin: 0;
  }
}

.main-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px 48px;
}

.active-filter {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding: 12px 16px;
  background: #f0f5ff;
  border-radius: 8px;

  .filter-hint {
    font-size: 14px;
    color: #666;
  }
}

.article-list {
  min-height: 300px;
}

.article-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #f0f0f0;

  &:hover {
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
    transform: translateY(-2px);
    border-color: #e0e0e0;
  }

  .article-card-body {
    display: flex;
    gap: 20px;
  }

  .article-info {
    flex: 1;
    min-width: 0;
  }

  .article-meta-top {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;

    .author-avatar {
      flex-shrink: 0;
    }

    .author-name {
      font-size: 14px;
      font-weight: 500;
      color: #333;
      text-decoration: none;

      &:hover {
        color: #667eea;
      }
    }

    .article-author {
      font-size: 13px;
      font-weight: 500;
      color: #555;
    }

    .article-date {
      font-size: 13px;
      color: #999;
      margin-left: auto;
    }
  }

  .article-title {
    font-size: 20px;
    font-weight: 600;
    color: #1a1a1a;
    margin: 0 0 10px;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .article-preview {
    font-size: 15px;
    color: #666;
    line-height: 1.6;
    margin: 0 0 16px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .article-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;
  }

  .article-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;

    .tag-item {
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        transform: scale(1.05);
      }
    }
  }

  .article-actions {
    display: flex;
    align-items: center;
    gap: 16px;

    .action-item {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 13px;
      color: #999;
    }

    .read-more {
      font-size: 14px;
      color: #667eea;
      font-weight: 500;
    }
  }

  .article-cover {
    flex-shrink: 0;
    width: 160px;
    height: 120px;
    border-radius: 8px;
    overflow: hidden;

    .cover-img {
      width: 100%;
      height: 100%;
    }

    .image-placeholder {
      width: 100%;
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #f5f5f5;
      color: #ccc;
    }
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 32px 0;
}

.sidebar {
  position: sticky;
  top: 24px;
}

.sidebar-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid #f0f0f0;

  .sidebar-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 16px;
    font-weight: 600;
    color: #1a1a1a;
    margin: 0 0 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid #f0f0f0;
  }
}

.tags-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  .sidebar-tag {
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      transform: scale(1.05);
    }
  }
}

.categories-list {
  display: flex;
  flex-direction: column;
  gap: 4px;

  .category-item {
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 14px;
    color: #555;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background: #f0f2ff;
      color: #667eea;
    }

    &.active {
      background: #667eea;
      color: #fff;
      font-weight: 500;
    }
  }
}

.quick-links {
  display: flex;
  flex-direction: column;
  gap: 4px;

  .quick-link-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 8px;
    color: #333;
    text-decoration: none;
    font-size: 14px;
    transition: all 0.2s;

    &:hover {
      background: #f5f7fa;
      color: #667eea;
    }

    .arrow {
      margin-left: auto;
      font-size: 12px;
    }
  }
}

// Transition
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}
.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .hero-banner {
    padding: 32px 0;

    .hero-title {
      font-size: 28px;
    }

    .hero-subtitle {
      font-size: 15px;
    }
  }

  .article-card {
    padding: 16px;

    .article-card-body {
      flex-direction: column-reverse;
    }

    .article-cover {
      width: 100%;
      height: 180px;
    }

    .article-title {
      font-size: 17px;
    }
  }
}
</style>