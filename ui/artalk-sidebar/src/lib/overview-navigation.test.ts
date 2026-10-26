// @vitest-environment jsdom
import { describe, expect, it } from 'vitest'
import { commentsTab, initialPage, overviewURL } from './overview-navigation'

describe('overview navigation', () => {
  it('defaults to overview only for administrators in the standalone dashboard', () => {
    expect(initialPage(true, false, '')).toBe('/overview')
    expect(initialPage(true, true, '')).toBe('/comments')
    expect(initialPage(false, false, '')).toBe('/comments')
    expect(initialPage(false, true, 'overview')).toBe('/comments')
    expect(initialPage(true, false, 'pages')).toBe('/pages')
    expect(initialPage(true, true, 'comments')).toBe('/comments')
    expect(initialPage(true, true, 'overview')).toBe('/overview')
  })

  it('accepts the pending deep link and validates comment tabs for the current role', () => {
    expect(commentsTab('pending', true)).toBe('pending')
    expect(commentsTab('pending', false)).toBe('pending')
    expect(commentsTab('personal_all', true)).toBe('personal_all')
    expect(commentsTab('personal_all', false)).toBe('all')
    expect(commentsTab('mentions', false)).toBe('mentions')
    expect(commentsTab('mentions', true)).toBe('all')
    expect(commentsTab(['pending'], true)).toBe('all')
    expect(commentsTab(undefined, true)).toBe('all')
  })

  it('opens comments on the original page and rejects unsafe schemes', () => {
    expect(overviewURL('https://example.com/post#old', 42)).toBe(
      'https://example.com/post#atk-comment-42',
    )
    expect(overviewURL('javascript:alert(1)')).toBeUndefined()
    expect(overviewURL('data:text/html,test')).toBeUndefined()
    expect(overviewURL('')).toBeUndefined()
    expect(overviewURL('https://[invalid')).toBeUndefined()
  })
})
