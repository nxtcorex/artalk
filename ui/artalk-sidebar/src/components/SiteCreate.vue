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
import { Input } from '@/components/ui/input'

const { t } = useI18n()

const props = defineProps<{
  initVal?: { name: string; urls: string }
}>()

const emit = defineEmits<{
  (evt: 'close'): void
  (evt: 'done', siteNew: ArtalkType.SiteData): void
}>()

const isLoading = ref(false)
const site = ref<{ name: string; urls: string }>({
  name: '',
  urls: '',
})

onMounted(() => {
  site.value.name = props.initVal?.name || ''
  site.value.urls = props.initVal?.urls || ''
})

async function submit() {
  const siteName = site.value.name.trim()
  const siteUrls = site.value.urls
    .trim()
    .split(',')
    .map((v) => v.trim())
    .filter((v) => !!v)

  if (siteName === '') {
    toast.warning(t('siteNameInputHint'))
    return
  }

  isLoading.value = true
  let s: ArtalkType.SiteData
  try {
    s = (
      await artalk!.ctx.getApi().sites.createSite({
        name: siteName,
        urls: siteUrls,
      })
    ).data
  } catch (err: any) {
    toast.error(err.message)
    console.error(err)
    return
  } finally {
    isLoading.value = false
  }

  emit('done', s)
}

function close() {
  emit('close')
}
</script>

<template>
  <Dialog :open="true" @update:open="(v) => !v && close()">
    <DialogContent class="max-w-md">
      <DialogHeader>
        <DialogTitle>{{ t('createSite') }}</DialogTitle>
        <DialogDescription>{{ t('createSiteHint') }}</DialogDescription>
      </DialogHeader>
      <form class="atk-site-add-form" @submit.prevent="submit()">
        <Input
          v-model="site.name"
          type="text"
          name="AtkSiteName"
          :placeholder="t('siteName')"
          autocomplete="off"
        />
        <Input
          v-model="site.urls"
          type="text"
          name="AtkSiteUrls"
          :placeholder="`${t('siteUrls')} (${t('multiSepHint')})`"
          autocomplete="off"
        />
        <DialogFooter>
          <Button type="button" variant="outline" :disabled="isLoading" @click="close()">
            {{ t('cancel') }}
          </Button>
          <Button type="submit" name="AtkSubmit" :disabled="isLoading">
            {{ t('add') }}
          </Button>
        </DialogFooter>
      </form>
      <LoadingLayer v-if="isLoading" />
    </DialogContent>
  </Dialog>
</template>

<style scoped lang="scss">
.atk-site-add-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
