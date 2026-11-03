<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { ChevronDown, ChevronsUpDown, LogOut } from 'lucide-vue-next'
import fallbackAvatar from '../assets/favicon.png'
import { useNavStore } from '../stores/nav'
import { useUserStore } from '../stores/user'
import { useNavigationMenu, NavGroupDefs } from './AppNavigationMenu'
import SiteSwitcher from './SiteSwitcher.vue'
import { loadSettingsTree, settingsNavGroups } from '@/lib/settings-tree'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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

const nav = useNavigationMenu()
const app = useNavStore()
const user = useUserStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const { sidebarCollapsed } = storeToRefs(app)
const { is_admin: isAdmin, avatar, name: userName, email } = storeToRefs(user)

const logoutConfirmOpen = ref(false)

const userInitial = computed(() =>
  (userName.value || email.value || 'U').charAt(0).toUpperCase(),
)

function confirmedLogout() {
  logoutConfirmOpen.value = false
  user.logout()
  nextTick(() => {
    router.replace('/login')
  })
}

const transferSections = [
  { key: 'import', label: 'import' },
  { key: 'export', label: 'export' },
]

const navGroups = computed(() =>
  NavGroupDefs.map((group) => ({
    key: group.key,
    items: group.names
      .map((name) => [name, nav.pages[name]] as const)
      .filter(([, page]) => !!page),
  })).filter((group) => group.items.length > 0),
)

const activeGroupPath = computed(() => {
  const value = (route.params as Record<string, string | string[] | undefined>).group
  return Array.isArray(value) ? value[0] : value || ''
})

const activeTransferSection = computed(() => {
  const value = (route.params as Record<string, string | string[] | undefined>).section
  const name = Array.isArray(value) ? value[0] : value || ''
  return name === 'export' ? 'export' : 'import'
})

const hasSubmenu = (pageName: string) => pageName === 'settings' || pageName === 'transfer'

function isNavEntryActive(entry: { key: string; paths: string[] }) {
  return entry.key === activeGroupPath.value || entry.paths.includes(activeGroupPath.value)
}

function goPage(pageName: string) {
  if (app.sidebarCollapsed && hasSubmenu(pageName)) app.expandSidebar()
  nav.goPage(pageName)
}

function goGroup(path: string) {
  router.push(`/settings/${path}`)
}

function goTransferSection(section: string) {
  router.push(`/transfer/${section}`)
}

watch(
  () => nav.curtPage,
  (page) => {
    if (page === 'settings' && isAdmin.value) loadSettingsTree()
  },
  { immediate: true },
)

watch([() => nav.curtPage, activeGroupPath, settingsNavGroups], async () => {
  if (nav.curtPage !== 'settings' || app.sidebarCollapsed) return
  await nextTick()
  document.querySelector('.app-sidebar .subnav-item.active')?.scrollIntoView({ block: 'center' })
})
</script>

<template>
  <aside class="app-sidebar" :class="{ collapsed: sidebarCollapsed }">
    <div class="sidebar-header">
      <SiteSwitcher v-if="isAdmin" />
      <div v-else class="sidebar-brand">
        <img :src="avatar || fallbackAvatar" alt="" class="brand-logo" />
        <span class="brand-name">{{ t('msgCenter') }}</span>
      </div>
    </div>

    <div class="sidebar-content">
      <div v-for="group in navGroups" :key="group.key" class="nav-group">
        <p class="group-label">{{ t(group.key) }}</p>
        <div class="nav-list">
          <template v-for="[pageName, page] in group.items" :key="pageName">
            <button
              type="button"
              class="nav-item"
              :class="{ active: pageName === nav.curtPage }"
              :title="sidebarCollapsed ? t(page.label) : undefined"
              @click="goPage(pageName)"
            >
              <component :is="page.icon" class="nav-icon" />
              <span class="nav-label">{{ t(page.label) }}</span>
              <ChevronDown
                v-if="hasSubmenu(pageName)"
                class="nav-caret"
                :class="{ 'rotate-180': nav.curtPage === pageName }"
              />
            </button>

            <Transition name="subnav">
              <div
                v-if="pageName === 'settings' && nav.curtPage === 'settings'"
                class="subnav"
              >
                <button
                  v-for="entry in settingsNavGroups"
                  :key="entry.key"
                  type="button"
                  class="subnav-item"
                  :class="{ active: isNavEntryActive(entry) }"
                  @click="goGroup(entry.key)"
                >
                  {{ entry.labelKey ? t(entry.labelKey) : entry.label }}
                </button>
              </div>
            </Transition>

            <Transition name="subnav">
              <div
                v-if="pageName === 'transfer' && nav.curtPage === 'transfer'"
                class="subnav"
              >
                <button
                  v-for="section in transferSections"
                  :key="section.key"
                  type="button"
                  class="subnav-item"
                  :class="{ active: section.key === activeTransferSection }"
                  @click="goTransferSection(section.key)"
                >
                  {{ t(section.label) }}
                </button>
              </div>
            </Transition>
          </template>
        </div>
      </div>
    </div>

    <div class="sidebar-footer">
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <button
            type="button"
            class="user-button"
            :title="sidebarCollapsed ? userName || email : undefined"
          >
            <Avatar class="user-avatar">
              <AvatarImage :src="avatar" :alt="userName" />
              <AvatarFallback class="rounded-lg">{{ userInitial }}</AvatarFallback>
            </Avatar>
            <div class="user-meta">
              <span class="user-name">{{ userName || t('user') }}</span>
              <span class="user-email">{{ email }}</span>
            </div>
            <ChevronsUpDown class="user-caret size-4 shrink-0" />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent side="right" align="end" class="w-56">
          <DropdownMenuLabel>
            <div class="menu-user">
              <span class="menu-name">{{ userName }}</span>
              <span class="menu-email">{{ email }}</span>
            </div>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem @click="logoutConfirmOpen = true">
            <LogOut class="mr-2 size-4" />
            {{ t('logout') }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>

    <AlertDialog :open="logoutConfirmOpen" @update:open="(v) => (logoutConfirmOpen = v)">
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

    <button
      type="button"
      class="sidebar-rail"
      :aria-label="t('toggleSidebar')"
      @click="app.toggleSidebar()"
    />
  </aside>
</template>

<style scoped lang="scss">
.app-sidebar {
  position: relative;
  display: flex;
  flex: none;
  flex-direction: column;
  width: 16rem;
  height: 100vh;
  overflow: hidden;
  background: var(--sidebar);
  color: var(--sidebar-foreground);
  transition: width 0.2s ease;

  &.collapsed {
    width: 3rem;
  }

  @media (max-width: 1023px) {
    display: none;
  }
}

.sidebar-header {
  padding: 8px;
}

.collapsed .sidebar-header :deep(.site-trigger) {
  justify-content: center;
  gap: 0;
  padding: 8px 0;
}

.collapsed .sidebar-header :deep(.site-name) {
  max-width: 0;
  opacity: 0;
}

.collapsed .sidebar-header :deep(.trigger-caret) {
  max-width: 0;
  margin: 0;
  opacity: 0;
}

.collapsed .sidebar-header :deep(.site-name),
.collapsed .sidebar-header :deep(.trigger-caret) {
  transition:
    opacity 0.15s ease,
    max-width 0.2s ease;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 8px;

  .brand-logo {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    object-fit: contain;
    flex-shrink: 0;
  }

  .brand-name {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
    font-weight: 600;
    transition:
      opacity 0.15s ease,
      max-width 0.2s ease;
  }

  .collapsed & {
    justify-content: center;
    padding: 8px 0;
  }

  .collapsed & .brand-name {
    max-width: 0;
    opacity: 0;
  }
}

.sidebar-footer {
  padding: 8px;
}

.user-button {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 44px;
  padding: 6px 8px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: inherit;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    gap 0.2s ease;

  &:hover,
  &[data-state='open'] {
    background: var(--sidebar-accent);
    color: var(--sidebar-accent-foreground);
  }

  .collapsed & {
    gap: 0;
    justify-content: center;
    padding: 6px 0;
  }
}

.user-avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 8px;
}

.user-meta {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  max-width: 200px;
  overflow: hidden;
  line-height: 1.2;
  text-align: left;
  transition:
    opacity 0.15s ease,
    max-width 0.2s ease;

  .collapsed & {
    max-width: 0;
    opacity: 0;
  }
}

.user-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 600;
}

.user-email {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: color-mix(in oklab, var(--sidebar-foreground) 60%, transparent);
  font-size: 12px;
}

.user-caret {
  max-width: 16px;
  margin-left: auto;
  color: color-mix(in oklab, var(--sidebar-foreground) 60%, transparent);
  transition:
    opacity 0.15s ease,
    max-width 0.2s ease,
    margin 0.2s ease;

  .collapsed & {
    max-width: 0;
    margin: 0;
    opacity: 0;
  }
}

.menu-user {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.menu-name {
  font-size: 13px;
  font-weight: 600;
}

.menu-email {
  color: var(--muted-foreground);
  font-size: 12px;
}

.sidebar-content {
  flex: 1;
  min-height: 0;
  padding: 0 8px 16px;
  overflow-y: auto;
  overflow-x: hidden;
}

.nav-group + .nav-group {
  margin-top: 16px;
  transition: margin-top 0.2s ease;

  .collapsed & {
    margin-top: 8px;
  }
}

.group-label {
  display: flex;
  align-items: center;
  height: 32px;
  margin: 0;
  padding: 0 8px;
  overflow: hidden;
  color: color-mix(in oklab, var(--sidebar-foreground) 70%, transparent);
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  transition:
    opacity 0.15s ease,
    height 0.2s ease,
    margin 0.2s ease;

  .collapsed & {
    height: 0;
    margin: 0;
    opacity: 0;
  }
}

.nav-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: inherit;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease,
    gap 0.2s ease;

  &:hover {
    background: var(--sidebar-accent);
    color: var(--sidebar-accent-foreground);
  }

  &.active {
    background: var(--sidebar-accent);
    color: var(--sidebar-accent-foreground);
    font-weight: 500;
  }

  .collapsed & {
    gap: 0;
    justify-content: center;
  }
}

.nav-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.nav-label {
  flex: 1;
  min-width: 0;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition:
    opacity 0.15s ease,
    max-width 0.2s ease;

  .collapsed & {
    max-width: 0;
    opacity: 0;
  }
}

.nav-caret {
  width: 16px;
  height: 16px;
  max-width: 16px;
  flex-shrink: 0;
  transition:
    opacity 0.15s ease,
    max-width 0.2s ease,
    transform 0.15s ease;

  .collapsed & {
    max-width: 0;
    margin: 0;
    opacity: 0;
  }
}

.subnav {
  display: flex;
  flex-direction: column;
  gap: 1px;
  max-height: 1200px;
  margin: 2px 0 4px 15px;
  padding: 2px 0 2px 9px;
  overflow: hidden;
  border-left: 1px solid var(--sidebar-border);
  opacity: 1;
  transition:
    max-height 0.25s ease,
    opacity 0.15s ease,
    margin 0.2s ease,
    padding 0.2s ease;
}

.collapsed .subnav {
  max-height: 0;
  margin-top: 0;
  margin-bottom: 0;
  padding-top: 0;
  padding-bottom: 0;
  opacity: 0;
  pointer-events: none;
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
  padding: 6px 8px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: color-mix(in oklab, var(--sidebar-foreground) 70%, transparent);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;

  &:hover {
    background: var(--sidebar-accent);
    color: var(--sidebar-accent-foreground);
  }

  &.active {
    background: var(--sidebar-accent);
    color: var(--sidebar-accent-foreground);
    font-weight: 500;
  }
}

.sidebar-rail {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 4px;
  border: 0;
  background: transparent;
  cursor: col-resize;
  transition: background-color 0.15s ease;

  &:hover {
    background: var(--sidebar-border);
  }
}
</style>
