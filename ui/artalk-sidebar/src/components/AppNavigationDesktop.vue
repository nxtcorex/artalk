<script lang="ts" setup>
/**
 * The page-level tab bar shown on desktop.
 */
import { Search } from 'lucide-vue-next'
import { useNavigationMenu } from './AppNavigationMenu'

const nav = useNavigationMenu()
const { t } = useI18n()
</script>

<template>
  <div v-if="!!Object.keys(nav.tabs).length" class="top-tabbar-wrap">
    <div class="top-tabbar">
      <div class="tabs">
        <button
          v-for="(tabLabel, tabName) in nav.tabs"
          :key="tabName"
          type="button"
          class="tab"
          :class="{ active: nav.curtTab === tabName }"
          @click="nav.goTab(tabName as string)"
        >
          {{ t(tabLabel) }}
        </button>
      </div>

      <button
        v-if="nav.isSearchEnabled"
        type="button"
        class="search-btn"
        :aria-label="t('searchHint')"
        @click="nav.showSearch()"
      >
        <Search class="size-4" />
      </button>
    </div>

    <AppNavigationSearch v-if="!nav.isMobile" :state="nav.searchState" />
  </div>
</template>

<style lang="scss" scoped>
.top-tabbar-wrap {
  position: relative;
  flex: none;
  padding: 4px 24px 0;

  @media (max-width: 1023px) {
    display: none;
  }
}

.top-tabbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.tabs {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  height: 36px;
  padding: 4px;
  border-radius: 8px;
  background: var(--muted);
}

.tab {
  height: 28px;
  padding: 0 12px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--muted-foreground);
  font-size: 14px;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;

  &:hover {
    color: var(--foreground);
  }

  &.active {
    background: var(--background);
    color: var(--foreground);
    font-weight: 500;
    box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  }
}

.search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: transparent;
  color: var(--muted-foreground);
  cursor: pointer;

  &:hover {
    background: var(--muted);
    color: var(--foreground);
  }
}
</style>
