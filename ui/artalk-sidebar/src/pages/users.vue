<script setup lang="ts">
import type { ArtalkType } from 'artalk'
import { storeToRefs } from 'pinia'
import { toast } from 'vue-sonner'
import { useNavStore } from '../stores/nav'
import { artalk } from '../global'
import Pagination from '../components/Pagination.vue'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

const nav = useNavStore()
const { curtTab } = storeToRefs(nav)
const users = ref<ArtalkType.UserDataForAdmin[]>([])
const { t } = useI18n()

const pageSize = ref(30)
const pageTotal = ref(0)
const pagination = ref<InstanceType<typeof Pagination>>()
const curtType = ref<'all' | 'admin' | 'in_conf' | undefined>('all')

const userEditorState = reactive({
  show: false,
  user: undefined as ArtalkType.UserDataForAdmin | undefined,
})
const search = ref('')

watch(curtTab, (tab) => {
  if (tab === 'create') {
    createUser()
  } else {
    curtType.value = tab as any
    fetchUsers(0)
    closeUserEditor()
  }
})

onMounted(() => {
  fetchUsers(0)

  nav.updateTabs(
    {
      all: 'all',
      admin: 'admin',
      create: 'create',
    },
    'all',
  )

  // Users search
  nav.enableSearch(
    (value: string) => {
      search.value = value
      fetchUsers(0)
    },
    () => {
      if (search.value === '') return
      search.value = ''
      fetchUsers(0)
    },
  )
})

watch(
  () => userEditorState.show,
  () => nav.scrollPageToTop(),
)

function fetchUsers(offset: number) {
  if (offset === 0) pagination.value?.reset()
  nav.setPageLoading(true)
  artalk?.ctx
    .getApi()
    .users.getUsers(curtType.value, {
      offset,
      limit: pageSize.value,
      search: search.value,
    })
    .then((res) => {
      pageTotal.value = res.data.count
      users.value = res.data.users
      nav.scrollPageToTop()
    })
    .finally(() => {
      nav.setPageLoading(false)
    })
}

function onChangePage(offset: number) {
  fetchUsers(offset)
}

function editUser(user: ArtalkType.UserDataForAdmin) {
  if (user.is_in_conf) {
    toast.warning(t('userInConfCannotEditHint'))
    return
  }

  userEditorState.show = true
  userEditorState.user = user
}

function createUser() {
  userEditorState.show = true
  userEditorState.user = undefined
}

function updateUser(user: ArtalkType.UserDataForAdmin) {
  const index = users.value.findIndex((u) => u.id === user.id)
  if (index != -1) {
    // Edit user
    const orgUser = users.value[index]
    Object.keys(user).forEach((key) => {
      ;(orgUser as any)[key] = (user as any)[key]
    })
  } else {
    // Create user
    fetchUsers(0)
  }

  closeUserEditor()
}

function closeUserEditor() {
  userEditorState.show = false
  userEditorState.user = undefined
}

const delConfirm = ref<{ open: boolean; user: ArtalkType.UserDataForAdmin | null }>({
  open: false,
  user: null,
})

function delUser(user: ArtalkType.UserDataForAdmin) {
  delConfirm.value = { open: true, user }
}

function onDelConfirm(open: boolean) {
  if (!open) delConfirm.value = { open: false, user: null }
}

function confirmedDelUser() {
  const user = delConfirm.value.user
  delConfirm.value = { open: false, user: null }
  if (!user) return

  artalk!.ctx
    .getApi()
    .users.deleteUser(user.id)
    .then(() => {
      const index = users.value.findIndex((u) => u.id === user.id)
      users.value.splice(index, 1)

      if (user.is_in_conf) {
        toast.info(t('userDeleteManuallyHint'))
      }
    })
    .catch((e: ArtalkType.FetchError) => {
      toast.error(e.message)
    })
}
</script>

<template>
  <div class="user-list-wrap">
    <div class="user-list">
      <div v-for="user in users" :key="user.id" class="user-item">
        <div class="user-main">
          <div class="title">
            {{ user.name }}
            <span class="badge-grp">
              <span
                v-if="user.badge_name"
                class="badge"
                :style="{ backgroundColor: user.badge_color }"
              >
                {{ user.badge_name }}
              </span>
              <span v-else-if="user.is_admin" class="badge admin" :title="t('userAdminHint')">
                {{ t('admin') }}
              </span>
              <span v-if="user.is_in_conf" class="badge in-conf" :title="t('userInConfHint')">
                {{ t('config') }}
              </span>
            </span>
          </div>
          <div class="sub">{{ user.email }}</div>
        </div>
        <div class="user-actions">
          <span @click="editUser(user)">{{ t('edit') }}</span>
          <span>{{ t('comment') }} ({{ user.comment_count }})</span>
          <span @click="delUser(user)">{{ t('delete') }}</span>
        </div>
      </div>
    </div>
    <Pagination
      ref="pagination"
      :page-size="pageSize"
      :total="pageTotal"
      :disabled="nav.isPageLoading"
      @change="onChangePage"
    />

    <UserEditor
      v-if="userEditorState.show"
      :user="userEditorState.user"
      @update="updateUser"
      @close="closeUserEditor"
    />

    <!-- Delete confirm -->
    <AlertDialog :open="delConfirm.open" @update:open="onDelConfirm">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ t('delete') }}</AlertDialogTitle>
          <AlertDialogDescription>
            <template v-if="delConfirm.user">
              {{ t('userDeleteConfirm', { name: delConfirm.user.name, email: delConfirm.user.email }) }}
            </template>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{{ t('cancel') }}</AlertDialogCancel>
          <AlertDialogAction
            class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            @click="confirmedDelUser()"
          >
            {{ t('delete') }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>

<style scoped lang="scss">
.user-list-wrap {
  .user-list {
    .user-item {
      padding: 15px 30px;
      transition: background-color 0.15s ease;

      &:hover {
        background: color-mix(in srgb, var(--atk-accent) 45%, transparent);
      }

      &:not(:last-child) {
        border-bottom: 1px solid var(--at-color-border);
      }

      .user-main {
        .title {
          display: flex;
          flex-direction: row;
          align-items: center;
          color: var(--at-color-deep);
          font-size: 20px;

          .badge-grp {
            display: flex;
            margin-left: 5px;
          }

          .badge {
            display: inline-flex;
            align-items: center;
            margin-left: 6px;
            height: 20px;
            padding: 0 6px;
            border: 1px solid transparent;
            border-radius: 6px;
            font-size: 12px;
            font-weight: 500;

            &.admin {
              background: var(--primary);
              color: var(--primary-foreground);
            }

            &.in-conf {
              background: var(--secondary);
              color: var(--secondary-foreground);
              border-color: var(--border);
            }
          }
        }

        .sub {
          color: var(--at-color-sub);
          font-size: 15px;
          margin-top: 5px;
        }
      }

      .user-actions {
        margin-top: 10px;

        & > span {
          cursor: pointer;
          color: var(--at-color-meta);
          font-size: 13px;

          &:not(:last-child) {
            margin-right: 16px;
          }

          &:hover {
            color: var(--atk-primary);
          }
        }
      }
    }
  }
}
</style>
