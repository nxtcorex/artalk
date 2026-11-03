<route lang="json">
{
  "meta": {
    "title": "Login - Artalk Sidebar"
  }
}
</route>

<script setup lang="ts">
import type { ArtalkType } from 'artalk'
import { getArtalk } from '../global'
import { useUserStore } from '../stores/user'

const router = useRouter()
const { t } = useI18n()

let userForm = ref({
  email: '',
  password: '',
})
let version = ref('')
let buildHash = ref('')
let loginErr = ref('')
let userSelector = ref<string[] | null>(null)

const captcha = ref<{ iframe: boolean; imgData: string } | null>(null)
const captchaValue = ref('')
const captchaErr = ref('')
const captchaBusy = ref(false)
const captchaInputRef = ref<{ $el?: HTMLElement } | null>(null)
const lastLoginName = ref('')
let captchaPollTimer: number | undefined

const captchaIframeURL = computed(() => {
  if (!captcha.value?.iframe) return ''
  const server = getArtalk()?.getConf().server || ''
  const theme = document.documentElement.classList.contains('atk-dark-mode') ? 'dark' : 'light'
  return `${server}/api/v2/captcha/?t=${Date.now()}&theme=${theme}`
})

onMounted(() => {
  getArtalk()!
    .ctx.getApi()
    .version.getVersion()
    .then((res) => {
      version.value = res.data.version
      buildHash.value = res.data.commit_hash
    })
})

function onFocus() {
  loginErr.value = ''
}

function login(username?: string) {
  loginErr.value = ''

  const artalk = getArtalk()
  if (!artalk) throw new Error('Artalk instance not initialized')

  if (captcha.value) {
    if (!captcha.value.iframe) verifyCaptcha(username)
    return
  }

  lastLoginName.value = username || ''

  artalk.ctx
    .getApi()
    .user.login({
      name: username || '',
      email: userForm.value.email,
      password: userForm.value.password,
    })
    .then((res) => {
      artalk.ctx.getUser().update({
        ...res.data.user,
        token: res.data.token,
      })
      useUserStore().sync()
      router.replace('/')
    })
    .catch((e: ArtalkType.FetchError) => {
      const data = (e.data || {}) as Record<string, any>
      if (data.need_name_select) {
        userSelector.value = data.need_name_select
      } else if (data.need_captcha) {
        showCaptcha(data)
      } else {
        loginErr.value = e.message || t('loginFailure')
      }
    })
}

function showCaptcha(data: Record<string, any>) {
  captchaErr.value = ''
  captchaValue.value = ''
  captcha.value = { iframe: !!data.iframe, imgData: data.img_data || '' }
  if (captcha.value.iframe) startCaptchaStatusPolling()
  else focusCaptchaInput()
}

function verifyCaptcha(username?: string) {
  const artalk = getArtalk()
  if (!artalk) return

  const value = captchaValue.value.trim()
  if (!value) {
    focusCaptchaInput()
    return
  }

  captchaBusy.value = true
  captchaErr.value = ''

  artalk.ctx
    .getApi()
    .captcha.verifyCaptcha({ value })
    .then(() => {
      captcha.value = null
      captchaValue.value = ''
      login(username ?? lastLoginName.value)
    })
    .catch((e: ArtalkType.FetchError) => {
      const data = (e.data || {}) as Record<string, any>
      captchaErr.value = e.message || t('loginFailure')
      if (data.img_data) captcha.value = { iframe: false, imgData: data.img_data }
      captchaValue.value = ''
      focusCaptchaInput()
    })
    .finally(() => {
      captchaBusy.value = false
    })
}

function refreshCaptcha() {
  const artalk = getArtalk()
  if (!artalk) return

  artalk.ctx
    .getApi()
    .captcha.getCaptcha()
    .then((res) => {
      captcha.value = { iframe: false, imgData: res.data.img_data }
      captchaValue.value = ''
    })
    .catch(() => {})
}

function startCaptchaStatusPolling() {
  stopCaptchaStatusPolling()
  captchaPollTimer = window.setInterval(async () => {
    if (!captcha.value?.iframe) {
      stopCaptchaStatusPolling()
      return
    }
    const artalk = getArtalk()
    if (!artalk) return
    try {
      const res = await artalk.ctx.getApi().captcha.getCaptchaStatus()
      if (res.data.is_pass) {
        stopCaptchaStatusPolling()
        captcha.value = null
        login(lastLoginName.value)
      }
    } catch {
      return
    }
  }, 1000)
}

function stopCaptchaStatusPolling() {
  if (captchaPollTimer === undefined) return
  window.clearInterval(captchaPollTimer)
  captchaPollTimer = undefined
}

function focusCaptchaInput() {
  if (window.innerWidth < 768) return
  nextTick(() => captchaInputRef.value?.$el?.focus())
}

onUnmounted(stopCaptchaStatusPolling)

function selectUser(username: string) {
  userSelector.value = null
  login(username)
}

const versionInfo = computed(() => {
  return `v${version.value + (buildHash.value ? ' / ' + buildHash.value : '')}`
})
</script>

<template>
  <div class="login-dialog">
    <a href="https://artalk.js.org/" target="_blank">
      <img class="logo" src="../assets/favicon.png" alt="logo" draggable="false" />
    </a>
    <Card class="login-card">
      <CardHeader class="text-center">
        <CardTitle class="text-xl">{{ t('login') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="login-form" @submit.prevent="login()">
          <Input
            v-model="userForm.email"
            type="text"
            :placeholder="t('email')"
            @focus="onFocus"
          />
          <Input
            v-model="userForm.password"
            type="password"
            :placeholder="t('password')"
            @focus="onFocus"
          />
          <div v-if="captcha" class="captcha-panel">
            <iframe
              v-if="captcha.iframe"
              class="captcha-iframe"
              :src="captchaIframeURL"
              :title="t('captchaTitle')"
              referrerpolicy="strict-origin-when-cross-origin"
            ></iframe>
            <div v-else class="captcha-row">
              <Input
                ref="captchaInputRef"
                v-model="captchaValue"
                class="captcha-input"
                autocomplete="off"
                :placeholder="t('captchaCheck')"
                @focus="(captchaErr = '')"
              />
              <img
                v-if="captcha.imgData"
                class="captcha-img"
                :src="captcha.imgData"
                :alt="t('captchaTitle')"
                :title="t('captchaRefresh')"
                @click="refreshCaptcha"
              />
            </div>
            <div v-if="!!captchaErr" class="err-msg atk-fade-in">{{ captchaErr }}</div>
          </div>
          <div v-if="!!loginErr" class="err-msg atk-fade-in">{{ loginErr }}</div>
          <Button type="submit" class="w-full" :disabled="captchaBusy">
            {{ t('login') }}
          </Button>
        </form>
      </CardContent>
    </Card>
    <div class="copyright">
      Powered by
      <a href="https://artalk.js.org" target="_blank">Artalk</a>
      ({{ versionInfo }})
    </div>

    <div v-if="userSelector" class="layer">
      <div class="user-selector atk-fade-in">
        <div class="text">{{ t('loginSelectHint') }}</div>
        <div class="user-list">
          <div v-for="(u, i) in userSelector" :key="i" class="item" @click="selectUser(u)">
            {{ u }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.login-dialog {
  z-index: 60;
  position: fixed;
  display: flex;
  flex-direction: column;
  place-items: center;
  justify-content: center;
  top: 0;
  left: 0;
  background: var(--background);
  width: 100%;
  height: 100%;

  .logo {
    user-select: none;
    display: inline-block;
    width: 44px;
    height: 44px;
    margin-bottom: 24px;
    border-radius: 8px;
  }

  .copyright {
    font-size: 13px;
    color: var(--muted-foreground);
    position: fixed;
    height: 30px;
    bottom: 0;
    width: 100%;
    text-align: center;
  }
}

.login-card {
  width: 360px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 12px;

  input {
    height: 40px;
  }

  button {
    margin-top: 8px;
  }

  .captcha-panel {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px;
    background: var(--muted);
    border: 1px solid var(--border);
    border-radius: calc(var(--radius) - 2px);
  }

  .captcha-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .captcha-input {
    flex: 1;
    min-width: 0;
  }

  .captcha-img {
    width: auto;
    max-width: 132px;
    height: 40px;
    flex-shrink: 0;
    background: #fff;
    border: 1px solid var(--border);
    border-radius: 6px;
    cursor: pointer;
  }

  .captcha-iframe {
    width: 100%;
    height: 320px;
    background: var(--background);
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .err-msg {
    display: flex;
    flex-direction: row;
    align-items: center;
    margin-top: 2px;
    color: var(--at-color-red);
    font-size: 13px;

    &::before {
      content: '';
      display: inline-block;
      vertical-align: bottom;
      height: 19px;
      width: 19px;
      margin-right: 6px;
      background-size: contain;
      background-position: center;
      background-repeat: no-repeat;
      background-image: url("data:image/svg+xml,%3Csvg width='19' height='19' viewBox='0 0 19 19' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath fill-rule='evenodd' clip-rule='evenodd' d='M9.5 15.8332C12.9978 15.8332 15.8333 12.9976 15.8333 9.49984C15.8333 6.00203 12.9978 3.1665 9.50001 3.1665C6.0022 3.1665 3.16667 6.00203 3.16667 9.49984C3.16667 12.9976 6.0022 15.8332 9.5 15.8332ZM17.4167 9.49984C17.4167 13.8721 13.8723 17.4165 9.5 17.4165C5.12775 17.4165 1.58334 13.8721 1.58334 9.49984C1.58334 5.12758 5.12775 1.58317 9.50001 1.58317C13.8723 1.58317 17.4167 5.12758 17.4167 9.49984ZM8.70834 13.4581L8.70834 11.8748L10.2917 11.8748L10.2917 13.4581L8.70834 13.4581ZM10.2917 5.54148L10.2917 11.0831L8.70834 11.0831L8.70834 5.54148L10.2917 5.54148Z' fill='%23F53F3F'/%3E%3C/svg%3E");
    }
  }
}

.layer {
  position: fixed;
  width: 100%;
  height: 100%;
  background: var(--at-color-bg-transl);
}

.user-selector {
  z-index: 999;
  position: fixed;
  left: 50%;
  top: 43.5%;
  transform: translate(-50%, -50%);
  background: var(--atk-popover);
  border: 1px solid var(--at-color-border);
  border-radius: var(--atk-radius, 8px);
  box-shadow:
    0 4px 6px -1px rgb(0 0 0 / 0.1),
    0 2px 4px -2px rgb(0 0 0 / 0.1);
  padding-bottom: 10px;
  width: 280px;
  overflow: hidden;

  & > .text {
    font-size: 15px;
    padding: 15px 20px;
  }

  .user-list {
    & > .item {
      cursor: pointer;
      padding: 10px 20px;
      transition: background-color 0.15s ease;

      &:hover {
        background: var(--atk-accent);
        color: var(--atk-accent-foreground);
      }
    }
  }
}
</style>
