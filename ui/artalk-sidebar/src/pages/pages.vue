<script setup lang="ts">
import { storeToRefs } from 'pinia'
import type { ArtalkType } from 'artalk'
import { Eraser, Flame, RefreshCw } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { artalk } from '../global'
import { useNavStore } from '../stores/nav'
import { useUserStore } from '../stores/user'
import Pagination from '../components/Pagination.vue'
import { Button } from '@/components/ui/button'

const nav = useNavStore()
const { site: curtSite } = storeToRefs(useUserStore())
const pages = ref<ArtalkType.PageData[]>([])
const curtEditPageID = ref<number | null>(null)
const { t } = useI18n()

const pageSize = ref(20)
const pageTotal = ref(0)
const search = ref('')
const pagination = ref<InstanceType<typeof Pagination>>()
const showActBarBorder = ref(false)
const refreshBtn = ref({
  isRun: false,
  statusText: '',
})

onMounted(() => {
  nav.updateTabs(
    {
      all: 'all',
    },
    'all',
  )

  // Users search
  nav.enableSearch(
    (value: string) => {
      search.value = value
      fetchPages(0)
    },
    () => {
      if (search.value === '') return
      search.value = ''
      fetchPages(0)
    },
  )

  fetchPages(0)

  // Refresh task status recovery
  getRefreshTaskStatus().then((d) => {
    if (d.is_progress === true) {
      refreshBtn.value.isRun = true
      refreshBtn.value.statusText = d.msg
      startRefreshTaskWatchdog()
    }
  })
})

watch(curtSite, () => {
  pagination.value?.reset()
  fetchPages(0)
})

onMounted(() => nav.scrollableArea?.addEventListener('scroll', scrollHandler))
onUnmounted(() => nav.scrollableArea?.removeEventListener('scroll', scrollHandler))

function scrollHandler() {
  showActBarBorder.value = nav.scrollableArea!.scrollTop > 10
}

function editPage(page: ArtalkType.PageData) {
  curtEditPageID.value = page.id
}

function fetchPages(offset: number) {
  if (offset === 0) pagination.value?.reset()
  nav.setPageLoading(true)
  artalk?.ctx
    .getApi()
    .pages.getPages({
      site_name: curtSite.value,
      offset: offset,
      limit: pageSize.value,
      search: search.value,
      has_comments: true,
    })
    .then((res) => {
      pageTotal.value = res.data.count
      pages.value = res.data.pages
      nav.scrollPageToTop()
    })
    .finally(() => {
      nav.setPageLoading(false)
    })
}

function onChangePage(offset: number) {
  fetchPages(offset)
}

function onPageItemUpdate(page: ArtalkType.PageData) {
  const index = pages.value.findIndex((p) => p.id === page.id)
  if (index != -1) {
    const orgPage = pages.value[index]
    Object.keys(page).forEach((key) => {
      ;(orgPage as any)[key] = (page as any)[key]
    })
  }
}

function onPageItemRemove(id: number) {
  const index = pages.value.findIndex((p) => p.id === id)
  pages.value.splice(index, 1)
}

async function getRefreshTaskStatus() {
  return (await artalk!.ctx.getApi().pages.getPageFetchStatus()).data
}

function startRefreshTaskWatchdog() {
  // TODO: Not perfect polling update status
  const timerID = window.setInterval(async () => {
    const d = await getRefreshTaskStatus()

    if (d.is_progress === false) {
      clearInterval(timerID)
      setRefreshTaskDone()
      return
    }

    refreshBtn.value.statusText = d.msg
  }, 1000)
}

function setRefreshTaskDone() {
  refreshBtn.value.statusText = t('updateComplete')
  window.setTimeout(() => {
    refreshBtn.value.isRun = false
  }, 1500)
}

async function refreshAllPages() {
  if (refreshBtn.value.isRun) return
  refreshBtn.value.isRun = true
  refreshBtn.value.statusText = t('updateReady')

  try {
    await artalk!.ctx.getApi().pages.fetchAllPages({
      site_name: curtSite.value,
    })
  } catch (err: any) {
    toast.error(err.msg)
    setRefreshTaskDone()
    return
  }

  startRefreshTaskWatchdog()
}

function cacheFlush() {
  artalk!.ctx
    .getApi()
    .cache.flushCache()
    .then((res) => toast.success(res.data.msg ?? 'OK'))
    .catch(() => toast.error(t('opFailed')))
}

function cacheWarm() {
  artalk!.ctx
    .getApi()
    .cache.warmUpCache()
    .then((res) => toast.success(res.data.msg ?? 'OK'))
    .catch(() => toast.error(t('opFailed')))
}

function openPage(url: string) {
  window.open(url)
}
</script>

<template>
  <div class="atk-page-list-wrap">
    <div class="atk-header-action-bar" :class="{ bordered: showActBarBorder }">
      <Button variant="outline" size="sm" :disabled="refreshBtn.isRun" @click="refreshAllPages()">
        <RefreshCw :class="['size-3.5', { 'animate-spin': refreshBtn.isRun }]" />
        {{ refreshBtn.isRun ? refreshBtn.statusText : t('updateTitle') }}
      </Button>
      <Button variant="outline" size="sm" @click="cacheFlush()">
        <Eraser class="size-3.5" />
        {{ t('cacheClear') }}
      </Button>
      <Button variant="outline" size="sm" @click="cacheWarm()">
        <Flame class="size-3.5" />
        {{ t('cacheWarm') }}
      </Button>
    </div>
    <div class="atk-page-list">
      <div v-for="page in pages" :key="page.id" class="atk-page-item">
        <div class="atk-page-main">
          <div class="atk-title" @click="openPage(page.url)">
            {{ page.title }}
          </div>
          <div class="atk-sub" @click="openPage(page.url)">{{ page.url }}</div>
        </div>
        <div class="atk-page-actions">
          <div class="atk-item atk-edit-btn" @click="editPage(page)">
            <i class="atk-icon atk-icon-edit"></i>
          </div>
        </div>
        <PageEditor
          v-if="curtEditPageID === page.id"
          :page="page"
          @close="curtEditPageID = null"
          @update="onPageItemUpdate"
          @remove="onPageItemRemove"
        />
      </div>
    </div>
    <Pagination
      ref="pagination"
      :page-size="pageSize"
      :total="pageTotal"
      :disabled="nav.isPageLoading"
      @change="onChangePage"
    />
  </div>
</template>

<style scoped lang="scss">
.atk-page-list-wrap {
  .atk-header-action-bar {
    position: sticky;
    top: 0;
    z-index: 10;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 24px;
    background: var(--background);
    border-bottom: 1px solid transparent;
    transition: 0.3s ease-out padding;

    &.bordered {
      border-color: var(--border);
    }
  }
}

.atk-page-list {
  .atk-page-item {
    display: flex;
    flex-direction: row;
    position: relative;
    min-height: 120px;
    align-items: center;
    transition: background-color 0.15s ease;

    &:hover {
      background: color-mix(in srgb, var(--atk-accent) 40%, transparent);
    }

    &:not(:last-child) {
      border-bottom: 1px solid var(--at-color-border);
    }
  }

  .atk-page-main {
    display: flex;
    flex-direction: column;
    flex: auto;
    padding: 20px 30px;

    .atk-title {
      color: var(--at-color-font);
      font-size: 21px;
      margin-bottom: 10px;
      cursor: pointer;
    }

    .atk-sub {
      color: var(--at-color-sub);
      font-size: 14px;
      cursor: pointer;
    }
  }

  :deep(.atk-page-actions) {
    @extend .atk-list-btn-actions;
  }
}
</style>
