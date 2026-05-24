<template>
  <div class="scanlines" aria-hidden="true"></div>
</template>

<script setup lang="ts">
// Pure CSS CRT scanline overlay — no JS needed
</script>

<style scoped>
.scanlines {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 9998;
  overflow: hidden;
}

.scanlines::before {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0, 0, 0, 0.04) 2px,
    rgba(0, 0, 0, 0.04) 4px
  );
}

.scanlines::after {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(
    ellipse at center,
    transparent 55%,
    rgba(0, 0, 0, 0.25) 100%
  );
}

@media (prefers-reduced-motion: no-preference) {
  .scanlines {
    animation: scanline-flicker 0.12s infinite alternate steps(1);
  }

  @keyframes scanline-flicker {
    0% { opacity: 0.97; }
    50% { opacity: 0.99; }
    100% { opacity: 1; }
  }
}

@media (prefers-reduced-motion: reduce) {
  .scanlines {
    opacity: 0.98;
  }
}
</style>
