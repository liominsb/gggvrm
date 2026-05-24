<template>
  <div
    :class="[
      'pixel-card',
      `pixel-card--${variant}`,
      { 'pixel-card--glow': glow, 'pixel-card--hoverable': hoverable }
    ]"
  >
    <!-- Corner notch decorations -->
    <span class="pixel-card__corner pixel-card__corner--tl"></span>
    <span class="pixel-card__corner pixel-card__corner--tr"></span>
    <span class="pixel-card__corner pixel-card__corner--bl"></span>
    <span class="pixel-card__corner pixel-card__corner--br"></span>

    <!-- Header -->
    <div v-if="title || $slots.header" class="pixel-card__header">
      <h3 v-if="title" class="pixel-card__title pixel-text-display">
        <span class="pixel-card__bracket">[</span>
        {{ title }}
        <span class="pixel-card__bracket">]</span>
      </h3>
      <div v-if="$slots.header" class="pixel-card__header-slot">
        <slot name="header" />
      </div>
    </div>

    <!-- Body -->
    <div class="pixel-card__body">
      <slot />
    </div>

    <!-- Footer -->
    <div v-if="$slots.footer" class="pixel-card__footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
  variant?: 'default' | 'accent' | 'pink' | 'gold' | 'purple'
  glow?: boolean
  hoverable?: boolean
}>(), {
  variant: 'default',
  glow: false,
  hoverable: true,
})
</script>

<style scoped lang="scss">
.pixel-card {
  position: relative;
  background: var(--pixel-bg-surface);
  border: var(--pixel-border-width) solid var(--pixel-border-default);
  padding: var(--pixel-space-md);
  overflow: hidden;
  transition: all var(--pixel-transition-base);

  &--hoverable {
    cursor: pointer;

    &:hover {
      border-color: var(--pixel-accent-mint-dim);
      transform: translateY(-2px);
    }
  }

  &--default {
    border-color: var(--pixel-border-default);
  }

  &--accent {
    border-color: var(--pixel-accent-mint-dim);
  }

  &--pink {
    border-color: var(--pixel-accent-pink-dim);
  }

  &--gold {
    border-color: var(--pixel-accent-gold-dim);
  }

  &--purple {
    border-color: var(--pixel-accent-purple-dim);
  }

  &--glow {
    &.pixel-card--accent {
      box-shadow: 0 0 12px rgba(0, 255, 255, 0.1);
    }

    &.pixel-card--pink {
      box-shadow: 0 0 12px rgba(255, 107, 157, 0.1);
    }

    &.pixel-card--gold {
      box-shadow: 0 0 12px rgba(255, 215, 0, 0.1);
    }

    &.pixel-card--purple {
      box-shadow: 0 0 12px rgba(179, 136, 255, 0.1);
    }
  }

  /* Scanline effect inside card */
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    background: repeating-linear-gradient(
      0deg,
      transparent,
      transparent 2px,
      rgba(0, 255, 255, 0.008) 2px,
      rgba(0, 255, 255, 0.008) 4px
    );
    pointer-events: none;
    z-index: 1;
  }
}

/* Corner notches */
.pixel-card__corner {
  position: absolute;
  width: 10px;
  height: 10px;
  z-index: 2;
  pointer-events: none;

  &::before,
  &::after {
    content: '';
    position: absolute;
    background: currentColor;
  }

  &--tl {
    top: -1px;
    left: -1px;
    color: var(--pixel-accent-mint);

    &::before { top: 0; left: 0; width: 10px; height: 2px; }
    &::after { top: 0; left: 0; width: 2px; height: 10px; }
  }

  &--tr {
    top: -1px;
    right: -1px;
    color: var(--pixel-accent-pink);

    &::before { top: 0; right: 0; width: 10px; height: 2px; }
    &::after { top: 0; right: 0; width: 2px; height: 10px; }
  }

  &--bl {
    bottom: -1px;
    left: -1px;
    color: var(--pixel-accent-gold);

    &::before { bottom: 0; left: 0; width: 10px; height: 2px; }
    &::after { bottom: 0; left: 0; width: 2px; height: 10px; }
  }

  &--br {
    bottom: -1px;
    right: -1px;
    color: var(--pixel-accent-purple);

    &::before { bottom: 0; right: 0; width: 10px; height: 2px; }
    &::after { bottom: 0; right: 0; width: 2px; height: 10px; }
  }
}

.pixel-card__header {
  margin-bottom: var(--pixel-space-md);
  position: relative;
  z-index: 2;
}

.pixel-card__title {
  font-size: var(--pixel-text-sm);
  color: var(--pixel-accent-mint);
  margin: 0;
}

.pixel-card__bracket {
  color: var(--pixel-text-muted);
}

.pixel-card__header-slot {
  margin-top: var(--pixel-space-sm);
}

.pixel-card__body {
  position: relative;
  z-index: 2;
}

.pixel-card__footer {
  margin-top: var(--pixel-space-md);
  padding-top: var(--pixel-space-sm);
  border-top: 1px solid var(--pixel-border-default);
  position: relative;
  z-index: 2;
}
</style>
