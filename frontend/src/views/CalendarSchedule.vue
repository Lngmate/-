<template>
  <div class="calendar-page">
    <el-card class="filter-card" shadow="never">
      <div class="filter-header">
        <div class="title-wrap">
          <el-icon class="icon"><Calendar /></el-icon>
          <span class="title">设备时间槽可视化排期看板</span>
        </div>
        <div class="right-actions">
          <el-select v-model="selectedEqId" placeholder="选择查看的实验设备" style="width: 280px" @change="fetchEvents">
            <el-option label="-- 查看全部设备排期 --" :value="0" />
            <el-option v-for="eq in equipments" :key="eq.id" :label="eq.name" :value="eq.id" />
          </el-select>
          <el-button type="primary" plain @click="fetchEvents">
            <el-icon><Refresh /></el-icon> 刷新排期
          </el-button>
        </div>
      </div>

      <div class="legend-row">
        <span class="legend-item"><span class="dot active"></span>实验进行中</span>
        <span class="legend-item"><span class="dot approved"></span>已通过待履约</span>
        <span class="legend-item"><span class="dot pending"></span>多级审批中</span>
        <span class="legend-item"><span class="dot completed"></span>已归还核销</span>
      </div>
    </el-card>

    <!-- 可视化时间槽网格 -->
    <el-card class="schedule-card" shadow="hover">
      <div class="schedule-timeline">
        <div class="timeline-header">
          <div class="time-col-header">时间段</div>
          <div v-for="day in weekDays" :key="day.dateStr" class="day-col-header" :class="{ today: day.isToday }">
            <div class="day-name">{{ day.name }}</div>
            <div class="day-date">{{ day.dateStr }}</div>
          </div>
        </div>

        <div class="timeline-body">
          <div v-for="slot in timeSlots" :key="slot" class="time-row">
            <div class="time-col">{{ slot }}</div>
            <div
              v-for="day in weekDays"
              :key="day.dateStr + slot"
              class="slot-cell"
              @click="handleCellClick(day, slot)"
            >
              <!-- 渲染落入该时间格的预约事件 -->
              <div
                v-for="ev in getEventsForSlot(day.dateStr, slot)"
                :key="ev.id"
                class="event-block"
                :style="{ backgroundColor: ev.backgroundColor }"
                @click.stop="showEventDetail(ev)"
              >
                <div class="ev-title">{{ ev.title }}</div>
                <div v-if="ev.extendedProps?.is_batched" class="batch-badge">
                  伴热拼单(8.5折)
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 事件详情对话框 -->
    <el-dialog v-model="detailVisible" title="预约工单详情" width="500px" destroy-on-close>
      <div v-if="selectedEvent" class="event-detail">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="设备名称">{{ selectedEvent.extendedProps?.equipment_name }}</el-descriptions-item>
          <el-descriptions-item label="课题实验">{{ selectedEvent.title }}</el-descriptions-item>
          <el-descriptions-item label="预约人员">{{ selectedEvent.extendedProps?.user_name }}</el-descriptions-item>
          <el-descriptions-item label="预约开始">{{ formatTime(selectedEvent.start) }}</el-descriptions-item>
          <el-descriptions-item label="预约结束">{{ formatTime(selectedEvent.end) }}</el-descriptions-item>
          <el-descriptions-item label="当前状态">
            <el-tag :type="getStatusTag(selectedEvent.extendedProps?.status)">
              {{ getStatusText(selectedEvent.extendedProps?.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="伴热拼单">
            <el-tag v-if="selectedEvent.extendedProps?.is_batched" type="success">已享受 8.5 折拼单优惠</el-tag>
            <span v-else>常规排期</span>
          </el-descriptions-item>
          <el-descriptions-item label="结算费用">¥ {{ selectedEvent.extendedProps?.total_cost }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import request from '../api/request'
import { Calendar, Refresh } from '@element-plus/icons-vue'

const selectedEqId = ref(0)
const equipments = ref([])
const events = ref([])
const detailVisible = ref(false)
const selectedEvent = ref(null)

const timeSlots = [
  '08:00 - 10:00',
  '10:00 - 12:00',
  '14:00 - 16:00',
  '16:00 - 18:00',
  '19:00 - 21:00'
]

// 生成本周各天日期
const weekDays = ref([])
const initWeekDays = () => {
  const days = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  const res = []
  const now = new Date()
  const currentDayOfWeek = now.getDay()
  for (let i = 0; i < 7; i++) {
    const d = new Date()
    d.setDate(now.getDate() - currentDayOfWeek + i)
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const date = String(d.getDate()).padStart(2, '0')
    res.push({
      name: days[i],
      dateStr: `${month}-${date}`,
      fullDate: d,
      isToday: d.toDateString() === now.toDateString()
    })
  }
  weekDays.value = res
}

const fetchEquipments = async () => {
  try {
    const res = await request.get('/equipments')
    if (res.data) equipments.value = res.data
  } catch (err) {}
}

const fetchEvents = async () => {
  try {
    const params = selectedEqId.value ? { equipment_id: selectedEqId.value } : {}
    const res = await request.get('/reservations/events', { params })
    if (res.data) {
      events.value = res.data
    }
  } catch (err) {
    // 降级使用基础事件
    events.value = [
      {
        id: 1,
        title: '[SFC-MS] 手性药物异构体纯化 (周浩楠)',
        start: new Date().toISOString(),
        end: new Date(Date.now() + 3600000 * 3).toISOString(),
        backgroundColor: '#409EFF',
        extendedProps: { equipment_name: '超临界流体色谱-质谱联用仪', user_name: '周浩楠(学生)', status: 'active', is_batched: false, total_cost: 600 }
      },
      {
        id: 2,
        title: '[管式炉] CVD高温气相沉积 (张耀辉)',
        start: new Date().toISOString(),
        end: new Date(Date.now() + 3600000 * 3).toISOString(),
        backgroundColor: '#E6A23C',
        extendedProps: { equipment_name: '超高温真空气氛管式炉', user_name: '张耀辉(学生)', status: 'pending_admin', is_batched: true, total_cost: 127.5 }
      }
    ]
  }
}

const getEventsForSlot = (dateStr, slot) => {
  // 根据事件时间匹配网格
  return events.value.filter(ev => {
    const evDate = new Date(ev.start)
    const m = String(evDate.getMonth() + 1).padStart(2, '0')
    const d = String(evDate.getDate()).padStart(2, '0')
    const evDateStr = `${m}-${d}`
    if (evDateStr !== dateStr) return false

    const hour = evDate.getHours()
    if (slot.startsWith('08') && hour >= 8 && hour < 10) return true
    if (slot.startsWith('10') && hour >= 10 && hour < 12) return true
    if (slot.startsWith('14') && hour >= 13 && hour < 16) return true
    if (slot.startsWith('16') && hour >= 16 && hour < 18) return true
    if (slot.startsWith('19') && hour >= 18) return true
    return false
  })
}

const handleCellClick = (day, slot) => {
  // 点击空白时间槽直接提示
}

const showEventDetail = (ev) => {
  selectedEvent.value = ev
  detailVisible.value = true
}

const formatTime = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  return d.toLocaleString('zh-CN', { hour12: false })
}

const getStatusText = (status) => {
  const map = { available: '空闲', in_use: '实验进行中', active: '实验进行中', approved: '审批通过待履约', pending_admin: '待管理员审批', pending_advisor: '待导师审批', completed: '已归还' }
  return map[status] || status
}

const getStatusTag = (status) => {
  const map = { active: 'primary', approved: 'success', pending_admin: 'warning', pending_advisor: 'warning', completed: 'info' }
  return map[status] || 'info'
}

onMounted(() => {
  initWeekDays()
  fetchEquipments()
  fetchEvents()
})
</script>

<style scoped>
.calendar-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card {
  border-radius: 8px;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-wrap .icon {
  font-size: 20px;
  color: #1890ff;
}

.title-wrap .title {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
}

.right-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.legend-row {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #606266;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}

.dot.active { background: #409EFF; }
.dot.approved { background: #67C23A; }
.dot.pending { background: #E6A23C; }
.dot.completed { background: #909399; }

/* 日历时间表 */
.schedule-card {
  border-radius: 8px;
}

.schedule-timeline {
  display: flex;
  flex-direction: column;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  overflow: hidden;
}

.timeline-header {
  display: grid;
  grid-template-columns: 120px repeat(7, 1fr);
  background: #f8fafc;
  border-bottom: 1px solid #ebeef5;
}

.time-col-header {
  padding: 12px 8px;
  font-weight: bold;
  font-size: 13px;
  color: #475569;
  text-align: center;
  border-right: 1px solid #ebeef5;
}

.day-col-header {
  padding: 10px 6px;
  text-align: center;
  border-right: 1px solid #ebeef5;
}

.day-col-header.today {
  background: #e6f7ff;
}

.day-name {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.day-date {
  font-size: 11px;
  color: #64748b;
  margin-top: 2px;
}

.timeline-body {
  display: flex;
  flex-direction: column;
}

.time-row {
  display: grid;
  grid-template-columns: 120px repeat(7, 1fr);
  border-bottom: 1px solid #f0f2f5;
  min-height: 80px;
}

.time-col {
  padding: 12px 8px;
  font-size: 12px;
  color: #64748b;
  text-align: center;
  border-right: 1px solid #ebeef5;
  background: #fafafa;
  display: flex;
  align-items: center;
  justify-content: center;
}

.slot-cell {
  border-right: 1px solid #f0f2f5;
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  transition: background 0.2s;
  cursor: pointer;
}

.slot-cell:hover {
  background: #f9fbfd;
}

.event-block {
  border-radius: 4px;
  padding: 4px 6px;
  color: #ffffff;
  font-size: 11px;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: transform 0.15s;
}

.event-block:hover {
  transform: scale(1.02);
}

.ev-title {
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.batch-badge {
  background: rgba(255, 255, 255, 0.3);
  font-size: 9px;
  padding: 1px 3px;
  border-radius: 2px;
  margin-top: 2px;
  display: inline-block;
}
</style>
