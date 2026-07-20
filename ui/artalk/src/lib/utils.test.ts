import { afterEach, describe, expect, it, vi } from 'vitest'
import { getCorrectUserAgent } from './utils'

const rawUserAgent =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/100.0.0.0 Safari/537.36'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('getCorrectUserAgent', () => {
  it('falls back to the raw user agent when platformVersion is missing', async () => {
    vi.stubGlobal('navigator', {
      userAgent: rawUserAgent,
      userAgentData: {
        platform: 'Windows',
        getHighEntropyValues: vi.fn().mockResolvedValue({}),
      },
    })

    await expect(getCorrectUserAgent()).resolves.toBe(rawUserAgent)
  })

  it('corrects the Windows version when platformVersion is available', async () => {
    const getHighEntropyValues = vi.fn().mockResolvedValue({ platformVersion: '13.0.0' })
    vi.stubGlobal('navigator', {
      userAgent: rawUserAgent,
      userAgentData: {
        platform: 'Windows',
        getHighEntropyValues,
      },
    })

    await expect(getCorrectUserAgent()).resolves.toContain('Windows NT 11.0')
    expect(getHighEntropyValues).toHaveBeenCalledWith(['platformVersion'])
  })
})
