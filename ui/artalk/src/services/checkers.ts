import type { ArtalkPlugin } from '@/types'
import { CheckerLauncher } from '@/components/checker'

export const CheckersService: ArtalkPlugin = (ctx) => {
  ctx.provide(
    'checkers',
    (api, apiHandlers, layers, user, config) => {
      const checkers = new CheckerLauncher({
        getApi: () => api,
        getLayers: () => layers,
        getUser: () => user,
        getEditorEl: () => ctx.getEl().querySelector<HTMLElement>('.atk-main-editor'),
        onReload: () => ctx.reload(),

        // make sure suffix with a slash, because it will be used as a base url when call `fetch`
        getCaptchaIframeURL: () => {
          const darkMode = config.get().darkMode
          const isDark =
            darkMode === 'auto'
              ? window.matchMedia('(prefers-color-scheme: dark)').matches
              : darkMode === true
          return `${config.get().server}/api/v2/captcha/?t=${+new Date()}&theme=${isDark ? 'dark' : 'light'}`
        },
      })

      apiHandlers.add('need_captcha', (res) => checkers.checkCaptcha(res))
      apiHandlers.add('need_login', () => checkers.checkAdmin({}))

      return checkers
    },
    ['api', 'apiHandlers', 'layers', 'user', 'config'] as const,
  )
}
