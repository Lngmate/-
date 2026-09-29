<template>
  <div class="dashboard-page">
    <!-- 顶部欢迎与指标卡片 -->
    <div class="metrics-row">
      <el-card class="metric-card" shadow="hover">
        <div class="metric-icon bg-blue">
          <el-icon><Cpu /></el-icon>
        </div>
        <div class="metric-info">
          <div class="metric-label">在册设备总数</div>
          <div class="metric-val">{{ stats.summary?.total_equipments || 5 }} <span class="unit">台</span></div>
        </div>
      </el-card>

      <el-card class="metric-card" shadow="hover">
        <div class="metric-icon bg-green">
          <el-icon><VideoPlay /></el-icon>
        </div>
        <div class="metric-info">
          <div class="metric-label">当前运行中</div>
          <div class="metric-val text-green">{{ stats.summary?.active_equipments || 1 }} <span class="unit">台</span></div>
        </div>
      </el-card>

      <el-card class="metric-card" shadow="hover">
        <div class="metric-icon bg-purple">
          <el-icon><Calendar /></el-icon>
        </div>
        <div class="metric-info">
          <div class="metric-label">今日预约工单</div>
          <div class="metric-val">{{ stats.summary?.today_reservations || 2 }} <span class="unit">单</span></div>
        </div>
      </el-card>

      <el-card class="metric-card" shadow="hover">
        <div class="metric-icon bg-orange">
          <el-icon><Timer /></el-icon>
        </div>
        <div class="metric-info">
          <div class="metric-label">历史累计机时</div>
          <div class="metric-val">1250.5 <span class="unit">小时</span></div>
        </div>
      </el-card>
    </div>

    <!-- 创新点成效亮点卡片：绿色节能与伴热拼单 -->
    <el-card class="eco-banner" shadow="never">
      <div class="eco-content">
        <div class="eco-left">
          <div class="eco-tag">
            <el-icon><Sunny /></el-icon> 特色创新：伴热/同参拼单节能成效
          </div>
          <div class="eco-desc">
            系统智能识别高温、真空相近实验进行伴热合并排期，大幅减少设备升降温反复耗电。已累计促成
            <strong>{{ stats.green_energy?.batched_count || 3 }}</strong> 次绿色拼单！
          </div>
        </div>
        <div class="eco-stats">
          <div class="stat-pill">
            <div class="num">{{ stats.green_energy?.saved_power_kwh || 55.5 }} <span>kWh</span></div>
            <div class="lbl">节约电能折算</div>
          </div>
          <div class="stat-pill">
            <div class="num">¥ {{ stats.green_energy?.saved_money || 135 }}</div>
            <div class="lbl">减免师生机时费</div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 图表网格 -->
    <div class="charts-grid">
      <!-- 设备利用率排行榜 -->
      <el-card class="chart-card" shadow="hover">
        <template #header>
          <div class="chart-header">
            <span><el-icon><Trophy /></el-icon> 仪器设备利用率排行 (累计运行机时)</span>
          </div>
        </template>
        <div ref="rankChartRef" class="chart-box"></div>
      </el-card>

      <!-- 课题组机时占比 -->
      <el-card class="chart-card" shadow="hover">
        <template #header>
          <div class="chart-header">
            <span><el-icon><PieChart /></el-icon> 课题组机时消耗占比分布</span>
          </div>
        </template>
        <div ref="groupChartRef" class="chart-box"></div>
      </el-card>

      <!-- 7日预约热度趋势 -->
      <el-card class="chart-card" shadow="hover">
        <template #header>
          <div class="chart-header">
            <span><el-icon><TrendCharts /></el-icon> 预约需求近7日热度走势</span>
          </div>
        </template>
        <div ref="trendChartRef" class="chart-box"></div>
      </el-card>

      <!-- 设备状态健康分布 -->
      <el-card class="chart-card" shadow="hover">
        <template #header>
          <div class="chart-header">
            <span><el-icon><DataAnalysis /></el-icon> 设备实时运行与就绪状态</span>
          </div>
        </template>
        <div ref="statusChartRef" class="chart-box"></div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import request from '../api/request'
import {
  Cpu,
  VideoPlay,
  Calendar,
  Timer,
  Sunny,
  Trophy,
  PieChart,
  TrendCharts,
  DataAnalysis
} from '@element-plus/icons-vue'

const stats = ref({
  summary: { total_equipments: 5, active_equipments: 1, total_reservations: 3, today_reservations: 2 },
  green_energy: { batched_count: 2, saved_power_kwh: 37.0, saved_money: 90.0 }
})

const rankChartRef = ref(null)
const groupChartRef = ref(null)
const trendChartRef = ref(null)
const statusChartRef = ref(null)

let chartInstances = []

const fetchDashboardData = async () => {
  try {
    const res = await request.get('/stats/dashboard')
    if (res.data) {
      stats.value = res.data
      initCharts(res.data)
    }
  } catch (err) {
    // 降级使用基础图表
    initFallbackCharts()
  }
}

const initCharts = (data) => {
  chartInstances.forEach(c => c.dispose())
  chartInstances = []

  // 1. 排行榜
  if (rankChartRef.value) {
    const c1 = echarts.init(rankChartRef.value)
    chartInstances.push(c1)
    const rankData = data.equipment_rank || [
      { name: '管式炉', total_hours: 420 },
      { name: 'FE-SEM', total_hours: 340.5 },
      { name: 'XRD', total_hours: 215 },
      { name: 'SFC-MS', total_hours: 180 },
      { name: '离心机', total_hours: 95 }
    ]
    c1.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: '3%', right: '8%', bottom: '3%', top: '5%', containLabel: true },
      xAxis: { type: 'value', name: '小时' },
      yAxis: { type: 'category', data: rankData.map(d => d.name).reverse() },
      series: [{
        type: 'bar',
        data: rankData.map(d => d.total_hours).reverse(),
        itemStyle: {
          color: new echarts.graphic.LinearGradient(1, 0, 0, 0, [
            { offset: 0, color: '#1890ff' },
            { offset: 1, color: '#69c0ff' }
          ]),
          borderRadius: [0, 4, 4, 0]
        },
        label: { show: true, position: 'right', formatter: '{c} h' }
      }]
    })
  }

  // 2. 课题组饼图
  if (groupChartRef.value) {
    const c2 = echarts.init(groupChartRef.value)
    chartInstances.push(c2)
    const groupData = (data.group_hours || [
      { name: '先进纳米材料组', total_hours_used: 124.5 },
      { name: '生物催化组', total_hours_used: 156 },
      { name: '超导物理组', total_hours_used: 89 }
    ]).map(g => ({ name: g.name, value: g.total_hours_used }))

    c2.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c} 小时 ({d}%)' },
      legend: { bottom: '0%' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        data: groupData
      }]
    })
  }

  // 3. 7日趋势
  if (trendChartRef.value) {
    const c3 = echarts.init(trendChartRef.value)
    chartInstances.push(c3)
    const trendData = data.trends || [
      { date: '09-14', count: 3 },
      { date: '09-15', count: 5 },
      { date: '09-16', count: 4 },
      { date: '09-17', count: 7 },
      { date: '09-18', count: 6 },
      { date: '09-19', count: 8 },
      { date: '09-20', count: 9 }
    ]
    c3.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '5%', bottom: '5%', top: '10%', containLabel: true },
      xAxis: { type: 'category', data: trendData.map(t => t.date) },
      yAxis: { type: 'value', minInterval: 1 },
      series: [{
        type: 'line',
        smooth: true,
        data: trendData.map(t => t.count),
        itemStyle: { color: '#52c41a' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(82, 196, 26, 0.4)' },
            { offset: 1, color: 'rgba(82, 196, 26, 0.05)' }
          ])
        }
      }]
    })
  }

  // 4. 设备状态环形图
  if (statusChartRef.value) {
    const c4 = echarts.init(statusChartRef.value)
    chartInstances.push(c4)
    c4.setOption({
      tooltip: { trigger: 'item' },
      legend: { orient: 'vertical', right: '5%', top: 'center' },
      series: [{
        type: 'pie',
        radius: ['50%', '75%'],
        center: ['40%', '50%'],
        data: [
          { value: 3, name: '空闲就绪', itemStyle: { color: '#67C23A' } },
          { value: 1, name: '实验运行中', itemStyle: { color: '#409EFF' } },
          { value: 1, name: '已有预约', itemStyle: { color: '#E6A23C' } }
        ]
      }]
    })
  }
}

const initFallbackCharts = () => {
  initCharts({
    equipment_rank: [
      { name: '管式炉', total_hours: 420 },
      { name: 'FE-SEM', total_hours: 340.5 },
      { name: 'XRD', total_hours: 215 },
      { name: 'SFC-MS', total_hours: 180 },
      { name: '离心机', total_hours: 95 }
    ]
  })
}

const handleResize = () => {
  chartInstances.forEach(c => c.resize())
}

onMounted(() => {
  fetchDashboardData()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chartInstances.forEach(c => c.dispose())
})
</script>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.metrics-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.metric-card {
  border-radius: 8px;
}

:deep(.metric-card .el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px;
}

.metric-icon {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  color: #fff;
}

.bg-blue { background: linear-gradient(135deg, #1890ff, #36cfc9); }
.bg-green { background: linear-gradient(135deg, #52c41a, #73d13d); }
.bg-purple { background: linear-gradient(135deg, #722ed1, #9254de); }
.bg-orange { background: linear-gradient(135deg, #fa8c16, #ffa940); }

.metric-info {
  display: flex;
  flex-direction: column;
}

.metric-label {
  font-size: 13px;
  color: #8c8c8c;
}

.metric-val {
  font-size: 24px;
  font-weight: bold;
  color: #262626;
  margin-top: 4px;
}

.metric-val .unit {
  font-size: 13px;
  font-weight: normal;
  color: #8c8c8c;
}

.text-green { color: #52c41a; }

/* 绿色节能卡片 */
.eco-banner {
  background: linear-gradient(135deg, #f6ffed 0%, #e6f7ff 100%);
  border: 1px solid #b7eb8f;
  border-radius: 8px;
}

.eco-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.eco-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #52c41a;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 20px;
  margin-bottom: 8px;
}

.eco-desc {
  font-size: 14px;
  color: #389e0d;
}

.eco-stats {
  display: flex;
  gap: 16px;
}

.stat-pill {
  background: #ffffff;
  border: 1px solid #d9f7be;
  border-radius: 8px;
  padding: 10px 18px;
  text-align: center;
  box-shadow: 0 2px 6px rgba(82, 196, 26, 0.1);
}

.stat-pill .num {
  font-size: 20px;
  font-weight: bold;
  color: #52c41a;
}

.stat-pill .num span {
  font-size: 12px;
}

.stat-pill .lbl {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 2px;
}

/* 图表网格 */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.chart-card {
  border-radius: 8px;
}

.chart-header {
  font-weight: 600;
  font-size: 14px;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 6px;
}

.chart-box {
  width: 100%;
  height: 280px;
}
</style>
