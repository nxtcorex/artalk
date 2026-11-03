<script lang="ts" setup>
/**
 * The part of the navigation bar that is displayed on mobile devices and the sidebar
 */
import { storeToRefs } from 'pinia'
import { ChevronDown, LogOut, Search } from 'lucide-vue-next'
import { NavGroupDefs, useNavigationMenu } from './AppNavigationMenu'
import SiteSwitcher from './SiteSwitcher.vue'
import { loadSettingsTree, settingsNavGroups } from '@/lib/settings-tree'
import { useNavStore } from '@/stores/nav'
import { useUserStore } from '@/stores/user'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const user = useUserStore()
const { is_admin: isAdmin, name: userName, email } = storeToRefs(user)
const { mobileNavOpen } = storeToRefs(useNavStore())

const settingsExpanded = ref(false)
const transferExpanded = ref(false)
const logoutConfirmOpen = ref(false)
const tabListEl = ref<HTMLElement | null>(null)
const nav = useNavigationMenu({
  onGoPage: () => {
    mobileNavOpen.value = false
  },
})

const navGroups = computed(() =>
  NavGroupDefs.map((group) => ({
    key: group.key,
    items: group.names
      .map((name) => [name, nav.pages[name]] as const)
      .filter(([, page]) => !!page),
  })).filter((group) => group.items.length > 0),
)

const userInitial = computed(() =>
  (userName.value || email.value || 'U').charAt(0).toUpperCase(),
)

function confirmedLogout() {
  logoutConfirmOpen.value = false
  mobileNavOpen.value = false
  user.logout()
  nextTick(() => {
    router.replace('/login')
  })
}

const transferSections = [
  { key: 'import', label: 'import' },
  { key: 'export', label: 'export' },
]

const activeTransferSection = computed(() => {
  const value = (route.params as Record<string, string | string[] | undefined>).section
  const name = Array.isArray(value) ? value[0] : value || ''
  return name === 'export' ? 'export' : 'import'
})

const activeGroupPath = computed(() => {
  const value = (route.params as Record<string, string | string[] | undefined>).group
  return Array.isArray(value) ? value[0] : value || ''
})

watch(mobileNavOpen, (open) => {
  if (!open) return
  if (nav.curtPage === 'settings') {
    settingsExpanded.value = true
    loadSettingsTree()
  }
  if (nav.curtPage === 'transfer') transferExpanded.value = true
})

function toggleSettings() {
  settingsExpanded.value = !settingsExpanded.value
  if (settingsExpanded.value) loadSettingsTree()
}

function toggleTransfer() {
  transferExpanded.value = !transferExpanded.value
}

function goGroupAndClose(path: string) {
  mobileNavOpen.value = false
  router.push(`/settings/${path}`)
}

function goTransferAndClose(section: string) {
  mobileNavOpen.value = false
  router.push(`/transfer/${section}`)
}

function isNavEntryActive(entry: { key: string; paths: string[] }) {
  return entry.key === activeGroupPath.value || entry.paths.includes(activeGroupPath.value)
}

/**
 * For tab list for horizontal scroll
 */
const wheelHandler = (evt: WheelEvent) => {
  evt.preventDefault()
  tabListEl.value && (tabListEl.value!.scrollLeft += evt.deltaY)
}
onMounted(() => tabListEl.value!.addEventListener('wheel', wheelHandler))
onUnmounted(() => tabListEl.value!.removeEventListener('wheel', wheelHandler))

function goPageAndClose(pageName: string) {
  nav.goPage(pageName)
  mobileNavOpen.value = false
}
</script>

<template>
  <div class="top-navigation">
    <Sheet v-model:open="mobileNavOpen">
      <SheetContent side="left" class="flex w-64 flex-col p-0" :show-close="false">
        <SheetHeader class="border-b p-2 text-left">
          <SheetTitle class="sr-only">{{ t('ctrlCenter') }}</SheetTitle>
          <SheetDescription class="sr-only">{{ t('navMenu') }}</SheetDescription>
          <SiteSwitcher v-if="isAdmin" />
          <div v-else class="sheet-brand">{{ t('msgCenter') }}</div>
        </SheetHeader>

        <nav class="sheet-nav">
          <div v-for="group in navGroups" :key="group.key" class="sheet-group">
            <p class="group-label">{{ t(group.key) }}</p>
            <template v-for="[pageName, page] in group.items" :key="pageName">
              <button
                type="button"
                class="nav-item"
                :class="{ active: pageName === nav.curtPage }"
                @click="
                  pageName === 'settings'
                    ? toggleSettings()
                    : pageName === 'transfer'
                      ? toggleTransfer()
                      : goPageAndClose(pageName as string)
                "
              >
                <component :is="page.icon" class="size-4 shrink-0" />
                <span class="truncate">{{ t(page.label) }}</span>
                <ChevronDown
                  v-if="pageName === 'settings' || pageName === 'transfer'"
                  class="ml-auto size-4 shrink-0 transition-transform"
                  :class="{
                    'rotate-180':
                      pageName === 'settings' ? settingsExpanded : transferExpanded,
                  }"
                />
              </button>

              <Transition name="subnav">
                <div v-if="pageName === 'settings' && settingsExpanded" class="subnav">
                  <button
                    v-for="entry in settingsNavGroups"
                    :key="entry.key"
                    type="button"
                    class="subnav-item"
                    :class="{ active: isNavEntryActive(entry) }"
                    @click="goGroupAndClose(entry.key)"
                  >
                    {{ entry.labelKey ? t(entry.labelKey) : entry.label }}
                  </button>
                </div>
              </Transition>

              <Transition name="subnav">
                <div v-if="pageName === 'transfer' && transferExpanded" class="subnav">
                  <button
                    v-for="section in transferSections"
                    :key="section.key"
                    type="button"
                    class="subnav-item"
                    :class="{ active: section.key === activeTransferSection }"
                    @click="goTransferAndClose(section.key)"
                  >
                    {{ t(section.label) }}
                  </button>
                </div>
              </Transition>
            </template>
          </div>
        </nav>

        <div class="sheet-footer">
          <div class="sheet-user">
            <span class="sheet-avatar">{{ userInitial }}</span>
            <div class="sheet-user-meta">
              <span class="sheet-name">{{ userName }}</span>
              <span class="sheet-email">{{ email }}</span>
            </div>
          </div>
          <button type="button" class="sheet-logout" @click="logoutConfirmOpen = true">
            <LogOut class="size-4" />
            {{ t('logout') }}
          </button>
        </div>

        <AlertDialog
          :open="logoutConfirmOpen"
          @update:open="(v) => (logoutConfirmOpen = v)"
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{{ t('logout') }}</AlertDialogTitle>
              <AlertDialogDescription>{{ t('logoutConfirm') }}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>{{ t('cancel') }}</AlertDialogCancel>
              <AlertDialogAction @click="confirmedLogout()">{{ t('logout') }}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </SheetContent>
    </Sheet>

    <div ref="tabListEl" class="tab-list">
      <button
        v-for="(tabLabel, tabName) in nav.tabs"
        :key="tabName"
        type="button"
        class="item"
        :class="{ active: nav.curtTab === tabName }"
        @click="nav.goTab(tabName as string)"
      >
        {{ t(tabLabel) }}
      </button>

      <button
        v-if="nav.isSearchEnabled"
        type="button"
        class="item search-btn"
        :aria-label="t('searchHint')"
        @click="nav.showSearch()"
      >
        <Search class="size-4" />
      </button>
    </div>

    <AppNavigationSearch v-if="nav.isMobile" :state="nav.searchState" />
  </div>
</template>

<style lang="scss" scoped>
.top-navigation {
  position: relative;
  display: flex;
  flex-direction: row;
  align-items: center;
  height: 48px;
  border-bottom: 1px solid var(--border);
  background: var(--background);

  @media (min-width: 1024px) {
    display: none;
  }
}

.tab-list {
  position: relative;
  display: flex;
  flex: auto;
  flex-direction: row;
  height: 100%;
  overflow-x: auto;
  overflow-y: hidden;

  &::-webkit-scrollbar {
    width: 0;
    height: 0;
    background: transparent;
  }
}

.item {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 0 14px;
  border: 0;
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
    background: color-mix(in oklab, var(--muted) 50%, transparent);
  }

  &.active {
    color: var(--foreground);
    font-weight: 500;
    box-shadow: inset 0 -2px 0 var(--foreground);
  }

  &.search-btn {
    margin-left: auto;
    padding: 0 14px;
  }
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 8px 12px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--muted-foreground);
  font-size: 14px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;

  &:hover {
    background: var(--muted);
    color: var(--foreground);
  }

  &.active {
    background: var(--muted);
    color: var(--foreground);
  }
}

.sheet-brand {
  padding: 8px;
  font-size: 14px;
  font-weight: 600;
}

.sheet-nav {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
  padding: 8px;
  overflow-y: auto;
}

.sheet-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.group-label {
  display: flex;
  align-items: center;
  height: 32px;
  margin: 0;
  padding: 0 8px;
  color: color-mix(in oklab, var(--foreground) 70%, transparent);
  font-size: 12px;
  font-weight: 500;
}

.sheet-footer {
  padding: 8px;
  border-top: 1px solid var(--border);
}

.sheet-user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
}

.sheet-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 8px;
  background: var(--secondary);
  color: var(--secondary-foreground);
  font-size: 13px;
  font-weight: 600;
}

.sheet-user-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
  line-height: 1.2;
}

.sheet-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 600;
}

.sheet-email {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted-foreground);
  font-size: 12px;
}

.sheet-logout {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  margin-top: 4px;
  padding: 8px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--muted-foreground);
  font-size: 14px;
  text-align: left;
  cursor: pointer;

  &:hover {
    background: var(--muted);
    color: var(--foreground);
  }
}

.subnav {
  display: flex;
  flex-direction: column;
  gap: 1px;
  max-height: 1200px;
  margin: 0 0 4px 19px;
  padding-left: 9px;
  overflow: hidden;
  border-left: 1px solid var(--border);
  transition:
    max-height 0.25s ease,
    opacity 0.15s ease,
    margin 0.2s ease,
    padding 0.2s ease;
}

.subnav-enter-active,
.subnav-leave-active {
  transition:
    max-height 0.25s ease,
    opacity 0.15s ease,
    margin 0.2s ease,
    padding 0.2s ease;
}

.subnav-enter-from,
.subnav-leave-to {
  max-height: 0;
  margin-top: 0;
  margin-bottom: 0;
  padding-top: 0;
  padding-bottom: 0;
  opacity: 0;
}

.subnav-item {
  width: 100%;
  padding: 7px 10px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--muted-foreground);
  font-size: 13px;
  text-align: left;
  cursor: pointer;

  &:hover {
    background: var(--muted);
    color: var(--foreground);
  }

  &.active {
    background: var(--muted);
    color: var(--foreground);
    font-weight: 500;
  }
}
</style>
