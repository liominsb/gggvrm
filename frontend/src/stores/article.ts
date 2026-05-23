import { defineStore } from 'pinia'
import { ref } from 'vue'
import { articleApi } from '@/api/article'
import type { ArticleListItem, ArticleDetail, Comment, ArticleCreateRequest, ArticleUpdateRequest } from '@/types/article'

/**
 * 将后端 gorm.Model 的大写 ID 归一化为小写 id
 * gorm.Model 的 ID 字段没有 json tag，Go encoding/json 会序列化为 "ID"
 */
function normalizeArticle(raw: any): any {
    if (!raw) return raw
    // gorm.Model 序列化为大写 ID/CreatedAt/UpdatedAt，前端期望小写 id/created_at/updated_at
    if (raw.ID !== undefined && raw.id === undefined) {
        raw.id = raw.ID
    }
    if (raw.CreatedAt !== undefined && raw.created_at === undefined) {
        raw.created_at = raw.CreatedAt
    }
    if (raw.UpdatedAt !== undefined && raw.updated_at === undefined) {
        raw.updated_at = raw.UpdatedAt
    }
    // 递归处理嵌套的 user 对象
    if (raw.user) {
        if (raw.user.ID !== undefined && raw.user.id === undefined) {
            raw.user.id = raw.user.ID
        }
        if (raw.user.CreatedAt !== undefined && raw.user.created_at === undefined) {
            raw.user.created_at = raw.user.CreatedAt
        }
        if (raw.user.UpdatedAt !== undefined && raw.user.updated_at === undefined) {
            raw.user.updated_at = raw.user.UpdatedAt
        }
    }
    // 递归处理 comments 数组
    if (Array.isArray(raw.comments)) {
        raw.comments.forEach((c: any) => {
            if (c && c.ID !== undefined && c.id === undefined) c.id = c.ID
            if (c && c.CreatedAt !== undefined && c.created_at === undefined) c.created_at = c.CreatedAt
            if (c && c.UpdatedAt !== undefined && c.updated_at === undefined) c.updated_at = c.UpdatedAt
        })
    }
    return raw
}

function normalizeArticles(raw: any[]): any[] {
    return (raw || []).map(normalizeArticle)
}

export const useArticleStore = defineStore('article', () => {
    // ======== State ========
    const articles = ref<ArticleListItem[]>([])
    const articlesCount = ref(0)
    const currentArticle = ref<ArticleDetail | null>(null)
    const comments = ref<Comment[]>([])
    const isLoading = ref(false)
    const feedArticles = ref<ArticleListItem[]>([])
    const feedCount = ref(0)
    const favoritedArticles = ref<ArticleListItem[]>([])
    const favoritedCount = ref(0)

    // ======== Actions ========

    /** 获取文章列表 — 对齐后端参数: page, page_size, category_id, tag_id, keyword */
    async function fetchArticles(params?: {
        tag_id?: number
        category_id?: number
        keyword?: string
        page?: number
        page_size?: number
        user_id?: number
    }) {
        isLoading.value = true
        try {
            const data = await articleApi.getArticles(params)
            articles.value = normalizeArticles(data.data || []) as ArticleListItem[]
            articlesCount.value = data.total || 0
            return data
        } finally {
            isLoading.value = false
        }
    }

    /** 获取关注 Feed */
    async function fetchFeed(params?: { limit?: number; offset?: number }) {
        isLoading.value = true
        try {
            const data = await articleApi.getFeedArticles(params)
            feedArticles.value = normalizeArticles(data.data || []) as ArticleListItem[]
            feedCount.value = data.total || 0
            return data
        } finally {
            isLoading.value = false
        }
    }

    /** 获取文章详情（后端用数字 ID，gorm.Model 序列化为大写 ID） */
    async function fetchArticle(id: string) {
        isLoading.value = true
        try {
            const data = await articleApi.getArticle(id)
            // 后端返回 { article, comments, likes }
            const normalized = normalizeArticle(data.article)
            currentArticle.value = normalized as ArticleDetail
            comments.value = (data.comments || []).map((c: any) => normalizeArticle(c)) as Comment[]
            return currentArticle.value
        } finally {
            isLoading.value = false
        }
    }

    /** 创建文章 — payload 必须一级扁平，tag_ids 纯数字数组 */
    async function createArticle(payload: ArticleCreateRequest) {
        const res = await articleApi.createArticle(payload)
        // Bug 5 修复：将新创建的文章插入列表最前端
        const newArticle = res.article || res
        if (newArticle && (newArticle.ID || newArticle.id)) {
            const normalized = normalizeArticle(newArticle) as ArticleListItem
            articles.value.unshift(normalized)
            articlesCount.value++
        }
        return res
    }

    /** 更新文章 — payload 必须一级扁平，tag_ids 纯数字数组 */
    async function updateArticle(id: string, payload: ArticleUpdateRequest) {
        const res = await articleApi.updateArticle(id, payload)
        if (res.article) {
            currentArticle.value = normalizeArticle(res.article) as unknown as ArticleDetail
        }
        return res
    }

    /** 删除文章（后端用数字 ID） */
    async function deleteArticle(id: string) {
        await articleApi.deleteArticle(id)
        articles.value = articles.value.filter((a) => String(a.id) !== id)
    }

    /** 获取评论 */
    async function fetchComments(articleId: string) {
        const data = await articleApi.getComments(articleId)
        comments.value = data || []
        return data
    }

    /** 创建评论 — 直接传递 { content } */
    async function createComment(articleId: string, content: string) {
        const data = await articleApi.createComment(articleId, { content })
        // 后端返回创建的评论对象，添加到列表
        if (data) {
            const normalized = normalizeArticle(data) as Comment
            comments.value.unshift(normalized)
        }
        return data
    }

    /** 删除评论 */
    async function deleteComment(commentId: number) {
        await articleApi.deleteComment(commentId)
        comments.value = comments.value.filter((c) => c.id !== commentId)
    }

    /** 收藏/取消收藏文章（后端是 toggle 接口，POST 一次即可切换） */
    async function toggleFavorite(articleId: string) {
        const data = await articleApi.toggleFavorite(articleId)
        if (data.is_favorited) {
            // Bug 3 修复：收藏时，将文章添加到收藏列表最前端
            const alreadyExists = favoritedArticles.value.some(
                (a) => String(a.id) === articleId
            )
            if (!alreadyExists) {
                // 优先从 currentArticle 获取（用户在文章详情页收藏），
                // 否则从文章列表中查找
                let source: any = null
                if (currentArticle.value && String(currentArticle.value.id) === articleId) {
                    source = currentArticle.value
                } else {
                    source = articles.value.find((a) => String(a.id) === articleId)
                }
                if (source) {
                    const copy = normalizeArticle({ ...source }) as ArticleListItem
                    favoritedArticles.value.unshift(copy)
                    favoritedCount.value++
                }
            }
        } else {
            // 取消收藏时，从收藏列表中移除
            favoritedArticles.value = favoritedArticles.value.filter(
                (a) => String(a.id) !== articleId
            )
            if (favoritedCount.value > 0) favoritedCount.value--
        }
        return data
    }

    /** 点赞/取消点赞文章（后端是 toggle 接口） */
    async function toggleLike(articleId: string) {
        const data = await articleApi.likeArticle(articleId)
        return data
    }

    /** 获取当前登录用户的收藏列表 */
    async function fetchUserFavorites(params?: { page?: number; page_size?: number }) {
        isLoading.value = true
        try {
            const data = await articleApi.getUserFavorites(params)
            favoritedArticles.value = normalizeArticles(data.data || []) as ArticleListItem[]
            favoritedCount.value = data.total || 0
            return data
        } finally {
            isLoading.value = false
        }
    }

    /** 根据用户ID获取收藏列表 */
    async function fetchUserFavoritesById(userId: number, params?: { page?: number; page_size?: number }) {
        isLoading.value = true
        try {
            const data = await articleApi.getUserFavoritesById(userId, params)
            favoritedArticles.value = normalizeArticles(data.data || []) as ArticleListItem[]
            favoritedCount.value = data.total || 0
            return data
        } finally {
            isLoading.value = false
        }
    }

    return {
        articles,
        articlesCount,
        currentArticle,
        comments,
        isLoading,
        feedArticles,
        feedCount,
        favoritedArticles,
        favoritedCount,
        fetchArticles,
        fetchFeed,
        fetchArticle,
        createArticle,
        updateArticle,
        deleteArticle,
        fetchComments,
        createComment,
        deleteComment,
        toggleFavorite,
        toggleLike,
        fetchUserFavorites,
        fetchUserFavoritesById,
    }
})
