<template>
  <div class="article-card" @click="$emit('click', article.id)">
    <div class="card-body">
      <div class="card-info">
        <div class="card-meta">
          <span class="meta-author">{{ article.username || `用户 #${article.user_id}` }}</span>
          <span class="meta-date">{{ formatDate(article.created_at) }}</span>
        </div>

        <h3 class="card-title">{{ article.title }}</h3>
        <p class="card-preview">{{ article.preview }}</p>

        <div class="card-footer">
          <div class="card-tags">
            <el-tag
              v-for="tag in article.tags"
              :key="tag"
              size="small"
              type="info"
              effect="plain"
              @click.stop="$emit('tag-click', tag)"
            >
              {{ tag }}
            </el-tag>
          </div>
          <div class="card-actions">
            <span class="action-fav">
              <el-icon><Star /></el-icon>
              {{ article.likes }}
            </span>
            <span class="read-more">阅读全文 →</span>
          </div>
        </div>
      </div>

      <div v-if="article.cover_img" class="card-cover">
        <el-image
          :src="getImageUrl(article.cover_img)"
          fit="cover"
          class="cover-img"
          lazy
        >
          <template #error>
            <div class="cover-error">
              <el-icon size="20"><Picture /></el-icon>
            </div>
          </template>
        </el-image>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ArticleListItem } from '@/types/article'
import { Star, Picture } from '@element-plus/icons-vue'

defineProps<{
  article: ArticleListItem
}>()

defineEmits<{
  (e: 'click', id: number): void
  (e: 'tag-click', tag: string): void
}>()

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
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days > 30) {
    return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
  }
  if (days > 0) return `${days} 天前`
  const hours = Math.floor(diff / (1000 * 60 * 60))
  if (hours > 0) return `${hours} 小时前`
  const minutes = Math.floor(diff / (1000 * 60))
  if (minutes > 0) return `${minutes} 分钟前`
  return '刚刚'
}
</script>

<style scoped lang="scss">
.article-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #f0f0f0;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
    transform: translateY(-2px);
    border-color: #e0e0e0;
  }

  .card-body {
    display: flex;
    gap: 16px;
  }

  .card-info {
    flex: 1;
    min-width: 0;
  }

  .card-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;

    .meta-avatar {
      flex-shrink: 0;
    }

    .meta-author {
      font-size: 13px;
      font-weight: 500;
      color: #333;
      text-decoration: none;

      &:hover {
        color: #667eea;
      }
    }

    .meta-date {
      font-size: 12px;
      color: #999;
      margin-left: auto;
    }
  }

  .card-title {
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

  .card-preview {
    font-size: 14px;
    color: #666;
    line-height: 1.6;
    margin: 0 0 12px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;
  }

  .card-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .card-actions {
    display: flex;
    align-items: center;
    gap: 12px;

    .action-fav {
      display: flex;
      align-items: center;
      gap: 3px;
      font-size: 12px;
      color: #999;
    }

    .read-more {
      font-size: 13px;
      color: #667eea;
      font-weight: 500;
    }
  }

  .card-cover {
    flex-shrink: 0;
    width: 140px;
    height: 105px;
    border-radius: 8px;
    overflow: hidden;

    .cover-img {
      width: 100%;
      height: 100%;
    }

    .cover-error {
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

@media (max-width: 768px) {
  .article-card {
    padding: 14px;

    .card-body {
      flex-direction: column-reverse;
    }

    .card-cover {
      width: 100%;
      height: 160px;
    }

    .card-title {
      font-size: 16px;
    }
  }
}
</style>