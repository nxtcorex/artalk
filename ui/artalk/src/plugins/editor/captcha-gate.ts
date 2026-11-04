import EditorPlugin from './_plug'
import type PlugKit from './_kit'
import type { CheckerCaptchaPayload, CheckerCtx } from '@/components/checker'

export default class CaptchaGate extends EditorPlugin {
  private disposed = false
  private gateCtx: CheckerCtx | null = null
  private requiresCode = false

  constructor(kit: PlugKit) {
    super(kit)

    this.kit.useMounted(() => {
      this.kit.useEvents().on('editor-submitted', this.onSubmitted)
      this.kit.useUI().$submitBtn.addEventListener('click', this.onSendClick, true)
      this.refresh()
    })

    this.kit.useUnmounted(() => {
      this.kit.useEvents().off('editor-submitted', this.onSubmitted)
      this.kit.useUI().$submitBtn.removeEventListener('click', this.onSendClick, true)
      this.disposed = true
    })
  }

  private onSubmitted = () => {
    this.refresh()
  }

  private onSendClick = (e: MouseEvent) => {
    if (!this.gateCtx || !this.requiresCode) return

    e.preventDefault()
    e.stopImmediatePropagation()

    const ctx = this.gateCtx
    this.setDisabled(true)

    ctx.submit().then((ok) => {
      if (ok) {
        this.kit.useEditor().submit()
        return
      }

      const input = document.querySelector<HTMLInputElement>(
        '.atk-checker-inline.captcha .atk-checker-input',
      )
      this.setDisabled(!input?.value.trim())
    })
  }

  private async refresh() {
    if (this.disposed) return

    const btn = this.kit.useUI().$submitBtn
    if (!btn.isConnected || btn.offsetParent === null) return

    this.gateCtx = null
    this.requiresCode = false

    let isPass: boolean
    try {
      const res = await this.kit.useApi().captcha.getCaptchaStatus()
      isPass = !!res.data.is_pass
    } catch {
      this.setDisabled(false)
      return
    }

    if (isPass) {
      this.setDisabled(false)
      return
    }

    this.setDisabled(true)

    const payload = await this.fetchPayload()
    if (!payload || this.disposed) {
      this.setDisabled(false)
      return
    }

    try {
      await this.kit.useCheckers().startCaptchaGate({
        ...payload,
        onMount: (panel, ctx) => this.onGateMount(panel, ctx),
      })
    } catch {
      this.setDisabled(true)
      return
    }

    this.gateCtx = null
    this.requiresCode = false
    if (!this.disposed) this.setDisabled(false)
  }

  private onGateMount(panel: HTMLElement, ctx: CheckerCtx) {
    this.gateCtx = ctx
    this.requiresCode = !panel.querySelector('.atk-checker-iframe-wrap')

    const input = panel.querySelector<HTMLInputElement>('.atk-checker-input')
    if (!input) {
      this.setDisabled(true)
      return
    }

    const update = () => this.setDisabled(!input.value.trim())
    input.addEventListener('input', update)
    ctx.addCleanup(() => input.removeEventListener('input', update))
    update()
  }

  private async fetchPayload(): Promise<CheckerCaptchaPayload | null> {
    const server = this.kit.useConf().server
    const resp = await fetch(`${server}/api/v2/captcha/?t=${Date.now()}`).catch(() => null)
    if (!resp || !resp.ok) return null

    const type = resp.headers.get('content-type') || ''
    if (type.includes('application/json')) {
      const data = await resp.json().catch(() => null)
      return data?.img_data ? { img_data: data.img_data } : null
    }

    return { iframe: '1' }
  }

  private setDisabled(disabled: boolean) {
    this.kit.useUI().$submitBtn.disabled = disabled
  }
}
