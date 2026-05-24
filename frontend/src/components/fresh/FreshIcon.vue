<template>
  <svg
    :width="size"
    :height="size"
    :viewBox="viewBox"
    xmlns="http://www.w3.org/2000/svg"
    :class="['fresh-icon', `fresh-icon--${color}`]"
    fill="none"
    stroke="currentColor"
    stroke-width="1.5"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <!-- Home -->
    <template v-if="name === 'home'">
      <path d="M3 8L8 3l5 5" />
      <path d="M4 6.5V13h8V6.5" />
      <path d="M9 13v-3H7v3" />
    </template>

    <!-- Compass -->
    <template v-if="name === 'compass'">
      <circle cx="8" cy="8" r="6.5" />
      <path d="M10.5 5.5L6.5 7l-2 4.5 4-1.5z" />
      <circle cx="8" cy="8" r="0.8" fill="currentColor" stroke="none" />
    </template>

    <!-- Chat -->
    <template v-if="name === 'chat'">
      <rect x="2.5" y="3.5" width="11" height="8" rx="2" />
      <path d="M5.5 14.5V11.5H4a1.5 1.5 0 01-1.5-1.5" opacity="0.4" />
      <path d="M7 8h4M7 6h2" opacity="0.4" />
    </template>

    <!-- Edit -->
    <template v-if="name === 'edit'">
      <path d="M11.5 2.5l2 2-8 8H3.5v-2l8-8z" />
      <path d="M10 4l2 2" opacity="0.5" />
    </template>

    <!-- User -->
    <template v-if="name === 'user'">
      <circle cx="8" cy="5" r="3" />
      <path d="M2.5 14c0-3 2.5-5.5 5.5-5.5S13.5 11 13.5 14" />
    </template>

    <!-- Settings -->
    <template v-if="name === 'settings'">
      <circle cx="8" cy="8" r="3" />
      <path d="M8 1v2M8 13v2M1 8h2M13 8h2M3.1 3.1l1.4 1.4M11.5 11.5l1.4 1.4M3.1 12.9l1.4-1.4M11.5 4.5l1.4-1.4" />
    </template>

    <!-- Star -->
    <template v-if="name === 'star'">
      <path d="M8 1.5l1.8 4.2 4.7.5-3.5 3.2 1 4.6L8 12l-4 2 1-4.6-3.5-3.2 4.7-.5z" />
    </template>

    <!-- Tag -->
    <template v-if="name === 'tag'">
      <path d="M2 2.5h7l5.5 5.5-7 7L2 9.5v-7z" />
      <circle cx="5" cy="5" r="0.8" fill="currentColor" stroke="none" />
    </template>

    <!-- Folder -->
    <template v-if="name === 'folder'">
      <path d="M2 4.5h4l1.5-1.5H14v9H2v-7.5z" />
    </template>

    <!-- Bell -->
    <template v-if="name === 'bell'">
      <path d="M8 2c-2 0-3.5 1.5-3.5 4 0 2-.5 3-1.5 4h10c-1-1-1.5-2-1.5-4 0-2.5-1.5-4-3.5-4z" />
      <path d="M9.5 12.5a1.5 1.5 0 01-3 0" />
    </template>

    <!-- Heart -->
    <template v-if="name === 'heart'">
      <path d="M8 13.5l-5-5.5C1.5 6.5 2 4 4 3s3.5 1 4 2c.5-1 2-2 4-2s2.5 3.5 1 5l-5 5.5z" />
    </template>

    <!-- Search -->
    <template v-if="name === 'search'">
      <circle cx="6.5" cy="6.5" r="5" />
      <path d="M10 10l4.5 4.5" />
    </template>

    <!-- Arrow Right -->
    <template v-if="name === 'arrow-right'">
      <path d="M3 8h10M11 5l4 3-4 3" />
    </template>

    <!-- Arrow Down -->
    <template v-if="name === 'arrow-down'">
      <path d="M8 3v10M5 11l3 4 3-4" />
    </template>

    <!-- Plus -->
    <template v-if="name === 'plus'">
      <path d="M8 2v12M2 8h12" />
    </template>

    <!-- Menu -->
    <template v-if="name === 'menu'">
      <path d="M2 4h12M2 8h12M2 12h12" />
    </template>

    <!-- Close -->
    <template v-if="name === 'close'">
      <path d="M3 3l10 10M13 3L3 13" />
    </template>

    <!-- Logout -->
    <template v-if="name === 'logout'">
      <path d="M5 2.5h-2v11h2M10 8H3M10 5.5l3.5 2.5-3.5 2.5" />
    </template>

    <!-- Document -->
    <template v-if="name === 'document'">
      <path d="M5 2.5h4l4.5 4.5v8.5H5v-13z" />
      <path d="M9 2.5V7h4.5" opacity="0.4" />
      <path d="M7 10h4M7 8h2" opacity="0.3" />
    </template>

    <!-- Lock -->
    <template v-if="name === 'lock'">
      <rect x="3" y="6.5" width="10" height="8" rx="1.5" />
      <path d="M5.5 6.5V4.5a2.5 2.5 0 015 0v2" />
      <circle cx="8" cy="10.5" r="0.8" fill="currentColor" stroke="none" />
    </template>

    <!-- Clock -->
    <template v-if="name === 'clock'">
      <circle cx="8" cy="8" r="6.5" />
      <path d="M8 5v3.5L10 10" />
    </template>

    <!-- Article -->
    <template v-if="name === 'article'">
      <rect x="2.5" y="2.5" width="11" height="11" rx="1.5" />
      <path d="M5.5 6h6M5.5 8.5h4M5.5 11h3" opacity="0.35" />
    </template>

    <!-- Plant -->
    <template v-if="name === 'plant'">
      <path d="M8 15V8" />
      <path d="M8 8c-1.5-2.5-1-6 0-6.5s1.5 4 0 6.5z" />
      <path d="M8 8c1.5-2.5 1-6 0-6.5S6.5 4 8 8z" />
      <path d="M6 7C4 5.5 2.5 6.5 2.5 7.5S4 8.5 6 7z" />
      <path d="M10 7c2-1.5 3.5-.5 3.5.5S12 8.5 10 7z" />
      <rect x="5.5" y="13" width="5" height="3" rx="1" opacity="0.7" />
    </template>

    <!-- Cloud -->
    <template v-if="name === 'cloud'">
      <path d="M5 10c-2 0-3.5-1.5-3.5-3.5S3 3 5 3c.3-2 2.5-3 4.2-2.2 1-.5 2.3-.2 3 .8 2 .2 3.3 2 3 3.8 1 .5 1.3 2 .8 3-.3.5-1 .7-1.5.5" />
      <circle cx="4.5" cy="7.5" r="0.4" fill="currentColor" stroke="none" />
      <circle cx="8" cy="5" r="0.3" fill="currentColor" stroke="none" />
    </template>

    <!-- Sparkle -->
    <template v-if="name === 'sparkle'">
      <path d="M8 2v2M8 12v2M2 8h2M12 8h2" />
      <path d="M4.5 4.5l1.5 1.5M10 10l1.5 1.5M4.5 11.5L6 10M10 6l1.5-1.5" opacity="0.5" />
      <circle cx="8" cy="8" r="1.5" fill="currentColor" stroke="none" opacity="0.3" />
    </template>
  </svg>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  name: 'home' | 'compass' | 'chat' | 'edit' | 'user' | 'settings' | 'star'
    | 'tag' | 'folder' | 'bell' | 'heart' | 'search' | 'arrow-right'
    | 'arrow-down' | 'plus' | 'menu' | 'close' | 'logout' | 'document'
    | 'lock' | 'clock' | 'article' | 'plant' | 'cloud' | 'sparkle'
  size?: number | string
  color?: 'mint' | 'pink' | 'apricot' | 'default'
}>(), {
  size: 20,
  color: 'default',
})

const viewBox = '0 0 16 16'
</script>

<style scoped>
.fresh-icon {
  display: inline-block;
  flex-shrink: 0;
  vertical-align: middle;
}

.fresh-icon--default { color: var(--fresh-text-secondary); }
.fresh-icon--mint { color: var(--fresh-mint); }
.fresh-icon--pink { color: var(--fresh-pink); }
.fresh-icon--apricot { color: var(--fresh-apricot); }
</style>
