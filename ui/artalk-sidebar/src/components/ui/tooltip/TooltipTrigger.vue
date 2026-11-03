<script setup lang="ts">
import type { TooltipTriggerProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { TooltipTrigger } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps<TooltipTriggerProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <TooltipTrigger v-bind="props" :class="cn(props.class)">
    <slot />
  </TooltipTrigger>
</template>
