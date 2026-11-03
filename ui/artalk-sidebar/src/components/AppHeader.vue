<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { Menu as MenuIcon, Moon, PanelLeft, Sun, X } from 'lucide-vue-next'
import { bootParams } from '../global'
import { useNavStore } from '../stores/nav'
import { useNavigationMenu } from './AppNavigationMenu'
import { Button } from '@/components/ui/button'

const nav = useNavStore()
const menu = useNavigationMenu()
const { t } = useI18n()
const { darkMode } = storeToRefs(nav)
const route = useRoute()

const isEmbedded = bootParams.embedded
const isLogin = computed(() => route.path === '/login')

const pageTitle = computed(() => {
  const page = menu.pages[menu.curtPage]
  return page ? t(page.label) : t('ctrlCenter')
})

function closeEmbeddedSidebar() {
  window.parent?.postMessage({ type: 'artalk:sidebar-close' }, '*')
}
</script>

<template>
  <header class="header" :class="{ elevated: isLogin && isEmbedded }">
    <Button
      variant="outline"
      size="icon"
      class="sidebar-trigger"
      :aria-label="t('toggleSidebar')"
      :title="t('toggleSidebar')"
      @click="nav.toggleSidebar()"
    >
      <PanelLeft class="size-4" />
    </Button>
    <div class="divider"></div>
    <h1 class="page-title">{{ pageTitle }}</h1>

    <Button
      v-if="!isLogin"
      variant="outline"
      class="mobile-nav-trigger"
      :aria-label="t('navMenu')"
      :title="t('navMenu')"
      @click="nav.mobileNavOpen = true"
    >
      <MenuIcon class="size-4" />
      <span class="mobile-nav-label">{{ pageTitle }}</span>
    </Button>

    <div class="header-actions">
      <Button
        variant="outline"
        size="icon"
        class="theme-toggle"
        :aria-label="t('toggleTheme')"
        :title="t('toggleTheme')"
        @click="nav.toggleDarkMode()"
      >
        <Sun v-if="darkMode" class="size-4" />
        <Moon v-else class="size-4" />
      </Button>

      <Button
        v-if="isEmbedded"
        variant="outline"
        size="icon"
        class="close-toggle"
        :aria-label="t('close')"
        :title="t('close')"
        @click="closeEmbeddedSidebar()"
      >
        <X class="size-4" />
      </Button>
    </div>
  </header>
</template>

<style scoped lang="scss">
.header {
  display: flex;
  flex: none;
  align-items: center;
  gap: 12px;
  height: 64px;
  padding: 0 16px;

  &.elevated {
    position: relative;
    z-index: 61;
  }
}

.sidebar-trigger {
  flex-shrink: 0;
}

.divider {
  width: 1px;
  height: 24px;
  background: var(--border);
}

.page-title {
  flex: 1;
  min-width: 0;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 16px;
  font-weight: 600;
}

.mobile-nav-trigger {
  display: none;
  gap: 8px;
  min-width: 0;
  padding: 0 12px;
  color: var(--foreground);
  font-weight: 500;

  .mobile-nav-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
  flex-shrink: 0;
}

.theme-toggle,
.close-toggle {
  color: var(--muted-foreground);
  cursor: pointer;

  &:hover {
    color: var(--foreground);
  }
}

@media (max-width: 1023px) {
  .header {
    gap: 8px;
    height: 56px;
  }

  .sidebar-trigger,
  .divider,
  .page-title {
    display: none;
  }

  .mobile-nav-trigger {
    display: inline-flex;
  }
}
</style>
