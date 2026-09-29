<template>
  <div class="credit-center-page">
    <div class="credit-hero">
      <el-card class="hero-card" shadow="hover">
        <div class="hero-left">
          <div class="score-circle" :style="{ borderColor: getBadgeColor(creditData.credit_score) }">
            <div class="score-num">{{ creditData.credit_score || 98 }}</div>
            <div class="score-lbl">科研信誉分</div>
          </div>
          <div class="hero-text">
            <div class="hero-name">{{ creditData.user_name || '张耀辉(学生)' }}</div>
            <div class="hero-tier">
              <el-tag :type="creditData.badge_color || 'success'" size="large">
                {{ creditData.rating || '五星卓越履约者' }}
              </el-tag>
            </div>
            <p class="hero-desc">
              信誉良好享有特权：可在黄金时段优先排队、部分精密仪器免导师二次复审。
            </p>
          </div>
        </div>

        <div class="hero-rules">
          <div class="rule-title">信誉分级权益与管控规则</div>
          <div class="rule-item">
            <span class="badge green">≥95分 (卓越)</span>
            <span class="desc">享有黄金时段优先排队权、免导师审批直通特权</span>
          </div>
          <div class="rule-item">
            <span class="badge blue">80~94分 (良好)</span>
            <span class="desc">享有标准机时预约权限，按规范履约流程执行</span>
          </div>
          <div class="rule-item">
            <span class="badge orange">60~79分 (预警)</span>
            <span class="desc">限制单周最大预约时长，暂停一级高精密设备预约</span>
          </div>
          <div class="rule-item">
            <span class="badge red">&lt;60分 (冻结)</span>
            <span class="desc">账号暂停设备预约权限，需向设备管理员提交申诉复核</span>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 履约奖惩流水 -->
    <el-card class="history-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <el-icon class="icon"><Medal /></el-icon>
            <span class="title">科研履约信誉奖惩流水档案</span>
          </div>
          <el-button type="primary" plain size="small" @click="fetchCreditDetail">
            <el-icon><Refresh /></el-icon> 刷新档案
          </el-button>
        </div>
      </template>

      <el-table :data="creditData.history || []" style="width: 100%">
        <el-table-column label="发生时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column prop="reason" label="奖惩原因与行为摘要" min-width="260" />

        <el-table-column label="积分变动" width="140">
          <template #default="{ row }">
            <span :class="row.delta_score >= 0 ? 'text-green font-bold' : 'text-red font-bold'">
              {{ row.delta_score >= 0 ? '+' : '' }}{{ row.delta_score }} 分
            </span>
          </template>
        </el-table-column>

        <el-table-column prop="score_after" label="变动后积分" width="140">
          <template #default="{ row }">
            <strong>{{ row.score_after }}</strong> 分
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import request from '../api/request'
import { Medal, Refresh } from '@element-plus/icons-vue'

const creditData = ref({
  user_name: '张耀辉(学生)',
  credit_score: 98,
  rating: '五星卓越 (享有高峰期优先排队与免导师审批特权)',
  badge_color: 'success',
  history: [
    {
      id: 1,
      created_at: new Date(Date.now() - 86400000 * 2).toISOString(),
      reason: '按时履约完成实验并如实记录设备完好状态',
      delta_score: 2,
      score_after: 98
    },
    {
      id: 2,
      created_at: new Date(Date.now() - 86400000 * 5).toISOString(),
      reason: '新学期科研诚信认证完成，基准分初始化',
      delta_score: 0,
      score_after: 96
    }
  ]
})

const fetchCreditDetail = async () => {
  try {
    const res = await request.get('/credits/detail')
    if (res.data) creditData.value = res.data
  } catch (err) {}
}

const getBadgeColor = (score) => {
  if (score >= 95) return '#52c41a'
  if (score >= 80) return '#1890ff'
  if (score >= 60) return '#fa8c16'
  return '#f5222d'
}

const formatDateTime = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  return d.toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => {
  fetchCreditDetail()
})
</script>

<style scoped>
.credit-center-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.hero-card {
  border-radius: 8px;
}

:deep(.hero-card .el-card__body) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px;
}

.hero-left {
  display: flex;
  align-items: center;
  gap: 24px;
}

.score-circle {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  border: 6px solid #52c41a;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #fafafa;
}

.score-num {
  font-size: 36px;
  font-weight: bold;
  color: #1f2937;
  line-height: 1;
}

.score-lbl {
  font-size: 11px;
  color: #6b7280;
  margin-top: 4px;
}

.hero-name {
  font-size: 20px;
  font-weight: bold;
  color: #111827;
}

.hero-tier {
  margin: 6px 0 8px 0;
}

.hero-desc {
  font-size: 13px;
  color: #64748b;
  margin: 0;
}

.hero-rules {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  width: 440px;
}

.rule-title {
  font-weight: 600;
  font-size: 13px;
  color: #1e293b;
  margin-bottom: 10px;
}

.rule-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  font-size: 12px;
}

.rule-item:last-child {
  margin-bottom: 0;
}

.rule-item .badge {
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.badge.green { background: #f6ffed; color: #52c41a; border: 1px solid #b7eb8f; }
.badge.blue { background: #e6f7ff; color: #1890ff; border: 1px solid #91d5ff; }
.badge.orange { background: #fff7e6; color: #fa8c16; border: 1px solid #ffd591; }
.badge.red { background: #fff1f0; color: #f5222d; border: 1px solid #ffa39e; }

.history-card {
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
.text-red { color: #f5222d; }
.font-bold { font-weight: bold; }
</style>
