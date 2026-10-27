<template>
  <el-dialog :model-value="modelValue" title="预约工单详情" width="min(560px, 92vw)" @update:model-value="emit('update:modelValue', $event)">
    <div v-loading="loading" class="reservation-detail-content">
      <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
      <el-descriptions v-else-if="reservation" :column="1" border>
        <el-descriptions-item label="工单编号">#{{ reservation.id }}</el-descriptions-item>
        <el-descriptions-item label="实验课题">{{ reservation.experiment_name }}</el-descriptions-item>
        <el-descriptions-item label="预约设备">{{ reservation.equipment_name }}</el-descriptions-item>
        <el-descriptions-item label="预约人员">{{ reservation.user_name }}</el-descriptions-item>
        <el-descriptions-item label="所属课题组">{{ reservation.group_name }}</el-descriptions-item>
        <el-descriptions-item label="预约开始">{{ formatTime(reservation.start_time) }}</el-descriptions-item>
        <el-descriptions-item label="预约结束">{{ formatTime(reservation.end_time) }}</el-descriptions-item>
        <el-descriptions-item label="当前状态"><el-tag>{{ statusText(reservation.status) }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="机时费用">¥{{ Number(reservation.total_cost || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item v-if="reservation.task_id" label="实验任务">#{{ reservation.task_id }} · 工序 {{ reservation.task_step }}</el-descriptions-item>
        <el-descriptions-item v-if="reservation.purpose" label="实验用途">{{ reservation.purpose }}</el-descriptions-item>
        <el-descriptions-item v-if="reservation.reject_reason" label="处理说明">{{ reservation.reject_reason }}</el-descriptions-item>
      </el-descriptions>
    </div>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">关闭</el-button>
      <el-button v-if="reservation && !error && canApprove" type="primary" @click="goTo('/approvals')">前往审批</el-button>
      <el-button v-else-if="reservation && !error && reservation.user_id === authStore.user?.id" type="primary" @click="goTo('/my-reservations')">前往履约</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import request from '../api/request'

const props = defineProps({ modelValue: Boolean, reservationId: [Number, String] })
const emit = defineEmits(['update:modelValue'])
const router = useRouter()
const authStore = useAuthStore()
const reservation = ref(null)
const loading = ref(false)
const error = ref('')
let requestVersion = 0
const canApprove = computed(() => {
  if (!reservation.value) return false
  const status = reservation.value.status
  return (status === 'pending_advisor' && (authStore.isAdvisor || authStore.isSystemAdmin)) ||
    (status === 'pending_admin' && (authStore.isEquipmentAdmin || authStore.isSystemAdmin))
})
const formatTime = value => new Date(value).toLocaleString('zh-CN', { hour12: false })
const statusText = status => ({ pending_advisor: '待导师审批', pending_admin: '待管理员审批', approved: '待履约签到', active: '实验进行中', completed: '已归还', rejected: '已驳回', cancelled: '已取消', no_show: '已爽约' }[status] || status)
function goTo(path) {
  emit('update:modelValue', false)
  router.push({ path, query: { reservation_id: reservation.value.id } })
}
watch(() => [props.modelValue, props.reservationId], async ([open, id]) => {
  const version = ++requestVersion
  if (!open) return
  reservation.value = null
  error.value = ''
  loading.value = true
  try {
    const response = await request.get(`/reservations/${id}`)
    if (version === requestVersion) reservation.value = response.data
  } catch (err) {
    if (version === requestVersion) error.value = err.response?.data?.message || '读取工单失败，请稍后重试'
  } finally { if (version === requestVersion) loading.value = false }
}, { immediate: true })
</script>

<style scoped>
.reservation-detail-content { min-height: 120px; overflow-wrap: anywhere; }
</style>
