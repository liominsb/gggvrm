<template>
  <div class="search-page">
    <div class="main-container">
      <!-- 搜索头部 -->
      <div class="search-header">
        <div class="search-bar-wrapper">
          <FreshIcon name="search" :size="18" color="mint" />
          <input
            ref="searchInputRef"
            v-model="keyword"
            class="search-input"
            type="text"
            placeholder="输入关键词，使用 AI 语义搜索..."
            @keyup.enter="handleSearch"
          />
          <FreshButton
            variant="mint"
            size="sm"
            :disabled="searching"
            @click="handleSearch"
          >
            搜索
          </FreshButton>
        </div>
        <p v-if="searchedKeyword" class="search-summary">
          「{{ searchedKeyword }}」的语义搜索结果，共 {{ results.length }} 条
        </p>
      </div>

      <!-- 搜索结果 -->
      <div v-loading="searching" class="results-list" element-loading-background="rgba(250,249,246,0.6)">
        <transition-group name="fade-list" tag="div">
          <article
            v-for="(item, idx) in results"
            :key="item.article_id"
            class="result-card card-soft accent-border-left"
            :style="{ animationDelay: idx * 0.04 + 's' }"
            @click="goToArticle(item.article_id)"
          >
            <div class="result-inner">
              <div class="result-icon">
                <FreshIcon name="article" :size="20" color="mint" />
              </div>
              <div class="result-body">
                <h3 class="result-title">{{ item.title }}</h3>
                <span class="result-hint">点击查看文章详情 →</span>
              </div>
            </div>
          </article>
        </transition-group>

        <!-- 空状态 -->
        <div v-if="!searching && searchedKeyword && results.length === 0" class="empty-state card-soft">
          <FreshIcon name="search" :size="40" />
          <p class="empty-title">没有找到相关文章</p>
          <p class="empty-desc">换个关键词试试吧</p>
        </div>

        <!-- 初始状态 -->
        <div v-if="!searching && !searchedKeyword" class="empty-state card-soft">
          <FreshIcon name="sparkle" :size="40" color="pink" />
          <p class="empty-title">AI 语义搜索</p>
          <p class="empty-desc">输入关键词，基于向量相似度为你匹配最相关的文章</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useArticleStore } from '@/stores/article'
import FreshIcon from '@/components/fresh/FreshIcon.vue'
import FreshButton from '@/components/fresh/FreshButton.vue'

const router = useRouter()
const route = useRoute()
const store = useArticleStore()

const keyword = ref('')
const searchedKeyword = ref('')
const searching = ref(false)
const searchInputRef = ref<HTMLInputElement | null>(null)

const results = ref<{ article_id: string; title: string }[]>([])

const handleSearch = async () => {
  const kw = keyword.value.trim()
  if (!kw) return
  searchedKeyword.value = kw
  searching.value = true
  try {
    await store.searchRag(kw)
    results.value = store.ragResults
    // 更新 URL query 参数，方便分享
    router.replace({ query: { keyword: kw } })
  } catch (error) {
    console.error('RAG 搜索失败:', error)
    results.value = []
  } finally {
    searching.value = false
  }
}

const goToArticle = (id: string) => {
  router.push(`/article/${id}`)
}

onMounted(async () => {
  // 如果 URL 带 keyword 参数，自动搜索
  const q = route.query.keyword
  if (q && typeof q === 'string') {
    keyword.value = q
    await handleSearch()
  }
  await nextTick()
  searchInputRef.value?.focus()
})
</script>

<style scoped lang="scss">
.main-container {
  max-width: var(--fresh-max-width);
  margin: 0 auto;
  padding: var(--fresh-space-xl) var(--fresh-space-lg) var(--fresh-space-3xl);
}

/* ===== 搜索头部 ===== */
.search-header {
  margin-bottom: var(--fresh-space-xl);
}

.search-bar-wrapper {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  padding: var(--fresh-space-md) var(--fresh-space-lg);
  background: var(--fresh-bg-surface);
  border-radius: var(--fresh-radius-xl);
  box-shadow: var(--fresh-shadow-sm);
  transition: box-shadow var(--fresh-transition-fast);

  &:focus-within {
    box-shadow: var(--fresh-shadow-md);
  }
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: var(--fresh-text-base);
  color: var(--fresh-text-primary);
  font-family: var(--fresh-font-display);

  &::placeholder {
    color: var(--fresh-text-muted);
  }
}

.search-summary {
  margin: var(--fresh-space-md) 0 0;
  font-size: var(--fresh-text-sm);
  color: var(--fresh-text-secondary);
}

/* ===== 搜索结果 ===== */
.results-list {
  min-height: 300px;
}

.result-card {
  padding: var(--fresh-space-lg);
  margin-bottom: var(--fresh-space-md);
  padding-left: calc(var(--fresh-space-lg) + 3px);
  cursor: pointer;
  transition: transform var(--fresh-transition-fast), box-shadow var(--fresh-transition-fast);

  &:hover {
    transform: translateY(-2px);
    box-shadow: var(--fresh-shadow-md);
  }
}

.result-inner {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-md);
}

.result-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: var(--fresh-radius-sm);
  background: var(--fresh-mint-light);
  display: flex;
  align-items: center;
  justify-content: center;
}

.result-body {
  flex: 1;
  min-width: 0;
}

.result-title {
  font-size: var(--fresh-text-lg);
  font-weight: 600;
  color: var(--fresh-text-primary);
  margin: 0 0 4px;
  line-height: 1.4;
}

.result-hint {
  font-size: var(--fresh-text-xs);
  color: var(--fresh-text-muted);
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
  color: var(--fresh-text-muted);
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
@media (max-width: 768px) {
  .search-bar-wrapper {
    flex-wrap: wrap;
  }
}
</style>
