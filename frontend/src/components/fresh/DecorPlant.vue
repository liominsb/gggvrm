<template>
  <div class="decor-plant" :class="`decor-plant--${variant}`" aria-hidden="true">
    <!-- Potted Plant -->
    <svg v-if="variant === 'potted'" width="120" height="120" viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
      <!-- Pot -->
      <path d="M35 90l5 15h40l5-15" stroke="var(--fresh-text-muted)" stroke-width="1.5" stroke-linecap="round"/>
      <rect x="38" y="75" width="44" height="15" rx="4" stroke="var(--fresh-mint)" stroke-width="1.5" fill="var(--fresh-mint-light)" opacity="0.6"/>
      <!-- Stem -->
      <path d="M60 75V45" stroke="var(--fresh-mint)" stroke-width="1.5" stroke-linecap="round"/>
      <!-- Leaves -->
      <path d="M60 60c-15-5-22-18-18-25 3-5 10-3 18 5" stroke="var(--fresh-mint)" stroke-width="1.5" stroke-linecap="round"/>
      <path d="M60 55c15-8 25-12 25-5s-8 12-25 15" stroke="var(--fresh-mint)" stroke-width="1.5" stroke-linecap="round"/>
      <path d="M60 50c-8-12-12-22-8-28 3-4 8 0 8 8" stroke="var(--fresh-mint)" stroke-width="1.2" stroke-linecap="round" opacity="0.6"/>
      <path d="M60 50c10-10 20-14 22-8 2 5-6 12-22 12" stroke="var(--fresh-mint)" stroke-width="1.2" stroke-linecap="round" opacity="0.6"/>
      <!-- Top leaf cluster -->
      <path d="M60 45c-3-5-5-10-3-12 2-2 5 1 3 12z" fill="var(--fresh-mint-light)" stroke="var(--fresh-mint)" stroke-width="1"/>
      <path d="M60 45c3-5 5-10 3-12-2-2-5 1-3 12z" fill="var(--fresh-mint-light)" stroke="var(--fresh-mint)" stroke-width="1"/>
      <!-- Soil dots -->
      <circle cx="48" cy="80" r="1.5" fill="var(--fresh-apricot)" opacity="0.5"/>
      <circle cx="55" cy="82" r="1" fill="var(--fresh-apricot)" opacity="0.4"/>
      <circle cx="65" cy="81" r="1.5" fill="var(--fresh-apricot)" opacity="0.5"/>
    </svg>

    <!-- Floating Cloud -->
    <svg v-if="variant === 'cloud'" width="140" height="80" viewBox="0 0 140 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M35 55c-15 0-28-10-28-24 0-13 10-22 22-24 2-13 14-20 24-15 6-8 18-6 24 2 12 0 23 8 23 20s-10 20-23 24" 
        stroke="var(--fresh-mint)" stroke-width="1.5" stroke-linecap="round" 
        fill="var(--fresh-mint-light)" opacity="0.4"/>
      <!-- Small accent clouds -->
      <path d="M100 25c-3 0-6-2-6-5 0-3 2-5 4-5 1-3 5-3 7-1 1-2 4-1 5 1 2 0 4 2 4 4s-2 4-4 5" 
        stroke="var(--fresh-pink)" stroke-width="1" stroke-linecap="round" 
        fill="var(--fresh-pink-light)" opacity="0.5"/>
    </svg>

    <!-- Geometric Shapes -->
    <svg v-if="variant === 'geometric'" width="100" height="100" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="30" cy="30" r="12" stroke="var(--fresh-mint)" stroke-width="1.5" fill="var(--fresh-mint-light)" opacity="0.5"/>
      <rect x="55" y="20" width="20" height="20" rx="4" stroke="var(--fresh-pink)" stroke-width="1.5" fill="var(--fresh-pink-light)" opacity="0.5"/>
      <path d="M25 65l15 20h-30z" stroke="var(--fresh-apricot)" stroke-width="1.5" fill="var(--fresh-apricot-light)" opacity="0.5"/>
      <circle cx="70" cy="70" r="5" fill="var(--fresh-mint)" opacity="0.3"/>
      <circle cx="15" cy="50" r="3" fill="var(--fresh-pink)" opacity="0.3"/>
      <circle cx="85" cy="45" r="2" fill="var(--fresh-apricot)" opacity="0.4"/>
    </svg>

    <!-- Simple Plant Sprout -->
    <svg v-if="variant === 'sprout'" width="60" height="80" viewBox="0 0 60 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M30 35v40" stroke="var(--fresh-mint)" stroke-width="1.5" stroke-linecap="round"/>
      <path d="M30 35c-8-12-15-22-20-25 3-3 10 3 20 15" stroke="var(--fresh-mint)" stroke-width="1.5" stroke-linecap="round"/>
      <path d="M30 35c8-12 15-22 20-25-3-3-10 3-20 15" stroke="var(--fresh-mint)" stroke-width="1.5" stroke-linecap="round"/>
      <circle cx="30" cy="20" r="4" fill="var(--fresh-mint-light)" stroke="var(--fresh-mint)" stroke-width="1"/>
      <path d="M30 75h-8l4-3 4 3z" fill="var(--fresh-apricot-light)" stroke="var(--fresh-apricot)" stroke-width="1" opacity="0.6"/>
    </svg>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  variant?: 'potted' | 'cloud' | 'geometric' | 'sprout'
}>(), {
  variant: 'potted',
})
</script>

<style scoped>
.decor-plant {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  user-select: none;
  pointer-events: none;
}

@media (prefers-reduced-motion: no-preference) {
  .decor-plant--cloud {
    animation: float 4s ease-in-out infinite;
  }

  .decor-plant--sprout {
    animation: float 3s ease-in-out infinite;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-6px); }
  }
}
</style>
