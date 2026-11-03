<script setup lang="ts">
import { toast } from 'vue-sonner'
import { useUserStore } from '../stores/user'

const { t } = useI18n()
const user = useUserStore()

const props = defineProps<{
  apiUrl: string
}>()

const emit = defineEmits<{
  /** Upload done event */
  (evt: 'done', filename: string): void
}>()

const { apiUrl } = toRefs(props)

let xhr: XMLHttpRequest | null = null
const fileInputEl = ref<HTMLInputElement | null>(null)
const remoteFilename = ref('')
const isUploading = ref(false)
const progress = ref(0)
const isDone = ref(false)

function reset() {
  remoteFilename.value = ''
  isUploading.value = false
  progress.value = 0
  isDone.value = false
  if (fileInputEl.value) fileInputEl.value.value = ''
}

async function startUploadFile(file: File) {
  remoteFilename.value = ''

  xhr = new XMLHttpRequest()

  // Progress bar
  xhr.upload.addEventListener('progress', (evt) => {
    if (evt.loaded === evt.total) {
      // Upload done
      progress.value = 100
      return
    }

    const fileSize = file.size
    if (evt.loaded <= fileSize) {
      // Uploading
      progress.value = Math.round((evt.loaded / fileSize) * 100)
    }
  })

  // Create form data for upload
  const formData = new FormData()
  formData.append('file', file)
  formData.append('token', user.token)

  // Start upload
  xhr.open('post', apiUrl.value)
  xhr.send(formData)

  // Update finished event
  xhr.onload = () => {
    const setErr = (msg: string): void => {
      reset()
      isUploading.value = false
      toast.error(`File upload failed: ${msg}`)
    }

    if (!xhr) {
      setErr('xhr instance is null')
      return
    }

    const ok = xhr.status >= 200 && xhr.status <= 299
    if (!ok) {
      setErr(`Response HTTP Code: ${xhr.status}, Body: ${xhr.response}`)
      return
    }

    let json: any
    try {
      json = JSON.parse(xhr.response)
    } catch (err) {
      console.error(err)
      setErr(`JSON parse error: ${err}`)
      return
    }

    if (!json.filename) {
      setErr(`Response filename is empty: ${xhr.response}`)
      return
    }

    isDone.value = true
    remoteFilename.value = json.filename
    isUploading.value = false
    emit('done', remoteFilename.value)
  }
}

function onFileInputChange() {
  const files = fileInputEl.value?.files
  if (!files || files.length === 0) return

  isUploading.value = true
  setTimeout(async () => {
    await startUploadFile(files[0])
    isUploading.value = false
  }, 80)
}

function abortUpload() {
  xhr?.abort()
  reset()
  isUploading.value = false
}
</script>

<template>
  <div class="atk-file-upload-group">
    <div v-show="!isUploading" class="atk-file-input-wrap atk-fade-in">
      <input
        ref="fileInputEl"
        type="file"
        name="AtkDataFile"
        accept=".artrans"
        @change="onFileInputChange()"
      />
      <div class="atk-desc">
        <slot v-if="!isDone" name="tip"></slot>
        <slot v-if="isDone" name="done-msg"></slot>
      </div>
    </div>
    <div v-show="isUploading" class="atk-uploading-wrap atk-fade-in">
      <div class="atk-progress">
        <div class="atk-bar" :style="{ width: `${progress}%` }"></div>
      </div>
      <div class="atk-status">
        {{ t('uploading') }}
        <span class="atk-curt">{{ progress }}%</span>
        ...
        <span class="atk-abort" @click="abortUpload()">{{ t('cancel') }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.atk-file-upload-group {
  width: 100%;
}

.atk-file-input-wrap {
  input[type='file'] {
    display: block;
    width: 100%;
    padding: 10px 12px;
    border: 1px dashed var(--border);
    border-radius: 8px;
    background: transparent;
    color: var(--muted-foreground);
    font-size: 14px;
    cursor: pointer;
    transition: border-color 0.15s ease;

    &:hover {
      border-color: var(--ring);
    }
  }

  input[type='file']::file-selector-button {
    margin-right: 12px;
    padding: 6px 12px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--secondary);
    color: var(--foreground);
    font: inherit;
    font-size: 13px;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  input[type='file']::file-selector-button:hover {
    background: var(--muted);
  }

  .atk-desc {
    margin-top: 8px;
    font-size: 13px;
    color: var(--muted-foreground);

    a {
      color: var(--foreground);
      text-decoration: underline;
      text-underline-offset: 2px;
    }
  }
}

.atk-uploading-wrap {
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 8px;

  .atk-progress {
    height: 6px;
    overflow: hidden;
    border-radius: 999px;
    background: var(--muted);
  }

  .atk-bar {
    height: 100%;
    border-radius: 999px;
    background: var(--primary);
    transition: width 0.3s ease;
  }

  .atk-status {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
    font-size: 13px;
    color: var(--muted-foreground);

    .atk-curt {
      color: var(--foreground);
      font-weight: 500;
    }

    .atk-abort {
      margin-left: auto;
      color: var(--foreground);
      cursor: pointer;

      &:hover {
        text-decoration: underline;
      }
    }
  }
}
</style>
