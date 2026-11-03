<script setup lang="ts">
import type { ArtalkType } from 'artalk'
import { toast } from 'vue-sonner'
import { artalk } from '../global'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const props = defineProps<{
  site: ArtalkType.SiteData
}>()

const emit = defineEmits<{
  (evt: 'close'): void
  (evt: 'update', page: ArtalkType.SiteData): void
  (evt: 'remove', id: number): void
}>()

const { site } = toRefs(props)
const isLoading = ref(false)
const editFieldKey = ref<keyof ArtalkType.SiteData | null>(null)
const editFieldVal = ref('')
const deleteConfirmOpen = ref(false)

const editFieldLabel = computed(() => {
  if (editFieldKey.value === 'name') return t('siteName')
  if (editFieldKey.value === 'urls') return t('siteUrls')
  return ''
})

const origFieldVal = computed(() => {
  if (editFieldKey.value === 'urls') return site.value.urls_raw || ''
  return String(editFieldKey.value ? site.value[editFieldKey.value!] || '' : '')
})

const { t } = useI18n()

function close() {
  emit('close')
}

function openURL(url: string) {
  window.open(url)
}

function rename() {
  editFieldKey.value = 'name'
  editFieldVal.value = site.value.name || ''
}

function editURL() {
  editFieldKey.value = 'urls'
  editFieldVal.value = site.value.urls_raw || ''
}

function del() {
  deleteConfirmOpen.value = false

  const del = async () => {
    isLoading.value = true
    try {
      await artalk!.ctx.getApi().sites.deleteSite(site.value.id)
    } catch (err: any) {
      toast.error(err.message)
      console.error(err)
      return
    } finally {
      isLoading.value = false
    }
    emit('remove', site.value.id)
  }
  del()
}

async function onFieldEditorYes() {
  if (!editFieldKey.value) return
  const val = editFieldVal.value

  if (editFieldVal.value !== origFieldVal.value) {
    isLoading.value = true
    let s: ArtalkType.SiteData
    try {
      let finalVal: string | string[] = val
      if (Array.isArray(site.value[editFieldKey.value]))
        finalVal = val
          .split(',')
          .map((v) => v.trim())
          .filter((v) => !!v)
      s = (
        await artalk!.ctx.getApi().sites.updateSite(site.value.id, {
          ...site.value,
          [editFieldKey.value]: finalVal,
        })
      ).data
    } catch (err: any) {
      toast.error(err.message)
      console.error(err)
      return false
    } finally {
      isLoading.value = false
    }
    emit('update', s)
  }

  editFieldKey.value = null
  return true
}

function onFiledEditorNo() {
  editFieldKey.value = null
}
</script>

<template>
  <Dialog :open="true" @update:open="(v) => !v && close()">
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>
          <span
            class="cursor-pointer hover:text-primary"
            title="Open site"
            @click="!site.first_url || openURL(site.first_url)"
          >
            {{ site.name }}
          </span>
        </DialogTitle>
        <DialogDescription v-if="site.urls.length" class="flex flex-wrap gap-1">
          <button
            v-for="(url, i) in site.urls"
            :key="i"
            type="button"
            class="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            @click="openURL(url)"
          >
            {{ url }}
          </button>
        </DialogDescription>
        <DialogDescription v-else class="sr-only">{{ t('siteUrls') }}</DialogDescription>
      </DialogHeader>

      <!-- Inline field editor (rename / edit URLs) -->
      <form
        v-if="!!editFieldKey"
        class="atk-field-editor"
        @submit.prevent="onFieldEditorYes()"
      >
        <div class="flex flex-col gap-1.5">
          <Label>{{ editFieldLabel }}</Label>
          <Input v-model="editFieldVal" type="text" autocomplete="off" />
        </div>
        <div class="flex gap-2">
          <Button type="submit" size="sm" :disabled="isLoading">{{ t('save') }}</Button>
          <Button type="button" size="sm" variant="outline" :disabled="isLoading" @click="onFiledEditorNo()">
            {{ t('cancel') }}
          </Button>
        </div>
      </form>

      <DialogFooter class="sm:justify-between">
        <Button
          type="button"
          variant="outline"
          :disabled="isLoading"
          @click="editFieldKey === 'urls' ? onFiledEditorNo() : editURL()"
        >
          {{ t('edit') }} URL
        </Button>
        <div class="flex gap-2">
          <Button
            type="button"
            variant="outline"
            :disabled="isLoading"
            @click="editFieldKey === 'name' ? onFiledEditorNo() : rename()"
          >
            {{ t('rename') }}
          </Button>
          <Button type="button" variant="destructive" :disabled="isLoading" @click="deleteConfirmOpen = true">
            {{ t('delete') }}
          </Button>
        </div>
      </DialogFooter>

      <LoadingLayer v-if="isLoading" />
    </DialogContent>
  </Dialog>

  <!-- Delete confirm -->
  <AlertDialog :open="deleteConfirmOpen" @update:open="(v) => (deleteConfirmOpen = v)">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t('delete') }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t('siteDeleteConfirm', { name: site.name }) }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t('cancel') }}</AlertDialogCancel>
        <AlertDialogAction class="bg-destructive text-destructive-foreground hover:bg-destructive/90" @click="del()">
          {{ t('delete') }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<style scoped lang="scss">
.atk-field-editor {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
</style>
