import { describe, expect, it } from 'vitest'
import { SensitiveConfigPaths, isSensitiveConfigPath } from './settings-sensitive'

describe('sensitive config paths', () => {
  it.each([
    'admin_notify.ding_talk.token',
    'cache.redis.password',
    'captcha.capjs.secret_key',
    'moderator.aliyun.access_key_secret',
    'moderator.tencent.secret_key',
    'auth.gitea_self.client_secret',
  ])('marks %s as sensitive', (path) => {
    expect(isSensitiveConfigPath(path)).toBe(true)
  })

  it.each([
    'admin.notify.bark.server',
    'admin_notify.bark.server',
    'moderator.akismet_key',
    'db.dsn',
  ])('does not mark %s as sensitive', (path) => {
    expect(SensitiveConfigPaths).not.toContain(path)
    expect(isSensitiveConfigPath(path)).toBe(false)
  })

  it('does not keep the old bark typo', () => {
    expect(SensitiveConfigPaths).not.toContain('admin.notify.bark.server')
  })
})
