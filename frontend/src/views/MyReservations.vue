<template>
  <div class="my-reservations-page">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <el-icon class="icon"><Document /></el-icon>
            <span class="title">我的机时预约与履约中心</span>
          </div>
          <div class="right-tools">
            <el-radio-group v-model="statusFilter" size="small" @change="fetchReservations">
              <el-radio-button value="">全部</el-radio-button>
              <el-radio-button value="approved">待履约签到</el-radio-button>
              <el-radio-button value="active">进行中</el-radio-button>
              <el-radio-button value="completed">已归还</el-radio-button>
            </el-radio-group>
            <el-button type="primary" plain size="small" @click="fetchReservations">
              <el-icon><Refresh /></el-icon> 刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-alert v-if="route.query.reservation_id" :title="`正在查看工单 #${route.query.reservation_id}`" type="info" :closable="false" style="margin-bottom: 12px">
        <el-button link type="primary" @click="router.replace('/my-reservations')">显示全部预约</el-button>
      </el-alert>
      <el-alert v-if="loadError" :title="loadError" type="error" :closable="false" style="margin-bottom: 12px" />
      <el-table :data="filteredReservations" style="width: 100%" v-loading="loading">
        <el-table-column prop="equipment_name" label="实验设备" min-width="180">
          <template #default="{ row }">
            <div class="eq-cell">
              <span class="name">{{ row.equipment_name }}</span>
              <span class="model">{{ row.equipment_model }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="experiment_name" label="实验课题名称" min-width="180">
          <template #default="{ row }">
            <div>{{ row.experiment_name }}</div>
            <el-link v-if="row.task_id" type="primary" href="/experiment-tasks">任务 #{{ row.task_id }} · 工序 {{ row.task_step }}</el-link>
          </template>
        </el-table-column>

        <el-table-column label="预约时间段" min-width="220">
          <template #default="{ row }">
            <div class="time-cell">
              <div>{{ formatDateTime(row.start_time) }}</div>
              <div class="arrow">至 {{ formatDateTime(row.end_time) }}</div>
              <div class="hours">({{ row.duration_hours }} 小时)</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="工单状态" width="130">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="机时费用" width="120">
          <template #default="{ row }">
            <div class="cost-cell">
              <span class="cost">¥ {{ row.total_cost }}</span>
              <el-tag v-if="row.is_batched" size="small" type="success" effect="plain">8.5折拼单</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="动态签到核销码" width="140">
          <template #default="{ row }">
            <div v-if="row.checkin_code && row.status !== 'rejected'" class="code-badge" @click="showQrModal(row)">
              <el-icon><Tickets /></el-icon> {{ row.checkin_code }}
            </div>
            <span v-else class="text-muted">--</span>
          </template>
        </el-table-column>

        <el-table-column label="现场履约操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showReservationDetail(row)">详情</el-button>
            <!-- 待履约 -> 现场扫码/输入签到码上机 -->
            <el-button
              v-if="row.status === 'approved'"
              type="success"
              size="small"
              @click="openCheckinDialog(row)"
            >
              现场签到
            </el-button>

            <!-- 待审核或待签到 -> 主动取消预约 -->
            <el-button
              v-if="row.status === 'pending_advisor' || row.status === 'pending_admin' || row.status === 'approved'"
              type="danger"
              size="small"
              plain
              @click="handleCancelReservation(row)"
            >
              取消预约
            </el-button>

            <!-- 进行中 -> 实验结束归还验收 -->
            <el-button
              v-if="row.status === 'active'"
              type="primary"
              size="small"
              @click="openCheckoutDialog(row)"
            >
              实验归还
            </el-button>

            <!-- 审批驳回查看原因 -->
            <el-popover
              v-if="row.status === 'rejected'"
              placement="top"
              title="驳回原因"
              :width="220"
              trigger="hover"
              :content="row.reject_reason || '资质或排期冲突'"
            >
              <template #reference>
                <el-button type="danger" link size="small">查看驳回原因</el-button>
              </template>
            </el-popover>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <ReservationDetailDialog v-model="reservationDetailVisible" :reservation-id="detailReservationId" />

    <!-- 签到核销对话框 -->
    <el-dialog v-model="checkinVisible" title="实验室现场上机签到核销" width="480px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="实验设备">
          <el-input :model-value="actionRow?.equipment_name" disabled />
        </el-form-item>
        <el-form-item label="核销签到码" required>
          <el-input v-model="checkinForm.code" placeholder="请输入6位现场核销码" maxlength="6" />
        </el-form-item>
        <el-form-item label="使用前状态确认">
          <el-input
            v-model="checkinForm.condition"
            type="textarea"
            placeholder="如：外观完好，真空度正常，温控仪表正常就绪"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="checkinVisible = false">取消</el-button>
        <el-button type="success" :loading="submitting" @click="handleCheckin">
          确认核销并开始实验
        </el-button>
      </template>
    </el-dialog>

    <!-- 归还验收对话框 -->
    <el-dialog v-model="checkoutVisible" title="实验结束设备归还与验收" width="480px" destroy-on-close>
      <el-alert
        type="info"
        title="按时归还奖励：按时完成实验并提交真实状态验收，系统将自动奖励 +2 点科研信誉积分！"
        show-icon
        :closable="false"
        style="margin-bottom: 16px"
      />
      <el-form label-width="110px">
        <el-form-item label="实验设备">
          <el-input :model-value="actionRow?.equipment_name" disabled />
        </el-form-item>
        <el-form-item label="归还状态确认" required>
          <el-input
            v-model="checkoutForm.condition"
            type="textarea"
            placeholder="如：样品已取出，腔体已清洁降温，关机断电，现场卫生良好"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="checkoutVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleCheckout">
          确认归还并结算
        </el-button>
      </template>
    </el-dialog>

    <!-- 动态核销二维码弹窗 -->
    <el-dialog v-model="qrVisible" title="现场扫码 / 核销凭证" width="360px" align-center destroy-on-close>
      <div class="qr-content">
        <div class="qr-box">
          <!-- 动态模拟二维码图形 -->
          <div class="fake-qr">
            <el-icon class="qr-icon"><Cpu /></el-icon>
            <div class="qr-scan-line"></div>
          </div>
        </div>
        <div class="qr-code-text">
          <span class="lbl">现场核销码：</span>
          <span class="code">{{ actionRow?.checkin_code }}</span>
        </div>
        <p class="qr-tip">请在到达实验室现场后，向机台扫码枪或管理员出示此核销凭证</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ReservationDetailDialog from '../components/ReservationDetailDialog.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../api/request'
import { Document, Refresh, Tickets, Cpu } from '@element-plus/icons-vue'

const reservations = ref([])
const route = useRoute()
const router = useRouter()
const loadError = ref('')
const reservationDetailVisible = ref(false)
const detailReservationId = ref(null)
const showReservationDetail = row => {
  detailReservationId.value = row.id
  reservationDetailVisible.value = true
}
const loading = ref(false)
const statusFilter = ref('')

const checkinVisible = ref(false)
const checkoutVisible = ref(false)
const qrVisible = ref(false)
const actionRow = ref(null)
const submitting = ref(false)

const checkinForm = ref({ code: '', condition: '设备外观正常，水电气源就绪' })
const checkoutForm = ref({ condition: '实验完成，腔体已清洁，已恢复常温并关机' })

const fetchReservations = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const res = await request.get('/reservations', { params: { scope: 'my' } })
    if (res.data) reservations.value = res.data
  } catch (err) {
    reservations.value = []
    loadError.value = err.response?.data?.message || '读取预约失败，请刷新重试'
  } finally {
    loading.value = false
  }
}

const filteredReservations = computed(() => {
  if (route.query.reservation_id) return reservations.value.filter(r => String(r.id) === String(route.query.reservation_id))
  if (!statusFilter.value) return reservations.value
  return reservations.value.filter(r => r.status === statusFilter.value)
})

const openCheckinDialog = (row) => {
  actionRow.value = row
  checkinForm.value.code = row.checkin_code
  checkinVisible.value = true
}

const handleCheckin = async () => {
  if (!checkinForm.value.code) {
    ElMessage.warning('请输入签到码')
    return
  }
  submitting.value = true
  try {
    const res = await request.post(`/reservations/${actionRow.value.id}/checkin`, {
      checkin_code: checkinForm.value.code,
      condition_check_in: checkinForm.value.condition
    })
    ElMessage.success(res.message || '现场核销签到成功！')
    checkinVisible.value = false
    fetchReservations()
  } catch (err) {
  } finally {
    submitting.value = false
  }
}

const openCheckoutDialog = (row) => {
  actionRow.value = row
  checkoutVisible.value = true
}

const handleCheckout = async () => {
  submitting.value = true
  try {
    const res = await request.post(`/reservations/${actionRow.value.id}/checkout`, {
      condition_check_out: checkoutForm.value.condition
    })
    ElMessage.success(res.message || '实验已归还并结算完成！')
    checkoutVisible.value = false
    fetchReservations()
  } catch (err) {
  } finally {
    submitting.value = false
  }
}

const handleCancelReservation = async (row) => {
  try {
    const { value: reason } = await ElMessageBox.prompt('请输入取消预约的原因（主动提前释放时间槽可获信誉积分奖励）:', '取消预约确认', {
      confirmButtonText: '确认取消',
      cancelButtonText: '暂不取消',
      inputPlaceholder: '如：样品制备延后 / 实验日程变更'
    })
    const res = await request.post(`/reservations/${row.id}/cancel`, { reason: reason || '用户主动取消' })
    ElMessage.success(res.message || '预约已成功取消')
    fetchReservations()
  } catch (e) {
    // cancelled
  }
}

const showQrModal = (row) => {
  actionRow.value = row
  qrVisible.value = true
}

const formatDateTime = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  return `${d.getMonth() + 1}-${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

const getStatusText = (status) => {
  const map = {
    pending_advisor: '待导师审批',
    pending_admin: '待管理员审批',
    approved: '待履约签到',
    active: '实验运行中',
    completed: '已归还',
    rejected: '已驳回',
    cancelled: '已取消',
    no_show: '爽约'
  }
  return map[status] || status
}

const getStatusTag = (status) => {
  const map = {
    pending_advisor: 'warning',
    pending_admin: 'warning',
    approved: 'success',
    active: 'primary',
    completed: 'info',
    rejected: 'danger',
    no_show: 'danger'
  }
  return map[status] || 'info'
}

onMounted(() => {
  fetchReservations()
})

watch(() => route.query.reservation_id, () => { statusFilter.value = ''; fetchReservations() })
</script>

<style scoped>
.my-reservations-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.box-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.right-tools {
  display: flex;
  align-items: center;
  gap: 12px;
}

.eq-cell {
  display: flex;
  flex-direction: column;
}

.eq-cell .name {
  font-weight: 600;
  color: #1f2937;
}

.eq-cell .model {
  font-size: 12px;
  color: #6b7280;
}

.time-cell {
  font-size: 12px;
  color: #374151;
}

.time-cell .hours {
  color: #9ca3af;
}

.cost-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.cost-cell .cost {
  font-weight: bold;
  color: #e65100;
}

.code-badge {
  background: #f0f7ff;
  border: 1px dashed #409eff;
  color: #409eff;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: monospace;
  font-weight: bold;
  font-size: 13px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.code-badge:hover {
  background: #e6f1fc;
}

.qr-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.fake-qr {
  width: 180px;
  height: 180px;
  background: #f8fafc;
  border: 2px solid #cbd5e1;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  margin-bottom: 16px;
}

.qr-icon {
  font-size: 64px;
  color: #1890ff;
}

.qr-scan-line {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 2px;
  background: #1890ff;
  box-shadow: 0 0 8px #1890ff;
  animation: scan 2s infinite linear;
}

@keyframes scan {
  0% { top: 0; }
  50% { top: 100%; }
  100% { top: 0; }
}

.qr-code-text {
  font-size: 16px;
  margin-bottom: 8px;
}

.qr-code-text .code {
  font-size: 22px;
  font-weight: bold;
  color: #1890ff;
  font-family: monospace;
}

.qr-tip {
  font-size: 12px;
  color: #94a3b8;
}
</style>
