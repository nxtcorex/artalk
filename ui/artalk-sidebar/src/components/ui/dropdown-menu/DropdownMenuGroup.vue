<script setup lang="ts">
import type { DropdownMenuGroupProps } from 'reka-ui'
import { DropdownMenuGroup, useForwardProps } from 'reka-ui'

const props = defineProps<DropdownMenuGroupProps>()

const delegatedProps = useForwardProps(props)
</script>

<template>
  <DropdownMenuGroup v-bind="delegatedProps">
    <slot />
  </DropdownMenuGroup>
</template>
