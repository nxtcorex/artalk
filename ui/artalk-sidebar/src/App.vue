<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useNavStore } from './stores/nav'
import { Toaster } from '@/components/ui/sonner'

const nav = useNavStore()
const { scrollableArea, darkMode, sidebarCollapsed } = storeToRefs(nav)

const query = window.matchMedia('(prefers-color-scheme: dark)')

onMounted(() => {
  query.addEventListener('change', onDarkModeChange)
  window.addEventListener('keydown', onKeydown)
})
onUnmounted(() => {
  query.removeEventListener('change', onDarkModeChange)
  window.removeEventListener('keydown', onKeydown)
})
const onDarkModeChange = (e: MediaQueryListEvent) => {
  // auto switch when localStorage is empty
  if (localStorage.getItem('ATK_SIDEBAR_DARK_MODE') === null) darkMode.value = e.matches
}

function onKeydown(e: KeyboardEvent) {
  if (!(e.metaKey || e.ctrlKey) || e.key.toLowerCase() !== 'b') return
  const target = e.target as HTMLElement | null
  if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA')) return
  e.preventDefault()
  nav.toggleSidebar()
}

// Sync the dark class to <html> so teleported layers (dialogs/popovers)
// outside `.app-wrap` still inherit the shadcn token values.
watch(
  darkMode,
  (v) => document.documentElement.classList.toggle('atk-dark-mode', v),
  { immediate: true },
)
</script>

<template>
  <div
    class="app-wrap artalk atk-sidebar"
    :class="{ 'atk-dark-mode': darkMode, 'sidebar-collapsed': sidebarCollapsed }"
  >
    <AppSidebar />

    <div class="inset">
      <AppHeader />
      <AppNavigation />

      <div class="main">
        <div ref="scrollableArea" class="atk-sidebar-inner">
          <div class="atk-sidebar-container">
            <router-view />
          </div>
        </div>
        <LoadingLayer v-if="nav.isPageLoading" />
      </div>
    </div>

    <Toaster position="top-center" :duration="3000" />
  </div>
</template>

<style scoped lang="scss">
$maxWidth: 1100px;

.app-wrap {
  --atk-sidebar-width: 16rem;
  display: flex;
  min-height: 100vh;
  background: var(--sidebar);
  color: var(--foreground);

  &.sidebar-collapsed {
    --atk-sidebar-width: 3rem;
  }
}

.inset {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  height: calc(100vh - 16px);
  margin: 8px 8px 8px 0;
  overflow: hidden;
  border-radius: calc(var(--radius) + 4px);
  background: var(--background);
  box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
}

.main {
  position: relative;
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.atk-sidebar-inner {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-bottom: 24px;
}

:deep(.atk-sidebar-container) {
  position: relative;
  width: 100%;
  max-width: $maxWidth;
  margin: 0 auto;
}

:deep(.atk-pagination-wrap) {
  z-index: 30;
  position: sticky;
  bottom: 0;
  background: var(--background);
  border-top: 1px solid var(--border);
}

@media (max-width: 1023px) {
  .inset {
    height: 100vh;
    margin: 0;
    border-radius: 0;
    box-shadow: none;
  }
}
</style>
