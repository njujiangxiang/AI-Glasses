import { computed, ref } from 'vue'
import { ApiError, apiGet } from '@/api/client'

export type LogEntry = {
  id: number
  time: string
  level: string
  source: string
  message: string
}

export type RecentLogsResponse = {
  stream_id: string
  entries: LogEntry[]
  gap?: boolean
  skipped?: number
  oldest_id?: number
  newest_id?: number
}

type Status = 'idle' | 'polling' | 'paused' | 'error' | 'forbidden'

type Options = {
  initialLimit?: number
  maxEntries?: number
  baseDelayMs?: number
  maxDelayMs?: number
  fetchLogs?: (path: string, init?: RequestInit) => Promise<RecentLogsResponse>
}

export function useRealtimeLogs(options: Options = {}) {
  const initialLimit = options.initialLimit ?? 200
  const maxEntries = options.maxEntries ?? 1000
  const baseDelayMs = options.baseDelayMs ?? 2000
  const maxDelayMs = options.maxDelayMs ?? 15000
  const fetchLogs = options.fetchLogs ?? ((path: string, init?: RequestInit) => apiGet<RecentLogsResponse>(path, init))

  const entries = ref<LogEntry[]>([])
  const keyword = ref('')
  const paused = ref(false)
  const autoScroll = ref(true)
  const loading = ref(false)
  const status = ref<Status>('idle')
  const errorMessage = ref('')
  const noPermission = ref(false)
  const currentStreamId = ref('')
  const lastId = ref(0)
  const lastLogTime = ref('')
  const gap = ref(false)
  const skipped = ref(0)

  let timer: ReturnType<typeof setTimeout> | null = null
  let inFlight = false
  let stopped = true
  let failures = 0
  let abortController: AbortController | null = null
  let seenIDs = new Set<number>()

  const filteredEntries = computed(() => {
    const needle = keyword.value.trim().toLowerCase()
    if (!needle) return entries.value
    return entries.value.filter((entry) =>
      [entry.time, entry.level, entry.source, entry.message].some((value) => value.toLowerCase().includes(needle))
    )
  })

  function start() {
    if (!stopped) return
    stopped = false
    paused.value = false
    schedule(0)
  }

  function stop() {
    stopped = true
    clearTimer()
    abortController?.abort()
    abortController = null
    inFlight = false
    loading.value = false
    if (status.value !== 'forbidden') status.value = 'idle'
  }

  async function refreshNow() {
    if (stopped || paused.value || inFlight || status.value === 'forbidden') return
    inFlight = true
    loading.value = true
    status.value = 'polling'
    abortController = new AbortController()
    try {
      const query = lastId.value > 0 ? `limit=${initialLimit}&after_id=${lastId.value}` : `limit=${initialLimit}`
      const response = await fetchLogs(`/api/admin/monitoring/logs/recent?${query}`, { signal: abortController.signal })
      if (stopped || paused.value) return
      applyResponse(response)
      failures = 0
      errorMessage.value = ''
      noPermission.value = false
      status.value = 'polling'
      schedule(baseDelayMs)
    } catch (error) {
      if (stopped || paused.value || isAbortError(error)) return
      if (error instanceof ApiError && error.status === 403) {
        noPermission.value = true
        errorMessage.value = '当前账号无权查看实时监控'
        status.value = 'forbidden'
        clearTimer()
        return
      }
      failures++
      status.value = 'error'
      errorMessage.value = '日志刷新失败，正在重试'
      schedule(Math.min(maxDelayMs, baseDelayMs * 2 ** failures))
    } finally {
      inFlight = false
      loading.value = false
      abortController = null
    }
  }

  function clear() {
    entries.value = []
    seenIDs = new Set<number>()
    lastId.value = 0
    lastLogTime.value = ''
    gap.value = false
    skipped.value = 0
  }

  function setPaused(value: boolean) {
    paused.value = value
    if (value) {
      clearTimer()
      abortController?.abort()
      status.value = 'paused'
      return
    }
    if (!stopped && status.value !== 'forbidden') {
      schedule(0)
    }
  }

  function applyResponse(response: RecentLogsResponse) {
    if (currentStreamId.value && response.stream_id !== currentStreamId.value) {
      clear()
    }
    currentStreamId.value = response.stream_id
    gap.value = Boolean(response.gap)
    skipped.value = response.skipped ?? 0
    const next = [...entries.value]
    for (const entry of response.entries || []) {
      if (seenIDs.has(entry.id)) continue
      seenIDs.add(entry.id)
      next.push(entry)
      if (entry.id > lastId.value) lastId.value = entry.id
      lastLogTime.value = entry.time
    }
    entries.value = next.slice(-maxEntries)
    seenIDs = new Set(entries.value.map((entry) => entry.id))
  }

  function schedule(delay: number) {
    clearTimer()
    if (stopped || paused.value || status.value === 'forbidden') return
    timer = setTimeout(() => void refreshNow(), delay)
  }

  function clearTimer() {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  return {
    entries,
    filteredEntries,
    keyword,
    paused,
    autoScroll,
    loading,
    status,
    errorMessage,
    noPermission,
    currentStreamId,
    lastId,
    lastLogTime,
    gap,
    skipped,
    start,
    stop,
    refreshNow,
    clear,
    setPaused
  }
}

function isAbortError(error: unknown) {
  return error instanceof DOMException && error.name === 'AbortError'
}
