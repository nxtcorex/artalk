<script setup lang="ts">
import type { ArtalkType } from 'artalk'
import { toast } from 'vue-sonner'
import { artalk } from '../global'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

const { t } = useI18n()

interface IUserEditData {
  id: number
  name: string
  email: string
  link: string
  password?: string
  badge_name: string
  badge_color: string
  is_admin: boolean
  receive_email: boolean
}

const props = defineProps<{
  user?: IUserEditData
}>()

const emit = defineEmits<{
  (evt: 'close'): void
  (evt: 'update', user: ArtalkType.UserDataForAdmin): void
}>()

const editUser = ref<IUserEditData>()

const isLoading = ref(false)
const isCreateMode = ref(false)

const showFullDetails = ref(false)

// reka-ui Select only accepts string values
const isAdminStr = computed({
  get: () => (editUser.value?.is_admin ? 'true' : 'false'),
  set: (v: string) => {
    if (editUser.value) editUser.value.is_admin = v === 'true'
  },
})

onBeforeMount(() => {
  if (!props.user) {
    isCreateMode.value = true
    editUser.value = {
      id: 0,
      name: '',
      email: '',
      link: '',
      password: '',
      badge_name: '',
      badge_color: '',
      is_admin: false,
      receive_email: true,
    }
  } else {
    editUser.value = {
      ...props.user,
      password: '',
    }
  }
})

function close() {
  emit('close')
}

function submit() {
  isLoading.value = true

  // The default badge color
  if (editUser.value!.badge_name !== '' && editUser.value!.badge_color === '') {
    editUser.value!.badge_color = '#52525b'
  }

  if (isCreateMode.value) {
    artalk!.ctx
      .getApi()
      .users.createUser({
        ...editUser.value!,
      })
      .then((res) => {
        toast.success(t('settingSaved'))
        emit('update', res.data)
      })
      .catch((e: ArtalkType.FetchError) => {
        toast.error(e.message)
      })
      .finally(() => {
        isLoading.value = false
      })
  } else {
    const user = editUser.value!
    artalk!.ctx
      .getApi()
      .users.updateUser(user.id, {
        ...editUser.value!,
      })
      .then((res) => {
        toast.success(t('settingSaved'))
        emit('update', res.data)
      })
      .catch((e: ArtalkType.FetchError) => {
        toast.error(e.message)
      })
      .finally(() => {
        isLoading.value = false
      })
  }
}
</script>

<template>
  <Dialog :open="true" @update:open="(v) => !v && close()">
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ isCreateMode ? t('userCreate') : t('userEdit') }}</DialogTitle>
        <DialogDescription v-if="!isCreateMode" class="user-log">
          <span class="block">
            <span class="log-label">{{ t('comments') }}</span>
            {{ (editUser as any).comment_count }}
          </span>
          <span class="block">
            <span class="log-label">{{ t('last') }} IP</span>
            {{ (editUser as any).last_ip || '-' }}
          </span>
          <span class="block">
            <span class="log-label">{{ t('last') }} UA</span>
            <template v-if="showFullDetails || !(editUser as any).last_ua">
              {{ (editUser as any).last_ua || '-' }}
            </template>
            <template v-else>
              <span class="cursor-pointer text-primary" @click="showFullDetails = true">
                {{ t('Show') }}
              </span>
            </template>
          </span>
        </DialogDescription>
        <DialogDescription v-else class="sr-only">
          {{ t('userCreate') }}
        </DialogDescription>
      </DialogHeader>
      <form v-if="editUser" class="user-edit-form" @submit.prevent="submit()">
        <div class="field">
          <Label class="required">{{ t('username') }}</Label>
          <Input v-model="editUser.name" type="text" autocomplete="off" />
        </div>
        <div class="field">
          <Label class="required">{{ t('email') }}</Label>
          <Input v-model="editUser.email" type="text" autocomplete="off" />
        </div>
        <div class="field">
          <Label>{{ t('link') }}</Label>
          <Input v-model="editUser.link" type="text" autocomplete="off" />
        </div>
        <div class="field">
          <Label>{{ t('badgeText') }}</Label>
          <Input v-model="editUser.badge_name" type="text" autocomplete="off" />
        </div>
        <div class="field">
          <Label>{{ t('badgeColor') }} (Color Hex)</Label>
          <Input v-model="editUser.badge_color" type="text" autocomplete="off" />
        </div>
        <div class="field">
          <Label class="required">{{ t('role') }}</Label>
          <Select v-model="isAdminStr">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="false">{{ t('normal') }}</SelectItem>
              <SelectItem value="true">{{ t('admin') }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div v-if="editUser.is_admin" class="field">
          <Label class="required">{{ t('password') }}</Label>
          <Input
            v-model="editUser.password"
            type="text"
            :placeholder="isCreateMode ? '' : `(${t('passwordEmptyHint')})`"
            autocomplete="off"
          />
        </div>
        <div class="field field-inline">
          <Switch v-model="editUser.receive_email" />
          <Label class="required">{{ t('emailNotify') }}</Label>
        </div>
        <Button type="submit" :disabled="isLoading">{{ t('save') }}</Button>
      </form>
      <LoadingLayer v-if="isLoading" />
    </DialogContent>
  </Dialog>
</template>

<style scoped lang="scss">
.user-log {
  padding: 12px 16px;
  border-radius: var(--atk-radius, 8px);
  font-size: 13px;
  background: var(--atk-muted);
  color: var(--atk-muted-foreground);
  line-height: 1.8;

  .log-label {
    display: inline-block;
    width: 5em;
  }
}

.user-edit-form {
  display: flex;
  flex-direction: column;
  gap: 14px;

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;

    &.field-inline {
      flex-direction: row;
      align-items: center;
      gap: 10px;
    }
  }

  .required::after {
    content: '*';
    color: var(--at-color-red);
    margin-left: 4px;
    vertical-align: top;
  }

  button[type='submit'] {
    margin-top: 8px;
    align-self: flex-start;
    min-width: 7em;
  }
}
</style>
