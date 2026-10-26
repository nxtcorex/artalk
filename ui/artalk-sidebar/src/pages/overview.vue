<route lang="json">
{
  "meta": {
    "title": "Overview - Artalk Sidebar",
    "adminOnly": true
  }
}
</route>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import type { ArtalkType } from 'artalk'
import { artalk } from '../global'
import { useNavStore } from '../stores/nav'
import { useUserStore } from '../stores/user'
import { overviewURL } from '../lib/overview-navigation'
import OverviewActivity from '../components/OverviewActivity.vue'

type Overview = Awaited<ReturnType<ArtalkType.Api['admin']['getOverview']>>['data']

const nav = useNavStore()
const { site } = storeToRefs(useUserStore())
const { t, locale } = useI18n()
const days = ref<7 | 30>(7)
const data = ref<Overview>()
const loading = ref(false)
const failed = ref(false)
const activityRefreshKey = ref(0)
let requestID = 0

onMounted(() => nav.updateTabs({}, ''))
onUnmounted(() => {
  requestID++
})

async function load() {
  const id = ++requestID
  loading.value = true
  failed.value = false
  data.value = undefined
  try {
    const response = await artalk!.ctx
      .getApi()
      .admin.getOverview({ site_name: site.value, days: days.value })
    if (id === requestID) data.value = response.data
  } catch {
    if (id === requestID) failed.value = true
  } finally {
    if (id === requestID) loading.value = false
  }
}

function refresh() {
  activityRefreshKey.value++
  load()
}

watch([site, days], load, { immediate: true })

const number = (value: number) => new Intl.NumberFormat(locale.value).format(value)
const time = (value: string) => value.replace('T', ' ').slice(0, 16)
const maximum = computed(() =>
  Math.max(1, ...(data.value?.trend.map((day) => day.published + day.pending) || [])),
)
const chartEmpty = computed(() =>
  data.value?.trend.every((day) => day.published + day.pending === 0),
)
const barWidth = computed(() => 600 / (data.value?.trend.length || days.value))
const height = (count: number) => (140 * count) / maximum.value
const dayLabel = (date: string) =>
  new Intl.DateTimeFormat(locale.value, {
    month: 'short',
    day: 'numeric',
    timeZone: 'UTC',
  }).format(new Date(`${date}T12:00:00Z`))

function chartLabel(day: Overview['trend'][number]) {
  return `${day.date}: ${t('overviewPublished')} ${number(day.published)}, ${t('pending')} ${number(day.pending)}`
}
</script>

<template>
  <main class="overview" :aria-busy="loading">
    <header class="overview-heading">
      <div>
        <p class="eyebrow">{{ site || t('allSites') }}</p>
        <h1>{{ t('overview') }}</h1>
      </div>
      <div class="controls">
        <label>
          <span class="sr-only">{{ t('overviewPeriod') }}</span>
          <select v-model="days" data-testid="overview-period">
            <option :value="7">{{ t('overviewDays', { days: 7 }) }}</option>
            <option :value="30">{{ t('overviewDays', { days: 30 }) }}</option>
          </select>
        </label>
        <button type="button" :disabled="loading" @click="refresh">
          {{ t('overviewRefresh') }}
        </button>
      </div>
    </header>

    <p v-if="loading" class="state" role="status">{{ t('overviewLoading') }}</p>
    <div v-else-if="failed" class="state error" role="alert">
      <p>{{ t('overviewFailed') }}</p>
      <button type="button" @click="load">{{ t('overviewRetry') }}</button>
    </div>

    <template v-else-if="data">
      <div class="metrics">
        <div class="metric">
          <span>{{ t('overviewToday') }}</span>
          <strong>{{ number(data.metrics.today_comments) }}</strong>
        </div>
        <RouterLink
          class="metric pending-metric"
          :to="{ path: '/comments', query: { tab: 'pending' } }"
        >
          <span>{{ t('overviewPending') }} <span aria-hidden="true">↗</span></span>
          <strong>{{ number(data.metrics.pending) }}</strong>
        </RouterLink>
        <div class="metric">
          <span>{{ t('overviewPublishedTotal') }}</span>
          <strong>{{ number(data.metrics.published) }}</strong>
        </div>
        <div class="metric">
          <span>{{ t('overviewParticipants') }}</span>
          <strong>{{ number(data.metrics.participants) }}</strong>
        </div>
      </div>

      <RouterLink
        v-if="data.metrics.pending > 0"
        class="review-notice"
        :to="{ path: '/comments', query: { tab: 'pending' } }"
      >
        <span>{{ t('overviewPendingNotice', { count: number(data.metrics.pending) }) }}</span>
        <span>{{ t('overviewReview') }} →</span>
      </RouterLink>

      <section class="panel trend-panel">
        <div class="panel-heading">
          <h2>{{ t('overviewTrend') }}</h2>
          <div class="legend">
            <span class="published-dot"></span>{{ t('overviewPublished')
            }}<span class="pending-dot"></span>{{ t('pending') }}
          </div>
        </div>
        <p v-if="chartEmpty" class="chart-empty">{{ t('overviewNoActivity') }}</p>
        <svg
          class="trend"
          viewBox="0 0 660 195"
          role="img"
          aria-labelledby="overview-chart-title overview-chart-description"
        >
          <title id="overview-chart-title">{{ t('overviewTrend') }}</title>
          <desc id="overview-chart-description">
            {{ t('overviewChartDescription', { days, timezone: data.timezone }) }}
          </desc>
          <g v-for="fraction in [0, 0.5, 1]" :key="fraction" class="grid">
            <line x1="45" x2="645" :y1="155 - fraction * 140" :y2="155 - fraction * 140" />
            <text x="38" :y="159 - fraction * 140" text-anchor="end">
              {{ number(Math.ceil(maximum * fraction)) }}
            </text>
          </g>
          <g
            v-for="(day, index) in data.trend"
            :key="day.date"
            tabindex="0"
            role="img"
            :aria-label="chartLabel(day)"
          >
            <title>{{ chartLabel(day) }}</title>
            <rect
              class="published-bar"
              :x="45 + index * barWidth + barWidth * 0.2"
              :y="155 - height(day.published)"
              :width="barWidth * 0.6"
              :height="height(day.published)"
              rx="1"
            />
            <rect
              class="pending-bar"
              :x="45 + index * barWidth + barWidth * 0.2"
              :y="155 - height(day.published + day.pending)"
              :width="barWidth * 0.6"
              :height="height(day.pending)"
              rx="1"
            />
            <text
              v-if="
                index === 0 ||
                index === data.trend.length - 1 ||
                index === Math.floor(data.trend.length / 2)
              "
              :x="45 + (index + 0.5) * barWidth"
              y="183"
              text-anchor="middle"
              class="date-label"
            >
              {{ dayLabel(day.date) }}
            </text>
          </g>
        </svg>
        <details class="daily-data">
          <summary>{{ t('overviewDailyData') }}</summary>
          <table>
            <caption class="sr-only">
              {{
                t('overviewTrend')
              }}
            </caption>
            <thead>
              <tr>
                <th scope="col">{{ t('overviewDate') }}</th>
                <th scope="col">{{ t('overviewPublished') }}</th>
                <th scope="col">{{ t('pending') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="day in data.trend" :key="day.date">
                <th scope="row">{{ day.date }}</th>
                <td>{{ number(day.published) }}</td>
                <td>{{ number(day.pending) }}</td>
              </tr>
            </tbody>
          </table>
        </details>
      </section>

      <div class="lists">
        <section class="panel">
          <div class="panel-heading">
            <h2>{{ t('overviewLatest') }}</h2>
            <RouterLink to="/comments">{{ t('overviewManage') }} →</RouterLink>
          </div>
          <p v-if="!data.latest_comments.length" class="empty">{{ t('noContent') }}</p>
          <ol v-else class="comment-list">
            <li v-for="comment in data.latest_comments" :key="comment.id">
              <div class="comment-heading">
                <strong>{{ comment.nick || t('overviewUnknownUser') }}</strong
                ><span class="status" :class="{ 'is-pending': comment.is_pending }">{{
                  comment.is_pending ? t('pending') : t('overviewPublished')
                }}</span>
              </div>
              <p class="excerpt">{{ comment.summary || t('noContent') }}</p>
              <a
                v-if="overviewURL(comment.page_url, comment.id)"
                :href="overviewURL(comment.page_url, comment.id)"
                target="_blank"
                rel="noopener noreferrer"
                class="page-link"
                >{{ comment.page_title || comment.page_key }}</a
              >
              <span v-else class="page-link">{{ comment.page_title || comment.page_key }}</span>
              <div class="meta">
                <span>{{ comment.site_name }}</span
                ><time :datetime="comment.created_at">{{ time(comment.created_at) }}</time>
              </div>
            </li>
          </ol>
        </section>
        <section class="panel">
          <div class="panel-heading">
            <h2>{{ t('overviewPopular') }}</h2>
            <span class="period-label">{{ t('overviewDays', { days }) }}</span>
          </div>
          <p v-if="!data.popular_pages.length" class="empty">{{ t('overviewNoActivity') }}</p>
          <ol v-else class="popular-list">
            <li v-for="(page, index) in data.popular_pages" :key="page.id">
              <span class="rank">{{ index + 1 }}</span>
              <div class="popular-page">
                <a
                  v-if="overviewURL(page.url)"
                  :href="overviewURL(page.url)"
                  target="_blank"
                  rel="noopener noreferrer"
                  >{{ page.title || page.key }}</a
                >
                <span v-else>{{ page.title || page.key }}</span>
                <small>{{ page.site_name }}</small>
              </div>
              <strong :aria-label="`${t('comments')}: ${number(page.comments)}`">{{
                number(page.comments)
              }}</strong>
            </li>
          </ol>
          <div class="secondary-metrics">
            <div>
              <span>{{ t('overviewCommentedPages') }}</span
              ><strong>{{ number(data.metrics.commented_pages) }}</strong>
            </div>
            <div>
              <span>{{ t('overviewPV') }}</span
              ><strong>{{ number(data.metrics.pv) }}</strong>
            </div>
          </div>
        </section>
      </div>
    </template>
    <OverviewActivity :site="site" :refresh-key="activityRefreshKey" />
    <footer v-if="data" class="overview-footer">
      {{ t('overviewUpdated', { time: time(data.generated_at), timezone: data.timezone }) }}
    </footer>
  </main>
</template>

<style scoped lang="scss">
.overview {
  --overview-accent: #3983de;
  --overview-pending: #ce8c23;
  padding: 28px;
  font-size: 14px;

  h1,
  h2,
  p {
    margin: 0;
  }
  h1 {
    font-size: 27px;
    font-weight: 600;
    margin-top: 5px;
  }
  h2 {
    font-size: 16px;
    font-weight: 600;
  }
  a {
    color: var(--at-color-font);
    text-decoration: none;
  }
  a:hover {
    color: var(--overview-accent);
  }
  button,
  select {
    font: inherit;
    color: var(--at-color-font);
    background: var(--at-color-bg);
    border: 1px solid var(--at-color-border);
    border-radius: 6px;
    padding: 8px 11px;
    cursor: pointer;
  }
  button:disabled {
    opacity: 0.5;
    cursor: default;
  }
  a:focus-visible,
  button:focus-visible,
  select:focus-visible,
  summary:focus-visible {
    outline: 2px solid var(--overview-accent);
    outline-offset: 3px;
  }
}
.overview-heading,
.panel-heading,
.controls,
.comment-heading,
.meta,
.review-notice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.overview-heading {
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.eyebrow,
.period-label,
.meta,
.overview-footer,
.popular-page small {
  color: var(--at-color-sub);
  font-size: 12px;
}
.eyebrow {
  overflow-wrap: anywhere;
}
.metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 20px;
}
.metric,
.panel {
  border: 1px solid var(--at-color-border);
  border-radius: 10px;
  background: var(--at-color-bg);
}
.metric {
  display: flex;
  flex-direction: column;
  padding: 18px;
  gap: 8px;
}
.metric > span {
  color: var(--at-color-sub);
  font-size: 13px;
}
.metric strong {
  font-size: clamp(24px, 3vw, 34px);
  font-weight: 600;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.pending-metric strong {
  color: var(--overview-pending);
}
.pending-metric:hover {
  border-color: var(--overview-pending);
}
.review-notice {
  padding: 13px 16px;
  border-radius: 7px;
  border: 1px solid var(--overview-pending);
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.review-notice span:last-child {
  color: var(--overview-pending);
  white-space: nowrap;
}
.panel {
  padding: 20px;
  min-width: 0;
}
.panel-heading {
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.panel-heading a {
  font-size: 12px;
  color: var(--overview-accent);
}
.trend-panel {
  margin-bottom: 20px;
}
.legend {
  display: flex;
  gap: 7px;
  align-items: center;
  color: var(--at-color-sub);
  font-size: 12px;
}
.published-dot,
.pending-dot {
  width: 8px;
  height: 8px;
  border-radius: 2px;
}
.published-dot {
  background: var(--overview-accent);
}
.pending-dot {
  background: var(--overview-pending);
  margin-left: 9px;
}
.trend {
  display: block;
  width: 100%;
  margin-top: 14px;
  overflow: visible;
}
.grid line {
  stroke: var(--at-color-border);
  stroke-dasharray: 3 4;
}
.grid text,
.date-label {
  fill: var(--at-color-sub);
  font-size: 10px;
}
.published-bar {
  fill: var(--overview-accent);
}
.pending-bar {
  fill: var(--overview-pending);
}
.chart-empty {
  text-align: center;
  color: var(--at-color-sub);
  margin-top: 14px !important;
}
.daily-data {
  font-size: 12px;
  color: var(--at-color-sub);
  margin-top: 8px;
}
.daily-data summary {
  cursor: pointer;
  width: fit-content;
}
.daily-data table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 12px;
  text-align: left;
}
.daily-data td,
.daily-data th {
  padding: 7px;
  border-bottom: 1px solid var(--at-color-border);
}
.lists {
  display: grid;
  grid-template-columns: 1.15fr 1fr;
  gap: 20px;
  align-items: start;
}
.comment-list,
.popular-list {
  margin: 0;
  padding: 0;
  list-style: none;
}
.comment-list li {
  padding: 16px 0;
}
.comment-list li + li,
.popular-list li + li {
  border-top: 1px solid var(--at-color-border);
}
.comment-list li:last-child {
  padding-bottom: 0;
}
.comment-heading strong {
  overflow-wrap: anywhere;
  min-width: 0;
}
.status {
  font-size: 11px;
  color: var(--at-color-sub);
  padding: 3px 6px;
  border-radius: 4px;
  background: var(--at-color-bg-grey);
  white-space: nowrap;
}
.status.is-pending {
  color: var(--overview-pending);
}
.excerpt {
  line-height: 1.6;
  margin: 9px 0 !important;
  overflow-wrap: anywhere;
}
.page-link {
  display: block;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--at-color-sub) !important;
}
.meta {
  margin-top: 7px;
  font-size: 11px;
  flex-wrap: wrap;
}
.meta span {
  overflow-wrap: anywhere;
}
.popular-list li {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 17px 0;
}
.rank {
  color: var(--at-color-sub);
  font-variant-numeric: tabular-nums;
  width: 16px;
  flex-shrink: 0;
}
.popular-page {
  flex: 1;
  min-width: 0;
}
.popular-page a,
.popular-page > span {
  display: block;
  overflow-wrap: anywhere;
  line-height: 1.5;
}
.popular-page small {
  display: block;
  margin-top: 4px;
  overflow-wrap: anywhere;
}
.secondary-metrics {
  border-top: 1px solid var(--at-color-border);
  margin-top: 14px;
  padding-top: 18px;
  display: grid;
  gap: 14px;
}
.secondary-metrics > div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.secondary-metrics span {
  font-size: 12px;
  color: var(--at-color-sub);
}
.overview-footer {
  text-align: right;
  margin-top: 18px;
  line-height: 1.6;
}
.state,
.empty {
  text-align: center;
  padding: 35px 0;
  color: var(--at-color-sub);
}
.error p {
  margin-bottom: 15px;
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
@media (max-width: 1200px) {
  .metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .lists {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 600px) {
  .overview {
    padding: 18px 14px;
  }
  .panel {
    padding: 16px;
  }
  .metrics {
    gap: 10px;
  }
  .metric {
    padding: 14px;
  }
  .controls {
    gap: 8px;
  }
}
</style>
