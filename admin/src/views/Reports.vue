<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <span class="card-title">巡检报告</span>
      <div>
        <el-input v-model="filters.keyword" placeholder="搜索报告" clearable style="width: 200px" @keyup.enter="load" />
        <el-button @click="load" style="margin-left: 8px">刷新</el-button>
      </div>
    </div>
    <el-table :data="items" stripe row-key="id" v-loading="loading">
      <el-table-column type="index" label="序号" width="70" align="center" :index="indexMethod" />
      <el-table-column prop="task_name" label="任务名称" min-width="150" />
      <el-table-column prop="point_name" label="巡检点位" width="120" />
      <el-table-column prop="equipment_name" label="设备" width="120" />
      <el-table-column label="节点数" width="80" align="center">
        <template #default="scope">{{ scope.row.node_count }}</template>
      </el-table-column>
      <el-table-column label="执行人" width="100">
        <template #default="scope">{{ scope.row.executor_name || '-' }}</template>
      </el-table-column>
      <el-table-column label="开始时间" width="160">
        <template #default="scope">{{ formatTime(scope.row.started_at) }}</template>
      </el-table-column>
      <el-table-column label="完成时间" width="160">
        <template #default="scope">{{ formatTime(scope.row.completed_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="scope">
          <el-button link type="primary" @click="viewDetail(scope.row)">详情</el-button>
          <el-button link type="primary" @click="downloadPDF(scope.row)">下载PDF</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager" v-if="total > filters.page_size">
      <el-pagination v-model:current-page="filters.page" :page-size="filters.page_size" :total="total" layout="total, prev, pager, next" @current-change="load" />
    </div>
  </el-card>

  <!-- 报告详情抽屉 -->
  <el-drawer v-model="detailVisible" title="巡检报告详情" size="650px">
    <div v-if="detailData">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="任务名称">{{ detailData.task.task_name }}</el-descriptions-item>
        <el-descriptions-item label="巡检点位">{{ detailData.task.point_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="设备名称">{{ detailData.task.equipment_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="作业区域">{{ detailData.task.inspect_area || '-' }}</el-descriptions-item>
        <el-descriptions-item label="巡检模板">{{ detailData.template_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="指派人">{{ detailData.assignee_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="执行人">{{ detailData.executor_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="AR眼镜编号">{{ detailData.task.glasses_sn || '-' }}</el-descriptions-item>
        <el-descriptions-item label="下发人">{{ detailData.task.assign_user || '-' }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(detailData.task.started_at) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(detailData.task.completed_at) }}</el-descriptions-item>
        <el-descriptions-item label="节点数量">{{ detailData.node_count }}</el-descriptions-item>
      </el-descriptions>

      <h4 style="margin: 20px 0 10px">节点执行详情 ({{ detailData.nodes.length }})</h4>
      <el-collapse accordion>
        <el-collapse-item v-for="(node, idx) in detailData.nodes" :key="node.id" :name="node.id">
          <template #title>
            <div style="display: flex; align-items: center; gap: 8px; width: 100%">
              <span style="color: #909399; font-size: 12px">{{ idx + 1 }}</span>
              <span>{{ node.name }}</span>
              <el-tag :type="nodeStatusType(node.status)" size="small">{{ nodeStatusLabel(node.status) }}</el-tag>
              <el-tag v-if="node.defects && node.defects.length > 0" type="danger" size="small">缺陷 {{ node.defects.length }}</el-tag>
            </div>
          </template>
          <div style="padding: 8px 0">
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="节点类型">{{ nodeTypeLabel(node.node_type) }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="nodeStatusType(node.status)" size="small">{{ nodeStatusLabel(node.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="执行时间" :span="2">{{ node.actual_exec_time ? formatTime(node.actual_exec_time) : '-' }}</el-descriptions-item>
            </el-descriptions>

            <div v-if="node.result" style="margin-top: 12px">
              <h5 style="margin: 0 0 8px; color: #606266">执行结果</h5>
              <el-descriptions :column="1" border size="small">
                <el-descriptions-item v-if="node.result.feedback_content" label="反馈内容">{{ node.result.feedback_content }}</el-descriptions-item>
                <el-descriptions-item v-if="node.result.text_note" label="备注">{{ node.result.text_note }}</el-descriptions-item>
                <el-descriptions-item v-if="node.result.algorithm_result" label="AI分析结果">{{ node.result.algorithm_result }}</el-descriptions-item>
                <el-descriptions-item v-if="node.result.query_result" label="实时查询">{{ node.result.query_result }}</el-descriptions-item>
                <el-descriptions-item v-if="node.result.is_abnormal" label="异常描述">
                  <span style="color: #f56c6c">{{ node.result.abnormal_desc }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="完成时间">{{ formatTime(node.result.completed_at) }}</el-descriptions-item>
              </el-descriptions>

              <!-- 附件列表 -->
              <div v-if="node.result.attachments && node.result.attachments.length > 0" style="margin-top: 8px">
                <h5 style="margin: 0 0 6px; color: #606266">附件 ({{ node.result.attachments.length }})</h5>
                <div style="display: flex; flex-wrap: wrap; gap: 8px">
                  <el-tag v-for="att in node.result.attachments" :key="att.id" size="small">
                    {{ att.file_name || att.content_type }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div v-else style="margin-top: 8px; color: #909399; font-size: 13px">未提交执行结果</div>

            <!-- 节点缺陷 -->
            <div v-if="node.defects && node.defects.length > 0" style="margin-top: 12px">
              <h5 style="margin: 0 0 8px; color: #f56c6c">缺陷记录 ({{ node.defects.length }})</h5>
              <el-table :data="node.defects" stripe size="small">
                <el-table-column prop="id" label="ID" width="60" />
                <el-table-column prop="description" label="缺陷描述" min-width="150" />
                <el-table-column label="状态" width="80">
                  <template #default="scope">
                    <el-tag size="small">{{ defectStatusLabel(scope.row.status) }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="close_reason" label="关闭原因" width="120">
                  <template #default="scope">{{ scope.row.close_reason || '-' }}</template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>

      <!-- 缺陷汇总 -->
      <div v-if="detailData.defects && detailData.defects.length > 0" style="margin-top: 20px">
        <h4 style="margin: 0 0 10px">缺陷汇总 ({{ detailData.defects.length }})</h4>
        <el-table :data="detailData.defects" stripe size="small">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="description" label="缺陷描述" min-width="200" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag size="small">{{ defectStatusLabel(scope.row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="close_reason" label="关闭原因" min-width="120">
            <template #default="scope">{{ scope.row.close_reason || '-' }}</template>
          </el-table-column>
          <el-table-column label="创建时间" width="160">
            <template #default="scope">{{ formatTime(scope.row.created_at) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div style="margin-top: 20px; text-align: right">
        <el-button type="primary" @click="downloadPDF(detailData.task)">下载PDF报告</el-button>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiGet } from '@/api/client'

type ReportItem = {
  id: number
  task_name: string
  point_name: string
  equipment_name: string
  inspect_area: string
  template_name: string
  assignee_name: string
  executor_name: string
  node_count: number
  started_at: string | null
  completed_at: string | null
  status: string
}

type NodeAttachment = {
  id: number
  file_name: string
  content_type: string
  size_bytes: number
}

type NodeResult = {
  id: number
  status: string
  feedback_content: string
  text_note: string
  algorithm_result: string
  query_result: string
  is_abnormal: boolean
  abnormal_desc: string
  completed_at: string
  attachments: NodeAttachment[]
}

type NodeDefect = {
  id: number
  description: string
  status: string
  close_reason: string
  created_at: string
  confirmed_at: string | null
  closed_at: string | null
}

type ReportNode = {
  id: number
  sort_order: number
  name: string
  node_type: string
  status: string
  is_mandatory: boolean
  is_required: boolean
  actual_exec_time: string | null
  result: NodeResult | null
  defects: NodeDefect[]
}

type ReportDetail = {
  task: {
    id: number
    task_name: string
    point_name: string
    equipment_name: string
    inspect_area: string
    template_id: number
    status: string
    assignee_id: number
    executor_id: number | null
    glasses_sn: string
    assign_user: string
    started_at: string | null
    completed_at: string | null
    [key: string]: unknown
  }
  template_name: string
  assignee_name: string
  executor_name: string
  nodes: ReportNode[]
  defects: NodeDefect[]
  node_count: number
}

const items = ref<ReportItem[]>([])
const total = ref(0)
const loading = ref(false)
const filters = reactive({ keyword: '', page: 1, page_size: 20 })
function indexMethod(rowIndex: number) { return (filters.page - 1) * filters.page_size + rowIndex + 1 }

const detailVisible = ref(false)
const detailData = ref<ReportDetail | null>(null)

function formatTime(t: string | null | undefined) {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return t
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })
}

function nodeTypeLabel(t: string) {
  const map: Record<string, string> = { text: '文本', read: '读取', check: '检查', photo: '拍照', video: '录像', audio: '录音' }
  return map[t] || t
}

function nodeStatusLabel(s: string) {
  const map: Record<string, string> = { pending: '待执行', completed: '已完成', skipped: '已跳过', abnormal: '异常' }
  return map[s] || s
}

function nodeStatusType(s: string) {
  const map: Record<string, string> = { pending: 'info', completed: 'success', skipped: 'warning', abnormal: 'danger' }
  return map[s] || ''
}

function defectStatusLabel(s: string) {
  const map: Record<string, string> = { pending_confirm: '待确认', reported: '已上报', confirmed: '已确认', closed: '已关闭' }
  return map[s] || s
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.keyword) params.set('keyword', filters.keyword)
    params.set('page', String(filters.page))
    params.set('page_size', String(filters.page_size))
    const result = await apiGet<{ items: ReportItem[]; total: number }>(`/api/admin/reports?${params}`)
    items.value = result.items
    total.value = result.total
  } catch (e: any) {
    ElMessage.error(e.message || '加载报告列表失败')
  } finally {
    loading.value = false
  }
}

async function viewDetail(row: ReportItem) {
  try {
    const result = await apiGet<ReportDetail>(`/api/admin/reports/${row.id}`)
    detailData.value = result
    detailVisible.value = true
  } catch (e: any) {
    ElMessage.error(e.message || '加载报告详情失败')
  }
}

async function downloadPDF(row: { id: number; task_name?: string }) {
  try {
    const token = localStorage.getItem('admin_token')
    const res = await fetch(`/api/admin/reports/${row.id}/pdf`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    if (!res.ok) {
      const errData = await res.json().catch(() => null)
      throw new Error(errData?.error?.message || 'PDF生成失败')
    }
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const fileName = `巡检报告_${row.task_name || row.id}.pdf`
    a.download = fileName
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success('PDF下载成功')
  } catch (e: any) {
    ElMessage.error(e.message || 'PDF下载失败')
  }
}

onMounted(() => {
  load()
})
</script>

<style scoped>
.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
