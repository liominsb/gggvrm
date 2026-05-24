<template>
  <div id="app" class="fresh-app">
    <AppHeader />
    <main class="app-main">
      <router-view v-slot="{ Component }">
        <transition name="page-fade" mode="out-in">
          <keep-alive :max="5">
            <component :is="Component" />
          </keep-alive>
        </transition>
      </router-view>
    </main>
    <AppFooter />
  </div>
</template>

<script setup lang="ts">
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
</script>

<style>
/* ============================================================
   FRESH APP — GLOBAL STYLES
   ============================================================ */
.fresh-app *,
.fresh-app *::before,
.fresh-app *::after {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  scroll-behavior: smooth;
}

body {
  font-family: var(--fresh-font-body);
  font-size: var(--fresh-text-base);
  line-height: 1.6;
  color: var(--fresh-text-primary);
  background: var(--fresh-bg-page);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

a {
  color: var(--fresh-mint);
  text-decoration: none;
  transition: color var(--fresh-transition-fast);
}

a:hover {
  color: var(--fresh-mint-hover);
}

img {
  max-width: 100%;
  height: auto;
}

::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--fresh-border-default);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--fresh-text-muted);
}

::selection {
  background: var(--fresh-mint-light);
  color: var(--fresh-text-primary);
}
</style>

<style scoped>
#app {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
}

.app-main {
  flex: 1;
  padding-top: var(--fresh-header-height);
}

/* Page transition */
.page-fade-enter-active {
  transition: opacity 0.2s var(--fresh-ease-out);
}

.page-fade-leave-active {
  transition: opacity 0.15s var(--fresh-ease-out);
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}

.page-fade-leave-to {
  opacity: 0;
}
</style>
