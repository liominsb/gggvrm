/** 通用错误响应 */
export interface ErrorResponse {
    error: string
    msg: string
}

/** 409 被踢下线响应 */
export interface KickedOutResponse {
    error: 'kicked_out'
    msg: '您已在其他设备登录，当前设备被迫下线'
}

/** 图片上传响应 */
export interface UploadResponse {
    urls: string[]
}

/** WebSocket 接收消息 */
export interface WSIncomingMessage {
    username: string
    content: string
}

/** WebSocket 发送消息 */
export interface WSOutgoingMessage {
    content: string
}

/** 关注/取消关注响应 */
export interface ProfileResponse {
    user: {
        id: number
        username: string
        email: string
        bio: string
        image: string
        following: boolean
        created_at: string
        updated_at: string
    }
}

/** 标签对象（后端返回原始数组） */
export interface Tag {
    id: number
    name: string
}

/** 分类 */
export interface Category {
    id: number
    name: string
    slug: string
}

/** 分类列表响应 */
export interface CategoriesResponse {
    categories: Category[]
}