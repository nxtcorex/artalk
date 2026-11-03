<script setup lang="ts">
import { Skeleton } from '@/components/ui/skeleton'

const props = defineProps<{
  transparentBg?: boolean
  timeout?: number
}>()

const showSkeleton = ref(false)
let timer: number | undefined

onMounted(() => {
  // skeleton delay to prevent flash
  // (no need to show anything if loading is fast)
  timer = window.setTimeout(() => {
    showSkeleton.value = true
    timer = undefined
  }, props.timeout || 700)
})

onUnmounted(() => {
  window.clearTimeout(timer)
})
</script>

<template>
  <div
    class="atk-loading atk-fade-in"
    :style="{ background: props.transparentBg ? 'transparent' : undefined }"
  >
    <div v-if="showSkeleton" class="atk-skeleton-group">
      <Skeleton class="h-5 w-[45%]" />
      <Skeleton class="h-4 w-[80%]" />
      <Skeleton class="h-4 w-[70%]" />
      <Skeleton class="h-4 w-[76%]" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.atk-skeleton-group {
  position: absolute;
  top: 15%;
  left: 0;
  right: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 0 25px;
}
</style>
