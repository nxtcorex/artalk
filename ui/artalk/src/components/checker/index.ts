import CaptchaChecker from './captcha'
import AdminChecker from './admin'
import type { Api } from '@/api'
import $t from '@/i18n'
import * as Utils from '@/lib/utils'
import type { UserManager, CheckerManager as ICheckerManager, LayerManager } from '@/types'

export interface CheckerCaptchaPayload extends CheckerPayload {
  img_data?: string
  iframe?: string
}

export interface CheckerPayload {
  onSuccess?: () => void
  onMount?: (dialogEl: HTMLElement) => void
  onCancel?: () => void
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
  constructor(private opts: CheckerLauncherOptions) {}

  public checkCaptcha: (payload: CheckerCaptchaPayload) => Promise<void> = wrapPromise((p) => {
    this.check(CaptchaChecker, p, (ctx) => {
      ctx.set('img_data', p.img_data)
      ctx.set('iframe', p.iframe)
    })
  })

  public checkAdmin: (payload: CheckerPayload) => Promise<void> = wrapPromise((p) => {
    this.check(AdminChecker, p)
  })

  public check(checker: Checker, payload: CheckerPayload, beforeCheck?: (c: CheckerCtx) => void) {
    const editorEl = this.opts.getEditorEl()
    if (!editorEl) throw new Error('editor element not found')
    editorEl.querySelector('.atk-checker-inline')?.remove()

    const isAdmin = checker === AdminChecker
    const panel = Utils.createElement(`
      <div class="atk-checker-inline ${isAdmin ? 'admin' : 'captcha'}">
        <div class="atk-checker-body"></div>
        <div class="atk-checker-actions">
          <input class="atk-checker-input" type="${checker.inputType || 'text'}" autocomplete="off" required>
          <button type="button" class="atk-checker-btn atk-checker-confirm">${$t('confirm')}</button>
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

    const host = isAdmin
      ? editorEl.querySelector('.atk-header')
      : editorEl.querySelector('.atk-textarea-wrap')
    host?.insertAdjacentElement('afterend', panel)

    const editorLoading = editorEl.querySelector<HTMLElement>(':scope > .atk-loading')
    if (editorLoading) editorLoading.style.display = 'none'

    const close = () => panel.remove()

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

    const submit = () => {
      const inputVal = $input.value.trim()

      $error.hidden = true
      $confirm.disabled = true
      $confirm.textContent = `${$t('loading')}...`

      checker
        .request(checkerCtx, inputVal)
        .then((data) => {
          close()

          if (checker.onSuccess) checker.onSuccess(checkerCtx, data, inputVal, panel)
          if (payload.onSuccess) payload.onSuccess()
        })
        .catch((err) => {
          $error.textContent = String(err.message || String(err))
          $error.hidden = false
          $confirm.disabled = false
          $confirm.textContent = confirmText

          if (checker.onError) checker.onError(checkerCtx, err, inputVal, panel)
        })
    }

    $confirm.onclick = submit
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

    if (payload.onMount) payload.onMount(panel)
  }
}

export interface Checker<T = any> {
  el?: HTMLElement
  inputType?: 'password' | 'text'
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
  hideInteractInput(): void
  triggerSuccess(): void
  cancel(): void
}
