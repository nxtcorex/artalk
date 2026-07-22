import { describe, expect, it, vi } from 'vitest'
import { Fetch } from './fetch'

describe('comment list fetch', () => {
  it('loads comments when the response does not include page data', async () => {
    const eventHandlers: Record<string, (payload?: any) => void> = {}
    const trigger = vi.fn()
    const onSuccess = vi.fn()
    const onError = vi.fn()
    const comment = { id: 1, content: 'sidebar comment' }

    const data = {
      getLoading: vi.fn(() => false),
      setLoading: vi.fn(),
      setListLastFetch: vi.fn(),
      loadComments: vi.fn(),
      updatePage: vi.fn(),
    }

    let resolveRequest!: (response: any) => void
    const api = {
      getUserFields: vi.fn(() => undefined),
      comments: {
        getComments: vi.fn(
          () =>
            new Promise((resolve) => {
              resolveRequest = resolve
            }),
        ),
      },
    }

    const ctx = {
      inject: vi.fn(() => ({
        get: () => ({
          pagination: { pageSize: 20 },
          flatMode: true,
          pageKey: '/sidebar',
          site: 'site',
        }),
      })),
      on: (name: string, handler: (payload?: any) => void) => {
        eventHandlers[name] = handler
      },
      getData: () => data,
      getApi: () => api,
      trigger,
    }

    Fetch(ctx as any)
    eventHandlers['list-fetch']({ offset: 0, onSuccess, onError })

    resolveRequest({
      data: {
        comments: [comment],
        count: 1,
        roots_count: 1,
      },
    })

    await vi.waitFor(() => expect(data.loadComments).toHaveBeenCalledWith([comment]))

    expect(data.updatePage).not.toHaveBeenCalled()
    expect(onSuccess).toHaveBeenCalledWith({
      comments: [comment],
      count: 1,
      roots_count: 1,
    })
    expect(onError).not.toHaveBeenCalled()
    expect(trigger).toHaveBeenCalledWith(
      'list-fetched',
      expect.objectContaining({
        data: { comments: [comment], count: 1, roots_count: 1 },
      }),
    )
  })
})
