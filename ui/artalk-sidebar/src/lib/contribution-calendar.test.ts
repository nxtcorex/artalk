import { describe, expect, it } from 'vitest'
import { contributionCalendar, contributionLevel } from './contribution-calendar'

function days(start: string, count: number) {
  const first = new Date(`${start}T00:00:00Z`)
  return Array.from({ length: count }, (_, index) => {
    const date = new Date(first)
    date.setUTCDate(date.getUTCDate() + index)
    return { date: date.toISOString().slice(0, 10), count: 0 }
  })
}

describe('contribution calendar', () => {
  it('aligns days in Sunday-first weeks without timezone shifts', () => {
    const calendar = contributionCalendar(days('2024-01-01', 366), '2024-12-31')
    expect(calendar.cells).toHaveLength(366)
    expect(calendar.cells[0]).toMatchObject({ date: '2024-01-01', column: 0, row: 1 })
    expect(calendar.cells[6]).toMatchObject({ date: '2024-01-07', column: 1, row: 0 })
    expect(calendar.cells[59].date).toBe('2024-02-29')
    expect(calendar.months.map((month) => month.date)).toHaveLength(12)
  })

  it('supports leap years spanning 54 weeks and marks future days', () => {
    const calendar = contributionCalendar(days('2000-01-01', 366), '2000-02-29')
    expect(calendar.columns).toBe(54)
    expect(calendar.cells[59].future).toBe(false)
    expect(calendar.cells[60].future).toBe(true)
    expect(calendar.cells[365]).toMatchObject({ date: '2000-12-31', column: 53, row: 0 })
  })

  it('assigns four green levels and leaves empty days at zero', () => {
    expect([0, 1, 5, 10, 15, 20].map((value) => contributionLevel(value, 20))).toEqual([
      0, 1, 1, 2, 3, 4,
    ])
    expect(contributionLevel(0, 0)).toBe(0)
    expect(contributionCalendar([], '')).toEqual({ cells: [], months: [], columns: 0 })
  })
})
