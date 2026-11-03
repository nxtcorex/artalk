<script setup lang="ts">
import type { TooltipProviderProps } from 'reka-ui'
import { TooltipProvider } from 'reka-ui'

const props = defineProps<TooltipProviderProps>()
</script>

<template>
  <TooltipProvider :delay-duration="props.delayDuration ?? 200" :skip-delay-duration="props.skipDelayDuration">
    <slot />
  </TooltipProvider>
</template>
