<template>
  <div class="article-detail-page">
    <div v-loading="loading" class="main-container">
      <template v-if="article">
        <!-- Article Header -->
        <div class="article-header">
          <h1 class="article-title">{{ article.title }}</h1>

          <div class="article-meta">
            <div class="author-section">
              <template v-if="article.user">
                <router-link :to="`/profile/${(article.user as any).ID ?? article.user.id}`" class="author-link">
                  <el-avatar
                    :size="44"
                    :src="getImageUrl(article.user.image)"
                    class="author-avatar"
                  />
                  <div class="author-info">
                    <span class="author-name">{{ article.user.username }}</span>
                    <span class="publish-date">{{ formatDate(article.created_at) }}</span>
                  </div>
                </router-link>
                <el-button
                  v-if="isAuthor"
                  type="primary"
                  plain
                  size="small"
                  @click="editArticle"
                >
                  <el-icon><EditPen /></el-icon>
                  编辑
                </el-button>
                <el-button
                  v-else-if="isLoggedIn"
                  :type="article.user.following ? 'default' : 'primary'"
                  size="small"
                  @click="toggleFollow"
                  :loading="followLoading"
                >
                  {{ article.user.following ? '已关注' : '+ 关注' }}
                </el-button>
              </template>
              <template v-else>
                <span class="publish-date">{{ formatDate(article.created_at) }}</span>
              </template>
            </div>

            <div class="article-tags">
              <el-tag
                v-for="tag in article.tags"
                :key="tag.id"
                type="info"
                class="tag-item"
                @click="goToTag(tag.name)"
              >
                {{ tag.name }}
              </el-tag>
            </div>
          </div>

          <!-- Cover Image -->
          <div v-if="(article as any).cover_img" class="cover-wrapper">
            <el-image
              :src="getImageUrl((article as any).cover_img)"
              fit="cover"
              class="cover-image"
            >
              <template #error>
                <div class="image-placeholder">
                  <el-icon size="48"><Picture /></el-icon>
                  <span>封面图加载失败</span>
                </div>
              </template>
            </el-image>
          </div>
        </div>

        <!-- Article Content -->
        <div class="article-body">
          <div class="article-content" v-html="renderedContent"></div>
        </div>

        <!-- Article Actions Bar -->
        <div class="article-actions-bar">
          <div class="actions-left">
            <el-button
              :type="article.favorited ? 'danger' : 'default'"
              :icon="Star"
              @click="toggleFavorite"
              :loading="favoriteLoading"
              round
            >
              {{ article.favorited ? '已收藏' : '收藏' }}
              <span class="count">{{ article.favorites_count }}</span>
            </el-button>
          </div>

          <div class="actions-right">
            <el-button
              v-if="isAuthor"
              type="danger"
              plain
              @click="handleDelete"
              :loading="deleteLoading"
              round
            >
              <el-icon><Delete /></el-icon>
              删除文章
            </el-button>
          </div>
        </div>

        <!-- Author Card -->
        <div v-if="article.user" class="author-card">
          <div class="author-card-body">
            <router-link :to="`/profile/${(article.user as any).ID ?? article.user.id}`" class="author-card-link">
              <el-avatar
                :size="56"
                :src="getImageUrl(article.user.image)"
                class="card-avatar"
              />
            </router-link>
            <div class="author-card-info">
              <router-link :to="`/profile/${(article.user as any).ID ?? article.user.id}`" class="card-name">
                {{ article.user.username }}
              </router-link>
              <p class="card-bio">{{ article.user.bio || '这个人很懒，什么都没写...' }}</p>
            </div>
            <el-button
              v-if="!isAuthor && isLoggedIn"
              :type="article.user.following ? 'default' : 'primary'"
              @click="toggleFollow"
              :loading="followLoading"
              round
            >
              {{ article.user.following ? '已关注' : '+ 关注' }}
            </el-button>
          </div>
        </div>

        <!-- Comments Section -->
        <div class="comments-section">
          <h3 class="section-title">
            <el-icon><ChatDotSquare /></el-icon>
            评论区 ({{ store.comments.length }})
          </h3>

          <!-- Comment Form -->
          <div v-if="isLoggedIn" class="comment-form">
            <el-input
              v-model="commentContent"
              type="textarea"
              :rows="3"
              placeholder="写下你的评论..."
              maxlength="500"
              show-word-limit
            />
            <div class="comment-form-actions">
              <el-button
                type="primary"
                :loading="commentSubmitting"
                :disabled="!commentContent.trim()"
                @click="submitComment"
              >
                发表评论
              </el-button>
            </div>
          </div>
          <div v-else class="comment-login-hint">
            <router-link to="/login" class="login-link">登录</router-link>后即可发表评论
          </div>

          <!-- Comment List -->
          <div v-if="store.comments.length > 0" class="comment-list">
            <div
              v-for="comment in store.comments"
              :key="comment.id"
              class="comment-item"
            >
              <div class="comment-header">
                <div class="comment-author-info">
                  <el-avatar :size="32" class="comment-avatar">
                    {{ (comment.user_id === currentUserId ? '我' : 'U').charAt(0) }}
                  </el-avatar>
                  <span class="comment-author-name">
                    {{ comment.user_id === currentUserId ? '我' : `用户 ${comment.user_id}` }}
                  </span>
                </div>
                <div class="comment-meta">
                  <span class="comment-time">{{ formatDate(comment.created_at) }}</span>
                  <el-button
                    v-if="comment.user_id === currentUserId || isAuthor"
                    type="danger"
                    text
                    size="small"
                    @click="handleDeleteComment(comment.id)"
                  >
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </div>
              </div>
              <div class="comment-body">{{ comment.content }}</div>
            </div>
          </div>
          <el-empty
            v-else
            description="暂无评论，来说两句吧"
            :image-size="80"
          />
        </div>
      </template>

      <!-- Error State -->
      <el-result
        v-if="!loading && !article"
        icon="warning"
        title="文章不存在"
        sub-title="该文章可能已被删除或你没有访问权限"
      >
        <template #extra>
          <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useArticleStore } from '@/stores/article'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/user'
import { articleApi } from '@/api/article'
import { ElMessageBox, ElMessage } from 'element-plus'
import type { ArticleDetail } from '@/types/article'
import {
  Star,
  Delete,
  EditPen,
  Picture,
  ChatDotSquare,
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const store = useArticleStore()
const authStore = useAuthStore()

const loading = ref(false)
const favoriteLoading = ref(false)
const followLoading = ref(false)
const deleteLoading = ref(false)
const article = ref<ArticleDetail | null>(null)
const commentContent = ref('')
const commentSubmitting = ref(false)
const currentUserId = computed(() => authStore.user?.id || 0)
const articleId = computed(() => String(route.params.id ?? route.params.slug ?? ''))

const isLoggedIn = computed(() => authStore.isLoggedIn)
const currentUser = computed(() => authStore.user?.username || '')
const isAuthor = computed(() => currentUser.value === article.value?.user?.username)

/** 拼接图片完整 URL */
const getImageUrl = (path: string): string => {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `http://localhost:3000${path}`
}

/** 格式化日期 */
const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/** 渲染 Markdown/HTML 内容 */
const renderedContent = computed(() => {
  if (!article.value?.content) return ''
  // 简单处理：后端返回的可能是 markdown 或 HTML
  // 如果是 markdown 需要额外引入 markdown-it 等库
  // 这里假设后端返回的是 HTML 或纯文本
  const content = article.value.content
  // 基础换行处理
  return content
    .replace(/\n\n/g, '</p><p>')
    .replace(/\n/g, '<br>')
    .replace(/^/, '<p>')
    .replace(/$/, '</p>')
})

/** 获取文章详情（含收藏状态 + 作者信息兜底） */
const fetchArticle = async () => {
  if (!articleId.value) {
    article.value = null
    return
  }
  loading.value = true
  try {
    await store.fetchArticle(articleId.value)
    article.value = store.currentArticle

    if (article.value) {
      // 后端详情接口依赖 Redis 缓存，旧缓存可能不含 user 字段
      // 如果 article.user 缺失但 user_id 存在，单独请求作者信息
      if (!article.value.user && article.value.user_id) {
        try {
          const profileRes = await userApi.getUserById(article.value.user_id)
          if (profileRes.user) {
            article.value.user = {
              id: getObjectId(profileRes.user),
              username: profileRes.user.username,
              bio: profileRes.user.bio || '',
              image: profileRes.user.image || '',
              following: false,
              created_at: profileRes.user.created_at || '',
              updated_at: profileRes.user.updated_at || '',
            } as any
          }
        } catch {
          console.warn('获取作者信息失败，关注按钮将不显示')
        }
      }

      // 获取作者的关注状态
      if (article.value.user && isLoggedIn.value && !isAuthor.value) {
        try {
          const authorId = getObjectId(article.value.user)
          if (authorId) {
            const followRes = await userApi.getFollowStatus(authorId)
            article.value.user.following = followRes.is_following
          }
        } catch {
          // 忽略关注状态获取失败
        }
      }

      // 获取收藏状态和收藏数（后端详情接口不返回这两个字段，需要单独请求）
      try {
        const [favStatus, favCount] = await Promise.all([
          articleApi.getFavoriteStatus(String(article.value.id)),
          articleApi.getFavoriteCount(String(article.value.id)),
        ])
        article.value.favorited = favStatus.is_favorited
        article.value.favorites_count = favCount.favorites
      } catch {
        if (article.value) {
          article.value.favorited = false
          article.value.favorites_count = 0
        }
      }
    }
  } catch (error) {
    console.error('获取文章详情失败:', error)
    article.value = null
  } finally {
    loading.value = false
  }
}

/** 切换收藏 */
const toggleFavorite = async () => {
  if (!isLoggedIn.value) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!article.value) return
  favoriteLoading.value = true
  try {
    const data = await store.toggleFavorite(String(article.value.id))
    // 后端返回 { message, is_favorited }
    article.value.favorited = data.is_favorited
    // 收藏数 +1 或 -1
    article.value.favorites_count = (article.value.favorites_count || 0) + (data.is_favorited ? 1 : -1)
    ElMessage.success(data.message || (data.is_favorited ? '收藏成功' : '取消收藏'))
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    favoriteLoading.value = false
  }
}

/** 兼容 gorm.Model 的 ID 字段（大写 ID 无 json tag） */
const getObjectId = (obj: any): number => obj?.id ?? obj?.ID ?? 0

/** 切换关注（使用用户 ID 调用后端 /user/:id/follow 路由） */
const toggleFollow = async () => {
  if (!isLoggedIn.value || !article.value?.user) return
  followLoading.value = true
  try {
    const userId = getObjectId(article.value.user)
    if (!userId) {
      ElMessage.error('无法获取用户ID')
      return
    }
    // 后端是 toggle 接口，POST 即可切换
    await userApi.followUserById(userId)
    article.value.user.following = !article.value.user.following
    ElMessage.success(article.value.user.following ? '已关注' : '已取消关注')
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    followLoading.value = false
  }
}

/** 编辑文章 */
const editArticle = () => {
  if (article.value) {
    router.push(`/editor/${article.value.id}`)
  }
}

/** 删除文章 */
const handleDelete = async () => {
  if (!article.value) return
  try {
    await ElMessageBox.confirm('确定要删除这篇文章吗？此操作不可撤销。', '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    deleteLoading.value = true
    await store.deleteArticle(String(article.value.id))
    ElMessage.success('文章已删除')
    router.push('/')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  } finally {
    deleteLoading.value = false
  }
}

/** 提交评论 */
const submitComment = async () => {
  if (!article.value || !commentContent.value.trim()) return
  commentSubmitting.value = true
  try {
    await store.createComment(String(article.value.id), commentContent.value.trim())
    commentContent.value = ''
    ElMessage.success('评论发表成功')
  } catch (error) {
    ElMessage.error('评论发表失败')
  } finally {
    commentSubmitting.value = false
  }
}

/** 删除评论 */
const handleDeleteComment = async (commentId: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这条评论吗？', '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await store.deleteComment(commentId)
    ElMessage.success('评论已删除')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

/** 跳转标签页 */
const goToTag = (tagName: string) => {
  router.push({ path: '/', query: { tag: tagName } })
}

onMounted(() => {
  fetchArticle()
})

// 监听路由参数变化，确保在 SPA 内导航时也能刷新文章
watch(
  articleId,
  (newArticleId) => {
    if (newArticleId) {
      fetchArticle()
    }
  }
)
</script>

<style scoped lang="scss">
.main-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 32px 24px 48px;
}

.article-header {
  margin-bottom: 32px;

  .article-title {
    font-size: 34px;
    font-weight: 700;
    color: #1a1a1a;
    line-height: 1.3;
    margin: 0 0 24px;
  }

  .article-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
    margin-bottom: 24px;

    .author-section {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .author-link {
      display: flex;
      align-items: center;
      gap: 12px;
      text-decoration: none;
      color: inherit;

      .author-info {
        display: flex;
        flex-direction: column;
        gap: 2px;
      }

      .author-name {
        font-size: 16px;
        font-weight: 600;
        color: #333;
      }

      .publish-date {
        font-size: 13px;
        color: #999;
      }
    }

    .article-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;

      .tag-item {
        cursor: pointer;
        transition: transform 0.2s;

        &:hover {
          transform: scale(1.05);
        }
      }
    }
  }

  .cover-wrapper {
    border-radius: 12px;
    overflow: hidden;
    margin-bottom: 8px;

    .cover-image {
      width: 100%;
      max-height: 420px;
      display: block;
    }

    .image-placeholder {
      width: 100%;
      height: 240px;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 8px;
      background: #f5f5f5;
      color: #ccc;

      span {
        font-size: 14px;
      }
    }
  }
}

.article-body {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  margin-bottom: 24px;
  border: 1px solid #f0f0f0;

  .article-content {
    font-size: 17px;
    line-height: 1.8;
    color: #333;
    word-break: break-word;

    :deep(p) {
      margin: 0 0 16px;
    }

    :deep(img) {
      max-width: 100%;
      border-radius: 8px;
      margin: 16px 0;
    }

    :deep(h1),
    :deep(h2),
    :deep(h3) {
      margin: 24px 0 12px;
      color: #1a1a1a;
    }

    :deep(a) {
      color: #667eea;
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }

    :deep(blockquote) {
      border-left: 4px solid #667eea;
      margin: 16px 0;
      padding: 12px 20px;
      background: #f8f9ff;
      border-radius: 0 8px 8px 0;
      color: #555;
    }

    :deep(code) {
      background: #f5f5f5;
      padding: 2px 6px;
      border-radius: 4px;
      font-size: 0.9em;
      color: #e83e8c;
    }

    :deep(pre) {
      background: #282c34;
      color: #abb2bf;
      padding: 20px;
      border-radius: 8px;
      overflow-x: auto;
      margin: 16px 0;

      code {
        background: none;
        color: inherit;
        padding: 0;
      }
    }
  }
}

.article-actions-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  margin-bottom: 32px;
  border-bottom: 1px solid #f0f0f0;

  .count {
    margin-left: 6px;
  }
}

.author-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 32px;
  border: 1px solid #f0f0f0;

  .author-card-body {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .author-card-link {
    flex-shrink: 0;
  }

  .author-card-info {
    flex: 1;
    min-width: 0;

    .card-name {
      font-size: 18px;
      font-weight: 600;
      color: #333;
      text-decoration: none;
      display: block;
      margin-bottom: 4px;

      &:hover {
        color: #667eea;
      }
    }

    .card-bio {
      font-size: 14px;
      color: #666;
      margin: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

.comments-section {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  border: 1px solid #f0f0f0;

  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 600;
    color: #1a1a1a;
    margin: 0 0 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid #f0f0f0;
  }

  .comment-form {
    margin-bottom: 24px;

    .comment-form-actions {
      display: flex;
      justify-content: flex-end;
      margin-top: 12px;
    }
  }

  .comment-login-hint {
    text-align: center;
    padding: 16px 0;
    color: #999;
    font-size: 14px;
    margin-bottom: 20px;
    background: #fafafa;
    border-radius: 8px;

    .login-link {
      color: #667eea;
      text-decoration: none;
      font-weight: 500;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  .comment-list {
    .comment-item {
      padding: 16px 0;
      border-bottom: 1px solid #f5f5f5;

      &:last-child {
        border-bottom: none;
      }

      .comment-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 8px;

        .comment-author-info {
          display: flex;
          align-items: center;
          gap: 10px;

          .comment-avatar {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #fff;
            font-size: 14px;
            font-weight: 600;
          }

          .comment-author-name {
            font-size: 14px;
            font-weight: 600;
            color: #333;
          }
        }

        .comment-meta {
          display: flex;
          align-items: center;
          gap: 8px;

          .comment-time {
            font-size: 12px;
            color: #999;
          }
        }
      }

      .comment-body {
        font-size: 15px;
        color: #444;
        line-height: 1.6;
        padding-left: 42px;
      }
    }
  }
}

@media (max-width: 768px) {
  .main-container {
    padding: 16px 16px 32px;
  }

  .article-header .article-title {
    font-size: 24px;
  }

  .article-body {
    padding: 20px;

    .article-content {
      font-size: 16px;
    }
  }

  .author-card .author-card-body {
    flex-wrap: wrap;
  }
}
</style>
