/** 文章列表项 —— 严格匹配后端 ArticleListResponse 结构 */
export interface ArticleListItem {
    id: number
    title: string
    preview: string
    likes: number
    views: number
    user_id: number
    username: string      // 作者用户名
    category_name: string
    tags: string[]       // 后端列表接口返回字符串数组，如 ["cs", "go"]
    cover_img: string
    created_at: string
}

/** 文章详情 —— 匹配后端 models.Article 结构 */
export interface ArticleDetail {
    id: number
    title: string
    content: string
    preview: string
    likes: number
    views: number
    user_id: number
    cover_img: string
    category_id: number | null
    category: { id: number; name: string; slug?: string } | null
    tags: Array<{ id: number; name: string }>
    comments: Comment[]
    user: Author             // 后端 json tag 是 "user"
    favorited?: boolean      // 后端可能不返回此字段
    favorites_count?: number // 后端可能不返回此字段
    created_at: string
    updated_at: string
}

/** 文章创建请求 —— 一级扁平结构，tag_ids 为纯数字数组 */
export interface ArticleCreateRequest {
    title: string
    content: string
    preview: string
    category_id: number | null
    tag_ids: number[]      // 🔥 锁死为纯数字数组
    cover_img: string
}

/** 文章更新请求 */
export interface ArticleUpdateRequest {
    title: string
    content: string
    preview: string
    category_id: number
    tag_ids: number[]      // 🔥 纯数字数组
    cover_img: string
}

/** 后端文章列表响应结构 */
export interface ArticlesResponse {
    data: ArticleListItem[]
    page: number
    page_size: number
    total: number
    total_pages: number
}

/** 后端文章详情响应结构 */
export interface ArticleDetailResponse {
    article: ArticleDetail
    comments: Comment[]
    likes: string    // 后端以字符串返回点赞数
}

/** 后端创建/更新文章响应结构 */
export interface ArticleMutationResponse {
    id: number
    title: string
    content: string
    preview: string
    category_id: number | null
    cover_img: string
    user_id: number
    created_at: string
    updated_at: string
}

export interface Author {
    id: number
    username: string
    bio: string
    image: string
    following: boolean
    created_at: string
    updated_at: string
}

export interface Comment {
    id: number
    article_id: number
    user_id: number
    content: string
    author?: Author
    created_at: string
}

/** 后端评论列表响应结构 */
export interface CommentsResponse {
    comments: Comment[]
}

/** 创建评论请求 */
export interface CommentCreateRequest {
    content: string
}

/** 创建评论响应（返回创建的评论对象） */
export type CommentResponse = Comment

/** 文章详情响应（兼容别名） */
export type ArticleResponse = ArticleDetailResponse

/** 文章列表响应（兼容别名） */
export type ArticleListResponse = ArticlesResponse
