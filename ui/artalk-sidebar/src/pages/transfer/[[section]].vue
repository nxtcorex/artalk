<route lang="json">
{
  "meta": {
    "title": "Transfer - Artalk Sidebar",
    "adminOnly": true
  }
}
</route>

<script setup lang="ts">
import { Download, FileUp } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { useNavStore } from '../../stores/nav'
import { useUserStore } from '../../stores/user'
import { artalk } from '../../global'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const nav = useNavStore()
const router = useRouter()
const route = useRoute()
const user = useUserStore()
const { t } = useI18n()

const sections = ['import', 'export']

const section = computed(() => {
  const value = (route.params as Record<string, string | string[] | undefined>).section
  const name = Array.isArray(value) ? value[0] : value || ''
  return sections.includes(name) ? name : 'import'
})

watch(
  () => (route.params as Record<string, string | string[] | undefined>).section,
  (value) => {
    const name = Array.isArray(value) ? value[0] : value || ''
    if (name && !sections.includes(name)) router.replace('/transfer/import')
  },
)

const importParams = ref({
  siteName: '',
  siteURL: '',
  payload: '',
})

const isLoading = ref(false)

const uploadApiURL = ref('')
const importTaskApiURL = ref('')

const uploadedFilename = ref('')

const importTaskStarted = ref(false)
const importTaskParams = ref<Record<string, string>>({})

const exportTaskStarted = ref(false)

onMounted(() => {
  nav.updateTabs({}, '')

  const current = (route.params as Record<string, string | string[] | undefined>).section
  if (!sections.includes(String(current || ''))) {
    router.replace('/transfer/import')
  }

  uploadApiURL.value = `${artalk?.ctx.conf.server}/api/v2/transfer/upload`
  importTaskApiURL.value = `${artalk?.ctx.conf.server}/api/v2/transfer/import`
})

function setError(msg: string) {
  toast.error(msg)
}

function fileUploaded(filename: string) {
  uploadedFilename.value = filename
}

function startImportTask() {
  if (!uploadedFilename.value) {
    setError(`Please upload a data file first`)
    return
  }

  const p = importParams.value
  const siteName = p.siteName.trim()
  const siteURL = p.siteURL.trim()
  const payload = p.payload.trim()

  // Prepare request params
  let rData: any = {}
  if (payload) {
    // Validate payload JSON
    try {
      rData = JSON.parse(payload)
    } catch (err) {
      setError(`Payload JSON invalid: ${err}`)
      return
    }

    if (typeof rData !== 'object' || Array.isArray(rData)) {
      setError(`Payload should be an object`)
      return
    }
  }
  if (siteName) rData.target_site_name = siteName
  if (siteURL) rData.target_site_url = siteURL
  rData.json_file = uploadedFilename.value

  // Create an import task
  importTaskParams.value = {
    ...rData,
    token: user.token,
  }
  importTaskStarted.value = true
}

function importTaskDone() {
  importTaskStarted.value = false
}

async function startExportTask() {
  if (exportTaskStarted.value) return
  exportTaskStarted.value = true
  isLoading.value = true
  try {
    const res = await artalk!.ctx.getApi().transfer.exportArtrans()
    downloadFile(`backup-${getYmdHisFilename()}.artrans`, res.data.artrans)
    toast.success(t('updateComplete'))
  } catch (err: any) {
    console.error(err)
    toast.error(`${String(err)}`)
  } finally {
    exportTaskStarted.value = false
    isLoading.value = false
  }
}

function downloadFile(filename: string, text: string) {
  const el = document.createElement('a')
  el.setAttribute('href', `data:text/json;charset=utf-8,${encodeURIComponent(text)}`)
  el.setAttribute('download', filename)
  el.style.display = 'none'
  document.body.appendChild(el)
  el.click()
  document.body.removeChild(el)
}

function getYmdHisFilename() {
  const date = new Date()

  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hours = date.getHours()
  const minutes = date.getMinutes()
  const seconds = date.getSeconds()

  return `${year}${month}${day}-${hours}${padWithZeros(minutes, 2)}${padWithZeros(seconds, 2)}`
}

function padWithZeros(vNumber: number, width: number) {
  let numAsString = vNumber.toString()
  while (numAsString.length < width) {
    numAsString = `0${numAsString}`
  }
  return numAsString
}

const artransferToolHint = computed(() =>
  t('artransferToolHint', { link: '__LINK__' }).replace(
    '__LINK__',
    `<a href="https://artalk.js.org/guide/transfer.html" target="_blank">${t('artransfer')}</a>`,
  ),
)
</script>

<template>
  <LoadingLayer v-if="isLoading" />
  <LogTerminal
    v-if="importTaskStarted"
    :api-url="importTaskApiURL"
    :req-params="importTaskParams"
    @back="importTaskDone()"
  />
  <div v-show="!importTaskStarted" class="transfer-page">
    <Card v-if="section === 'import'" class="transfer-card">
      <CardHeader>
        <CardTitle class="text-base">{{ t('import') }}</CardTitle>
        <!-- eslint-disable-next-line vue/no-v-html -->
        <p class="hint" v-html="artransferToolHint" />
      </CardHeader>
      <CardContent class="space-y-5">
        <div class="field">
          <Label>{{ t('dataFile') }}</Label>
          <FileUploader :api-url="uploadApiURL" @done="fileUploaded">
            <template #done-msg>
              {{ t('uploadReadyToImport') }}
            </template>
          </FileUploader>
        </div>
        <div class="field">
          <Label>{{ t('targetSiteName') }}</Label>
          <Input
            v-model="importParams.siteName"
            type="text"
            name="AtkSiteName"
            :placeholder="t('inputHint')"
            autocomplete="off"
          />
        </div>
        <div class="field">
          <Label>{{ t('targetSiteURL') }}</Label>
          <Input
            v-model="importParams.siteURL"
            type="text"
            name="AtkSiteURL"
            :placeholder="t('inputHint')"
            autocomplete="off"
          />
        </div>
        <div class="field">
          <Label>{{ t('payload') }} ({{ t('optional') }})</Label>
          <Textarea v-model="importParams.payload" name="AtkPayload" class="min-h-[100px]" />
          <span class="desc">
            <a href="https://artalk.js.org/guide/transfer.html" target="_blank">
              {{ t('moreDetails') }}
            </a>
          </span>
        </div>
        <Button name="AtkSubmit" :disabled="isLoading" @click="startImportTask()">
          <FileUp class="size-4" />
          {{ t('import') }}
        </Button>
      </CardContent>
    </Card>

    <Card v-else class="transfer-card">
      <CardHeader>
        <CardTitle class="text-base">{{ t('export') }}</CardTitle>
        <p class="hint">{{ t('exportHint') }}</p>
      </CardHeader>
      <CardContent>
        <Button :disabled="isLoading" @click="startExportTask()">
          <Download class="size-4" />
          {{ t('export') }}
        </Button>
      </CardContent>
    </Card>
  </div>
</template>

<style scoped lang="scss">
.transfer-page {
  padding: 24px;

  @media (max-width: 1023px) {
    padding: 16px;
  }
}

.transfer-card {
  max-width: 640px;
}

.hint {
  margin: 6px 0 0;
  font-size: 14px;
  color: var(--muted-foreground);

  a {
    color: var(--foreground);
    text-decoration: underline;
    text-underline-offset: 2px;
  }
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.desc {
  font-size: 13px;
  color: var(--muted-foreground);

  a {
    color: var(--foreground);
    text-decoration: underline;
    text-underline-offset: 2px;
  }
}

button[name='AtkSubmit'] {
  align-self: flex-start;
  min-width: 8em;
}
</style>
