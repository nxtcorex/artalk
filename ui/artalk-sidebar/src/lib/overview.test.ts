// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createApp, defineComponent, h, nextTick, type PropType } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import Overview from '../pages/overview.vue'
import { useUserStore } from '../stores/user'
import { en } from '../i18n-en'

const mocks = vi.hoisted(() => ({ getOverview: vi.fn(), updateTabs: vi.fn() }))
vi.mock('../components/OverviewActivity.vue', () => ({ default: { render: () => null } }))
vi.mock('../global', () => ({
  artalk: { ctx: { getApi: () => ({ admin: { getOverview: mocks.getOverview } }) } },
}))
vi.mock('../stores/nav', () => ({ useNavStore: () => ({ updateTabs: mocks.updateTabs }) }))
vi.mock('../stores/user', async () => {
  const { defineStore } = await import('pinia')
  return { useUserStore: defineStore('overview-test-user', { state: () => ({ site: 'Site A' }) }) }
})

function payload(count = 12) {
  return {
    data: {
      timezone: 'Asia/Shanghai',
      generated_at: '2026-10-26T12:00:00+08:00',
      metrics: {
        today_comments: count,
        pending: 2,
        published: 100,
        participants: 5,
        commented_pages: 3,
        pv: 900,
      },
      trend: [{ date: '2026-10-26', published: 10, pending: 2 }],
      latest_comments: [
        {
          id: 42,
          nick: 'Reader',
          summary: '<b>plain text</b>',
          page_title: 'My page',
          page_key: '/post',
          page_url: 'https://example.com/post',
          site_name: 'Site A',
          created_at: '2026-10-26T11:00:00+08:00',
          is_pending: true,
        },
      ],
      popular_pages: [
        {
          id: 3,
          title: 'My page',
          key: '/post',
          url: 'https://example.com/post',
          site_name: 'Site A',
          comments: 10,
        },
      ],
    },
  }
}

const RouterLink = defineComponent({
  props: {
    to: {
      type: [String, Object] as PropType<string | { path: string; query: { tab: string } }>,
      required: true,
    },
  },
  setup(props, { slots }) {
    return () =>
      h(
        'a',
        {
          href:
            typeof props.to === 'string' ? props.to : `${props.to.path}?tab=${props.to.query.tab}`,
        },
        slots.default?.(),
      )
  },
})

const cleanups: (() => void)[] = []
function mount() {
  const pinia = createPinia()
  const container = document.createElement('div')
  document.body.append(container)
  const app = createApp(Overview)
  app.use(pinia).use(createI18n({ legacy: false, locale: 'en', messages: { en } }))
  app.component('RouterLink', RouterLink)
  app.mount(container)
  cleanups.push(() => {
    app.unmount()
    container.remove()
  })
  return { container, user: useUserStore(pinia) }
}

async function flush() {
  await Promise.resolve()
  await nextTick()
}
beforeEach(() => {
  mocks.getOverview.mockReset()
  mocks.updateTabs.mockReset()
})
afterEach(() => {
  cleanups.splice(0).forEach((cleanup) => cleanup())
})

describe('overview', () => {
  it('loads the selected site and links pending comments to moderation', async () => {
    mocks.getOverview.mockResolvedValue(payload())
    const { container } = mount()
    await flush()
    expect(mocks.getOverview).toHaveBeenCalledWith({ site_name: 'Site A', days: 7 })
    expect(mocks.updateTabs).toHaveBeenCalledWith({}, '')
    expect(container.querySelector('.pending-metric')?.getAttribute('href')).toBe(
      '/comments?tab=pending',
    )
    expect(container.querySelector('.excerpt')?.textContent).toBe('<b>plain text</b>')
    expect(container.querySelector('.excerpt b')).toBeNull()
    expect(container.querySelector('.page-link')?.getAttribute('href')).toBe(
      'https://example.com/post#atk-comment-42',
    )
    expect(container.querySelectorAll('table tbody tr')).toHaveLength(1)
  })

  it('reloads when the period changes', async () => {
    mocks.getOverview.mockResolvedValue(payload())
    const { container } = mount()
    await flush()
    const select = container.querySelector('select')!
    select.value = '30'
    select.dispatchEvent(new Event('change'))
    await flush()
    expect(mocks.getOverview).toHaveBeenLastCalledWith({ site_name: 'Site A', days: 30 })
  })

  it('ignores a stale response after switching sites', async () => {
    let resolveFirst!: (value: ReturnType<typeof payload>) => void
    mocks.getOverview.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFirst = resolve
      }),
    )
    mocks.getOverview.mockResolvedValueOnce(payload(222))
    const { container, user } = mount()
    user.site = 'Site B'
    await flush()
    expect(container.querySelector('.metric strong')?.textContent).toBe('222')
    resolveFirst(payload(111))
    await flush()
    expect(container.querySelector('.metric strong')?.textContent).toBe('222')
  })

  it('shows an error with retry and recovers to the empty state', async () => {
    mocks.getOverview.mockRejectedValueOnce(new Error('offline'))
    const { container } = mount()
    await flush()
    expect(container.querySelector('[role="alert"]')).not.toBeNull()
    const empty = payload(0)
    empty.data.metrics.pending = 0
    empty.data.latest_comments = []
    empty.data.popular_pages = []
    empty.data.trend[0].published = 0
    empty.data.trend[0].pending = 0
    mocks.getOverview.mockResolvedValueOnce(empty)
    container.querySelector<HTMLButtonElement>('[role="alert"] button')!.click()
    await flush()
    expect(container.querySelector('[role="alert"]')).toBeNull()
    expect(container.querySelector('.review-notice')).toBeNull()
    expect(container.querySelector('.chart-empty')?.textContent).toBe('No comments in this period')
    expect(container.querySelector('.empty')?.textContent).toBe('No Content')
  })
})
