/** 登录请求 */
export interface LoginRequest {
    username: string
    password: string
}

/** 注册请求 */
export interface RegisterRequest {
    username: string
    password: string
    bio?: string
    image?: string
}

/** 刷新 Token 请求 */
export interface RefreshTokenRequest {
    access_token: string
    refresh_token: string
}

/** 用户信息 */
export interface User {
    id: number
    username: string
    bio: string
    image: string
    created_at: string
    updated_at: string
}

/** 认证响应（登录/注册/刷新 Token 共用） */
export interface AuthResponse {
    user: User
    access_token: string
    refresh_token: string
}

/** Token 响应（登录/注册/刷新 Token 共用） */
export interface TokenResponse {
    token: string
    refreshToken: string
}

/** 纯用户响应 */
export interface UserResponse {
    user: User
}
