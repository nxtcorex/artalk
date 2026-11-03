<script setup lang="ts">
import settings, {
  formatOptionValue,
  KEYWORD_FILE_SEPARATOR_PATH,
  patchOptionValue,
  type OptionNode,
} from '../lib/settings'
import { isSensitiveConfigPath } from '@/lib/settings-sensitive'

const props = defineProps<{
  node: OptionNode
}>()

const value = ref<any>('')
const sensitiveHidden = ref(true)

const { t } = useI18n()

const separatorSuggestions = ['\\n', '\\r\\n', '\\t', '|']
const isKeywordFileSeparator = computed(() => props.node.path === KEYWORD_FILE_SEPARATOR_PATH)
const separatorListId = computed(() =>
  isKeywordFileSeparator.value ? 'atk-keyword-separator-options' : undefined,
)
const envRef = computed(() => settings.get().getEnvRefByPath(props.node.path))
const disabled = computed(() => envRef.value !== null)

onBeforeMount(() => {
  // initial value
  value.value = formatOptionValue(settings.get().getCustom(props.node.path), props.node)
})

function onChange(event: Event) {
  if (event.currentTarget instanceof HTMLInputElement && !event.currentTarget.checkValidity()) {
    event.currentTarget.reportValidity()
    return
  }

  commit()
}

function commit() {
  const v = patchOptionValue(value.value, props.node)
  settings.get().setCustom(props.node.path, v)
  // console.log('[SET]', props.node.path, v)
}

const envVariableName = computed(() => settings.get().getEnvKeyByPath(props.node.path))
const isSensitive = computed(() => isSensitiveConfigPath(props.node.path))

// boolean options are stored as 'true' / 'false' strings
const valueBool = computed({
  get: () => value.value === true || value.value === 'true',
  set: (v: boolean) => {
    value.value = v ? 'true' : 'false'
  },
})

function toggleSensitiveHidden() {
  sensitiveHidden.value = !sensitiveHidden.value
}
</script>

<template>
  <div class="pf-item">
    <div class="info">
      <div class="title" :title="envVariableName">{{ node.title }}</div>
      <div v-if="node.subTitle" class="sub-title">{{ node.subTitle }}</div>
    </div>

    <div class="value">
      <div v-if="disabled" class="disable-note">
        {{ t('envVarControlHint', { key: envVariableName }) }}
      </div>

      <!-- Array -->
      <template v-if="node.type === 'array'">
        <PreferenceArr :node="node" />
      </template>

      <!-- Dropdown -->
      <template v-else-if="node.selector">
        <Select v-model="value" :disabled="disabled" @update:model-value="commit()">
          <SelectTrigger class="w-[200px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="(item, i) in node.selector" :key="i" :value="item">
              {{ item }}
            </SelectItem>
          </SelectContent>
        </Select>
      </template>

      <!-- Toggle -->
      <template v-else-if="node.type === 'boolean'">
        <Switch
          v-model="valueBool"
          :disabled="disabled"
          @update:model-value="commit()"
        />
      </template>

      <!-- Text -->
      <template v-else>
        <input
          v-model="value"
          :type="!isSensitive || !sensitiveHidden ? 'text' : 'password'"
          :list="separatorListId"
          :required="isKeywordFileSeparator"
          :disabled="disabled"
          @change="onChange"
        />
        <datalist v-if="separatorListId" :id="separatorListId">
          <option v-for="item in separatorSuggestions" :key="item" :value="item" />
        </datalist>
        <div v-if="isSensitive" class="input-suffix">
          <div class="hidden-switch" @click="toggleSensitiveHidden()">
            <i :class="['atk-icon', `atk-icon-eye-${sensitiveHidden ? 'off' : 'on'}`]" />
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped lang="scss">
.pf-item {
  display: flex;
  flex-direction: row;
  margin-bottom: 20px;

  & > .info {
    display: flex;
    justify-content: center;
    flex-direction: column;
    flex: 1;
    padding-right: 20px;

    .title {
    }

    .sub-title {
      font-size: 14px;
      margin-top: 4px;
      color: var(--muted-foreground);
    }
  }

  & > .value {
    position: relative;
    flex: 1;
    display: flex;
    flex-direction: row;
    justify-content: flex-end;
    align-items: center;
    min-height: 35px;

    input:not([type='checkbox']) {
      height: 36px;
      max-width: 260px;
      width: 100%;
      padding: 0 10px;
      border: 1px solid var(--atk-border);
      border-radius: 8px;
      background: transparent;
      font-size: 14px;
      outline: none;
      transition:
        border-color 0.15s ease,
        box-shadow 0.15s ease;

      &:focus {
        border-color: var(--atk-ring);
        box-shadow: 0 0 0 3px color-mix(in oklab, var(--ring) 30%, transparent);
      }

      &:disabled {
        cursor: not-allowed;
        opacity: 0.5;
      }
    }

    &:hover {
      .disable-note {
        opacity: 1;
      }
    }

    .disable-note {
      z-index: 999;
      padding: 3px 8px;
      background: rgba(105, 113, 130, 0.9);
      color: #fff;
      position: absolute;
      top: -30px;
      font-size: 13px;
      left: 0;
      opacity: 0;
      transition: opacity 0.2s;
    }

    .input-suffix {
      margin-left: 5px;
    }

    .hidden-switch {
      cursor: pointer;
      padding-left: 10px;

      .atk-icon {
        &::after {
          background-color: var(--muted-foreground);
        }

        &.atk-icon-eye-on::after {
          mask-image: url('@/assets/icon-eye-on.svg');
        }

        &.atk-icon-eye-off::after {
          mask-image: url('@/assets/icon-eye-off.svg');
        }
      }
    }
  }
}
</style>
