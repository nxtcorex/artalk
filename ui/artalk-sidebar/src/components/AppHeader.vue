<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { Moon, PanelLeft, Sun } from 'lucide-vue-next'
import { useNavStore } from '../stores/nav'
import { useNavigationMenu } from './AppNavigationMenu'
import { Button } from '@/components/ui/button'

const nav = useNavStore()
const menu = useNavigationMenu()
const { t } = useI18n()
const { darkMode } = storeToRefs(nav)

const pageTitle = computed(() => {
  const page = menu.pages[menu.curtPage]
  return page ? t(page.label) : t('ctrlCenter')
})
</script>

<template>
  <header class="header">
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

.theme-toggle {
  margin-left: auto;
  flex-shrink: 0;
  color: var(--muted-foreground);
  cursor: pointer;

  &:hover {
    color: var(--foreground);
  }
}

@media (max-width: 1023px) {
  .header {
    height: 56px;
    justify-content: flex-end;
  }

  .sidebar-trigger,
  .divider,
  .page-title {
    display: none;
  }
}
</style>
