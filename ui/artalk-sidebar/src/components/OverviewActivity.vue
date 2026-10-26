<script setup lang="ts">
import type { ArtalkType } from 'artalk'
import { artalk } from '../global'
import { contributionCalendar } from '../lib/contribution-calendar'

type Activity = Awaited<ReturnType<ArtalkType.Api['admin']['getCommentActivity']>>['data']
const props = withDefaults(defineProps<{ site: string; refreshKey?: number }>(), { refreshKey: 0 })
const { t, locale } = useI18n()
const year = ref(0)
const data = ref<Activity>()
const years = ref<number[]>([])
const loading = ref(false)
const failed = ref(false)
const scroller = ref<HTMLElement>()
const calendarEl = ref<SVGElement>()
const focusedDate = ref('')
const tooltip = ref<{ label: string; left: number; top: number }>()
let requestID = 0

async function load() {
  const id = ++requestID
  loading.value = true
  failed.value = false
  data.value = undefined
  tooltip.value = undefined
  try {
    const response = await artalk!.ctx.getApi().admin.getCommentActivity({
      site_name: props.site,
      ...(year.value ? { year: year.value } : {}),
    })
    if (id !== requestID) return
    data.value = response.data
    years.value = response.data.years
    const activeDays = response.data.days.filter(
      (day) => day.date <= response.data.generated_at.slice(0, 10),
    )
    focusedDate.value = activeDays[activeDays.length - 1]?.date || ''
    loading.value = false
    await nextTick()
    if (id === requestID && scroller.value) {
      scroller.value.scrollLeft = year.value ? 0 : scroller.value.scrollWidth
    }
  } catch {
    if (id === requestID) failed.value = true
  } finally {
    if (id === requestID) loading.value = false
  }
}

watch([() => props.site, () => props.refreshKey, year], load, { immediate: true })
onUnmounted(() => {
  requestID++
})

const availableYears = computed(() =>
  [...new Set([...years.value, ...(year.value ? [year.value] : [])])].sort((a, b) => b - a),
)
const calendar = computed(() =>
  contributionCalendar(data.value?.days || [], data.value?.generated_at.slice(0, 10) || ''),
)
const total = computed(() => data.value?.days.reduce((sum, day) => sum + day.count, 0) || 0)
const number = (value: number) => new Intl.NumberFormat(locale.value).format(value)
const formatDate = (value: string, options: Intl.DateTimeFormatOptions) =>
  new Intl.DateTimeFormat(locale.value, {
    ...options,
    timeZone: 'UTC',
  }).format(new Date(`${value}T00:00:00Z`))
const label = (day: { date: string; count: number }) =>
  t('overviewActivityDay', {
    date: formatDate(day.date, { year: 'numeric', month: 'long', day: 'numeric' }),
    count: number(day.count),
  })

function showDay(event: Event, day: { date: string; count: number }) {
  const rect = (event.currentTarget as SVGElement).getBoundingClientRect()
  tooltip.value = {
    label: label(day),
    left: Math.max(120, Math.min(window.innerWidth - 120, rect.left + rect.width / 2)),
    top: rect.top - 8,
  }
}

function moveFocus(event: KeyboardEvent, index: number) {
  const offsets: Record<string, number> = {
    ArrowLeft: -7,
    ArrowRight: 7,
    ArrowUp: -1,
    ArrowDown: 1,
  }
  if (!(event.key in offsets) && event.key !== 'Home' && event.key !== 'End') return
  event.preventDefault()
  const cells = calendar.value.cells.filter((day) => !day.future)
  const next =
    event.key === 'Home'
      ? 0
      : event.key === 'End'
        ? cells.length - 1
        : Math.max(0, Math.min(cells.length - 1, index + offsets[event.key]))
  const day = cells[next]
  if (!day) return
  focusedDate.value = day.date
  nextTick(() => calendarEl.value?.querySelector<SVGElement>(`[data-date="${day.date}"]`)?.focus())
}
</script>

<template>
  <section class="activity" :aria-busy="loading" aria-labelledby="overview-activity-title">
    <header class="activity-heading">
      <h2 id="overview-activity-title">{{ t('overviewActivity') }}</h2>
      <label>
        <span class="sr-only">{{ t('overviewActivityYear') }}</span>
        <select v-model="year" data-testid="activity-year">
          <option :value="0">{{ t('overviewActivityPastYear') }}</option>
          <option v-for="option in availableYears" :key="option" :value="option">
            {{ option }}
          </option>
        </select>
      </label>
    </header>
    <p v-if="loading" class="activity-state" role="status">{{ t('overviewLoading') }}</p>
    <div v-else-if="failed" class="activity-state" role="alert">
      <p>{{ t('overviewActivityFailed') }}</p>
      <button type="button" @click="load">{{ t('overviewRetry') }}</button>
    </div>
    <template v-else-if="data">
      <p class="activity-total">{{ t('overviewActivityTotal', { count: number(total) }) }}</p>
      <div ref="scroller" class="calendar-scroll" @scroll="tooltip = undefined">
        <svg
          ref="calendarEl"
          class="calendar"
          :viewBox="`0 0 ${40 + calendar.columns * 14} 128`"
          role="group"
          :aria-label="t('overviewActivity')"
        >
          <text
            v-for="month in calendar.months"
            :key="month.date"
            :x="42 + month.column * 14"
            y="12"
            class="calendar-label"
          >
            {{ formatDate(month.date, { month: 'short' }) }}
          </text>
          <text v-for="row in [1, 3, 5]" :key="row" x="0" :y="33 + row * 14" class="calendar-label">
            {{ formatDate(`2024-01-${String(7 + row).padStart(2, '0')}`, { weekday: 'short' }) }}
          </text>
          <rect
            v-for="(day, index) in calendar.cells"
            :key="day.date"
            :x="42 + day.column * 14"
            :y="24 + day.row * 14"
            width="11"
            height="11"
            rx="2"
            class="calendar-day"
            :class="{ future: day.future }"
            :data-level="day.level"
            :data-date="day.date"
            :tabindex="!day.future && day.date === focusedDate ? 0 : -1"
            :aria-hidden="day.future || undefined"
            :aria-label="day.future ? undefined : label(day)"
            role="img"
            @pointerenter="!day.future && showDay($event, day)"
            @pointerleave="tooltip = undefined"
            @focus="showDay($event, day)"
            @blur="tooltip = undefined"
            @click="!day.future && showDay($event, day)"
            @keydown="moveFocus($event, index)"
          />
        </svg>
      </div>
      <div class="activity-footer">
        <span class="activity-range"
          >{{ formatDate(data.start_date, { year: 'numeric', month: 'short', day: 'numeric' }) }} —
          {{ formatDate(data.end_date, { year: 'numeric', month: 'short', day: 'numeric' }) }}</span
        >
        <div
          class="activity-legend"
          :aria-label="`${t('overviewActivityLess')} — ${t('overviewActivityMore')}`"
        >
          <span>{{ t('overviewActivityLess') }}</span>
          <span
            v-for="level in [0, 1, 2, 3, 4]"
            :key="level"
            class="legend-cell"
            :data-level="level"
            aria-hidden="true"
          ></span>
          <span>{{ t('overviewActivityMore') }}</span>
        </div>
      </div>
      <div
        v-if="tooltip"
        class="activity-tooltip"
        :style="{ left: `${tooltip.left}px`, top: `${tooltip.top}px` }"
        aria-hidden="true"
      >
        {{ tooltip.label }}
      </div>
    </template>
  </section>
</template>

<style scoped lang="scss">
.activity {
  --activity-0: #ebedf0;
  --activity-1: #9be9a8;
  --activity-2: #40c463;
  --activity-3: #30a14e;
  --activity-4: #216e39;
  margin-top: 20px;
  padding: 20px;
  border: 1px solid var(--at-color-border);
  border-radius: 10px;
  background: var(--at-color-bg);
  min-width: 0;
}
.atk-dark-mode .activity {
  --activity-0: #161b22;
  --activity-1: #0e4429;
  --activity-2: #006d32;
  --activity-3: #26a641;
  --activity-4: #39d353;
}
.activity-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 18px;
}
.activity-heading h2 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}
.activity-heading select,
.activity-state button {
  font: inherit;
  font-size: 13px;
  color: var(--at-color-font);
  background: var(--at-color-bg);
  border: 1px solid var(--at-color-border);
  border-radius: 6px;
  padding: 7px 10px;
  cursor: pointer;
}
.activity-total {
  margin: 0 0 16px;
  font-size: 14px;
  font-weight: 500;
}
.calendar-scroll {
  overflow-x: auto;
  padding-bottom: 4px;
}
.calendar {
  display: block;
  width: 100%;
  height: auto;
  min-width: 760px;
  overflow: visible;
  max-width: none;
}
.calendar-label {
  fill: var(--at-color-sub);
  font-size: 10px;
}
.calendar-day {
  fill: var(--activity-0);
  stroke: rgb(27 31 35 / 6%);
  stroke-width: 1px;
}
.legend-cell {
  background: var(--activity-0);
  display: inline-block;
  width: 11px;
  height: 11px;
  border-radius: 2px;
  box-shadow: inset 0 0 0 1px rgb(27 31 35 / 6%);
}
[data-level='1'] {
  fill: var(--activity-1);
  background: var(--activity-1);
}
[data-level='2'] {
  fill: var(--activity-2);
  background: var(--activity-2);
}
[data-level='3'] {
  fill: var(--activity-3);
  background: var(--activity-3);
}
[data-level='4'] {
  fill: var(--activity-4);
  background: var(--activity-4);
}
.calendar-day:hover,
.calendar-day:focus {
  stroke: var(--at-color-font);
  stroke-width: 1.5px;
  outline: none;
}
.calendar-day.future {
  visibility: hidden;
}
.activity-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 10px;
  color: var(--at-color-sub);
  font-size: 11px;
}
.activity-legend {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
  white-space: nowrap;
}
.activity-legend span:first-child {
  margin-right: 3px;
}
.activity-legend span:last-child {
  margin-left: 3px;
}
.activity-tooltip {
  position: fixed;
  z-index: 10000;
  transform: translate(-50%, -100%);
  max-width: 220px;
  padding: 8px 10px;
  border-radius: 6px;
  color: var(--at-color-bg);
  background: var(--at-color-font);
  text-align: center;
  line-height: 1.5;
  font-size: 12px;
  pointer-events: none;
}
.activity-state {
  min-height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 12px;
  color: var(--at-color-sub);
  font-size: 13px;
}
.activity-state p {
  margin: 0;
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
@media (max-width: 600px) {
  .activity {
    padding: 16px;
  }
}
</style>
