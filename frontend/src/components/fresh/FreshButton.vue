<template>
  <button
    :class="[
      'fresh-btn',
      `fresh-btn--${variant}`,
      `fresh-btn--${size}`,
      { 'fresh-btn--block': block, 'fresh-btn--disabled': disabled }
    ]"
    :disabled="disabled"
    :type="nativeType"
    @click="$emit('click', $event)"
  >
    <span v-if="icon" class="fresh-btn__icon">
      <FreshIcon :name="icon" :size="iconSize" :color="iconColor" />
    </span>
    <span v-if="$slots.default" class="fresh-btn__text">
      <slot />
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import FreshIcon from './FreshIcon.vue'

const props = withDefaults(defineProps<{
  variant?: 'mint' | 'pink' | 'apricot' | 'ghost' | 'outline'
  size?: 'sm' | 'md' | 'lg'
  block?: boolean
  disabled?: boolean
  nativeType?: 'button' | 'submit' | 'reset'
  icon?: 'home' | 'compass' | 'chat' | 'edit' | 'user' | 'settings' | 'star'
    | 'tag' | 'folder' | 'bell' | 'heart' | 'search' | 'arrow-right'
    | 'arrow-down' | 'plus' | 'menu' | 'close' | 'logout' | 'document'
    | 'lock' | 'clock' | 'article' | 'plant' | 'cloud' | 'sparkle'
}>(), {
  variant: 'mint',
  size: 'md',
  block: false,
  disabled: false,
  nativeType: 'button',
})

defineEmits<{ click: [event: MouseEvent] }>()

const iconSize = computed(() => props.size === 'sm' ? 14 : props.size === 'lg' ? 20 : 16)
const iconColor = computed(() => {
  if (props.variant === 'ghost' || props.variant === 'outline') return 'mint'
  return 'default'
})
</script>

<style scoped lang="scss">
.fresh-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-family: var(--fresh-font-display);
  font-weight: 500;
  cursor: pointer;
  border: 1.5px solid transparent;
  border-radius: var(--fresh-radius-sm);
  transition: all var(--fresh-transition-fast);
  user-select: none;
  outline: none;
  white-space: nowrap;
  letter-spacing: 0.02em;
  line-height: 1;

  &:focus-visible {
    outline: 2px solid var(--fresh-mint);
    outline-offset: 2px;
  }

  &:active:not(:disabled) {
    transform: scale(0.97);
  }

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  /* Sizes */
  &--sm {
    padding: 6px 14px;
    font-size: 13px;
  }

  &--md {
    padding: 9px 20px;
    font-size: 14px;
  }

  &--lg {
    padding: 11px 28px;
    font-size: 16px;
  }

  &--block {
    display: flex;
    width: 100%;
  }

  /* Mint */
  &--mint {
    background: var(--fresh-mint);
    color: #fff;

    &:hover:not(:disabled) {
      background: var(--fresh-mint-hover);
      box-shadow: 0 2px 12px rgba(136, 201, 161, 0.35);
    }
  }

  /* Pink */
  &--pink {
    background: var(--fresh-pink);
    color: #fff;

    &:hover:not(:disabled) {
      background: var(--fresh-pink-hover);
      box-shadow: 0 2px 12px rgba(232, 160, 180, 0.35);
    }
  }

  /* Apricot */
  &--apricot {
    background: var(--fresh-apricot);
    color: #fff;

    &:hover:not(:disabled) {
      background: var(--fresh-apricot-hover);
      box-shadow: 0 2px 12px rgba(240, 200, 160, 0.35);
    }
  }

  /* Ghost */
  &--ghost {
    background: transparent;
    border-color: transparent;
    color: var(--fresh-text-secondary);

    &:hover:not(:disabled) {
      background: var(--fresh-bg-hover);
      color: var(--fresh-text-primary);
    }
  }

  /* Outline */
  &--outline {
    background: transparent;
    border-color: var(--fresh-border-default);
    color: var(--fresh-text-primary);

    &:hover:not(:disabled) {
      border-color: var(--fresh-mint);
      color: var(--fresh-mint-hover);
      background: var(--fresh-mint-light);
    }
  }
}

.fresh-btn__icon {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.fresh-btn__text {
  white-space: nowrap;
}
</style>
