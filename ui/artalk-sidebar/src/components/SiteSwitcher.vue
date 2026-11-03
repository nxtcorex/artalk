<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { ChevronsUpDown } from 'lucide-vue-next'
import { useNavStore } from '../stores/nav'
import { useUserStore } from '../stores/user'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

const nav = useNavStore()
const { sites } = storeToRefs(nav)
const { site: curtSite } = storeToRefs(useUserStore())

const open = ref(false)
const router = useRouter()
const { t } = useI18n()

onMounted(() => {
  nav.refreshSites()
})

interface IDisplaySite {
  label: string
  name: string
  logoText: string
}

const displaySites = computed(() => {
  const displays: IDisplaySite[] = []
  displays.push({ label: t('allSites'), name: '', logoText: '_' })
  sites.value.forEach((site) => {
    displays.push({
      label: site.name,
      name: site.name,
      logoText: site.name.substring(0, 1),
    })
  })
  displays.push({
    label: t('siteManage'),
    name: '__SITE_MANAGEMENT__',
    logoText: '+',
  })
  return displays
})

function switchSite(siteName: string) {
  if (siteName === '__SITE_MANAGEMENT__') {
    router.replace('/sites')
  } else {
    curtSite.value = siteName
  }

  open.value = false
}
</script>

<template>
  <DropdownMenu v-model:open="open">
    <DropdownMenuTrigger as-child>
      <button
        type="button"
        class="site-trigger"
        :class="{ active: open }"
        :aria-label="t('siteManage')"
      >
        <span class="trigger-logo">{{ curtSite.substring(0, 1) || '_' }}</span>
        <span class="site-name">{{ curtSite || t('allSites') }}</span>
        <ChevronsUpDown class="trigger-caret size-4 shrink-0" />
      </button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" class="w-60 max-w-[80vw]">
      <DropdownMenuItem
        v-for="(site, i) in displaySites"
        :key="i"
        class="gap-3"
        :class="{ 'bg-accent text-accent-foreground': curtSite === site.name }"
        @click="switchSite(site.name)"
      >
        <span class="site-logo">{{ site.logoText }}</span>
        <span class="truncate">{{ site.label }}</span>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

<style scoped lang="scss">
.site-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  height: 44px;
  padding: 8px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: inherit;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    gap 0.2s ease;

  &:hover,
  &.active {
    background: var(--sidebar-accent);
    color: var(--sidebar-accent-foreground);
  }
}

.trigger-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--sidebar-primary);
  color: var(--sidebar-primary-foreground);
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
  user-select: none;
}

.site-name {
  flex: 1;
  min-width: 0;
  max-width: 200px;
  overflow: hidden;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  font-weight: 600;
  transition:
    opacity 0.15s ease,
    max-width 0.2s ease;
}

.trigger-caret {
  max-width: 16px;
  margin-left: auto;
  color: color-mix(in oklab, var(--sidebar-foreground) 60%, transparent);
  transition:
    opacity 0.15s ease,
    max-width 0.2s ease,
    margin 0.2s ease;
}

.site-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  border-radius: 4px;
  background: var(--muted);
  color: var(--muted-foreground);
  font-size: 12px;
  text-align: center;
}
</style>
