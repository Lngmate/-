<template>
  <div class="approval-page">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <el-icon class="icon"><Stamp /></el-icon>
            <span class="title">多级审批流转工作台</span>
            <el-tag size="small" type="primary">
              当前角色：{{ roleName }}
            </el-tag>
          </div>
          <el-button type="primary" plain size="small" @click="handleRefresh">
            <el-icon><Refresh /></el-icon> 刷新数据
          </el-button>
        </div>
      </template>

      <el-alert
        v-if="authStore.isStudent"
        type="warning"
        title="温馨提示：您当前以【学生】角色登录，无审批权限。请点击顶部导航栏右上角的【角色切换】，切换为【导师/组长】或【设备管理员】体验多级审批流程！"
        show-icon
        :closable="false"
        style="margin-bottom: 16px"
      />

      <el-alert v-if="route.query.reservation_id" :title="`正在查看工单 #${route.query.reservation_id} 的审批待办`" type="info" :closable="false" style="margin-bottom: 12px">
        <span v-if="!loading && !focusedPendingList.length">该工单已处理或不在当前角色的待办中。</span>
        <el-button link type="primary" @click="router.replace('/approvals')">显示全部待办</el-button>
      </el-alert>
      <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <!-- 选项卡 1：待我审批 -->
        <el-tab-pane :label="`待我审批待办 (${pendingList.length})`" name="pending">
          <el-table :data="focusedPendingList" style="width: 100%" v-loading="loading">
            <el-table-column prop="experiment_name" label="实验课题名称" min-width="180" />
            
            <el-table-column prop="equipment_name" label="预约设备" min-width="180" />

            <el-table-column label="申请人" width="160">
              <template #default="{ row }">
                <div><strong>{{ row.user_name }}</strong></div>
                <div class="sub-text">{{ row.group_name }}</div>
              </template>
            </el-table-column>

            <el-table-column label="预约时间段" min-width="200">
              <template #default="{ row }">
                <div>{{ formatDateTime(row.start_time) }}</div>
                <div>至 {{ formatDateTime(row.end_time) }}</div>
                <div class="sub-text">({{ row.duration_hours }} 小时)</div>
              </template>
            </el-table-column>

            <el-table-column label="扣费经费" width="120">
              <template #default="{ row }">
                <span class="cost">¥ {{ row.total_cost }}</span>
                <div v-if="row.is_batched" class="badge-sub">伴热拼单8.5折</div>
              </template>
            </el-table-column>

            <el-table-column label="工单状态" width="130">
              <template #default="{ row }">
                <el-tag type="warning">{{ getStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>

            <el-table-column label="审批操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button type="success" size="small" @click="openApproveModal(row)">
                  通过
                </el-button>
                <el-button type="danger" size="small" @click="openRejectModal(row)">
                  驳回
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 选项卡 2：审批流转历史日志 -->
        <el-tab-pane label="审批流转历史审计日志" name="history">
          <el-table :data="historyList" style="width: 100%" v-loading="historyLoading">
            <el-table-column prop="id" label="日志ID" width="80" />
            <el-table-column prop="equipment_name" label="申请设备" min-width="170" />
            <el-table-column prop="experiment_name" label="实验课题" min-width="180" />
            <el-table-column label="申请人" width="150">
              <template #default="{ row }">
                <div><strong>{{ row.applicant_name }}</strong></div>
                <div class="sub-text">{{ row.group_name }}</div>
              </template>
            </el-table-column>
            <el-table-column label="审批人" width="160">
              <template #default="{ row }">
                <div>{{ row.approver_name }}</div>
                <el-tag size="small" :type="row.approver_role === 'advisor' ? 'primary' : 'warning'">
                  {{ row.approver_role === 'advisor' ? '导师审批' : '管理员审批' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="审批决定" width="110">
              <template #default="{ row }">
                <el-tag :type="row.action === 'approve' ? 'success' : 'danger'">
                  {{ row.action === 'approve' ? '同意通过' : '驳回拒绝' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="comment" label="审批批注 / 意见" min-width="200" />
            <el-table-column label="审批时间" width="160">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 审批通过弹窗 -->
    <el-dialog v-model="approveVisible" title="审批通过确认" width="460px" destroy-on-close>
      <p style="margin-bottom: 12px; color: #475569;">
        确定核准【{{ actionRow?.user_name }}】关于【{{ actionRow?.experiment_name }}】的实验机时预约申请？
      </p>
      <el-form label-width="80px">
        <el-form-item label="审批批注">
          <el-input v-model="approveComment" type="textarea" placeholder="填写审批通过批注，如：资质与经费已确认核验（选填）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="success" :loading="submitting" @click="handleApprove">
          确认通过
        </el-button>
      </template>
    </el-dialog>

    <!-- 审批驳回弹窗 -->
    <el-dialog v-model="rejectVisible" title="审批驳回" width="460px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="驳回原因" required>
          <el-input
            v-model="rejectComment"
            type="textarea"
            rows="3"
            placeholder="请详细说明驳回原因，如课题经费不足、实验安全操作未认证等"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="submitting" @click="handleReject">
          确认驳回
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '../api/request'
import { Stamp, Refresh } from '@element-plus/icons-vue'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeTab = ref('pending')
const pendingList = ref([])
const focusedPendingList = computed(() => route.query.reservation_id ? pendingList.value.filter(row => String(row.id) === String(route.query.reservation_id)) : pendingList.value)
const historyList = ref([])
const loading = ref(false)
const historyLoading = ref(false)
const submitting = ref(false)

const approveVisible = ref(false)
const rejectVisible = ref(false)
const actionRow = ref(null)
const approveComment = ref('')
const rejectComment = ref('')

const roleName = computed(() => {
  const map = {
    student: '普通学生',
    advisor: '导师/课题组长',
    equipment_admin: '设备管理员',
    system_admin: '系统管理员'
  }
  return map[authStore.role] || authStore.role
})

watch(() => authStore.role, () => {
  fetchPendingList()
  fetchHistoryList()
})

watch(() => route.query.reservation_id, id => {
  if (id) {
    activeTab.value = 'pending'
    fetchPendingList()
  }
})

const fetchPendingList = async () => {
  loading.value = true
  try {
    const res = await request.get('/approvals/pending')
    if (res.data) {
      pendingList.value = res.data
    }
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const fetchHistoryList = async () => {
  historyLoading.value = true
  try {
    const res = await request.get('/approvals/history')
    if (res.data) {
      historyList.value = res.data
    }
  } catch (err) {
    console.error(err)
  } finally {
    historyLoading.value = false
  }
}

const handleRefresh = () => {
  if (activeTab.value === 'pending') {
    fetchPendingList()
  } else {
    fetchHistoryList()
  }
}

const handleTabClick = (tab) => {
  if (tab.paneName === 'history' && historyList.value.length === 0) {
    fetchHistoryList()
  }
}

const openApproveModal = (row) => {
  actionRow.value = row
  approveComment.value = '排期合规，同意机时预约'
  approveVisible.value = true
}

const openRejectModal = (row) => {
  actionRow.value = row
  rejectComment.value = ''
  rejectVisible.value = true
}

const handleApprove = async () => {
  submitting.value = true
  try {
    const res = await request.post(`/approvals/${actionRow.value.id}/approve`, {
      comment: approveComment.value
    })
    ElMessage.success(res.message || '审批通过成功！')
    approveVisible.value = false
    fetchPendingList()
    fetchHistoryList()
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
}

const handleReject = async () => {
  if (!rejectComment.value) {
    ElMessage.warning('请填写驳回原因')
    return
  }
  submitting.value = true
  try {
    const res = await request.post(`/approvals/${actionRow.value.id}/reject`, {
      comment: rejectComment.value
    })
    ElMessage.success(res.message || '已成功驳回')
    rejectVisible.value = false
    fetchPendingList()
    fetchHistoryList()
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
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
    approved: '已通过'
  }
  return map[status] || status
}

onMounted(() => {
  fetchPendingList()
  fetchHistoryList()
})
</script>

<style scoped>
.approval-page {
  padding: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-wrap .icon {
  font-size: 20px;
  color: #1890ff;
}

.title-wrap .title {
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
}

.sub-text {
  font-size: 11px;
  color: #64748b;
  margin-top: 2px;
}

.cost {
  font-weight: 700;
  color: #f59e0b;
}

.badge-sub {
  font-size: 10px;
  color: #10b981;
  font-weight: 600;
}
</style>
