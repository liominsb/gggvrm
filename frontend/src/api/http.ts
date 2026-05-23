import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import router from '@/router'

/** 后端基地址（开发环境通过 vite proxy 代理，生产环境直接配置） */
const BASE_URL = import.meta.env.DEV ? '' : 'http://localhost:8080'

const http: AxiosInstance = axios.create({
    baseURL: BASE_URL,
    timeout: 15000,
    headers: {
        'Content-Type': 'application/json',
    },
})

/** 是否正在刷新 Token */
let isRefreshing = false

/** 等待刷新完成的请求队列 */
let pendingRequests: Array<{
    resolve: (token: string) => void
    reject: (error: any) => void
}> = []

/** 处理等待队列 */
function processPendingRequests(token: string, error?: any) {
    pendingRequests.forEach((promise) => {
        if (error) {
            promise.reject(error)
        } else {
            promise.resolve(token)
        }
    })
    pendingRequests = []
}

// ============ 请求拦截器 ============
http.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        // 对于 /api/v1/* 路径自动附加 Authorization
        if (config.url?.startsWith('/api/v1/') && !config.headers.Authorization) {
            const token = sessionStorage.getItem('access_token')
            if (token) {
                config.headers.Authorization = `Bearer ${token}`
            }
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// ============ 响应拦截器 ============
http.interceptors.response.use(
    (response: AxiosResponse) => {
        return response
    },
    async (error: AxiosError<any>) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

        // ----- 409 Conflict: 单设备登录 - 被踢下线 -----
        if (error.response?.status === 409 && error.response?.data?.error === 'kicked_out') {
            // 清除本地状态
            const authStore = useAuthStore()
            authStore.clearAuth()

            // 弹窗强制提示
            ElMessage.error(error.response.data.msg || '您已在其他设备登录，当前设备被迫下线')

            // 路由跳转到登录页
            router.push('/login')
            return Promise.reject(error)
        }

        // ----- 401 Unauthorized: 尝试无感刷新 Token -----
        if (error.response?.status === 401 && !originalRequest._retry) {
            // 如果请求的是刷新接口本身，不再重试
            if (originalRequest.url?.includes('/api/auth/refreshTokens')) {
                const authStore = useAuthStore()
                authStore.clearAuth()
                router.push('/login')
                return Promise.reject(error)
            }

            // 如果请求的是登录/注册接口，不刷新
            if (originalRequest.url?.includes('/api/auth/login') || originalRequest.url?.includes('/api/auth/register')) {
                return Promise.reject(error)
            }

            originalRequest._retry = true

            if (isRefreshing) {
                // 已经在刷新中，将请求加入队列等待
                return new Promise<string>((resolve, reject) => {
                    pendingRequests.push({ resolve, reject })
                }).then((token) => {
                    originalRequest.headers.Authorization = `Bearer ${token}`
                    return http(originalRequest)
                }).catch((err) => {
                    return Promise.reject(err)
                })
            }

            isRefreshing = true

            const refreshTokenValue = sessionStorage.getItem('refresh_token')
            const accessTokenValue = sessionStorage.getItem('access_token')

            if (!refreshTokenValue || !accessTokenValue) {
                // 没有 Token，直接跳登录
                const authStore = useAuthStore()
                authStore.clearAuth()
                router.push('/login')
                isRefreshing = false
                return Promise.reject(error)
            }

            try {
                // 直接用 axios 发请求，避免触发 http 实例的拦截器导致死循环
                const { data } = await axios.post(`${BASE_URL}/api/auth/refreshTokens`, {
                    access_token: accessTokenValue,
                    refreshtoken: refreshTokenValue,
                })

                // 后端返回 { token, refreshToken }，token 带有 "Bearer " 前缀，需要去掉
                const newAccessToken: string = data.token.replace(/^Bearer\s+/i, '')
                const newRefreshToken: string = data.refreshToken.replace(/^Bearer\s+/i, '')

                // 更新本地存储
                sessionStorage.setItem('access_token', newAccessToken)
                sessionStorage.setItem('refresh_token', newRefreshToken)

                // 更新 Pinia 状态
                const authStore = useAuthStore()
                authStore.setTokens(newAccessToken, newRefreshToken)

                // 处理等待队列
                processPendingRequests(newAccessToken)

                // 重发原始请求
                originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
                return http(originalRequest)
            } catch (refreshError) {
                // 刷新失败，清除状态跳登录
                processPendingRequests('', refreshError)
                const authStore = useAuthStore()
                authStore.clearAuth()
                router.push('/login')
                return Promise.reject(refreshError)
            } finally {
                isRefreshing = false
            }
        }

        return Promise.reject(error)
    }
)

export default http