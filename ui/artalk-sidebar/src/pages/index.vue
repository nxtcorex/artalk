<route lang="json">
{
  "meta": {
    "title": "Artalk Sidebar"
  }
}
</route>

<script setup lang="ts">
import { bootParams, isOpenFromSidebar } from '../global'
import { useUserStore } from '../stores/user'
import { initialPage } from '../lib/overview-navigation'

const router = useRouter()

onMounted(() => {
  if (bootParams.view) {
    const splitted = bootParams.view.split('|')
    if (splitted[0]) bootParams.view = splitted[0]
    if (splitted[1]) bootParams.viewParams = JSON.parse(splitted[1])
  }

  router.replace(initialPage(useUserStore().is_admin, isOpenFromSidebar(), bootParams.view))
})
</script>

<template>
  <div></div>
</template>
