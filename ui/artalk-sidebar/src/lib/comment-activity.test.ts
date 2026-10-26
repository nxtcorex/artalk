// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createApp, h, nextTick, ref } from 'vue'
import { createI18n } from 'vue-i18n'
import OverviewActivity from '../components/OverviewActivity.vue'
import { en } from '../i18n-en'

const api = vi.hoisted(() => ({ getCommentActivity: vi.fn() }))
vi.mock('../global', () => ({ artalk: { ctx: { getApi: () => ({ admin: api }) } } }))

function payload(count = 5) {
  const first = new Date('2025-10-27T00:00:00Z')
  return {
    data: {
      timezone: 'Asia/Shanghai',
      generated_at: '2026-10-26T12:00:00+08:00',
      start_date: '2025-10-27',
      end_date: '2026-10-26',
      years: [2026, 2025, 2024],
      days: Array.from({ length: 365 }, (_, index) => {
        const date = new Date(first)
        date.setUTCDate(date.getUTCDate() + index)
        return { date: date.toISOString().slice(0, 10), count: index === 364 ? count : 0 }
      }),
    },
  }
}

const cleanups: (() => void)[] = []
function mount() {
  const site = ref('Site A')
  const refreshKey = ref(0)
  const app = createApp({
    setup: () => () => h(OverviewActivity, { site: site.value, refreshKey: refreshKey.value }),
  })
  app.use(createI18n({ legacy: false, locale: 'en', messages: { en } }))
  const container = document.createElement('div')
  document.body.append(container)
  app.mount(container)
  cleanups.push(() => {
    app.unmount()
    container.remove()
  })
  return { container, site, refreshKey }
}
async function flush() {
  await Promise.resolve()
  await nextTick()
  await nextTick()
}
beforeEach(() => api.getCommentActivity.mockReset())
afterEach(() => cleanups.splice(0).forEach((cleanup) => cleanup()))

describe('comment activity calendar', () => {
  it('defaults to the last year, renders every day and exposes the daily count', async () => {
    api.getCommentActivity.mockResolvedValue(payload())
    const { container } = mount()
    await flush()
    expect(api.getCommentActivity).toHaveBeenCalledWith({ site_name: 'Site A' })
    expect(container.querySelectorAll('.calendar-day')).toHaveLength(365)
    expect(container.querySelector('.activity-total')?.textContent).toBe('5 comments')
    const today = container.querySelector<SVGElement>('[data-date="2026-10-26"]')!
    expect(today.getAttribute('data-level')).toBe('4')
    expect(today.getAttribute('aria-label')).toContain('5 comments')
    today.dispatchEvent(new Event('pointerenter'))
    await nextTick()
    expect(container.querySelector('.activity-tooltip')?.textContent).toContain('October 26, 2026')
    today.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowLeft', bubbles: true }))
    await flush()
    expect(container.querySelector('[tabindex="0"]')?.getAttribute('data-date')).toBe('2026-10-19')
  })

  it('changes calendar year and preserves it when refreshing', async () => {
    api.getCommentActivity.mockResolvedValue(payload())
    const { container, refreshKey } = mount()
    await flush()
    const select = container.querySelector('select')!
    expect([...select.options].map((option) => option.value)).toEqual(['0', '2026', '2025', '2024'])
    select.value = '2024'
    select.dispatchEvent(new Event('change'))
    await flush()
    expect(api.getCommentActivity).toHaveBeenLastCalledWith({ site_name: 'Site A', year: 2024 })
    refreshKey.value++
    await flush()
    expect(api.getCommentActivity).toHaveBeenLastCalledWith({ site_name: 'Site A', year: 2024 })
  })

  it('ignores responses from a previously selected site', async () => {
    let resolveFirst!: (value: ReturnType<typeof payload>) => void
    api.getCommentActivity.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFirst = resolve
      }),
    )
    api.getCommentActivity.mockResolvedValueOnce(payload(22))
    const { container, site } = mount()
    site.value = 'Site B'
    await flush()
    expect(container.querySelector('.activity-total')?.textContent).toBe('22 comments')
    resolveFirst(payload(11))
    await flush()
    expect(container.querySelector('.activity-total')?.textContent).toBe('22 comments')
  })

  it('retries a failed load and displays an empty calendar', async () => {
    api.getCommentActivity.mockRejectedValueOnce(new Error('offline'))
    const { container } = mount()
    await flush()
    expect(container.querySelector('[role="alert"]')).not.toBeNull()
    api.getCommentActivity.mockResolvedValueOnce(payload(0))
    container.querySelector<HTMLButtonElement>('[role="alert"] button')!.click()
    await flush()
    expect(container.querySelector('[role="alert"]')).toBeNull()
    expect(container.querySelector('.activity-total')?.textContent).toBe('0 comments')
    expect(container.querySelectorAll('.calendar-day[data-level="0"]')).toHaveLength(365)
  })
})
