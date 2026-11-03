<script setup lang="ts">
import type { ArtalkType } from 'artalk'
import { Plus } from 'lucide-vue-next'
import { useNavStore } from '../stores/nav'
import { artalk, bootParams } from '../global'
import { Button } from '@/components/ui/button'

const nav = useNavStore()
const sites = ref<ArtalkType.SiteData[]>([])
const curtEditSite = ref<ArtalkType.SiteData | null>(null)
const showSiteCreate = ref(false)
const siteCreateInitVal = ref()
const { t } = useI18n()

onMounted(() => {
  nav.updateTabs({}, 'sites')

  nav.setPageLoading(true)
  artalk?.ctx
    .getApi()
    .sites.getSites()
    .then((res) => {
      sites.value = res.data.sites
    })
    .finally(() => {
      nav.setPageLoading(false)
    })

  // Open site create dialog by view params (from URL query)
  const vp = bootParams.viewParams
  if (vp && vp.create_name && vp.create_urls) {
    siteCreateInitVal.value = { name: vp.create_name, urls: vp.create_urls }
    showSiteCreate.value = true
    nextTick(() => {
      siteCreateInitVal.value = null
      bootParams.viewParams = null
    })
  }
})

function create() {
  curtEditSite.value = null
  showSiteCreate.value = true
}

const sitesGrouped = computed(() => {
  if (sites.value.length === 0) return []

  const grp: ArtalkType.SiteData[][] = []
  let j = -1
  for (let i = 0; i < sites.value.length; i++) {
    const item = sites.value[i]
    if (i % 4 === 0) {
      // Each row has 4 items
      grp.push([])
      j++
    }
    grp[j].push(item)
  }
  return grp
})

function edit(site: ArtalkType.SiteData) {
  showSiteCreate.value = false
  curtEditSite.value = site
}

function onNewSiteCreated(siteNew: ArtalkType.SiteData) {
  sites.value.push(siteNew)
  showSiteCreate.value = false
  nav.refreshSites()
}

function onSiteItemUpdate(site: ArtalkType.SiteData) {
  const index = sites.value.findIndex((s) => s.id === site.id)
  if (index != -1) {
    const orgSite = sites.value[index]
    Object.keys(site).forEach((key) => {
      ;(orgSite as any)[key] = (site as any)[key]
    })
  }
  nav.refreshSites()
}

function onSiteItemRemove(id: number) {
  const index = sites.value.findIndex((p) => p.id === id)
  sites.value.splice(index, 1)
  curtEditSite.value = null
  nav.refreshSites()
}
</script>

<template>
  <div class="atk-site-list">
    <div class="atk-header">
      <div class="atk-title">{{ t('siteCount', { count: sites.length }) }}</div>
      <div class="atk-actions">
        <Button variant="ghost" size="icon" :aria-label="t('createSite')" @click="create()">
          <Plus />
        </Button>
      </div>
    </div>
    <SiteCreate
      v-if="showSiteCreate"
      :init-val="siteCreateInitVal"
      @close="showSiteCreate = false"
      @done="onNewSiteCreated"
    />
    <SiteEditor
      v-if="curtEditSite !== null"
      :site="curtEditSite"
      @close="curtEditSite = null"
      @update="onSiteItemUpdate"
      @remove="onSiteItemRemove"
    />
    <div class="atk-site-rows-wrap">
      <div v-for="(ss, i) in sitesGrouped" :key="i" class="atk-site-row">
        <div
          v-for="site in ss"
          :key="site.id"
          class="atk-site-item"
          :class="{ 'atk-active': curtEditSite === site }"
          @click="edit(site)"
        >
          <div class="atk-site-logo">{{ site.name.substring(0, 1) }}</div>
          <div class="atk-site-name">{{ site.name }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.atk-site-list {
  & > .atk-header {
    display: flex;
    flex-direction: row;
    padding: 15px 30px;
    align-items: center;

    .atk-title {
      flex: auto;
      padding-right: 10px;
    }

    .atk-actions {
      display: flex;
      flex-direction: row;

      .atk-item {
        display: flex;
        height: 30px;
        width: 30px;
        justify-content: center;
        align-items: center;
        user-select: none;
        cursor: pointer;
        border-radius: var(--atk-radius-sm, 6px);
        transition: background-color 0.15s ease;

        &:hover {
          background: var(--atk-accent);
          color: var(--atk-accent-foreground);
        }
      }
    }
  }

  .atk-site-rows-wrap {
    position: relative;

    .atk-site-row {
      display: flex;
      flex-direction: row;
      padding: 10px 20px;
      padding-bottom: 0;
    }

    .atk-site-item {
      display: flex;
      flex-basis: 25%;
      flex-direction: column;
      align-items: center;
      padding-bottom: 5px;
      user-select: none;
      cursor: pointer;
      border: 1px solid transparent;
      border-radius: var(--atk-radius, 8px);
      transition:
        background-color 0.15s ease,
        border-color 0.15s ease;

      .atk-site-logo {
        margin: 15px;
        text-align: center;
        font-size: 20px;
        height: 65px;
        width: 65px;
        line-height: 65px;
        background: var(--atk-secondary);
        color: var(--atk-secondary-foreground);
        border-radius: var(--atk-radius-lg, 12px);
      }

      .atk-site-name {
        text-align: center;
        font-size: 15px;
        color: var(--at-color-sub);
        padding: 0 17px;
        word-break: break-word;
      }

      &.atk-active {
        background-color: var(--atk-accent);
        border: 1px solid var(--at-color-border);
        margin-top: -1px;

        .atk-site-name {
          color: var(--at-color-deep);
        }

        .atk-site-logo {
          background: var(--atk-primary);
          color: var(--atk-primary-foreground);
        }
      }

      &:hover {
        background-color: color-mix(in srgb, var(--atk-accent) 45%, transparent);

        .atk-site-name {
          color: var(--at-color-font);
        }
      }
    }
  }
}
</style>
