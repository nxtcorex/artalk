export interface ActivityDay {
  date: string
  count: number
}

export function contributionLevel(count: number, maximum: number) {
  return count > 0 ? Math.min(4, Math.ceil((count * 4) / Math.max(1, maximum))) : 0
}

export function contributionCalendar(days: ActivityDay[], today: string) {
  const offset = days.length ? new Date(`${days[0].date}T00:00:00Z`).getUTCDay() : 0
  const maximum = Math.max(1, ...days.map((day) => day.count))
  const cells = days.map((day, index) => ({
    ...day,
    column: Math.floor((index + offset) / 7),
    row: (index + offset) % 7,
    level: contributionLevel(day.count, maximum),
    future: day.date > today,
  }))
  const months = cells
    .filter((cell, index) => index === 0 || cell.date.endsWith('-01'))
    .filter(
      (cell, index, all) => index === all.length - 1 || all[index + 1].column - cell.column >= 3,
    )
  return { cells, months, columns: Math.ceil((days.length + offset) / 7) }
}
