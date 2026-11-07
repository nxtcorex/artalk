import type { Checker } from '.'
import $t from '@/i18n'
import * as Utils from '@/lib/utils'

const AdminChecker: Checker<{ token: string }> = {
  inputType: 'password',
  submitText: () => $t('login'),
  inputPlaceholder: () => $t('password'),

  async request(checker, inputVal) {
    return (
      await checker.getApi().user.login({
        name: checker.getUser().getData().name,
        email: checker.getUser().getData().email,
        password: inputVal,
      })
    ).data
  },

  body() {
    return Utils.createElement(
      `<div class="atk-checker-head"><div class="atk-checker-title">${$t('login')}</div><div class="atk-checker-desc">${$t('adminCheck')}</div></div>`,
    )
  },

  onSuccess(checker, res) {
    checker.getUser().update({
      is_admin: true,
      token: res?.token || checker.getUser().getData().token,
    })
    checker.getOpts().onReload()
  },

  onError(checker, err, inputVal, formEl) {},
}

export default AdminChecker
