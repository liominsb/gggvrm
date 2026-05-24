<template>
  <div
    :class="['fresh-card', {
      'fresh-card--accent': accent,
      'fresh-card--hoverable': hoverable,
      'fresh-card--compact': compact,
    }]"
  >
    <div v-if="title || $slots.header" class="fresh-card__header">
      <h3 v-if="title" class="fresh-card__title">
        <slot name="title-icon" />
        {{ title }}
      </h3>
      <slot name="header" />
    </div>

    <div class="fresh-card__body">
      <slot />
    </div>

    <div v-if="$slots.footer" class="fresh-card__footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
  accent?: 'mint' | 'pink' | 'apricot' | 'none'
  hoverable?: boolean
  compact?: boolean
}>(), {
  accent: 'none',
  hoverable: true,
  compact: false,
})
</script>

<style scoped lang="scss">
.fresh-card {
  background: var(--fresh-bg-surface);
  border-radius: var(--fresh-radius-lg);
  box-shadow: var(--fresh-shadow-sm);
  padding: var(--fresh-space-lg);
  transition: box-shadow var(--fresh-transition-base),
              transform var(--fresh-transition-base);

  &--hoverable {
    cursor: pointer;

    &:hover {
      box-shadow: var(--fresh-shadow-md);
      transform: translateY(-2px);
    }
  }

  &--compact {
    padding: var(--fresh-space-md);
  }

  &--accent {
    position: relative;

    &.fresh-card--mint::before {
      background: var(--fresh-mint);
    }

    &.fresh-card--pink::before {
      background: var(--fresh-pink);
    }

    &.fresh-card--apricot::before {
      background: var(--fresh-apricot);
    }

    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: var(--fresh-space-lg);
      bottom: var(--fresh-space-lg);
      width: 3px;
      border-radius: 0 3px 3px 0;
      opacity: 0.6;
    }
  }
}

.fresh-card__header {
  margin-bottom: var(--fresh-space-md);
}

.fresh-card__title {
  display: flex;
  align-items: center;
  gap: var(--fresh-space-sm);
  font-size: var(--fresh-text-sm);
  font-weight: 600;
  color: var(--fresh-text-primary);
  margin: 0;
  letter-spacing: 0.02em;
}

.fresh-card__body {
  position: relative;
}

.fresh-card__footer {
  margin-top: var(--fresh-space-md);
  padding-top: var(--fresh-space-sm);
  border-top: 1px solid var(--fresh-border-light);
}
</style>
