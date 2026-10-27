<template>
  <div class="calendar-page">
    <el-card class="filter-card" shadow="never">
      <div class="filter-header">
        <div class="title-wrap"><el-icon><Calendar /></el-icon><h2>设备时间槽可视化排期看板</h2></div>
        <div class="right-actions">
          <el-select v-model="selectedEqId" filterable placeholder="选择实验设备" @change="fetchEvents">
            <el-option label="全部设备" :value="0" />
            <el-option v-for="eq in equipments" :key="eq.id" :label="eq.name" :value="eq.id" />
          </el-select>
          <el-button type="primary" plain :loading="loading" @click="fetchEvents"><el-icon><Refresh /></el-icon>刷新排期</el-button>
        </div>
      </div>
      <div class="week-toolbar">
        <div class="week-navigation"><el-button aria-label="上一周" @click="weekOffset--">上一周</el-button><el-button @click="weekOffset = 0">本周</el-button><el-button aria-label="下一周" @click="weekOffset++">下一周</el-button></div>
        <strong>{{ weekRange }}</strong>
      </div>
      <div class="legend-row">
        <span><i class="dot active"></i>实验进行中</span><span><i class="dot approved"></i>已通过待履约</span><span><i class="dot pending"></i>多级审批中</span><span><i class="dot completed"></i>已归还核销</span><span><i class="dot cancelled"></i>已取消或驳回</span>
      </div>
    </el-card>

    <el-card class="schedule-card" shadow="never" v-loading="loading">
      <el-alert v-if="loadError" :title="loadError" type="error" :closable="false" show-icon class="calendar-message" />
      <p class="calendar-hint">{{ weekEvents.length ? `本周共 ${weekEvents.length} 条预约，点击色块查看详情。` : '当前周暂无预约。' }} 时间均为本地时间。</p>
      <div class="calendar-scroll" tabindex="0" aria-label="设备周排期，可横向滚动">
        <div class="schedule-timeline">
          <div class="timeline-header">
            <div class="time-col-header">时间段</div>
            <div v-for="day in weekDays" :key="day.dateKey" class="day-col-header" :class="{ today: day.isToday }"><div>{{ day.name }}</div><time>{{ day.dateLabel }}</time></div>
          </div>
          <div class="timeline-body">
            <div v-for="slot in timeSlots" :key="slot" class="time-row">
              <div class="time-col">{{ slotLabel(slot) }}</div>
              <div v-for="day in weekDays" :key="day.dateKey + slot" class="slot-cell" :data-date="day.dateKey" :data-hour="slot">
                <button v-for="ev in getEventsForSlot(day, slot)" :key="ev.id" type="button" class="event-block" :style="{ backgroundColor: ev.backgroundColor }" :title="`${ev.title}\n${formatTime(ev.start)} 至 ${formatTime(ev.end)}`" @click="showEventDetail(ev)">
                  <span class="ev-title">{{ ev.title }}</span>
                  <span class="ev-time">{{ eventTimeLabel(ev, day, slot) }}</span>
                  <span v-if="ev.extendedProps?.is_batched" class="batch-badge">伴热拼单 8.5 折</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="预约工单详情" width="min(540px, 92vw)" destroy-on-close>
      <el-descriptions v-if="selectedEvent" :column="1" border size="small" class="event-detail">
        <el-descriptions-item label="工单编号">#{{ selectedEvent.id }}</el-descriptions-item>
        <el-descriptions-item label="设备名称">{{ selectedEvent.extendedProps?.equipment_name }}</el-descriptions-item>
        <el-descriptions-item label="课题实验">{{ selectedEvent.title }}</el-descriptions-item>
        <el-descriptions-item label="预约人员">{{ selectedEvent.extendedProps?.user_name }}</el-descriptions-item>
        <el-descriptions-item label="预约开始">{{ formatTime(selectedEvent.start) }}</el-descriptions-item>
        <el-descriptions-item label="预约结束">{{ formatTime(selectedEvent.end) }}</el-descriptions-item>
        <el-descriptions-item label="当前状态"><el-tag :type="getStatusTag(selectedEvent.extendedProps?.status)">{{ getStatusText(selectedEvent.extendedProps?.status) }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="伴热拼单">{{ selectedEvent.extendedProps?.is_batched ? '已享受 8.5 折拼单优惠' : '常规排期' }}</el-descriptions-item>
        <el-descriptions-item label="结算费用">¥{{ Number(selectedEvent.extendedProps?.total_cost || 0).toFixed(2) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import request from '../api/request'
import { Calendar, Refresh } from '@element-plus/icons-vue'

const selectedEqId = ref(0)
const equipments = ref([])
const events = ref([])
const detailVisible = ref(false)
const selectedEvent = ref(null)
const loading = ref(false)
const loadError = ref('')
const weekOffset = ref(0)
const timeSlots = Array.from({ length: 12 }, (_, index) => index * 2)
let fetchVersion = 0

const localDateKey = date => `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
const weekDays = computed(() => {
  const now = new Date()
  const monday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  monday.setDate(monday.getDate() - (monday.getDay() + 6) % 7 + weekOffset.value * 7)
  return Array.from({ length: 7 }, (_, index) => {
    const date = new Date(monday)
    date.setDate(date.getDate() + index)
    return { name: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'][index], date, dateKey: localDateKey(date), dateLabel: `${date.getMonth() + 1}月${date.getDate()}日`, isToday: localDateKey(date) === localDateKey(now) }
  })
})
const weekRange = computed(() => `${weekDays.value[0].dateKey} 至 ${weekDays.value[6].dateKey}`)
const weekEvents = computed(() => {
  const start = weekDays.value[0].date.getTime()
  const end = new Date(weekDays.value[6].date)
  end.setDate(end.getDate() + 1)
  return events.value.filter(event => new Date(event.start).getTime() < end.getTime() && new Date(event.end).getTime() > start)
})
const slotBounds = (day, slot) => {
  const start = new Date(day.date)
  const end = new Date(day.date)
  start.setHours(slot, 0, 0, 0)
  end.setHours(slot + 2, 0, 0, 0)
  return [start.getTime(), end.getTime()]
}
const getEventsForSlot = (day, slot) => {
  const [start, end] = slotBounds(day, slot)
  return weekEvents.value.filter(event => new Date(event.start).getTime() < end && new Date(event.end).getTime() > start)
}
const clock = date => `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
const slotLabel = slot => `${String(slot).padStart(2, '0')}:00–${String(slot + 2).padStart(2, '0')}:00`
const eventTimeLabel = (event, day, slot) => {
  const [start, end] = slotBounds(day, slot)
  const eventStart = new Date(event.start).getTime()
  const eventEnd = new Date(event.end).getTime()
  return `${eventStart < start ? '续 ' : ''}${clock(new Date(Math.max(start, eventStart)))}–${eventEnd >= end && slot === 22 ? '24:00' : clock(new Date(Math.min(end, eventEnd)))}${eventEnd > end ? ' 续' : ''}`
}

async function fetchEvents() {
  const version = ++fetchVersion
  loading.value = true
  loadError.value = ''
  try {
    const params = selectedEqId.value ? { equipment_id: selectedEqId.value } : {}
    const response = await request.get('/reservations/events', { params })
    if (version !== fetchVersion) return
    events.value = (response.data || []).filter(event => Number.isFinite(new Date(event.start).getTime()) && new Date(event.end) > new Date(event.start)).sort((a, b) => new Date(a.start) - new Date(b.start))
  } catch (error) {
    if (version !== fetchVersion) return
    events.value = []
    loadError.value = error.response?.data?.message || '读取排期失败，请刷新重试'
  } finally { if (version === fetchVersion) loading.value = false }
}
const showEventDetail = event => { selectedEvent.value = event; detailVisible.value = true }
const formatTime = value => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : ''
const getStatusText = status => ({ active: '实验进行中', approved: '审批通过待履约', pending_admin: '待管理员审批', pending_advisor: '待导师审批', completed: '已归还', cancelled: '已取消', rejected: '已驳回', no_show: '已爽约' }[status] || status)
const getStatusTag = status => ({ active: 'primary', approved: 'success', pending_admin: 'warning', pending_advisor: 'warning', completed: 'info', cancelled: 'info', rejected: 'danger', no_show: 'danger' }[status] || 'info')
onMounted(async () => {
  await Promise.all([
    fetchEvents(),
    request.get('/equipments').then(response => { equipments.value = response.data || [] }).catch(() => {})
  ])
})
</script>

<style scoped>
.calendar-page { display: flex; flex-direction: column; gap: 16px; min-width: 0; max-width: 100%; }
.filter-card, .schedule-card { min-width: 0; border-radius: 8px; }
.filter-header { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px; margin-bottom: 16px; }
.title-wrap { display: flex; align-items: center; gap: 8px; min-width: 0; }
.title-wrap .el-icon { font-size: 20px; color: #1890ff; flex-shrink: 0; }
.title-wrap h2 { font-size: 16px; font-weight: 600; color: #2c3e50; margin: 0; }
.right-actions { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; min-width: 0; }
.right-actions .el-select { width: 280px; max-width: 100%; }
.week-toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 12px 20px; margin-bottom: 16px; font-size: 14px; color: #475569; }
.week-navigation { display: flex; gap: 8px; }
.week-navigation .el-button + .el-button { margin-left: 0; }
.legend-row { display: flex; flex-wrap: wrap; gap: 10px 16px; font-size: 12px; color: #606266; }
.legend-row > span { display: flex; align-items: center; gap: 6px; }
.dot { width: 10px; height: 10px; border-radius: 50%; display: inline-block; }
.dot.active { background: #409eff; } .dot.approved { background: #67c23a; } .dot.pending { background: #e6a23c; } .dot.completed { background: #909399; } .dot.cancelled { background: #c0c4cc; }
.calendar-message { margin-bottom: 12px; }
.calendar-hint { color: #64748b; font-size: 13px; margin: 0 0 14px; }
.calendar-scroll { width: 100%; max-width: 100%; overflow-x: auto; border: 1px solid #e2e8f0; border-radius: 6px; }
.calendar-scroll:focus-visible { outline: 2px solid #1890ff; outline-offset: 2px; }
.schedule-timeline { min-width: 1050px; width: 100%; }
.timeline-header, .time-row { display: grid; grid-template-columns: 100px repeat(7, minmax(0, 1fr)); }
.timeline-header { background: #f8fafc; border-bottom: 1px solid #e2e8f0; }
.time-col-header, .day-col-header { padding: 12px 6px; text-align: center; font-size: 13px; font-weight: 600; border-right: 1px solid #e2e8f0; min-width: 0; }
.day-col-header.today { background: #e6f7ff; color: #096dd9; }
.day-col-header time { display: block; font-size: 12px; font-weight: 400; color: #64748b; margin-top: 4px; }
.time-row { border-bottom: 1px solid #f0f2f5; min-height: 80px; }
.time-row:last-child { border-bottom: none; }
.time-col { padding: 10px 4px; font-size: 11px; color: #64748b; text-align: center; border-right: 1px solid #e2e8f0; background: #fafafa; display: flex; align-items: center; justify-content: center; }
.slot-cell { border-right: 1px solid #f0f2f5; padding: 5px; display: flex; flex-direction: column; gap: 5px; min-width: 0; max-height: 240px; overflow-y: auto; }
.slot-cell:last-child, .day-col-header:last-child { border-right: none; }
.event-block { display: block; flex-shrink: 0; width: 100%; min-width: 0; max-width: 100%; border: none; border-radius: 4px; padding: 6px; text-align: left; color: white; font: inherit; font-size: 11px; cursor: pointer; overflow: hidden; }
.event-block:hover { filter: brightness(.94); }
.event-block:focus-visible { outline: 2px solid #096dd9; outline-offset: 1px; }
.ev-title { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; overflow-wrap: anywhere; line-height: 1.5; }
.ev-time { display: block; font-size: 10px; margin-top: 4px; line-height: 1.5; }
.batch-badge { display: inline-block; max-width: 100%; background: #ffffff30; border-radius: 2px; padding: 1px 3px; font-size: 10px; margin-top: 4px; overflow-wrap: anywhere; }
.event-detail { overflow-wrap: anywhere; }
@media (max-width: 700px) { .right-actions { width: 100%; } .right-actions .el-select { flex: 1; min-width: 160px; } }
</style>
