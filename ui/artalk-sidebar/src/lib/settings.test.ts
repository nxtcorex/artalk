import { describe, expect, it } from 'vitest'
import YAML from 'yaml'
import { Settings } from './settings'

function createSettings(template: string, custom = '') {
  const settings = new Settings(YAML.parseDocument(template))
  settings.setCustoms(custom)
  return settings
}

describe('settings env values', () => {
  it('keeps env values that contain equals signs', () => {
    const settings = createSettings('db:\n  dsn: ""\n')
    settings.setEnvs(['ATK_DB_DSN=mysql://user:pass@example.com/artalk?token=a=b=='])

    expect(settings.getEnvByPath('db.dsn')).toBe('mysql://user:pass@example.com/artalk?token=a=b==')
  })

  it('treats an empty env value as controlling the option', () => {
    const settings = createSettings(
      'site_url: "https://template.example"\n',
      'site_url: "https://custom.example"\n',
    )
    settings.setEnvs(['ATK_SITE_URL='])

    expect(settings.getEnv('ATK_SITE_URL')).toBe('')
    expect(settings.getEnvByPath('site_url')).toBe('')
    expect(settings.getEnvRefByPath('site_url')).toEqual({ key: 'ATK_SITE_URL', value: '' })
    expect(settings.getCustom('site_url')).toBe('')
  })

  it('recognizes indexed array env values', () => {
    const settings = createSettings('trusted_domains: []\n')
    settings.setEnvs([
      'ATK_TRUSTED_DOMAINS_1=https://b.example',
      'ATK_TRUSTED_DOMAINS_0=https://a.example',
    ])

    expect(settings.getEnvRefByPath('trusted_domains')?.key).toBe('ATK_TRUSTED_DOMAINS_0')
    expect(settings.getCustom('trusted_domains')).toEqual([
      'https://a.example',
      'https://b.example',
    ])
  })

  it('recognizes double-underscore compatibility env aliases', () => {
    const settings = createSettings('ip_region:\n  db_path: ./data/ip2region.xdb\n')
    settings.setEnvs(['ATK_IP__REGION_DB__PATH=/data/ip.xdb'])

    expect(settings.getEnvRefByPath('ip_region.db_path')).toEqual({
      key: 'ATK_IP__REGION_DB__PATH',
      value: '/data/ip.xdb',
    })
    expect(settings.getCustom('ip_region.db_path')).toBe('/data/ip.xdb')
  })
})

describe('settings custom values', () => {
  it('falls back to the template default when the custom YAML omits a field', () => {
    const settings = createSettings('db:\n  port: 3306\n', 'db: {}\n')

    expect(settings.getCustom('db.port')).toBe(3306)
  })

  it('keeps explicit falsy custom values instead of using defaults', () => {
    const settings = createSettings('captcha:\n  enabled: true\n', 'captcha:\n  enabled: false\n')

    expect(settings.getCustom('captcha.enabled')).toBe(false)
  })
})
