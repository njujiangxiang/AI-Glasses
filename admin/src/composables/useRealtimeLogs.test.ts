import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { ApiError } from '@/api/client'
import { useRealtimeLogs, type RecentLogsResponse } from './useRealtimeLogs'

function response(stream: string, ids: number[]): RecentLogsResponse {
  return {
    stream_id: stream,
    entries: ids.map((id) => ({ id, time: `2026-06-20T00:00:0${id}.000Z`, level: 'LOG', source: 'test', message: `line ${id}` }))
  }
}

describe('useRealtimeLogs', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('loads initial logs and polls with after_id', async () => {
    const fetchLogs = vi.fn()
      .mockResolvedValueOnce(response('a', [1, 2]))
      .mockResolvedValueOnce(response('a', [3]))
    const logs = useRealtimeLogs({ fetchLogs })

    logs.start()
    await vi.runOnlyPendingTimersAsync()
    expect(logs.entries.value.map((entry) => entry.id)).toEqual([1, 2])
    expect(logs.lastId.value).toBe(2)

    await vi.advanceTimersByTimeAsync(2000)
    expect(fetchLogs.mock.calls[1][0]).toContain('after_id=2')
    expect(logs.entries.value.map((entry) => entry.id)).toEqual([1, 2, 3])
    logs.stop()
  })

  it('dedupes entries and caps displayed rows', async () => {
    const fetchLogs = vi.fn()
      .mockResolvedValueOnce(response('a', [1, 2, 3]))
      .mockResolvedValueOnce(response('a', [3, 4, 5]))
    const logs = useRealtimeLogs({ fetchLogs, maxEntries: 3 })

    logs.start()
    await vi.runOnlyPendingTimersAsync()
    await vi.advanceTimersByTimeAsync(2000)

    expect(logs.entries.value.map((entry) => entry.id)).toEqual([3, 4, 5])
    logs.stop()
  })

  it('resets state when stream_id changes', async () => {
    const fetchLogs = vi.fn()
      .mockResolvedValueOnce(response('a', [10, 11]))
      .mockResolvedValueOnce(response('b', [1, 2]))
    const logs = useRealtimeLogs({ fetchLogs })

    logs.start()
    await vi.runOnlyPendingTimersAsync()
    await vi.advanceTimersByTimeAsync(2000)

    expect(logs.currentStreamId.value).toBe('b')
    expect(logs.entries.value.map((entry) => entry.id)).toEqual([1, 2])
    expect(logs.lastId.value).toBe(2)
    logs.stop()
  })

  it('prevents overlapping requests', async () => {
    let resolveFetch!: (value: RecentLogsResponse) => void
    const fetchLogs = vi.fn(() => new Promise<RecentLogsResponse>((resolve) => { resolveFetch = resolve }))
    const logs = useRealtimeLogs({ fetchLogs })

    logs.start()
    await vi.advanceTimersByTimeAsync(0)
    void logs.refreshNow()
    expect(fetchLogs).toHaveBeenCalledTimes(1)

    resolveFetch(response('a', [1]))
    await nextTick()
    await Promise.resolve()
    logs.stop()
  })

  it('pauses, resumes, and cleans up timers', async () => {
    const fetchLogs = vi.fn()
      .mockResolvedValueOnce(response('a', [1]))
      .mockResolvedValueOnce(response('a', [2]))
    const logs = useRealtimeLogs({ fetchLogs })

    logs.start()
    await vi.runOnlyPendingTimersAsync()
    logs.setPaused(true)
    await vi.advanceTimersByTimeAsync(5000)
    expect(fetchLogs).toHaveBeenCalledTimes(1)

    logs.setPaused(false)
    await vi.runOnlyPendingTimersAsync()
    expect(fetchLogs).toHaveBeenCalledTimes(2)
    logs.stop()
    await vi.advanceTimersByTimeAsync(5000)
    expect(fetchLogs).toHaveBeenCalledTimes(2)
  })

  it('backs off on failure and resets after recovery', async () => {
    const fetchLogs = vi.fn()
      .mockRejectedValueOnce(new Error('network'))
      .mockResolvedValueOnce(response('a', [1]))
    const logs = useRealtimeLogs({ fetchLogs, baseDelayMs: 100, maxDelayMs: 500 })

    logs.start()
    await vi.runOnlyPendingTimersAsync()
    expect(logs.status.value).toBe('error')
    expect(logs.errorMessage.value).toBe('日志刷新失败，正在重试')

    await vi.advanceTimersByTimeAsync(200)
    expect(logs.status.value).toBe('polling')
    expect(logs.errorMessage.value).toBe('')
    logs.stop()
  })

  it('stops retrying on 403', async () => {
    const fetchLogs = vi.fn().mockRejectedValue(new ApiError(403, 'forbidden'))
    const logs = useRealtimeLogs({ fetchLogs, baseDelayMs: 100 })

    logs.start()
    await vi.runOnlyPendingTimersAsync()

    expect(logs.status.value).toBe('forbidden')
    expect(logs.noPermission.value).toBe(true)
    expect(logs.errorMessage.value).toBe('当前账号无权查看实时监控')
    await vi.advanceTimersByTimeAsync(1000)
    expect(fetchLogs).toHaveBeenCalledTimes(1)
  })
})
