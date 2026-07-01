<template>
  <div>
    <div class="stat-grid">
      <el-card shadow="never" v-for="item in stats" :key="item.label">
        <div class="stat-label">{{ item.label }}</div>
        <div class="stat-value">{{ item.value }}</div>
        <div class="stat-label">{{ item.note }}</div>
      </el-card>
    </div>
    <el-card shadow="never">
      <template #header><span class="card-title">常用功能</span></template>
      <el-space wrap>
        <el-button type="primary" @click="$router.push('/templates')">新建任务模板</el-button>
        <el-button @click="$router.push('/plans')">创建周期计划</el-button>
        <el-button @click="$router.push('/tasks')">查看任务进度</el-button>
        <el-button @click="$router.push('/defects')">处理异常缺陷</el-button>
      </el-space>
    </el-card>
  </div>
</template>

<script setup lang="ts">
const stats = [
  { label: '今日任务', value: '0', note: '由计划自动生成' },
  { label: '执行中', value: '0', note: '眼镜端已开始' },
  { label: '逾期任务', value: '0', note: '每 5 分钟检查' },
  { label: '待确认缺陷', value: '0', note: '异常上报生成' }
]
</script>
