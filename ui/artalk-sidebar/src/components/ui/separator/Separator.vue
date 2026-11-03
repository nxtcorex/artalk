<script setup lang="ts">
import type { SeparatorProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { Separator } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<
    SeparatorProps & { class?: HTMLAttributes['class'], label?: string }
  >(),
  {
    orientation: 'horizontal',
    decorative: true,
  },
)
</script>

<template>
  <div
    :class="
      cn(
        'flex shrink-0 select-none items-center gap-2 text-sm text-muted-foreground',
        orientation === 'vertical' ? 'flex-col' : 'flex-row',
        props.class,
      )
    "
  >
    <Separator
      :orientation="orientation"
      :decorative="decorative"
      class="bg-border shrink-0"
      :class="
        orientation === 'vertical' ? 'h-full w-px' : 'h-px w-full'
      "
    />
    <div
      v-if="label"
      class="text-sm leading-none text-muted-foreground"
    >
      {{ label }}
    </div>
  </div>
</template>
