<script setup lang="ts">
import type { OptionNode } from '../lib/settings'

defineProps<{
  node: OptionNode
}>()

const hiddenNodes = ['admin_users']
</script>

<template>
  <div
    v-if="!hiddenNodes.includes(node.name)"
    class="pf-grp"
    :class="[`level-${node.level}`, 'expand']"
  >
    <div
      v-if="node.level > 0 && (node.type === 'object' || node.type === 'array')"
      class="pf-head"
    >
      <div class="title">{{ node.title }}</div>
      <div v-if="!!node.subTitle" class="sub-title">{{ node.subTitle }}</div>
    </div>
    <div class="pf-body">
      <!-- Grp -->
      <template v-if="node.items">
        <PreferenceGrp v-for="n in node.items" :key="n.path" :node="n" />
      </template>

      <!-- Item -->
      <PreferenceItem v-else :node="node" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.pf-grp {
  background: var(--at-color-bg);
  margin-bottom: 10px;
  border-radius: 4px;
}

.pf-grp.level-1 {
  & > .pf-head {
    margin-top: 30px;
    margin-bottom: 20px;

    .title {
      user-select: none;
      font-size: 1.4em;
      font-weight: bold;
      padding-left: 12px;

      &::before {
        position: absolute;
        top: 50%;
        transform: translateY(-50%);
        transition: height ease 0.2s;
        left: -10px;
        content: '';
        height: 10px;
        width: 10px;
        background: var(--muted-foreground);
        border-radius: 2px;
      }
    }
  }

  & > .pf-body {
  }

  &.expand > .pf-head .title::before {
    height: 25px;
  }
}

.pf-grp.level-2 {
  & > .pf-head {
    margin-top: 15px;
    margin-bottom: 20px;

    .title {
      position: relative;
      font-weight: bold;
      font-size: 1.2em;
    }
  }
}

.pf-grp.level-1,
.pf-grp.level-2 {
  & > .pf-head > .sub-title {
    padding: 0 10px 0 10px;
    margin-left: 4px;
    margin-top: 15px;
    border-left: 2px solid var(--at-color-border);
  }
}

.pf-grp.level-3 {
  margin-left: 10px;

  & > .pf-head {
    margin-top: 15px;
    margin-bottom: 20px;

    & > .title {
      position: relative;
      font-weight: bold;
      font-size: 0.9em;
    }
  }

  & > .pf-body {
    margin-left: 15px;
  }
}

.pf-head {
  & > .title {
    position: relative;
  }

  & > .sub-title {
    font-size: 14px;
    margin-top: 5px;
  }
}
</style>
