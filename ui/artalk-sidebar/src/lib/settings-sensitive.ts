/**
 * List of sensitive config paths.
 *
 * (which should be hidden in the UI)
 */
export const SensitiveConfigPaths = [
  'app_key',
  'admin_notify.ding_talk.secret',
  'admin_notify.ding_talk.token',
  'admin_notify.line.channel_access_token',
  'admin_notify.line.channel_secret',
  'admin_notify.lark.webhook_url',
  'admin_notify.slack.oauth_token',
  'admin_notify.telegram.api_token',
  'admin_notify.webhook.url',
  'auth.apple.client_secret',
  'auth.auth0.client_secret',
  'auth.discord.client_secret',
  'auth.facebook.client_secret',
  'auth.gitea.client_secret',
  'auth.gitea_self.client_secret',
  'auth.github.client_secret',
  'auth.gitlab.client_secret',
  'auth.gitlab_self.client_secret',
  'auth.google.client_secret',
  'auth.line.client_secret',
  'auth.mastodon.client_secret',
  'auth.microsoft.client_secret',
  'auth.patreon.client_secret',
  'auth.slack.client_secret',
  'auth.tiktok.client_secret',
  'auth.twitter.client_secret',
  'auth.wechat.client_secret',
  'auth.steam.api_key',
  'cache.redis.password',
  'captcha.geetest.captcha_key',
  'captcha.capjs.secret_key',
  'captcha.hcaptcha.secret_key',
  'captcha.recaptcha.secret_key',
  'captcha.turnstile.secret_key',
  'db.password',
  'email.ali_dm.access_key_secret',
  'email.smtp.password',
  'img_upload.s3.access_key_id',
  'img_upload.s3.secret_access_key',
  'img_upload.s3.session_token',
  'moderator.aliyun.access_key_secret',
  'moderator.jev.api_key',
  'moderator.jev.headers',
  'moderator.llm.api_key',
  'moderator.llm.headers',
  'moderator.tencent.secret_key',
]

export function isSensitiveConfigPath(path: string) {
  return (
    SensitiveConfigPaths.includes(path) ||
    path.startsWith('moderator.jev.headers.') ||
    path.startsWith('moderator.llm.headers.')
  )
}
