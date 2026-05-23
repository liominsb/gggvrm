import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { WSIncomingMessage, WSOutgoingMessage } from '@/types/common'

/** 后端每 30 秒发 Ping，60 秒无心跳则断开 */
const HEARTBEAT_INTERVAL = 25000 // 前端主动发送 Ping 的间隔（25秒，留 5 秒容错）
const RECONNECT_DELAY = 3000 // 重连延迟
const MAX_RECONNECT_ATTEMPTS = 5 // 最大重连次数

export const useChatStore = defineStore('chat', () => {
    // ======== State ========
    const messages = ref<WSIncomingMessage[]>([])
    const isConnected = ref(false)
    const isConnecting = ref(false)

    let ws: WebSocket | null = null
    let heartbeatTimer: ReturnType<typeof setInterval> | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let manualClose = false
    let reconnectAttempts = 0
    let forceRefreshOnReconnect = false // 标记：重连时需要强制刷新 token

    // ======== Actions ========

    /**
     * 刷新 Token（静默刷新，失败则跳登录）
     * @param force 是否强制刷新（忽略 token 过期时间检查）
     */
    async function ensureFreshToken(force = false): Promise<string | null> {
        const token = sessionStorage.getItem('access_token')
        if (!token) return null

        // 非强制模式下，检查 JWT 是否即将过期
        if (!force) {
            try {
                const payload = JSON.parse(atob(token.split('.')[1]))
                const now = Math.floor(Date.now() / 1000)
                // 如果 token 还有 2 分钟以上有效期，直接使用
                if (payload.exp && payload.exp > now + 120) {
                    return token
                }
            } catch {
                // 解析失败，尝试使用现有 token
                return token
            }
        }

        // Token 即将过期/已过期，或强制刷新模式，尝试刷新
        const refreshToken = sessionStorage.getItem('refresh_token')
        if (!refreshToken) return token

        try {
            const BASE_URL = import.meta.env.DEV ? '' : 'http://localhost:8080'
            const resp = await fetch(`${BASE_URL}/api/auth/refreshTokens`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    access_token: token,
                    refreshtoken: refreshToken,
                }),
            })
            if (resp.ok) {
                const data = await resp.json()
                const newAccess = data.token.replace(/^Bearer\s+/i, '')
                const newRefresh = data.refreshToken.replace(/^Bearer\s+/i, '')
                sessionStorage.setItem('access_token', newAccess)
                sessionStorage.setItem('refresh_token', newRefresh)
                // 同步更新 Pinia auth store，避免其他模块使用旧 token
                try {
                    const { useAuthStore } = await import('@/stores/auth')
                    const authStore = useAuthStore()
                    authStore.setTokens(newAccess, newRefresh)
                } catch {
                    // auth store 不可用时忽略（不影响 WS 功能）
                }
                return newAccess
            }
        } catch {
            // 刷新失败，返回原有 token（让后端判断）
        }
        return token
    }

    /** 建立 WebSocket 连接 */
    async function connect() {
        if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
            return
        }

        manualClose = false
        isConnecting.value = true

        // 如果上次连接因认证失败关闭，强制刷新 token
        const shouldForceRefresh = forceRefreshOnReconnect
        forceRefreshOnReconnect = false

        // 连接前确保 token 新鲜
        const token = await ensureFreshToken(shouldForceRefresh)

        // 通过 Vite 代理连接（开发环境走 localhost:5173 的 /api 代理，生产环境走同源）
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const wsBase = `${protocol}//${window.location.host}`
        const wsUrl = token
            ? `${wsBase}/api/v1/ws?token=${encodeURIComponent(token)}`
            : `${wsBase}/api/v1/ws`

        ws = new WebSocket(wsUrl)

        ws.onopen = () => {
            isConnected.value = true
            isConnecting.value = false
            reconnectAttempts = 0 // 连接成功，重置重连计数
            forceRefreshOnReconnect = false // 连接成功，清除强制刷新标记
            startHeartbeat()
            console.log('[WebSocket] 连接成功')
        }

        ws.onmessage = (event: MessageEvent) => {
            try {
                const data: WSIncomingMessage = JSON.parse(event.data)
                messages.value.push(data)
            } catch {
                console.warn('[WebSocket] 收到非 JSON 消息:', event.data)
            }
        }

        ws.onerror = (event) => {
            console.error('[WebSocket] 错误:', event)
        }

        ws.onclose = (event) => {
            isConnected.value = false
            isConnecting.value = false
            stopHeartbeat()
            console.log(`[WebSocket] 连接关闭: code=${event.code}, reason=${event.reason}`)

            // 非手动关闭时自动重连
            if (!manualClose) {
                // 认证失败导致的断开（自定义关闭码 4000-4999），直接停止重连
                if (event.code >= 4000 && event.code < 5000) {
                    console.warn(`[WebSocket] 认证失败 (code=${event.code})，停止重连`)
                    return
                }
                // code 1006 = 异常关闭（可能是 HTTP 401 导致握手失败）
                // 第一次异常关闭时，标记下次重连强制刷新 token
                if (event.code === 1006 && reconnectAttempts === 0) {
                    console.log('[WebSocket] 疑似认证失败，下次重连将强制刷新 token')
                    forceRefreshOnReconnect = true
                }
                scheduleReconnect()
            }
        }
    }

    /**
     * 前端主动发送 Ping 保持心跳
     * 后端每 30 秒发 PingMessage，浏览器 WebSocket 底层会自动回 Pong。
     * 但我们额外每隔 25 秒从前端也发一个 Ping（文本帧），双重保障。
     */
    function startHeartbeat() {
        stopHeartbeat()
        heartbeatTimer = setInterval(() => {
            if (ws && ws.readyState === WebSocket.OPEN) {
                // 发送一个空文本帧作为心跳（后端会忽略非 JSON 格式的文本消息）
                // 或者也可以发 JSON Ping：ws.send(JSON.stringify({ type: 'ping' }))
                ws.send('')
            }
        }, HEARTBEAT_INTERVAL)
    }

    function stopHeartbeat() {
        if (heartbeatTimer) {
            clearInterval(heartbeatTimer)
            heartbeatTimer = null
        }
    }

    /** 定时重连 */
    function scheduleReconnect() {
        if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
            console.warn(`[WebSocket] 已达最大重连次数(${MAX_RECONNECT_ATTEMPTS})，停止重连`)
            return
        }
        reconnectAttempts++
        if (reconnectTimer) clearTimeout(reconnectTimer)
        reconnectTimer = setTimeout(() => {
            console.log(`[WebSocket] 尝试重连 (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})...`)
            connect()
        }, RECONNECT_DELAY)
    }

    /** 发送消息 */
    function sendMessage(content: string) {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            console.warn('[WebSocket] 未连接，无法发送消息')
            return false
        }

        const msg: WSOutgoingMessage = { content }
        ws.send(JSON.stringify(msg))
        return true
    }

    /** 断开连接 */
    function disconnect() {
        manualClose = true
        stopHeartbeat()
        if (reconnectTimer) {
            clearTimeout(reconnectTimer)
            reconnectTimer = null
        }
        if (ws) {
            ws.close()
            ws = null
        }
        isConnected.value = false
        isConnecting.value = false
    }

    /** 清空消息记录 */
    function clearMessages() {
        messages.value = []
    }

    return {
        messages,
        isConnected,
        isConnecting,
        connect,
        disconnect,
        sendMessage,
        clearMessages,
    }
})