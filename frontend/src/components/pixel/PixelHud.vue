<template>
  <!-- Top HUD Bar -->
  <div class="pixel-hud pixel-hud--top" aria-label="System status bar">
    <div class="pixel-hud__inner">
      <!-- Left: System Status -->
      <div class="pixel-hud__section">
        <span class="pixel-hud__label">SYS</span>
        <span class="pixel-hud__dot pixel-hud__dot--online"></span>
        <span class="pixel-hud__value">ONLINE</span>
        <span class="pixel-hud__sep">|</span>
        <span class="pixel-hud__label">LV</span>
        <span class="pixel-hud__value">{{ userLevel }}</span>
      </div>

      <!-- Center: Clock / Title -->
      <div class="pixel-hud__section pixel-hud__center">
        <span class="pixel-hud__title pixel-text-hud">
          {{ pageTitle }}
        </span>
      </div>

      <!-- Right: Stats -->
      <div class="pixel-hud__section">
        <span v-if="notifications > 0" class="pixel-hud__alert">
          <span class="pixel-hud__label">MSG</span>
          <span class="pixel-hud__value pixel-hud__value--alert">{{ notifications }}</span>
        </span>
        <span class="pixel-hud__sep">|</span>
        <span class="pixel-hud__label">TIME</span>
        <span class="pixel-hud__value">{{ currentTime }}</span>
      </div>
    </div>
  </div>

  <!-- Bottom HUD Bar (contextual) -->
  <div v-if="showBottom" class="pixel-hud pixel-hud--bottom" aria-label="Context bar">
    <div class="pixel-hud__inner">
      <div class="pixel-hud__section">
        <span class="pixel-hud__hint">[F1] HELP</span>
        <span class="pixel-hud__sep">|</span>
        <span class="pixel-hud__hint">[ESC] BACK</span>
      </div>
      <div class="pixel-hud__section">
        <span class="pixel-hud__breadcrumb">
          <slot name="breadcrumb">
            <span class="pixel-hud__value">/home</span>
          </slot>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

withDefaults(defineProps<{
  showBottom?: boolean
  notifications?: number
}>(), {
  showBottom: true,
  notifications: 0,
})

const route = useRoute()
const authStore = useAuthStore()
const now = ref(new Date())
let timer: ReturnType<typeof setInterval> | null = null

const userLevel = computed(() => {
  return authStore.isLoggedIn ? '05' : '00'
})

const pageTitle = computed(() => {
  const name = String(route.name || 'HOME')
  return `[ ${name.toUpperCase()} ]`
})

const currentTime = computed(() => {
  return now.value.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
})

onMounted(() => {
  timer = setInterval(() => {
    now.value = new Date()
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped lang="scss">
.pixel-hud {
  position: fixed;
  left: 0;
  right: 0;
  z-index: var(--z-hud);
  background: rgba(12, 12, 29, 0.92);
  border-bottom: 1px solid var(--pixel-border-default);
  backdrop-filter: blur(4px);
  pointer-events: none;

  &--top {
    top: 0;
    height: 28px;
  }

  &--bottom {
    bottom: 0;
    height: 28px;
    border-bottom: none;
    border-top: 1px solid var(--pixel-border-default);
  }
}

.pixel-hud__inner {
  max-width: var(--pixel-max-width);
  margin: 0 auto;
  padding: 0 var(--pixel-space-md);
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.pixel-hud__section {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.pixel-hud__center {
  flex: 1;
  justify-content: center;
}

.pixel-hud__label {
  font-family: var(--pixel-font-display);
  font-size: 7px;
  color: var(--pixel-text-muted);
  letter-spacing: 0.1em;
}

.pixel-hud__value {
  font-family: var(--pixel-font-display);
  font-size: 7px;
  color: var(--pixel-accent-mint);

  &--alert {
    color: var(--pixel-accent-pink);
  }
}

.pixel-hud__dot {
  width: 6px;
  height: 6px;
  border: 1px solid currentColor;

  &--online {
    color: var(--pixel-success);
    background: var(--pixel-success);
  }
}

.pixel-hud__sep {
  font-family: var(--pixel-font-display);
  font-size: 7px;
  color: var(--pixel-text-muted);
  margin: 0 2px;
}

.pixel-hud__title {
  font-size: 7px;
  color: var(--pixel-accent-gold);
  letter-spacing: 0.15em;
}

.pixel-hud__alert {
  display: flex;
  align-items: center;
  gap: 3px;
}

.pixel-hud__hint {
  font-family: var(--pixel-font-display);
  font-size: 7px;
  color: var(--pixel-text-muted);
  letter-spacing: 0.1em;
}

.pixel-hud__breadcrumb {
  font-family: var(--pixel-font-display);
  font-size: 7px;
  color: var(--pixel-text-muted);
}

@media (max-width: 768px) {
  .pixel-hud__center {
    display: none;
  }

  .pixel-hud__hint {
    display: none;
  }
}
</style>
