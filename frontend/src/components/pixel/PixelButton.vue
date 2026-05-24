<template>
  <button
    :class="[
      'pixel-btn',
      `pixel-btn--${variant}`,
      `pixel-btn--${size}`,
      { 'pixel-btn--block': block, 'pixel-btn--disabled': disabled }
    ]"
    :disabled="disabled"
    :type="nativeType"
    @click="$emit('click', $event)"
  >
    <span v-if="icon" class="pixel-btn__icon">
      <PixelIcon :name="icon" :size="iconSize" :color="iconColor" />
    </span>
    <span v-if="$slots.default" class="pixel-btn__text">
      <slot />
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import PixelIcon from './PixelIcon.vue'

const props = withDefaults(defineProps<{
  variant?: 'mint' | 'pink' | 'gold' | 'purple' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  block?: boolean
  disabled?: boolean
  nativeType?: 'button' | 'submit' | 'reset'
  icon?: 'house' | 'compass' | 'chat' | 'edit' | 'user' | 'settings' | 'star'
    | 'tag' | 'folder' | 'bell' | 'heart' | 'search' | 'arrow-right'
    | 'arrow-down' | 'plus' | 'menu' | 'close' | 'logout' | 'document'
    | 'lock' | 'clock' | 'article' | 'dashboard' | 'pixel-book'
    | 'pixel-sparkle' | 'pixel-shield' | 'pixel-coin'
}>(), {
  variant: 'mint',
  size: 'md',
  block: false,
  disabled: false,
  nativeType: 'button',
})

defineEmits<{ click: [event: MouseEvent] }>()

const iconSize = computed(() => props.size === 'sm' ? 12 : props.size === 'lg' ? 20 : 16)
const iconColor = computed(() => {
  if (props.variant === 'ghost') return 'mint'
  return 'white'
})
</script>

<style scoped lang="scss">
.pixel-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--pixel-space-sm);
  font-family: var(--pixel-font-display);
  font-size: var(--pixel-text-xs);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  border: 2px solid transparent;
  position: relative;
  transition: all var(--pixel-transition-fast);
  user-select: none;
  outline: none;
  white-space: nowrap;

  &:focus-visible {
    outline: 2px solid var(--pixel-accent-mint);
    outline-offset: 2px;
  }

  &:active:not(:disabled) {
    transform: scale(0.98);
  }

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  /* Size variants */
  &--sm {
    padding: var(--pixel-space-xs) var(--pixel-space-sm);
    font-size: 7px;
  }

  &--md {
    padding: var(--pixel-space-sm) var(--pixel-space-md);
    font-size: var(--pixel-text-xs);
  }

  &--lg {
    padding: var(--pixel-space-sm) var(--pixel-space-lg);
    font-size: var(--pixel-text-sm);
  }

  &--block {
    display: flex;
    width: 100%;
  }

  /* Color variants */
  &--mint {
    background: rgba(0, 255, 255, 0.1);
    border-color: var(--pixel-accent-mint-dim);
    color: var(--pixel-accent-mint);

    &:hover:not(:disabled) {
      background: rgba(0, 255, 255, 0.2);
      box-shadow: var(--pixel-glow-mint);
    }
  }

  &--pink {
    background: rgba(255, 107, 157, 0.1);
    border-color: var(--pixel-accent-pink-dim);
    color: var(--pixel-accent-pink);

    &:hover:not(:disabled) {
      background: rgba(255, 107, 157, 0.2);
      box-shadow: var(--pixel-glow-pink);
    }
  }

  &--gold {
    background: rgba(255, 215, 0, 0.1);
    border-color: var(--pixel-accent-gold-dim);
    color: var(--pixel-accent-gold);

    &:hover:not(:disabled) {
      background: rgba(255, 215, 0, 0.2);
      box-shadow: var(--pixel-glow-gold);
    }
  }

  &--purple {
    background: rgba(179, 136, 255, 0.1);
    border-color: var(--pixel-accent-purple-dim);
    color: var(--pixel-accent-purple);

    &:hover:not(:disabled) {
      background: rgba(179, 136, 255, 0.2);
      box-shadow: var(--pixel-glow-purple);
    }
  }

  &--ghost {
    background: transparent;
    border-color: transparent;
    color: var(--pixel-text-secondary);

    &:hover:not(:disabled) {
      color: var(--pixel-accent-mint);
      background: rgba(0, 255, 255, 0.05);
    }
  }
}

.pixel-btn__icon {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.pixel-btn__text {
  line-height: 1;
}
</style>
