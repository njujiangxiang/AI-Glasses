<template>
  <el-card shadow="never" class="monitor-card">
    <div class="page-toolbar">
      <div>
        <span class="card-title">实时监控</span>
        <p class="page-subtitle">显示当前 API 进程最近日志，页面只读，不能执行命令。</p>
      </div>
      <el-space wrap>
        <el-tag :type="statusTagType">{{ statusText }}</el-tag>
        <span class="stat">{{ entries.length }} 条</span>
        <span class="stat">最后日志：{{ lastLogTime || '暂无' }}</span>
        <el-button @click="clear">清屏</el-button>
        <el-button :loading="loading" @click="refreshNow">刷新</el-button>
      </el-space>
    </div>

    <el-alert v-if="errorMessage" :title="errorMessage" :type="noPermission ? 'warning' : 'error'" show-icon :closable="false" class="monitor-alert" />
    <el-alert v-if="gap" :title="`日志缓存已滚动，可能跳过 ${skipped || '部分'} 条日志`" type="warning" show-icon :closable="false" class="monitor-alert" />

    <div class="monitor-controls">
      <el-input v-model="keyword" clearable placeholder="搜索时间、级别、来源或消息" class="keyword-input" />
      <el-space>
        <span>自动滚动</span>
        <el-switch v-model="autoScroll" />
        <span>暂停接收</span>
        <el-switch v-model="pausedModel" />
      </el-space>
    </div>

    <section ref="logPanel" aria-label="实时日志输出" class="log-panel">
      <el-empty v-if="filteredEntries.length === 0" description="暂无日志，等待后端输出…" />
      <div v-for="entry in filteredEntries" :key="`${currentStreamId}-${entry.id}`" class="log-line">
        <span class="log-time">{{ formatTime(entry.time) }}</span>
        <span class="log-level" :class="levelClass(entry.level)">{{ entry.level }}</span>
        <span class="log-source">{{ entry.source }}</span>
        <span class="log-message">{{ entry.message }}</span>
      </div>
    </section>
  </el-card>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRealtimeLogs } from '@/composables/useRealtimeLogs'

const logPanel = ref<HTMLElement | null>(null)
const {
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
  lastLogTime,
  gap,
  skipped,
  start,
  stop,
  refreshNow,
  clear,
  setPaused
} = useRealtimeLogs()

const pausedModel = computed({
  get: () => paused.value,
  set: (value: boolean) => setPaused(value)
})

const statusText = computed(() => {
  if (status.value === 'forbidden') return '无权限'
  if (status.value === 'paused') return '已暂停'
  if (status.value === 'error') return '请求失败'
  if (loading.value) return '刷新中'
  return '正常'
})

const statusTagType = computed(() => {
  if (status.value === 'forbidden' || status.value === 'paused') return 'warning'
  if (status.value === 'error') return 'danger'
  return 'success'
})

watch(
  () => entries.value.length,
  async () => {
    const panel = logPanel.value
    if (!panel || !autoScroll.value) return
    const nearBottom = panel.scrollHeight - panel.scrollTop - panel.clientHeight < 32
    if (!nearBottom) return
    await nextTick()
    panel.scrollTop = panel.scrollHeight
  }
)

onMounted(start)
onUnmounted(stop)

function formatTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString()
}

function levelClass(level: string) {
  const normalized = level.toUpperCase()
  return {
    'is-error': normalized.includes('ERROR'),
    'is-warn': normalized.includes('WARN'),
    'is-info': !normalized.includes('ERROR') && !normalized.includes('WARN')
  }
}
</script>

<style scoped>
.monitor-card {
  min-height: 100%;
}

.page-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 16px;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
}

.page-subtitle {
  margin: 6px 0 0;
  color: #6b7280;
  font-size: 13px;
}

.stat {
  color: #4b5563;
  font-size: 13px;
}

.monitor-alert {
  margin-bottom: 12px;
}

.monitor-controls {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
  margin-bottom: 12px;
}

.keyword-input {
  max-width: 360px;
}

.log-panel {
  min-height: 520px;
  max-height: calc(100vh - 280px);
  overflow: auto;
  border-radius: 10px;
  background: #0f172a;
  color: #d1d5db;
  padding: 14px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  font-size: 12px;
  line-height: 1.7;
}

.log-line {
  display: grid;
  grid-template-columns: 90px 72px 90px minmax(0, 1fr);
  gap: 10px;
  white-space: pre-wrap;
  word-break: break-word;
}

.log-time {
  color: #94a3b8;
}

.log-level {
  font-weight: 700;
}

.log-level.is-error {
  color: #f87171;
}

.log-level.is-warn {
  color: #facc15;
}

.log-level.is-info {
  color: #93c5fd;
}

.log-source {
  color: #a78bfa;
}

.log-message {
  color: #e5e7eb;
}
</style>
