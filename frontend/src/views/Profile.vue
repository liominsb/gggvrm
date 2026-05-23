<template>
  <div class="profile-page">
    <div v-loading="loading" class="main-container">
      <template v-if="profileUser">
        <!-- Profile Header -->
        <div class="profile-header">
          <div class="profile-cover-bg"></div>
          <div class="profile-info">
            <el-avatar
              :size="88"
              :src="getImageUrl((profileUser as any).image)"
              class="profile-avatar"
            />
            <div class="profile-details">
              <h1 class="profile-username">{{ profileUser.username }}</h1>
              <p class="profile-bio">{{ profileUser.bio || '这个人很懒，什么都没写...' }}</p>
              <div class="profile-stats">
                <div class="stat-item">
                  <span class="stat-value">{{ total }}</span>
                  <span class="stat-label">文章</span>
                </div>
                <div class="stat-item">
                  <span class="stat-value">{{ followCounts.followers_count }}</span>
                  <span class="stat-label">粉丝</span>
                </div>
                <div class="stat-item">
                  <span class="stat-value">{{ followCounts.following_count }}</span>
                  <span class="stat-label">关注</span>
                </div>
              </div>
            </div>
            <div v-if="!isOwnProfile && isLoggedIn" class="profile-actions">
              <el-button
                :type="(profileUser as any).following ? 'default' : 'primary'"
                size="large"
                :loading="followLoading"
                @click="toggleFollow"
                round
              >
                {{ (profileUser as any).following ? '已关注' : '+ 关注' }}
              </el-button>
              <el-button
                size="large"
                @click="$router.push('/chat')"
                round
              >
                <el-icon><ChatDotRound /></el-icon>
                私信
              </el-button>
            </div>
            <div v-if="isOwnProfile" class="profile-actions">
              <el-button
                size="large"
                @click="$router.push('/settings')"
                round
              >
                <el-icon><Setting /></el-icon>
                编辑资料
              </el-button>
            </div>
          </div>
        </div>

        <!-- Content Tabs -->
        <div class="content-section">
          <el-tabs v-model="activeTab" class="profile-tabs" @tab-change="handleTabChange">
            <!-- 文章 Tab -->
            <el-tab-pane name="articles">
              <template #label>
                <span class="tab-label">
                  <el-icon><Document /></el-icon>
                  文章
                </span>
              </template>

              <div class="articles-list">
                <div
                  v-for="article in articles"
                  :key="article.id"
                  class="article-item"
                  @click="goToArticle(article.id)"
                >
                  <div class="article-body">
                    <div class="article-meta">
                      <span class="article-date">{{ formatDate(article.created_at) }}</span>
                    </div>
                    <h2 class="article-title">{{ article.title }}</h2>
                    <p class="article-preview">{{ article.preview }}</p>
                    <div class="article-footer">
                      <div class="article-tags">
                        <el-tag
                          v-for="tag in article.tags"
                          :key="getTagName(tag)"
                          size="small"
                          type="info"
                          @click.stop="goToTag(getTagName(tag))"
                        >
                          {{ getTagName(tag) }}
                        </el-tag>
                      </div>
                      <span class="article-likes">
                        <el-icon><Star /></el-icon>
                        {{ article.likes }}
                      </span>
                    </div>
                  </div>
                  <div v-if="article.cover_img" class="article-cover">
                    <el-image
                      :src="getImageUrl(article.cover_img)"
                      fit="cover"
                      class="cover-img"
                      lazy
                    />
                  </div>
                </div>

                <el-empty
                  v-if="!articlesLoading && articles.length === 0"
                  :description="isOwnProfile ? '你还没有发布文章' : '该用户还没有发布文章'"
                  :image-size="100"
                >
                  <el-button
                    v-if="isOwnProfile"
                    type="primary"
                    @click="$router.push('/editor')"
                  >
                    写第一篇文章
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
            </el-tab-pane>

            <!-- 收藏 Tab -->
            <el-tab-pane name="favorited">
              <template #label>
                <span class="tab-label">
                  <el-icon><Star /></el-icon>
                  收藏
                </span>
              </template>

              <div class="articles-list">
                <div
                  v-for="article in favoritedArticles"
                  :key="article.id"
                  class="article-item"
                  @click="goToArticle(article.id)"
                >
                  <div class="article-body">
                    <div class="article-meta">
                      <span class="article-author">{{ article.username || `用户 ${article.user_id}` }}</span>
                      <span class="article-date">{{ formatDate(article.created_at) }}</span>
                    </div>
                    <h2 class="article-title">{{ article.title }}</h2>
                    <p class="article-preview">{{ article.preview }}</p>
                    <div class="article-footer">
                      <div class="article-tags">
                        <el-tag
                          v-for="tag in article.tags"
                          :key="getTagName(tag)"
                          size="small"
                          type="info"
                          @click.stop="goToTag(getTagName(tag))"
                        >
                          {{ getTagName(tag) }}
                        </el-tag>
                      </div>
                      <span class="article-likes">
                        <el-icon><Star /></el-icon>
                        {{ article.likes }}
                      </span>
                    </div>
                  </div>
                  <div v-if="article.cover_img" class="article-cover">
                    <el-image
                      :src="getImageUrl(article.cover_img)"
                      fit="cover"
                      class="cover-img"
                      lazy
                    />
                  </div>
                </div>

                <el-empty
                  v-if="!favLoading && favoritedArticles.length === 0"
                  :description="isOwnProfile ? '你还没有收藏文章' : '该用户还没有收藏文章'"
                  :image-size="100"
                />
              </div>

              <div v-if="favTotal > pageSize" class="pagination-wrapper">
                <el-pagination
                  v-model:current-page="favCurrentPage"
                  :page-size="pageSize"
                  :total="favTotal"
                  layout="prev, pager, next"
                  background
                  @current-change="handleFavPageChange"
                />
              </div>
            </el-tab-pane>

            <!-- 关注 Tab -->
            <el-tab-pane name="following">
              <template #label>
                <span class="tab-label">
                  <el-icon><User /></el-icon>
                  关注
                </span>
              </template>

              <div class="follow-list">
                <div
                  v-for="user in followingList"
                  :key="user.id"
                  class="follow-item"
                  @click="goToProfile(user.id)"
                >
                  <el-avatar
                    :size="48"
                    :src="getImageUrl(user.image)"
                    class="follow-avatar"
                  />
                  <div class="follow-info">
                    <span class="follow-username">{{ user.username }}</span>
                    <span class="follow-bio">{{ user.bio || '这个人很懒，什么都没写...' }}</span>
                  </div>
                </div>

                <el-empty
                  v-if="!followingLoading && followingList.length === 0"
                  description="还没有关注任何人"
                  :image-size="100"
                />
              </div>

              <div v-if="followingTotal > pageSize" class="pagination-wrapper">
                <el-pagination
                  v-model:current-page="followingCurrentPage"
                  :page-size="pageSize"
                  :total="followingTotal"
                  layout="prev, pager, next"
                  background
                  @current-change="handleFollowingPageChange"
                />
              </div>
            </el-tab-pane>

            <!-- 粉丝 Tab -->
            <el-tab-pane name="followers">
              <template #label>
                <span class="tab-label">
                  <el-icon><UserFilled /></el-icon>
                  粉丝
                </span>
              </template>

              <div class="follow-list">
                <div
                  v-for="user in followersList"
                  :key="user.id"
                  class="follow-item"
                  @click="goToProfile(user.id)"
                >
                  <el-avatar
                    :size="48"
                    :src="getImageUrl(user.image)"
                    class="follow-avatar"
                  />
                  <div class="follow-info">
                    <span class="follow-username">{{ user.username }}</span>
                    <span class="follow-bio">{{ user.bio || '这个人很懒，什么都没写...' }}</span>
                  </div>
                </div>

                <el-empty
                  v-if="!followersLoading && followersList.length === 0"
                  description="还没有粉丝"
                  :image-size="100"
                />
              </div>

              <div v-if="followersTotal > pageSize" class="pagination-wrapper">
                <el-pagination
                  v-model:current-page="followersCurrentPage"
                  :page-size="pageSize"
                  :total="followersTotal"
                  layout="prev, pager, next"
                  background
                  @current-change="handleFollowersPageChange"
                />
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </template>

      <!-- User Not Found -->
      <el-result
        v-if="!loading && !profileUser"
        icon="warning"
        title="用户不存在"
        sub-title="该用户可能已被删除或不存在"
      >
        <template #extra>
          <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onActivated, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useArticleStore } from '@/stores/article'
import { userApi } from '@/api/user'
import { ElMessage } from 'element-plus'
import type { ArticleListItem } from '@/types/article'
import type { FollowCountsResponse } from '@/api/user'
import {
  Document,
  Star,
  Setting,
  ChatDotRound,
  User,
  UserFilled,
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const store = useArticleStore()

const loading = ref(false)
const followLoading = ref(false)
const profileUser = ref<any>(null)

// 文章相关
const articles = ref<ArticleListItem[]>([])
const articlesLoading = ref(false)
const currentPage = ref(1)
const total = ref(0)

// 收藏相关
const favoritedArticles = ref<ArticleListItem[]>([])
const favLoading = ref(false)
const favCurrentPage = ref(1)
const favTotal = ref(0)

// 关注相关
const followingList = ref<any[]>([])
const followingLoading = ref(false)
const followingCurrentPage = ref(1)
const followingTotal = ref(0)

// 粉丝相关
const followersList = ref<any[]>([])
const followersLoading = ref(false)
const followersCurrentPage = ref(1)
const followersTotal = ref(0)

// 关注/粉丝数
const followCounts = reactive<FollowCountsResponse>({
  following_count: 0,
  followers_count: 0,
})

const pageSize = ref(10)
const activeTab = ref('articles')

/** 兼容 gorm.Model 的 ID 字段（大写 ID 无 json tag） */
const getObjectId = (obj: any): number => obj?.id ?? obj?.ID ?? 0

const isLoggedIn = computed(() => authStore.isLoggedIn)
const isOwnProfile = computed(() => {
  if (!profileUser.value || !authStore.user) return false
  return getObjectId(authStore.user) === getObjectId(profileUser.value)
})

/** 拼接图片完整 URL */
const getImageUrl = (path: string): string => {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `http://localhost:3000${path}`
}

/** 格式化日期 */
const formatDate = (dateStr: string): string => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days > 30) {
    return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
  }
  if (days > 0) return `${days} 天前`
  return '今天'
}

/** 获取标签名（兼容字符串和对象格式） */
const getTagName = (tag: any): string => {
  if (typeof tag === 'string') return tag
  return tag?.name || ''
}

/** 获取用户资料（通过用户 ID 调用后端 /api/v1/user/:id） */
const fetchProfile = async () => {
  const userId = Number(route.params.id)
  if (!userId || isNaN(userId)) {
    profileUser.value = null
    return
  }
  loading.value = true
  try {
    const res = await userApi.getUserById(userId)
    profileUser.value = res.user

    // 如果不是自己的主页，获取关注状态
    if (isLoggedIn.value && !isOwnProfile.value) {
      try {
        const followRes = await userApi.getFollowStatus(getObjectId(profileUser.value))
        ;(profileUser.value as any).following = followRes.is_following
      } catch {
        // 忽略关注状态获取失败
      }
    }
  } catch (error: any) {
    console.error('[Profile] getUserById error:', error?.message || error)
    profileUser.value = null
  } finally {
    loading.value = false
  }
}

/** 获取关注/粉丝数 */
const fetchFollowCounts = async () => {
  const userId = Number(route.params.id)
  if (!userId) return
  try {
    const data = await userApi.getFollowCounts(userId)
    followCounts.following_count = data.following_count
    followCounts.followers_count = data.followers_count
  } catch (error) {
    console.error('获取关注数失败:', error)
  }
}

/** 获取用户文章（按用户ID过滤） */
const fetchUserArticles = async () => {
  const userId = Number(route.params.id)
  if (!userId) return
  articlesLoading.value = true
  try {
    await store.fetchArticles({
      user_id: userId,
      page: currentPage.value,
      page_size: pageSize.value,
    })
    articles.value = store.articles
    total.value = store.articlesCount
  } catch (error) {
    console.error('获取文章失败:', error)
  } finally {
    articlesLoading.value = false
  }
}

/** 获取收藏文章 —— 使用用户ID专用收藏接口 */
const fetchFavoritedArticles = async () => {
  const userId = Number(route.params.id)
  if (!userId) return
  favLoading.value = true
  try {
    await store.fetchUserFavoritesById(userId, {
      page: favCurrentPage.value,
      page_size: pageSize.value,
    })
    favoritedArticles.value = store.favoritedArticles
    favTotal.value = store.favoritedCount
  } catch (error) {
    console.error('获取收藏失败:', error)
  } finally {
    favLoading.value = false
  }
}

/** 获取关注列表 */
const fetchFollowing = async () => {
  const userId = Number(route.params.id)
  if (!userId) return
  followingLoading.value = true
  try {
    const data = await userApi.getFollowing(userId, {
      page: followingCurrentPage.value,
      page_size: pageSize.value,
    })
    followingList.value = data.data || []
    followingTotal.value = data.total || 0
  } catch (error) {
    console.error('获取关注列表失败:', error)
  } finally {
    followingLoading.value = false
  }
}

/** 获取粉丝列表 */
const fetchFollowers = async () => {
  const userId = Number(route.params.id)
  if (!userId) return
  followersLoading.value = true
  try {
    const data = await userApi.getFollowers(userId, {
      page: followersCurrentPage.value,
      page_size: pageSize.value,
    })
    followersList.value = data.data || []
    followersTotal.value = data.total || 0
  } catch (error) {
    console.error('获取粉丝列表失败:', error)
  } finally {
    followersLoading.value = false
  }
}

/** 切换关注（使用用户 ID 调用后端 /user/:id/follow 路由） */
const toggleFollow = async () => {
  if (!profileUser.value) return
  followLoading.value = true
  try {
    const userId = getObjectId(profileUser.value)
    const wasFollowing = (profileUser.value as any).following
    if (wasFollowing) {
      await userApi.unfollowUserById(userId)
    } else {
      await userApi.followUserById(userId)
    }
    ;(profileUser.value as any).following = !wasFollowing

    // 策略 A：即时更新本地状态，无需等待重新拉取
    if (!wasFollowing) {
      // 关注成功 → 粉丝数 +1
      followCounts.followers_count++
      // 如果当前在粉丝 tab，把当前用户 unshift 进粉丝列表
      if (activeTab.value === 'followers' && authStore.user) {
        const alreadyInList = followersList.value.some(
          (u) => getObjectId(u) === getObjectId(authStore.user)
        )
        if (!alreadyInList) {
          followersList.value.unshift({
            id: getObjectId(authStore.user),
            username: authStore.user.username,
            image: authStore.user.image || '',
            bio: authStore.user.bio || '',
          })
          followersTotal.value++
        }
      }
      // 如果当前在关注 tab（查看的是对方的关注列表），刷新
      if (activeTab.value === 'following') {
        fetchFollowing()
      }
    } else {
      // 取消关注 → 粉丝数 -1
      followCounts.followers_count = Math.max(0, followCounts.followers_count - 1)
      // 如果当前在粉丝 tab，从列表中移除当前用户
      if (activeTab.value === 'followers' && authStore.user) {
        const myId = getObjectId(authStore.user)
        followersList.value = followersList.value.filter(
          (u) => getObjectId(u) !== myId
        )
        if (followersTotal.value > 0) followersTotal.value--
      }
      // 如果当前在关注 tab，刷新
      if (activeTab.value === 'following') {
        fetchFollowing()
      }
    }
    ElMessage.success((profileUser.value as any).following ? '已关注' : '已取消关注')
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    followLoading.value = false
  }
}

const goToArticle = (id: number) => router.push(`/article/${id}`)
const goToTag = (tag: string) => router.push({ path: '/', query: { tag } })
const goToProfile = (id: number) => router.push(`/profile/${id}`)

/** Tab 切换处理 */
const handleTabChange = (tab: string | number) => {
  const tabName = String(tab)
  if (tabName === 'articles') {
    fetchUserArticles()
  } else if (tabName === 'favorited') {
    fetchFavoritedArticles()
  } else if (tabName === 'following') {
    fetchFollowing()
  } else if (tabName === 'followers') {
    fetchFollowers()
  }
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchUserArticles()
  window.scrollTo({ top: 400, behavior: 'smooth' })
}

const handleFavPageChange = (page: number) => {
  favCurrentPage.value = page
  fetchFavoritedArticles()
  window.scrollTo({ top: 400, behavior: 'smooth' })
}

const handleFollowingPageChange = (page: number) => {
  followingCurrentPage.value = page
  fetchFollowing()
  window.scrollTo({ top: 400, behavior: 'smooth' })
}

const handleFollowersPageChange = (page: number) => {
  followersCurrentPage.value = page
  fetchFollowers()
  window.scrollTo({ top: 400, behavior: 'smooth' })
}

/** 初始化页面数据 */
const initPage = () => {
  currentPage.value = 1
  favCurrentPage.value = 1
  followingCurrentPage.value = 1
  followersCurrentPage.value = 1
  activeTab.value = 'articles'
  fetchProfile()
  fetchUserArticles()
  fetchFollowCounts()
}

watch(
  () => route.params.id,
  () => {
    initPage()
  }
)

onMounted(() => {
  initPage()
})

/** Bug 1、2、3、5 修复：KeepAlive 场景下，组件被重新激活时刷新数据 */
onActivated(() => {
  fetchFollowCounts()
  handleTabChange(activeTab.value)
})
</script>

<style scoped lang="scss">
.main-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 0 24px 48px;
}

.profile-header {
  margin-bottom: 32px;

  .profile-cover-bg {
    height: 160px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 0 0 20px 20px;
    margin: 0 -24px;
  }

  .profile-info {
    display: flex;
    align-items: flex-start;
    gap: 20px;
    margin-top: -44px;
    padding: 0 16px;
    flex-wrap: wrap;
  }

  .profile-avatar {
    border: 4px solid #fff;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    flex-shrink: 0;
  }

  .profile-details {
    flex: 1;
    min-width: 0;
    padding-top: 48px;

    .profile-username {
      font-size: 26px;
      font-weight: 700;
      color: #1a1a1a;
      margin: 0 0 8px;
    }

    .profile-bio {
      font-size: 15px;
      color: #666;
      margin: 0 0 12px;
      line-height: 1.5;
    }

    .profile-stats {
      display: flex;
      gap: 24px;

      .stat-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 2px;

        .stat-value {
          font-size: 20px;
          font-weight: 700;
          color: #1a1a1a;
        }

        .stat-label {
          font-size: 13px;
          color: #999;
        }
      }
    }
  }

  .profile-actions {
    display: flex;
    gap: 12px;
    padding-top: 48px;
  }
}

.content-section {
  .profile-tabs {
    :deep(.el-tabs__header) {
      margin-bottom: 24px;
    }

    :deep(.el-tabs__item) {
      font-size: 15px;
      padding: 0 24px;
      height: 48px;
      line-height: 48px;
    }

    .tab-label {
      display: flex;
      align-items: center;
      gap: 6px;
    }
  }
}

.articles-list {
  min-height: 200px;
}

.article-item {
  display: flex;
  gap: 20px;
  padding: 20px;
  background: #fff;
  border-radius: 12px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #f0f0f0;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
    transform: translateY(-1px);
    border-color: #e0e0e0;
  }

  .article-body {
    flex: 1;
    min-width: 0;
  }

  .article-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .article-author {
      font-size: 13px;
      color: #667eea;
      text-decoration: none;
      font-weight: 500;

      &:hover {
        text-decoration: underline;
      }
    }

    .article-date {
      font-size: 13px;
      color: #999;
    }

    .following-tag {
      font-size: 11px;
      color: #52c41a;
      background: #f6ffed;
      padding: 1px 8px;
      border-radius: 4px;
    }
  }

  .article-title {
    font-size: 18px;
    font-weight: 600;
    color: #1a1a1a;
    margin: 0 0 8px;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .article-preview {
    font-size: 14px;
    color: #666;
    line-height: 1.6;
    margin: 0 0 12px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .article-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .article-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .article-likes {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 13px;
      color: #999;
    }
  }

  .article-cover {
    flex-shrink: 0;
    width: 140px;
    height: 105px;
    border-radius: 8px;
    overflow: hidden;

    .cover-img {
      width: 100%;
      height: 100%;
    }
  }
}

/* 关注/粉丝列表样式 */
.follow-list {
  min-height: 200px;
}

.follow-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: #fff;
  border-radius: 12px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #f0f0f0;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
    transform: translateY(-1px);
    border-color: #e0e0e0;
  }

  .follow-avatar {
    flex-shrink: 0;
  }

  .follow-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;

    .follow-username {
      font-size: 16px;
      font-weight: 600;
      color: #1a1a1a;
    }

    .follow-bio {
      font-size: 13px;
      color: #999;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 32px 0;
}

@media (max-width: 768px) {
  .profile-header {
    .profile-cover-bg {
      height: 120px;
    }

    .profile-info {
      flex-direction: column;
      align-items: center;
      text-align: center;
      padding: 0;
    }

    .profile-details {
      padding-top: 12px;

      .profile-stats {
        justify-content: center;
      }
    }

    .profile-actions {
      padding-top: 0;
      justify-content: center;
      width: 100%;
    }
  }

  .article-item {
    flex-direction: column-reverse;
    padding: 16px;

    .article-cover {
      width: 100%;
      height: 160px;
    }
  }
}
</style>
