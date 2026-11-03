import { computed, shallowRef } from 'vue'
import YAML from 'yaml'
import { artalk } from '../global'
import settings, { type OptionNode } from './settings'

export const settingsTree = shallowRef<OptionNode>()
let pending: Promise<void> | null = null

export function loadSettingsTree(): Promise<void> {
  if (settingsTree.value) return Promise.resolve()
  if (!pending) {
    pending = Promise.all([
      artalk!.ctx.getApi().settings.getSettingsTemplate(''),
      artalk!.ctx.getApi().settings.getSettings(),
    ]).then(([template, custom]) => {
      const yamlObj = YAML.parseDocument(template.data.yaml)
      settingsTree.value = settings.init(yamlObj).getTree()
      settings.get().setCustoms(custom.data.yaml)
      settings.get().setEnvs(custom.data.envs)
    })
  }
  return pending
}

const hiddenGroups = ['admin_users']

export const basicSettingsPaths = [
  'host',
  'port',
  'app_key',
  'debug',
  'locale',
  'timezone',
  'site_default',
  'site_url',
  'login_timeout',
]

export type SettingsNavEntry = {
  key: string
  label: string
  labelKey?: string
  paths: string[]
}

export const settingsGroups = computed(() =>
  (settingsTree.value?.items ?? []).filter((node) => !hiddenGroups.includes(node.name)),
)

export const basicSettingsNodes = computed(() =>
  settingsGroups.value.filter((node) => basicSettingsPaths.includes(node.path)),
)

export const settingsNavGroups = computed<SettingsNavEntry[]>(() => {
  const basic = basicSettingsNodes.value
  const entry: SettingsNavEntry[] = []
  if (basic.length) {
    entry.push({
      key: 'basic',
      label: '',
      labelKey: 'generalSettings',
      paths: basic.map((node) => node.path),
    })
  }
  settingsGroups.value
    .filter((node) => !basicSettingsPaths.includes(node.path))
    .forEach((node) => entry.push({ key: node.path, label: node.title, paths: [node.path] }))
  return entry
})
