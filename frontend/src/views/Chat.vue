<template>
  <div class="chat-page">
    <div class="main-container">
      <!-- Not Logged In State -->
      <div v-if="!isLoggedIn" class="login-prompt">
        <el-card class="prompt-card" shadow="hover">
          <div class="prompt-content">
            <el-icon class="prompt-icon" size="48"><ChatDotRound /></el-icon>
            <h2>登录后加入聊天室</h2>
            <p>登录你的账号，与其他用户实时交流</p>
            <div class="prompt-actions">
              <el-button type="primary" size="large" @click="$router.push('/login')">
                立即登录
              </el-button>
              <el-button size="large" @click="$router.push('/register')">
                注册账号
              </el-button>
            </div>
          </div>
        </el-card>
      </div>

      <!-- Chat Room -->
      <div v-else class="chat-container">
        <!-- Connection Status Bar -->
        <div class="status-bar" :class="statusClass">
          <div class="status-left">
            <span class="status-dot"></span>
            <span class="status-text">{{ statusText }}</span>
          </div>
          <el-button
            v-if="!wsStore.isConnected && !wsStore.isConnecting"
            size="small"
            type="primary"
            @click="reconnect"
          >
            重新连接
          </el-button>
        </div>

        <!-- Messages Area -->
        <div ref="messagesContainer" class="messages-area" @scroll="handleScroll">
          <!-- Load More -->
          <div v-if="hasMore" class="load-more">
            <el-button text size="small" @click="scrollToTop">
              查看更早的消息
            </el-button>
          </div>

          <div
            v-for="(msg, index) in messages"
            :key="index"
            class="message-item"
            :class="{ 'message-mine': msg.username === currentUser }"
          >
            <!-- System Message -->
            <template v-if="msg.type === 'system'">
              <div class="system-message">
                <span>{{ msg.content }}</span>
              </div>
            </template>

            <!-- Normal Message -->
            <template v-else>
              <el-avatar
                :size="36"
                class="msg-avatar"
                :style="{ backgroundColor: getAvatarColor(msg.username) }"
              >
                {{ msg.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              <div class="msg-body">
                <div class="msg-header">
                  <span class="msg-username">{{ msg.username }}</span>
                  <span class="msg-time">{{ formatTime(msg.timestamp) }}</span>
                </div>
                <div class="msg-content">
                  {{ msg.content }}
                </div>
              </div>
            </template>
          </div>

          <!-- Empty State -->
          <div v-if="messages.length === 0 && !wsStore.isConnecting" class="empty-chat">
            <el-icon size="48" color="#ddd"><ChatLineSquare /></el-icon>
            <p>暂无消息，发送第一条消息开始聊天吧！</p>
          </div>
        </div>

        <!-- Input Area -->
        <div class="input-area">
          <div class="input-wrapper">
            <el-input
              v-model="inputMessage"
              type="textarea"
              :rows="1"
              :autosize="{ minRows: 1, maxRows: 4 }"
              placeholder="输入消息，按 Enter 发送..."
              resize="none"
              class="message-input"
              @keydown.enter.exact.prevent="sendMessage"
            />
            <el-button
              type="primary"
              :disabled="!inputMessage.trim() || !wsStore.isConnected"
              :loading="wsStore.isConnecting"
              class="send-btn"
              @click="sendMessage"
            >
              <el-icon><Promotion /></el-icon>
              发送
            </el-button>
          </div>
          <div class="input-hint">
            <span v-if="wsStore.isConnected" class="hint-connected">
              <el-icon><CircleCheck /></el-icon>
              已连接
            </span>
            <span v-else class="hint-disconnected">
              <el-icon><WarningFilled /></el-icon>
              未连接，消息将无法发送
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/websocket'
import {
  ChatDotRound,
  ChatLineSquare,
  Promotion,
  CircleCheck,
  WarningFilled,
} from '@element-plus/icons-vue'

interface ChatMessage {
  type: 'system' | 'message'
  username?: string
  content: string
  timestamp?: string
}

const authStore = useAuthStore()
const wsStore = useChatStore()

const messagesContainer = ref<HTMLElement | null>(null)
const inputMessage = ref('')
const messages = ref<ChatMessage[]>([])
const hasMore = ref(false)

const isLoggedIn = computed(() => authStore.isLoggedIn)
const currentUser = computed(() => authStore.user?.username || '')

const statusClass = computed(() => {
  if (wsStore.isConnected) return 'status-connected'
  if (wsStore.isConnecting) return 'status-connecting'
  return 'status-disconnected'
})

const statusText = computed(() => {
  if (wsStore.isConnected) return '已连接'
  if (wsStore.isConnecting) return '连接中...'
  return '未连接'
})

/** 生成头像颜色 */
const getAvatarColor = (username: string = ''): string => {
  const colors = [
    '#667eea', '#764ba2', '#f093fb', '#f5576c',
    '#4facfe', '#00f2fe', '#43e97b', '#fa709a',
    '#a18cd1', '#fbc2eb', '#fccb90', '#d57eeb',
  ]
  let hash = 0
  for (let i = 0; i < username.length; i++) {
    hash = username.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

/** 格式化时间 */
const formatTime = (timestamp?: string): string => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${hours}:${minutes}`
}

/** 滚动到底部 */
const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const scrollToTop = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = 0
  }
}

const handleScroll = () => {
  // 可扩展：实现加载更多历史消息
}

/** 发送消息 */
const sendMessage = () => {
  const content = inputMessage.value.trim()
  if (!content || !wsStore.isConnected) return

  // 发送格式严格为 {"content": "消息内容"}，不含 username（后端自动从 JWT 解析）
  wsStore.sendMessage(content)
  inputMessage.value = ''
}

/** 重新连接 */
const reconnect = () => {
  wsStore.disconnect()
  wsStore.connect()
}

/** 监听 WebSocket store 中的消息变化 */
watch(
  () => wsStore.messages.length,
  (newLen, oldLen) => {
    if (newLen > oldLen) {
      const latestMsg = wsStore.messages[newLen - 1]
      messages.value.push({
        type: 'message',
        username: latestMsg.username,
        content: latestMsg.content,
        timestamp: new Date().toISOString(),
      })
      scrollToBottom()
    }
  }
)

onMounted(() => {
  if (isLoggedIn.value) {
    messages.value.push({
      type: 'system',
      content: `欢迎 ${currentUser.value} 进入聊天室`,
      timestamp: new Date().toISOString(),
    })
    wsStore.connect()
  }
})

onBeforeUnmount(() => {
  wsStore.disconnect()
})
</script>

<style scoped lang="scss">
.main-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 32px 24px 48px;
}

.login-prompt {
  margin-top: 24px;

  .prompt-card {
    border-radius: 16px;
    text-align: center;
    padding: 20px;
  }

  .prompt-content {
    padding: 32px 0;
  }

  .prompt-icon {
    color: #43e97b;
    margin-bottom: 16px;
  }

  h2 {
    font-size: 22px;
    color: #1a1a1a;
    margin: 0 0 8px;
  }

  p {
    font-size: 15px;
    color: #666;
    margin: 0 0 24px;
  }

  .prompt-actions {
    display: flex;
    justify-content: center;
    gap: 12px;
  }
}

.chat-container {
  margin-top: 24px;
  background: #fff;
  border-radius: 16px;
  border: 1px solid #f0f0f0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 140px);
  min-height: 480px;
}

.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 20px;
  font-size: 13px;
  border-bottom: 1px solid #f0f0f0;

  .status-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
  }

  &.status-connected {
    background: #f0fff4;

    .status-dot {
      background: #52c41a;
      box-shadow: 0 0 6px rgba(82, 196, 26, 0.5);
    }

    .status-text {
      color: #52c41a;
    }
  }

  &.status-connecting {
    background: #fffbe6;

    .status-dot {
      background: #faad14;
      animation: pulse 1.5s infinite;
    }

    .status-text {
      color: #faad14;
    }
  }

  &.status-disconnected {
    background: #fff2f0;

    .status-dot {
      background: #ff4d4f;
    }

    .status-text {
      color: #ff4d4f;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #fafbfc;
  scroll-behavior: smooth;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #ddd;
    border-radius: 3px;

    &:hover {
      background: #ccc;
    }
  }
}

.load-more {
  text-align: center;
  padding: 8px 0;
  margin-bottom: 12px;
}

.message-item {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  animation: messageIn 0.3s ease;

  &.message-mine {
    flex-direction: row-reverse;

    .msg-body {
      align-items: flex-end;
    }

    .msg-header {
      flex-direction: row-reverse;
    }

    .msg-content {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
      border-radius: 16px 4px 16px 16px;
    }
  }

  .msg-avatar {
    flex-shrink: 0;
    color: #fff;
    font-weight: 600;
    font-size: 14px;
  }

  .msg-body {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-width: 70%;
  }

  .msg-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .msg-username {
    font-size: 13px;
    font-weight: 600;
    color: #333;
  }

  .msg-time {
    font-size: 11px;
    color: #bbb;
  }

  .msg-content {
    background: #fff;
    padding: 10px 16px;
    border-radius: 4px 16px 16px 16px;
    font-size: 14px;
    line-height: 1.6;
    color: #333;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
    word-break: break-word;
  }
}

.system-message {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;

  span {
    background: #f0f0f0;
    color: #999;
    font-size: 12px;
    padding: 4px 16px;
    border-radius: 12px;
  }
}

.empty-chat {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 12px;

  p {
    font-size: 14px;
    color: #ccc;
    margin: 0;
  }
}

.input-area {
  border-top: 1px solid #f0f0f0;
  padding: 16px 20px;
  background: #fff;

  .input-wrapper {
    display: flex;
    gap: 12px;
    align-items: flex-end;
  }

  .message-input {
    flex: 1;

    :deep(.el-textarea__inner) {
      border-radius: 20px;
      padding: 8px 16px;
      font-size: 14px;
      resize: none;
      box-shadow: none;

      &:focus {
        box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
      }
    }
  }

  .send-btn {
    border-radius: 20px;
    padding: 8px 20px;
    height: auto;
    flex-shrink: 0;
  }

  .input-hint {
    margin-top: 8px;
    font-size: 12px;
    display: flex;
    align-items: center;
    gap: 4px;

    .hint-connected {
      color: #52c41a;
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .hint-disconnected {
      color: #ff4d4f;
      display: flex;
      align-items: center;
      gap: 4px;
    }
  }
}

@keyframes messageIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .chat-container {
    height: calc(100vh - 100px);
    min-height: 400px;
    border-radius: 12px;
  }

  .message-item .msg-body {
    max-width: 80%;
  }
}
</style>