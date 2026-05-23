import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/auth'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
    // ======== State ========
    const user = ref<User | null>(null)
    const accessToken = ref<string>(sessionStorage.getItem('access_token') || '')
    const refreshToken = ref<string>(sessionStorage.getItem('refresh_token') || '')

    // ======== Getters ========
    const isLoggedIn = computed(() => !!accessToken.value && !!user.value)
    const username = computed(() => user.value?.username || '')
    const userImage = computed(() => user.value?.image || '')

    // ======== Actions ========

    /** 设置用户信息 */
    function setUser(u: User) {
        user.value = u
    }

    /** 设置 Token */
    function setTokens(access: string, refresh: string) {
        accessToken.value = access
        refreshToken.value = refresh
        sessionStorage.setItem('access_token', access)
        sessionStorage.setItem('refresh_token', refresh)
    }

    /** 登录 */
    async function login(usernameOrEmail: string, password: string) {
        const data = await authApi.login({
            username: usernameOrEmail,
            password,
        })
        // 后端登录返回 "Bearer <jwt>" 格式的 token，去掉 "Bearer " 前缀只存储纯 JWT
        const pureToken = data.token.replace(/^Bearer\s+/i, '')
        const pureRefresh = data.refreshToken.replace(/^Bearer\s+/i, '')
        setTokens(pureToken, pureRefresh)
        // 登录后拉取当前用户信息，但拉取失败不清理登录态
        // （刚登录成功，Token 一定有效，fetchCurrentUser 失败大概率是网络问题或并发时序问题）
        try {
            await fetchCurrentUser()
        } catch {
            // 忽略：登录成功但拉取用户失败，Token 已经存好了，后续页面会重试
        }
        return data
    }

    /** 注册 */
    async function register(username: string, password: string) {
        const data = await authApi.register({
            username,
            password,
        })
        // 后端注册返回 "Bearer <jwt>" 格式的 token，去掉 "Bearer " 前缀只存储纯 JWT
        const pureToken = data.token.replace(/^Bearer\s+/i, '')
        const pureRefresh = data.refreshToken.replace(/^Bearer\s+/i, '')
        setTokens(pureToken, pureRefresh)
        // 注册后拉取当前用户信息，但拉取失败不清理登录态
        try {
            await fetchCurrentUser()
        } catch {
            // 忽略：注册成功但拉取用户失败，Token 已经存好了
        }
        return data
    }

    /** 获取当前用户信息（用于初始化时恢复状态） */
    async function fetchCurrentUser() {
        try {
            const data = await authApi.getCurrentUser()
            setUser(data.user)
            return data.user
        } catch (err: any) {
            // 只有在明确的 401 且拦截器已经完成了刷新重试仍失败的情况下才清空登录态
            // 如果是网络错误等其他异常，不应清空登录态，以免误杀刚登录成功的用户
            // 注意：http 拦截器在 refreshTokens 失败时已经会调用 clearAuth() + 跳转登录页
            // 所以这里不需要再重复 clearAuth()
            return null
        }
    }

    /** 更新用户资料 */
    async function updateProfile(data: { username: string; image?: string; bio?: string }) {
        const res = await authApi.updateProfile(data)
        if (res.user) {
            setUser(res.user)
        }
        return res
    }

    /** 修改密码 */
    async function changePassword(data: { old_password: string; new_password: string }) {
        return await authApi.changePassword(data)
    }

    /** 清除认证状态（登出/被踢下线时调用） */
    function clearAuth() {
        user.value = null
        accessToken.value = ''
        refreshToken.value = ''
        sessionStorage.removeItem('access_token')
        sessionStorage.removeItem('refresh_token')
    }

    /** 登出 */
    async function logout() {
        clearAuth()
    }

    return {
        // state
        user,
        accessToken,
        refreshToken,
        // getters
        isLoggedIn,
        username,
        userImage,
        // actions
        setUser,
        setTokens,
        login,
        register,
        fetchCurrentUser,
        updateProfile,
        changePassword,
        clearAuth,
        logout,
    }
})
