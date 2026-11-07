import CaptchaChecker from './captcha'
import AdminChecker from './admin'
import type { Api } from '@/api'
import $t from '@/i18n'
import * as Utils from '@/lib/utils'
import type { UserManager, CheckerManager as ICheckerManager, Layer, LayerManager } from '@/types'

export interface CheckerCaptchaPayload extends CheckerPayload {
  img_data?: string
  iframe?: string
}

export interface CheckerPayload {
  onSuccess?: () => void
  onMount?: (dialogEl: HTMLElement, ctx: CheckerCtx) => void
  onCancel?: () => void
  gate?: boolean
}

export interface CheckerLauncherOptions {
  getApi: () => Api
  getLayers: () => LayerManager
  getUser: () => UserManager
  getEditorEl: () => HTMLElement | null
  getCaptchaIframeURL: () => string
  onReload: () => void
}

function wrapPromise<P extends CheckerPayload = CheckerPayload>(fn: (p: P) => void) {
  return (payload: P) =>
    new Promise<void>((resolve, reject) => {
      const cancelFn = payload.onCancel
      payload.onCancel = () => {
        cancelFn && cancelFn()
        reject(new Error('user canceled the checker'))
      }
      const successFn = payload.onSuccess
      payload.onSuccess = () => {
        successFn && successFn()
        resolve()
      }
      fn(payload)
    })
}

/**
 * Checker Launcher
 */
export class CheckerLauncher implements ICheckerManager {
  private adminLayer: Layer | null = null

  constructor(private opts: CheckerLauncherOptions) {}

  public checkCaptcha: (payload: CheckerCaptchaPayload) => Promise<void> = wrapPromise((p) => {
    this.check(CaptchaChecker, p, (ctx) => {
      ctx.set('img_data', p.img_data)
      ctx.set('iframe', p.iframe)
    })
  })

  public startCaptchaGate: (payload: CheckerCaptchaPayload) => Promise<void> = wrapPromise((p) => {
    this.check(CaptchaChecker, { ...p, gate: true }, (ctx) => {
      ctx.set('img_data', p.img_data)
      ctx.set('iframe', p.iframe)
    })
  })

  public checkAdmin: (payload: CheckerPayload) => Promise<void> = wrapPromise((p) => {
    this.check(AdminChecker, p)
  })

  public check(checker: Checker, payload: CheckerPayload, beforeCheck?: (c: CheckerCtx) => void) {
    const isAdmin = checker === AdminChecker
    const editorEl = this.opts.getEditorEl()
    if (!editorEl && !isAdmin) throw new Error('editor element not found')

    if (editorEl && !isAdmin) {
      editorEl.querySelectorAll('.atk-checker-inline').forEach((el) => el.remove())
    }

    const submitLabel =
      typeof checker.submitText === 'function' ? checker.submitText() : checker.submitText
    const inputPlaceholder =
      typeof checker.inputPlaceholder === 'function'
        ? checker.inputPlaceholder()
        : checker.inputPlaceholder
    const panel = Utils.createElement(`
      <div class="atk-checker-inline ${isAdmin ? 'admin' : 'captcha'}">
        <div class="atk-checker-body"></div>
        <div class="atk-checker-actions">
          <input class="atk-checker-input" type="${checker.inputType || 'text'}" autocomplete="off" placeholder="${inputPlaceholder || ''}" required>
          <button type="button" class="atk-checker-btn atk-checker-confirm">${submitLabel || $t('confirm')}</button>
          <button type="button" class="atk-checker-btn atk-checker-cancel">${$t('cancel')}</button>
        </div>
        <div class="atk-checker-error" hidden></div>
      </div>
    `)

    const $body = panel.querySelector<HTMLElement>('.atk-checker-body')!
    const $actions = panel.querySelector<HTMLElement>('.atk-checker-actions')!
    const $input = panel.querySelector<HTMLInputElement>('.atk-checker-input')!
    const $confirm = panel.querySelector<HTMLButtonElement>('.atk-checker-confirm')!
    const $cancel = panel.querySelector<HTMLButtonElement>('.atk-checker-cancel')!
    const $error = panel.querySelector<HTMLElement>('.atk-checker-error')!
    const confirmText = $confirm.textContent

    let modalLayer: Layer | null = null

    if (isAdmin) {
      const modal = Utils.createElement('<div class="atk-layer-item atk-checker-modal"></div>')
      modal.appendChild(panel)

      this.adminLayer?.destroy()
      modalLayer = this.opts.getLayers().create('checker', modal)
      modalLayer.setAllowMaskClose(false)
      modalLayer.show()
      this.adminLayer = modalLayer
    } else if (editorEl) {
      const bottom = editorEl.querySelector('.atk-bottom')
      const sendItem = bottom?.querySelector('.atk-item.atk-bottom-right')
      if (bottom && sendItem) bottom.insertBefore(panel, sendItem)
      else bottom?.insertAdjacentElement('beforebegin', panel)
    }

    const editorLoading = editorEl?.querySelector<HTMLElement>(':scope > .atk-loading')
    if (editorLoading) editorLoading.style.display = 'none'

    const cleanups: (() => void)[] = []
    const close = () => {
      cleanups.forEach((fn) => fn())
      if (modalLayer) {
        modalLayer.destroy()
        if (this.adminLayer === modalLayer) this.adminLayer = null
        modalLayer = null
      } else {
        panel.remove()
      }
    }

    const checkerStore: CheckerStore = {}
    let hideInteractInput = false
    const checkerCtx: CheckerCtx = {
      set: (key, val) => {
        checkerStore[key] = val
      },
      get: (key) => checkerStore[key],
      getOpts: () => this.opts,
      getUser: () => this.opts.getUser(),
      getApi: () => this.opts.getApi(),
      submit: () => submit(),
      addCleanup: (fn) => {
        cleanups.push(fn)
      },
      hideInteractInput: () => {
        hideInteractInput = true
      },
      triggerSuccess: () => {
        close()
        if (checker.onSuccess) checker.onSuccess(checkerCtx, '', '', panel)
        if (payload.onSuccess) payload.onSuccess()
      },
      cancel: () => {
        close()
        if (payload.onCancel) payload.onCancel()
      },
    }

    if (beforeCheck) beforeCheck(checkerCtx)

    $body.appendChild(checker.body(checkerCtx))

    if (hideInteractInput) $actions.style.display = 'none'
    if (payload.gate) {
      $cancel.style.display = 'none'
      $confirm.style.display = 'none'
    }

    const submit = (): Promise<boolean> => {
      const inputVal = $input.value.trim()

      $error.hidden = true
      $confirm.disabled = true
      $confirm.textContent = `${$t('loading')}...`

      return checker
        .request(checkerCtx, inputVal)
        .then((data) => {
          close()

          if (checker.onSuccess) checker.onSuccess(checkerCtx, data, inputVal, panel)
          if (payload.onSuccess) payload.onSuccess()
          return true
        })
        .catch((err) => {
          $error.textContent = String(err.message || String(err))
          $error.hidden = false
          $confirm.disabled = false
          $confirm.textContent = confirmText

          if (checker.onError) checker.onError(checkerCtx, err, inputVal, panel)
          return false
        })
    }

    $confirm.onclick = () => {
      submit()
    }
    $cancel.onclick = () => {
      close()
      if (payload.onCancel) payload.onCancel()
    }
    $input.onkeyup = (evt) => {
      if (evt.key === 'Enter' || evt.keyCode === 13) {
        evt.preventDefault()
        submit()
      }
    }
    $input.onfocus = () => {
      $error.hidden = true
    }

    if (window.innerWidth >= 768) setTimeout(() => $input.focus(), 80)

    if (payload.onMount) payload.onMount(panel, checkerCtx)
    if (isAdmin) this.loadAuthProviders(panel, checkerCtx)
  }

  private async loadAuthProviders(panel: HTMLElement, ctx: CheckerCtx) {
    let data: Record<string, any>
    try {
      data = (await this.opts.getApi().conf.getSocialLoginProviders()).data as Record<string, any>
    } catch {
      return
    }
    if (!panel.isConnected) return

    const providers: Record<string, any>[] = (data?.providers || []).filter(
      (p: Record<string, any>) => p?.name && p.name !== 'email' && p.name !== 'skip' && p.path,
    )
    if (!providers.length) return

    const authEl = Utils.createElement(
      `<div class="atk-checker-auth"><div class="atk-checker-divider">${$t('orContinueWith')}</div><div class="atk-checker-providers"></div></div>`,
    )
    const list = authEl.querySelector<HTMLElement>('.atk-checker-providers')!

    providers.forEach((p) => {
      const btn = document.createElement('button')
      btn.type = 'button'
      btn.className = 'atk-checker-provider'

      if (p.icon) {
        const img = document.createElement('img')
        img.src = String(p.icon)
        img.alt = ''
        btn.appendChild(img)
      }

      const label = document.createElement('span')
      label.textContent = String(p.label || p.name)
      btn.appendChild(label)

      btn.onclick = () => {
        this.loginWithProvider(String(p.path), data.callback_origin, panel, ctx)
      }
      list.appendChild(btn)
    })

    panel.insertBefore(authEl, panel.querySelector('.atk-checker-error'))
  }

  private loginWithProvider(
    path: string,
    callbackOrigin: string | undefined,
    panel: HTMLElement,
    ctx: CheckerCtx,
  ) {
    const url = new URL(path, window.location.href).href
    const origin = new URL(callbackOrigin || url, window.location.href).origin

    const width = 1020
    const height = 618
    const left = (window.innerWidth - width) / 2
    const top = (window.innerHeight - height) / 2
    const win = window.open(
      url,
      '_blank',
      `width=${width},height=${height},left=${left},top=${top},scrollbars=no,resizable=no,status=no,location=no,toolbar=no,menubar=no`,
    )

    if (!win) {
      this.showPanelError(panel, 'Login popup blocked')
      return
    }

    const onMessage = async (e: MessageEvent) => {
      if (e.source !== win || e.origin !== origin) return
      if (e.data?.type !== 'ATK_AUTH_CALLBACK' || typeof e.data.payload !== 'string') return
      if (!e.data.payload) return

      cleanup()
      win.close()
      await this.finishProviderLogin(e.data.payload, panel, ctx)
    }

    const timer = window.setInterval(() => {
      if (win.closed) {
        cleanup()
        return
      }
      win.postMessage({ type: 'ATK_LOGIN' }, origin)
    }, 1000)

    const cleanup = () => {
      window.clearInterval(timer)
      window.removeEventListener('message', onMessage)
    }

    window.addEventListener('message', onMessage)
    ctx.addCleanup(cleanup)
  }

  private async finishProviderLogin(token: string, panel: HTMLElement, ctx: CheckerCtx) {
    if (!panel.isConnected) return

    const user = ctx.getUser()
    user.update({ token })

    try {
      const res = await this.opts.getApi().user.getUser()
      const u = res.data.user
      if (!u) throw new Error($t('adminRequired'))

      user.update({
        name: u.name,
        email: u.email,
        link: u.link,
        is_admin: u.is_admin,
      })

      if (!u.is_admin) throw new Error($t('adminRequired'))

      ctx.triggerSuccess()
    } catch (e: any) {
      this.showPanelError(panel, e?.message || String(e))
    }
  }

  private showPanelError(panel: HTMLElement, message: string) {
    const el = panel.querySelector<HTMLElement>('.atk-checker-error')
    if (!el) return
    el.textContent = message
    el.hidden = false
  }
}

export interface Checker<T = any> {
  el?: HTMLElement
  inputType?: 'password' | 'text'
  submitText?: string | (() => string)
  inputPlaceholder?: string | (() => string)
  body: (checker: CheckerCtx) => HTMLElement
  request: (checker: CheckerCtx, inputVal: string) => Promise<T>
  onSuccess?: (checker: CheckerCtx, respData: T, inputVal: string, formEl: HTMLElement) => void
  onError?: (checker: CheckerCtx, errData: any, inputVal: string, formEl: HTMLElement) => void
}

interface CheckerStore {
  val?: string
  img_data?: string
  iframe?: string
}

export interface CheckerCtx {
  get<K extends keyof CheckerStore>(key: K): CheckerStore[K]
  set<K extends keyof CheckerStore>(key: K, val: CheckerStore[K]): void
  getOpts(): CheckerLauncherOptions
  getApi(): Api
  getUser(): UserManager
  submit(): Promise<boolean>
  addCleanup(fn: () => void): void
  hideInteractInput(): void
  triggerSuccess(): void
  cancel(): void
}
