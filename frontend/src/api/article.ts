import http from './http'
import type {
    ArticleCreateRequest,
    ArticleUpdateRequest,
    ArticleListResponse,
    ArticleResponse,
    Comment,
    CommentCreateRequest,
} from '@/types/article'

/** 后端 toggle 收藏返回结构 */
interface ToggleFavoriteResponse {
    message: string
    is_favorited: boolean
}

/** 后端收藏数返回结构 */
interface FavoriteCountResponse {
    favorites: number
}

/** 后端收藏状态返回结构 */
interface FavoriteStatusResponse {
    is_favorited: boolean
}

export const articleApi = {
    /**
     * 获取文章列表
     * 后端路由: GET /api/v1/articles?page=&page_size=&category_id=&tag_id=&keyword=
     */
    async getArticles(params?: {
        tag_id?: number
        category_id?: number
        keyword?: string
        page?: number
        page_size?: number
        user_id?: number
    }): Promise<ArticleListResponse> {
        const res = await http.get<ArticleListResponse>('/api/v1/articles', { params })
        return res.data
    },

    /** 获取关注 Feed 流 */
    async getFeedArticles(params?: {
        limit?: number
        offset?: number
    }): Promise<ArticleListResponse> {
        const res = await http.get<ArticleListResponse>('/api/v1/articles/feed', { params })
        return res.data
    },

    /**
     * 获取文章详情（后端用数字 ID）
     * 后端路由: GET /api/v1/article/:id
     * 后端返回: { article, comments, likes }
     */
    async getArticle(id: string): Promise<ArticleResponse> {
        const res = await http.get<ArticleResponse>(`/api/v1/article/${id}`)
        return res.data
    },

    /**
     * 创建文章 - 一级扁平结构
     * 后端路由: POST /api/v1/articles
     */
    async createArticle(data: ArticleCreateRequest): Promise<any> {
        const res = await http.post('/api/v1/articles', data)
        return res.data
    },

    /**
     * 更新文章 - 一级扁平结构，不包裹在 { article: ... } 中
     * 后端路由: PUT /api/v1/article/:id
     * 后端返回: { message, article }
     */
    async updateArticle(id: string, data: ArticleUpdateRequest): Promise<any> {
        const res = await http.put(`/api/v1/article/${id}`, data)
        return res.data
    },

    /**
     * 删除文章
     * 后端路由: DELETE /api/v1/article/:id
     */
    async deleteArticle(id: string): Promise<void> {
        await http.delete(`/api/v1/article/${id}`)
    },

    /**
     * 获取文章评论
     * 后端路由: GET /api/v1/article/:id/comments
     * 后端直接返回 Comment[] 数组
     */
    async getComments(articleId: string): Promise<Comment[]> {
        const res = await http.get<any>(`/api/v1/article/${articleId}/comments`)
        const raw = res.data
        return Array.isArray(raw) ? raw : (raw?.comments || [])
    },

    /**
     * 创建评论
     * 后端路由: POST /api/v1/article/:id/comment
     * 请求体: { content: "..." }
     * 后端返回创建的评论对象
     */
    async createComment(articleId: string, data: CommentCreateRequest): Promise<Comment> {
        const res = await http.post<Comment>(`/api/v1/article/${articleId}/comment`, data)
        return res.data
    },

    /**
     * 删除评论
     * 后端路由: DELETE /api/v1/comment/:id
     */
    async deleteComment(commentId: number): Promise<void> {
        await http.delete(`/api/v1/comment/${commentId}`)
    },

    /**
     * 收藏/取消收藏文章（toggle 接口，POST 一次即可切换状态）
     * 后端路由: POST /api/v1/article/:id/favorite
     * 后端返回: { message, is_favorited }
     */
    async toggleFavorite(articleId: string): Promise<ToggleFavoriteResponse> {
        const res = await http.post<ToggleFavoriteResponse>(`/api/v1/article/${articleId}/favorite`)
        return res.data
    },

    /**
     * 获取文章收藏状态
     * 后端路由: GET /api/v1/article/:id/favorite
     */
    async getFavoriteStatus(articleId: string): Promise<FavoriteStatusResponse> {
        const res = await http.get<FavoriteStatusResponse>(`/api/v1/article/${articleId}/favorite`)
        return res.data
    },

    /**
     * 获取文章收藏数
     * 后端路由: GET /api/v1/article/:id/favorites/count
     */
    async getFavoriteCount(articleId: string): Promise<FavoriteCountResponse> {
        const res = await http.get<FavoriteCountResponse>(`/api/v1/article/${articleId}/favorites/count`)
        return res.data
    },

    /**
     * 获取当前用户的收藏列表
     * 后端路由: GET /api/v1/user/favorites?page=&page_size=
     */
    async getUserFavorites(params?: { page?: number; page_size?: number }): Promise<ArticleListResponse> {
        const res = await http.get<ArticleListResponse>('/api/v1/user/favorites', { params })
        return res.data
    },

    /**
     * 根据用户ID获取收藏列表
     * 后端路由: GET /api/v1/user/:id/favorites?page=&page_size=
     */
    async getUserFavoritesById(userId: number, params?: { page?: number; page_size?: number }): Promise<ArticleListResponse> {
        const res = await http.get<ArticleListResponse>(`/api/v1/user/${userId}/favorites`, { params })
        return res.data
    },

    /**
     * 点赞/取消点赞文章（toggle 接口）
     * 后端路由: POST /api/v1/article/:id/like
     */
    async likeArticle(articleId: string): Promise<any> {
        const res = await http.post(`/api/v1/article/${articleId}/like`)
        return res.data
    },

    /**
     * 获取文章点赞数
     * 后端路由: GET /api/v1/article/:id/like
     */
    async getLikeCount(articleId: string): Promise<any> {
        const res = await http.get(`/api/v1/article/${articleId}/like`)
        return res.data
    },
}
