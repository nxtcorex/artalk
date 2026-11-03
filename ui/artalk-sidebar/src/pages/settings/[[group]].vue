<route lang="json">
{
  "meta": {
    "title": "Settings - Artalk Sidebar",
    "adminOnly": true
  }
}
</route>

<script setup lang="ts">
import { toast } from 'vue-sonner'
import { useNavStore } from '../../stores/nav'
import { artalk } from '../../global'
import settings from '../../lib/settings'
import {
  basicSettingsNodes,
  basicSettingsPaths,
  loadSettingsTree,
  settingsGroups,
  settingsNavGroups,
  settingsTree,
} from '../../lib/settings-tree'
import LoadingLayer from '../../components/LoadingLayer.vue'
import { Button } from '@/components/ui/button'

const nav = useNavStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const isLoading = ref(false)

const groupParam = computed(() => {
  const value = (route.params as Record<string, string | string[] | undefined>).group
  return Array.isArray(value) ? value[0] : value || ''
})

const activeNode = computed(() =>
  settingsGroups.value.find((node) => node.path === groupParam.value),
)

const renderNodes = computed(() => {
  const nodes =
    groupParam.value === 'basic'
      ? basicSettingsNodes.value
      : activeNode.value
        ? [activeNode.value]
        : []
  return nodes.flatMap((node) => (node.items ? node.items : [node]))
})

watch(groupParam, (value) => {
  if (basicSettingsPaths.includes(value)) {
    router.replace('/settings/basic')
    return
  }
  if (value === 'basic') return
  if (value && settingsGroups.value.some((node) => node.path === value)) return
  router.replace(`/settings/${settingsNavGroups.value[0]?.key ?? 'basic'}`)
})

onMounted(() => {
  nav.updateTabs({}, '')

  loadSettingsTree().then(() => {
    if (basicSettingsPaths.includes(groupParam.value)) {
      router.replace('/settings/basic')
      return
    }
    const valid =
      groupParam.value === 'basic' ||
      settingsGroups.value.some((node) => node.path === groupParam.value)
    if (!valid) router.replace(`/settings/${settingsNavGroups.value[0]?.key ?? 'basic'}`)
  })
})

function save() {
  let yamlStr: string
  try {
    yamlStr = settings.get().getCustoms().value?.toString() || ''
  } catch (err) {
    toast.error('YAML export error: ' + err)
    console.error(err)
    return
  }

  if (!yamlStr) {
    toast.error('YAML export error: data is empty')
    return
  }

  if (isLoading.value) return
  isLoading.value = true
  artalk!.ctx
    .getApi()
    .settings.applySettings({
      yaml: yamlStr,
    })
    .then(() => {
      toast.success(t('settingSaved'))
    })
    .catch((err) => {
      console.error(err)
      toast.error(t('settingSaveFailed') + ': ' + err)
    })
    .finally(() => {
      isLoading.value = false
    })
}
</script>

<template>
  <div class="settings">
    <div v-if="settingsTree && renderNodes.length" class="settings-content">
      <PreferenceGrp v-for="node in renderNodes" :key="node.path" :node="node" />
      <div class="notice">{{ t('settingNotice') }}</div>
    </div>

    <div class="act-bar">
      <div class="atk-sidebar-container">
        <div class="status-text"></div>
        <Button class="save-btn" :disabled="isLoading" @click="save()">
          <i class="atk-icon atk-icon-yes" />
          {{ t('apply') }}
        </Button>
      </div>
      <LoadingLayer v-if="isLoading" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.settings-content {
  padding: 20px 24px 80px;

  @media (max-width: 1023px) {
    padding: 16px 16px 80px;
  }
}

.notice {
  font-size: 13px;
  background: var(--muted);
  color: var(--muted-foreground);
  border-radius: 6px;
  text-align: center;
  padding: 8px 10px;
  margin-top: 10px;
  margin-bottom: 20px;
}

.act-bar {
  z-index: 40;
  position: fixed;
  left: var(--atk-sidebar-width);
  right: 8px;
  bottom: 8px;
  height: 56px;
  padding: 0 20px;
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: 0 0 calc(var(--radius) + 4px) calc(var(--radius) + 4px);
  transition: left 0.2s ease;

  @media (max-width: 1023px) {
    left: 0;
    right: 0;
    bottom: 0;
    border-right: 0;
    border-bottom: 0;
    border-left: 0;
    border-radius: 0;
  }

  .atk-sidebar-container {
    display: flex;
    height: 100%;
    align-items: center;
    justify-content: space-between;
    flex-direction: row;
  }

  .status-text {
    padding: 0 5px;
    flex: 1;
  }

  .save-btn {
    font-size: 14px;

    i {
      margin-right: 8px;

      &::after {
        background-color: var(--primary-foreground);
      }
    }
  }
}
</style>
