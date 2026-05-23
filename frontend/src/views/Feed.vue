<template>
  <div class="feed-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <el-icon><Compass /></el-icon>
          我的关注
        </h1>
        <p class="page-desc">查看你关注的作者发布的最新文章</p>
      </div>
    </div>

    <div class="main-container">
      <!-- Not Logged In State -->
      <div v-if="!isLoggedIn" class="login-prompt">
        <el-card class="prompt-card" shadow="hover">
          <div class="prompt-content">
            <el-icon class="prompt-icon" size="48"><UserFilled /></el-icon>
            <h2>登录后查看关注动态</h2>
            <p>登录你的账号，关注感兴趣的作者，获取最新内容推送</p>
            <div class="prompt-actions">
              <el-button type="primary" size="large" @click="$router.push('/login')">
                立即登录
              </el-button>
              <el-button size="large" @click="$router.push('/register')">
                注册账号
              </el-button>
            </div>
          </div>
        </el-card>
      </div>

      <!-- Logged In Content -->
      <template v-else>
        <!-- Feed List -->
        <div v-loading="loading" class="feed-list">
          <transition-group name="fade" tag="div">
            <div
              v-for="article in feedArticles"
              :key="article.id"
              class="feed-card"
              @click="goToArticle(article.id)"
            >
              <div class="feed-card-header">
                <div class="author-info">
                  <span class="author-name">{{ article.username || `用户 #${article.user_id}` }}</span>
                  <span class="article-time">{{ formatDate(article.created_at) }}</span>
                </div>
              </div>

              <div class="feed-card-body">
                <div class="feed-text">
                  <h2 class="feed-title">{{ article.title }}</h2>
                  <p class="feed-preview">{{ article.preview }}</p>
                  <div class="feed-tags">
                    <el-tag
                      v-for="tag in article.tags"
                      :key="tag"
                      size="small"
                      type="info"
                      class="tag-item"
                      @click.stop="goToTag(tag)"
                    >
                      {{ tag }}
                    </el-tag>
                  </div>
                </div>
                <div v-if="(article as any).cover_img" class="feed-cover">
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

              <div class="feed-card-footer">
                <span class="footer-action">
                  <el-icon><Star /></el-icon>
                  {{ article.likes }} 点赞
                </span>
                <span class="read-more-link">阅读全文 →</span>
              </div>
            </div>
          </transition-group>

          <!-- Empty State -->
          <el-empty
            v-if="!loading && feedArticles.length === 0"
            description="暂无关注动态"
            :image-size="120"
          >
            <template #description>
              <p>你关注的作者还没有发布文章</p>
              <p>去发现更多有趣的作者吧！</p>
            </template>
            <el-button type="primary" @click="$router.push('/')">
              浏览推荐文章
            </el-button>
          </el-empty>
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
import {
  Compass,
  UserFilled,
  Star,
  Picture,
} from '@element-plus/icons-vue'

const router = useRouter()
const store = useArticleStore()
const authStore = useAuthStore()

const loading = ref(false)
const feedArticles = ref<ArticleListItem[]>([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const isLoggedIn = computed(() => authStore.isLoggedIn)

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

/** 获取 Feed 列表 */
const fetchFeed = async () => {
  if (!isLoggedIn.value) return
  loading.value = true
  try {
    await store.fetchFeed({
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
    })
    feedArticles.value = store.feedArticles
    total.value = store.articlesCount
  } catch (error) {
    console.error('获取 Feed 失败:', error)
  } finally {
    loading.value = false
  }
}

/** 跳转文章详情 */
const goToArticle = (id: number) => {
  router.push(`/article/${id}`)
}

/** 按标签跳转 */
const goToTag = (tag: string) => {
  router.push({ path: '/', query: { tag } })
}

/** 分页 */
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchFeed()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  fetchFeed()
})
</script>

<style scoped lang="scss">
.page-header {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  padding: 40px 0;
  margin-bottom: 32px;

  .header-content {
    max-width: 800px;
    margin: 0 auto;
    padding: 0 24px;
  }

  .page-title {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 32px;
    font-weight: 700;
    color: #fff;
    margin: 0 0 8px;
  }

  .page-desc {
    font-size: 16px;
    color: rgba(255, 255, 255, 0.85);
    margin: 0;
  }
}

.main-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 0 24px 48px;
}

.login-prompt {
  margin-top: 20px;

  .prompt-card {
    border-radius: 16px;
    text-align: center;
    padding: 20px;
  }

  .prompt-content {
    padding: 32px 0;
  }

  .prompt-icon {
    color: #667eea;
    margin-bottom: 16px;
  }

  h2 {
    font-size: 22px;
    color: #1a1a1a;
    margin: 0 0 8px;
  }

  p {
    font-size: 15px;
    color: #666;
    margin: 0 0 24px;
  }

  .prompt-actions {
    display: flex;
    justify-content: center;
    gap: 12px;
  }
}

.feed-card {
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

  .feed-card-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;

    .author-avatar {
      flex-shrink: 0;
    }

    .author-info {
      display: flex;
      flex-direction: column;
      gap: 2px;
      flex: 1;
      min-width: 0;
    }

    .author-name {
      font-size: 15px;
      font-weight: 600;
      color: #333;
      text-decoration: none;

      &:hover {
        color: #667eea;
      }
    }

    .article-time {
      font-size: 13px;
      color: #999;
    }

    .following-badge {
      flex-shrink: 0;
    }
  }

  .feed-card-body {
    display: flex;
    gap: 20px;
    margin-bottom: 16px;

    .feed-text {
      flex: 1;
      min-width: 0;
    }

    .feed-title {
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

    .feed-preview {
      font-size: 15px;
      color: #666;
      line-height: 1.6;
      margin: 0 0 12px;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .feed-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;

      .tag-item {
        cursor: pointer;

        &:hover {
          transform: scale(1.05);
        }
      }
    }

    .feed-cover {
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

  .feed-card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 12px;
    border-top: 1px solid #f5f5f5;

    .footer-action {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 13px;
      color: #999;
    }

    .read-more-link {
      font-size: 14px;
      color: #667eea;
      font-weight: 500;
    }
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 32px 0;
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
  .page-header {
    padding: 28px 0;

    .page-title {
      font-size: 24px;
    }
  }

  .feed-card {
    padding: 16px;

    .feed-card-body {
      flex-direction: column-reverse;
    }

    .feed-cover {
      width: 100%;
      height: 180px;
    }

    .feed-title {
      font-size: 17px;
    }
  }
}
</style>