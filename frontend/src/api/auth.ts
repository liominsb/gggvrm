import http from './http'
import type {
    LoginRequest,
    RegisterRequest,
    RefreshTokenRequest,
    TokenResponse,
    UserResponse,
} from '@/types/auth'

export const authApi = {
    /** 登录 */
    async login(data: LoginRequest): Promise<TokenResponse> {
        const res = await http.post<TokenResponse>('/api/auth/login', data)
        return res.data
    },

    /** 注册 */
    async register(data: RegisterRequest): Promise<TokenResponse> {
        const res = await http.post<TokenResponse>('/api/auth/register', data)
        return res.data
    },

    /** 刷新 Token */
    async refreshTokens(data: RefreshTokenRequest): Promise<TokenResponse> {
        const res = await http.post<TokenResponse>('/api/auth/refreshTokens', data)
        return res.data
    },

    /** 获取当前用户 */
    async getCurrentUser(): Promise<UserResponse> {
        const res = await http.get<UserResponse>('/api/v1/user')
        return res.data
    },

    /** 更新个人资料（用户名） */
    async updateProfile(data: { username: string; image?: string; bio?: string }): Promise<UserResponse> {
        const res = await http.put<UserResponse>('/api/v1/user/profile', data)
        return res.data
    },

    /** 修改密码 */
    async changePassword(data: { old_password: string; new_password: string }): Promise<any> {
        const res = await http.put('/api/v1/user/password', data)
        return res.data
    },
}
