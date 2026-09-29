<template>
  <div class="group-settlement-page">
    <!-- 课题组虚拟账户概览卡片 -->
    <div class="overview-grid">
      <el-card class="balance-card" shadow="hover">
        <div class="balance-header">
          <span class="lbl">所属课题组</span>
          <el-tag type="info">{{ currentGroup.department || '材料科学与工程学院' }}</el-tag>
        </div>
        <div class="group-title">{{ currentGroup.name || '先进纳米材料课题组' }}</div>
        <div class="leader-info">课题组长 / 导师：{{ currentGroup.leader_name || '李教授' }}</div>

        <div class="balance-main">
          <div class="balance-num">
            <span class="currency">¥</span>
            <span class="amount">{{ currentGroup.balance?.toFixed(2) || '24872.50' }}</span>
          </div>
          <div class="balance-sub">可用虚拟科研经费与机时配额</div>
        </div>

        <div class="btn-wrap">
          <el-button type="primary" @click="rechargeVisible = true">
            <el-icon><Plus /></el-icon> 经费划拨 / 充值
          </el-button>
        </div>
      </el-card>

      <el-card class="stat-card" shadow="hover">
        <div class="stat-title">机时消耗统计与概况</div>
        <div class="stat-item">
          <div class="stat-lbl">累计消耗机时</div>
          <div class="stat-val">{{ currentGroup.total_hours_used || 124.5 }} <span class="unit">小时</span></div>
        </div>
        <div class="stat-item">
          <div class="stat-lbl">平均小时成本</div>
          <div class="stat-val">¥ 145.2 <span class="unit">/ 小时</span></div>
        </div>
        <div class="stat-item">
          <div class="stat-lbl">跨组共享结算加成</div>
          <div class="stat-val text-green">免加价优惠 (校内协同示范组)</div>
        </div>
      </el-card>
    </div>

    <!-- 虚拟结算明细流水 -->
    <el-card class="tx-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <el-icon class="icon"><List /></el-icon>
            <span class="title">虚拟经费扣减与消费流水明细</span>
          </div>
          <el-button type="primary" plain size="small" @click="fetchTransactions">
            <el-icon><Refresh /></el-icon> 刷新流水
          </el-button>
        </div>
      </template>

      <el-table :data="transactions" style="width: 100%" v-loading="loading">
        <el-table-column label="发生时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column prop="type" label="交易类型" width="140">
          <template #default="{ row }">
            <el-tag :type="row.amount >= 0 ? 'success' : 'primary'">
              {{ getTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="description" label="摘要说明" min-width="240" />

        <el-table-column label="变动金额" width="140">
          <template #default="{ row }">
            <span :class="row.amount >= 0 ? 'text-green font-bold' : 'text-red font-bold'">
              {{ row.amount >= 0 ? '+' : '' }}{{ row.amount.toFixed(2) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="变动后余额" width="140">
          <template #default="{ row }">
            ¥ {{ row.balance_after.toFixed(2) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 充值弹窗 -->
    <el-dialog v-model="rechargeVisible" title="课题组虚拟经费划拨 / 充值" width="460px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="充值金额" required>
          <el-input-number v-model="rechargeAmount" :min="100" :step="500" style="width: 100%" />
        </el-form-item>
        <el-form-item label="划转说明">
          <el-input v-model="rechargeDesc" placeholder="如：学院国家重点实验室培育经费划拨" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleRecharge">
          确认划拨
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '../api/request'
import { Plus, List, Refresh } from '@element-plus/icons-vue'

const currentGroup = ref({})
const transactions = ref([])
const loading = ref(false)
const rechargeVisible = ref(false)
const rechargeAmount = ref(5000)
const rechargeDesc = ref('')
const submitting = ref(false)

const fetchGroupData = async () => {
  try {
    const res = await request.get('/groups/1')
    if (res.data?.group) currentGroup.value = res.data.group
  } catch (err) {
    currentGroup.value = {
      id: 1,
      name: '先进纳米材料课题组',
      leader_name: '李教授',
      department: '材料科学与工程学院',
      balance: 24872.5,
      total_hours_used: 124.5
    }
  }
}

const fetchTransactions = async () => {
  loading.value = true
  try {
    const res = await request.get('/groups/transactions', { params: { group_id: 1 } })
    if (res.data) transactions.value = res.data
  } catch (err) {
    transactions.value = [
      {
        id: 1,
        created_at: new Date().toISOString(),
        type: 'reservation_fee',
        amount: -127.5,
        balance_after: 24872.5,
        description: '预约管式炉机时费(伴热拼单8.5折)'
      },
      {
        id: 2,
        created_at: new Date(Date.now() - 86400000 * 3).toISOString(),
        type: 'recharge',
        amount: 10000.0,
        balance_after: 25000.0,
        description: '国家重点实验室机时补贴专项'
      }
    ]
  } finally {
    loading.value = false
  }
}

const handleRecharge = async () => {
  submitting.value = true
  try {
    const res = await request.post('/groups/1/recharge', {
      amount: rechargeAmount.value,
      desc: rechargeDesc.value || '课题组机时配额充值'
    })
    ElMessage.success(res.message || '经费划拨成功！')
    rechargeVisible.value = false
    fetchGroupData()
    fetchTransactions()
  } catch (err) {
  } finally {
    submitting.value = false
  }
}

const formatDateTime = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  return d.toLocaleString('zh-CN', { hour12: false })
}

const getTypeText = (type) => {
  const map = {
    reservation_fee: '机时消费',
    recharge: '经费划拨',
    refund: '撤销退款'
  }
  return map[type] || type
}

onMounted(() => {
  fetchGroupData()
  fetchTransactions()
})
</script>

<style scoped>
.group-settlement-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.overview-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
}

.balance-card {
  border-radius: 8px;
  background: linear-gradient(135deg, #ffffff 0%, #f0f7ff 100%);
}

.balance-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.group-title {
  font-size: 20px;
  font-weight: bold;
  color: #1e293b;
  margin: 6px 0 2px 0;
}

.leader-info {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 16px;
}

.balance-main {
  margin-bottom: 16px;
}

.balance-num {
  font-size: 32px;
  font-weight: bold;
  color: #0284c7;
}

.currency {
  font-size: 22px;
  margin-right: 4px;
}

.balance-sub {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}

.stat-card {
  border-radius: 8px;
  display: flex;
  flex-direction: column;
}

.stat-title {
  font-weight: bold;
  font-size: 15px;
  color: #1e293b;
  margin-bottom: 16px;
}

.stat-item {
  margin-bottom: 14px;
}

.stat-lbl {
  font-size: 12px;
  color: #64748b;
}

.stat-val {
  font-size: 18px;
  font-weight: bold;
  color: #334155;
  margin-top: 2px;
}

.stat-val .unit {
  font-size: 12px;
  font-weight: normal;
  color: #94a3b8;
}

.tx-card {
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

.text-green { color: #52c41a; }
.text-red { color: #f56c6c; }
.font-bold { font-weight: bold; }
</style>
