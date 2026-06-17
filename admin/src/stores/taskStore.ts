
// Task Sheet Store - simple localStorage persistence
import { reactive, watch } from 'vue'

export interface TaskSheet {
  id: number
  code: string
  name: string
  orgCode: string
  orgName: string
  planDate: string
  ownerName: string
  priority: 'normal' | 'urgent'
  estimatedHours: number
  remark: string
  status: 'draft' | 'submitted' | 'completed' | 'voided'
  detailCount: number
  updatedAt: string
  details: TaskDetail[]
}

export interface TaskDetail {
  localId: number
  pointName: string
  deviceName: string
  workContent: string
  standardHours: number
  riskLevel: 'low' | 'medium' | 'high'
}

const STORAGE_KEY = 'task-sheets-data'

// Initial sample data
const initialData: TaskSheet[] = [
  {
    id: 1,
    code: 'TASK-20260614-001',
    name: 'A区变电站 AI眼镜巡检作业',
    orgCode: 'ROOT',
    orgName: '默认单位',
    planDate: '2026-06-15',
    ownerName: '巡检班组长',
    priority: 'normal',
    estimatedHours: 2,
    remark: '明细中高风险点位提交前需要班组长复核。',
    status: 'draft',
    detailCount: 3,
    updatedAt: '2026-06-14 14:32',
    details: [
      { localId: 1, pointName: 'A区一号柜', deviceName: '开关柜 KYN28', workContent: '红外测温与外观检查', standardHours: 0.5, riskLevel: 'medium' },
      { localId: 2, pointName: '', deviceName: '主变压器', workContent: '油温与声音巡检', standardHours: 1, riskLevel: 'high' },
      { localId: 3, pointName: '电缆夹层', deviceName: '电缆桥架', workContent: '积水与异物检查', standardHours: 0.5, riskLevel: 'medium' }
    ]
  },
  {
    id: 2,
    code: 'TASK-20260613-009',
    name: 'B区电缆夹层复检',
    orgCode: 'ROOT',
    orgName: '默认单位',
    planDate: '2026-06-14',
    ownerName: '巡检员',
    priority: 'urgent',
    estimatedHours: 3,
    remark: '',
    status: 'submitted',
    detailCount: 5,
    updatedAt: '2026-06-13 18:20',
    details: [
      { localId: 1, pointName: 'B区一层', deviceName: '电缆支架', workContent: '外观检查', standardHours: 0.5, riskLevel: 'low' },
      { localId: 2, pointName: 'B区二层', deviceName: '高压柜', workContent: '红外测温', standardHours: 1, riskLevel: 'medium' },
      { localId: 3, pointName: 'B区三层', deviceName: '变压器', workContent: '声音检测', standardHours: 1, riskLevel: 'high' },
      { localId: 4, pointName: 'B区四层', deviceName: '开关柜', workContent: '状态检查', standardHours: 0.5, riskLevel: 'low' },
      { localId: 5, pointName: 'B区五层', deviceName: '电缆沟', workContent: '积水检查', standardHours: 1, riskLevel: 'medium' }
    ]
  },
  {
    id: 3,
    code: 'TASK-20260612-006',
    name: '主变区域常规巡视',
    orgCode: 'ROOT',
    orgName: '默认单位',
    planDate: '2026-06-12',
    ownerName: '巡检班组长',
    priority: 'normal',
    estimatedHours: 4,
    remark: '',
    status: 'completed',
    detailCount: 4,
    updatedAt: '2026-06-12 17:10',
    details: []
  }
]

// Load from localStorage or use initial data
function loadData(): TaskSheet[] {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored)
    }
  } catch (e) {
    console.error('Failed to load task sheets:', e)
  }
  return initialData
}

// Save to localStorage
function saveData(data: TaskSheet[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  } catch (e) {
    console.error('Failed to save task sheets:', e)
  }
}

// Generate task code
function generateTaskCode(): string {
  const now = new Date()
  const dateStr = now.getFullYear().toString() +
    (now.getMonth() + 1).toString().padStart(2, '0') +
    now.getDate().toString().padStart(2, '0')
  const todayTasks = taskSheets.filter(t => t.code.startsWith(`TASK-${dateStr}-`))
  const seq = (todayTasks.length + 1).toString().padStart(3, '0')
  return `TASK-${dateStr}-${seq}`
}

// Format date
function formatDateTime(date: Date = new Date()): string {
  return `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

// Reactive state
const taskSheets = reactive<TaskSheet[]>(loadData())

// Watch for changes and auto-save
watch(taskSheets, () => {
  saveData(taskSheets)
}, { deep: true })

// Store operations
export const taskStore = {
  // Get all task sheets
  getAll(): TaskSheet[] {
    return taskSheets
  },

  // Get by id
  getById(id: number): TaskSheet | undefined {
    return taskSheets.find(t => t.id === id)
  },

  // Create new draft
  create(): TaskSheet {
    const newId = Math.max(0, ...taskSheets.map(t => t.id)) + 1
    const newTask: TaskSheet = {
      id: newId,
      code: '', // Will be generated when saved
      name: '',
      orgCode: 'ROOT',
      orgName: '默认单位',
      planDate: new Date().toISOString().split('T')[0],
      ownerName: '巡检班组长',
      priority: 'normal',
      estimatedHours: 2,
      remark: '',
      status: 'draft',
      detailCount: 0,
      updatedAt: formatDateTime(),
      details: []
    }
    return newTask
  },

  // Save (insert or update)
  save(task: TaskSheet) {
    const index = taskSheets.findIndex(t => t.id === task.id)
    task.updatedAt = formatDateTime()
    task.detailCount = task.details.length

    if (index >= 0) {
      taskSheets[index] = { ...task }
    } else {
      if (!task.code) {
        task.code = generateTaskCode()
      }
      taskSheets.push({ ...task })
    }
  },

  // Delete
  delete(id: number) {
    const index = taskSheets.findIndex(t => t.id === id)
    if (index >= 0) {
      taskSheets.splice(index, 1)
    }
  },

  // Copy
  copy(fromId: number): TaskSheet {
    const source = taskSheets.find(t => t.id === fromId)
    if (!source) {
      throw new Error('Source task not found')
    }
    const newId = Math.max(0, ...taskSheets.map(t => t.id)) + 1
    const copied: TaskSheet = {
      ...source,
      id: newId,
      code: '',
      name: source.name ? `${source.name} (复制)` : '',
      status: 'draft',
      updatedAt: formatDateTime(),
      details: source.details.map(d => ({ ...d, localId: Date.now() + Math.random() }))
    }
    return copied
  },

  // Submit (change status to submitted)
  submit(id: number) {
    const task = taskSheets.find(t => t.id === id)
    if (task) {
      task.status = 'submitted'
      task.updatedAt = formatDateTime()
      if (!task.code) {
        task.code = generateTaskCode()
      }
    }
  },

  // Withdraw (change status back to draft)
  withdraw(id: number) {
    const task = taskSheets.find(t => t.id === id)
    if (task && task.status === 'submitted') {
      task.status = 'draft'
      task.updatedAt = formatDateTime()
    }
  }
}
