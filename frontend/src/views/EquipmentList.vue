<template>
  <div class="equipment-list-page">
    <!-- 顶部操作与筛选栏 -->
    <el-card class="filter-card" shadow="never">
      <div class="filter-row">
        <div class="left-tools">
          <el-radio-group v-model="viewMode" size="default">
            <el-radio-button value="map">
              <el-icon><Location /></el-icon> 2D 孪生地图视图
            </el-radio-button>
            <el-radio-button value="cards">
              <el-icon><Menu /></el-icon> 设备卡片列表
            </el-radio-button>
          </el-radio-group>
        </div>

        <div class="right-filters">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索设备名称/型号/位置..."
            prefix-icon="Search"
            clearable
            style="width: 240px"
            @input="fetchEquipments"
          />
          <el-select v-model="filters.category" placeholder="设备分类" clearable style="width: 180px" @change="fetchEquipments">
            <el-option label="显微镜与微观表征" value="显微镜与微观表征" />
            <el-option label="晶体结构与物性表征" value="晶体结构与物性表征" />
            <el-option label="高温合成与薄膜制备" value="高温合成与薄膜制备" />
            <el-option label="生化分离与质谱色谱" value="生化分离与质谱色谱" />
          </el-select>
          <el-select v-model="filters.status" placeholder="运行状态" clearable style="width: 140px" @change="fetchEquipments">
            <el-option label="空闲就绪" value="available" />
            <el-option label="实验运行中" value="in_use" />
            <el-option label="维保检修" value="maintenance" />
          </el-select>
        </div>
      </div>
    </el-card>

    <!-- 视图 1: 2D 数字孪生平面地图（创新点） -->
    <LabMap2D v-if="viewMode === 'map'" :equipments="equipments" @book="openBookDialog" />

    <!-- 视图 2: 设备卡片列表 -->
    <div v-else class="cards-grid">
      <el-card v-for="eq in equipments" :key="eq.id" class="eq-card" shadow="hover">
        <div class="eq-card-header">
          <div class="name-box">
            <h3 class="eq-name">{{ eq.name }}</h3>
            <span class="eq-model">{{ eq.model }}</span>
          </div>
          <el-tag :type="getStatusTag(eq.status)" effect="light">
            {{ getStatusText(eq.status) }}
          </el-tag>
        </div>

        <div class="eq-body">
          <div class="info-line">
            <el-icon><Location /></el-icon>
            <span>{{ eq.location }} ({{ eq.lab_room }})</span>
          </div>
          <div class="info-line">
            <el-icon><Money /></el-icon>
            <span class="rate">¥ {{ eq.hourly_rate }} / 小时</span>
            <span class="sub-text">({{ eq.group_name }})</span>
          </div>

          <!-- 健康度预测进度条（创新点） -->
          <div class="health-block">
            <div class="health-lbl">
              <span>健康指数 (Predictive Health)</span>
              <span :style="{ color: getHealthColor(eq.health_score), fontWeight: 'bold' }">{{ eq.health_score }} 分</span>
            </div>
            <el-progress :percentage="eq.health_score" :color="getHealthColor(eq.health_score)" :show-text="false" />
          </div>

          <!-- 审批策略与特性标签 -->
          <div class="tags-row">
            <el-tag size="small" type="info">
              {{ eq.approval_type === 'none' ? '免审直通' : (eq.approval_type === 'dual' ? '双审 (导师+管理员)' : '单审 (管理员)') }}
            </el-tag>
            <el-tag v-if="eq.support_batching" size="small" type="success" effect="plain">
              支持伴热拼单(8.5折)
            </el-tag>
            <el-tag size="small" :type="eq.safety_level >= 2 ? 'warning' : ''">
              安全 {{ eq.safety_level }} 级
            </el-tag>
          </div>

          <p class="specs-text">{{ eq.specs }}</p>
        </div>

        <div class="eq-footer">
          <el-button
            type="primary"
            style="width: 100%"
            :disabled="eq.status === 'maintenance' || eq.status === 'faulty'"
            @click="openBookDialog(eq)"
          >
            <el-icon><Calendar /></el-icon> 立即预约此设备
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- 预约弹窗 -->
    <el-dialog v-model="bookDialogVisible" title="提交设备机时预约申请" width="600px" destroy-on-close>
      <el-form :model="bookForm" label-width="110px">
        <el-form-item label="目标设备">
          <el-input :model-value="selectedEq?.name + ' (' + selectedEq?.model + ')'" disabled />
        </el-form-item>
        <el-form-item label="所属实验室">
          <el-input :model-value="selectedEq?.location" disabled />
        </el-form-item>
        <el-form-item label="审批流类型">
          <el-tag :type="selectedEq?.approval_type === 'none' ? 'success' : 'primary'">
            {{ selectedEq?.approval_type === 'none' ? '免审直通（提交即生效）' : (selectedEq?.approval_type === 'dual' ? '双级审批（导师审经费资质 + 管理员审机时）' : '单级审批（设备管理员排期审核）') }}
          </el-tag>
        </el-form-item>
        <el-form-item label="预约时间段" required>
          <el-date-picker
            v-model="timeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm"
            :default-time="[new Date(2000, 1, 1, 9, 0, 0), new Date(2000, 1, 1, 12, 0, 0)]"
            @change="handleTimeChange"
          />
        </el-form-item>

        <!-- 创新点：工况参数（用于触发伴热/同参拼单检测） -->
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="设定工况温度">
              <el-input-number v-model="bookForm.target_temp" :min="0" :max="1800" placeholder="℃" @change="checkBatchOpportunity" />
              <span class="unit-text"> ℃</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="需要抽真空">
              <el-switch v-model="bookForm.vacuum_needed" @change="checkBatchOpportunity" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 创新点：伴热拼单优惠提示卡片 -->
        <el-alert
          v-if="batchAlert.has_opportunity"
          type="success"
          :title="batchAlert.reason"
          show-icon
          :closable="false"
          style="margin-bottom: 16px"
        >
          <template #default>
            邻近实验：<strong>{{ batchAlert.adjacent_experiment }}</strong> (设定 {{ batchAlert.adjacent_temp }}℃)<br />
            绿色节能减碳奖励：<strong>机时费享受 {{ batchAlert.discount_rate * 10 }} 折优惠！</strong>
          </template>
        </el-alert>

        <el-form-item label="实验课题名称" required>
          <el-input v-model="bookForm.experiment_name" placeholder="如：二维纳米材料高温退火处理" />
        </el-form-item>
        <el-form-item label="实验用途简述">
          <el-input v-model="bookForm.purpose" type="textarea" :rows="2" placeholder="请简要说明实验目的与样品特征" />
        </el-form-item>

        <el-form-item label="预计机时费用">
          <div class="cost-calc">
            <span class="cost-num">¥ {{ estimatedCost.toFixed(2) }}</span>
            <span v-if="batchAlert.has_opportunity" class="discount-badge">已享 8.5 折节能优惠</span>
            <span class="cost-tip">(将从所属课题组账户虚拟扣除)</span>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="bookDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitBooking">
          确认提交预约
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '../api/request'
import LabMap2D from '../components/LabMap2D.vue'
import {
  Location,
  Menu,
  Search,
  Money,
  Calendar
} from '@element-plus/icons-vue'

const viewMode = ref('map')
const equipments = ref([])
const filters = ref({ keyword: '', category: '', status: '' })

const bookDialogVisible = ref(false)
const selectedEq = ref(null)
const timeRange = ref([])
const submitting = ref(false)

const bookForm = ref({
  experiment_name: '',
  purpose: '',
  target_temp: 0,
  vacuum_needed: false
})

const batchAlert = ref({
  has_opportunity: false,
  discount_rate: 1.0,
  reason: '',
  adjacent_experiment: '',
  adjacent_temp: 0
})

const fetchEquipments = async () => {
  try {
    const res = await request.get('/equipments', { params: filters.value })
    if (res.data) {
      equipments.value = res.data
    }
  } catch (err) {
    // 降级使用内置数据集
    equipments.value = [
      { id: 1, name: '场发射扫描电子显微镜(FE-SEM)', model: 'JEOL JSM-7800F', category: '显微镜与微观表征', location: '理科实验大楼 301 室', lab_room: 'Lab-301', map_x: 22, map_y: 30, safety_level: 2, hourly_rate: 180, status: 'available', approval_type: 'dual', group_name: '先进纳米材料课题组', health_score: 92, specs: '分辨率: 0.8nm @ 15kV; 配备牛津能谱仪(EDS)' },
      { id: 2, name: '超高温真空气氛管式炉', model: 'NBD-T1700-60IT', category: '高温合成与热处理', location: '先进材料馆 104 室', lab_room: 'Lab-104', map_x: 55, map_y: 28, safety_level: 3, hourly_rate: 50, status: 'available', approval_type: 'single', group_name: '先进纳米材料课题组', health_score: 86, support_batching: true, specs: '最高工作温度: 1700℃; 温控精度: ±1℃' },
      { id: 3, name: 'X射线多晶粉末衍射仪(XRD)', model: 'Bruker D8 ADVANCE', category: '晶体结构分析', location: '理科实验大楼 305 室', lab_room: 'Lab-305', map_x: 78, map_y: 35, safety_level: 2, hourly_rate: 120, status: 'available', approval_type: 'single', group_name: '超导物理与量子计算组', health_score: 95, support_batching: true, specs: 'Cu靶陶瓷X射线管; LynxEye探测器' },
      { id: 4, name: '超临界流体色谱-质谱联用仪(SFC-MS)', model: 'Waters ACQUITY UPC2', category: '色谱与质谱分离', location: '化学生物楼 208 室', lab_room: 'Lab-208', map_x: 35, map_y: 75, safety_level: 2, hourly_rate: 200, status: 'in_use', approval_type: 'dual', group_name: '生物催化与精准合成组', health_score: 89, specs: '超临界CO2流动相; 配备Q-Da质量检测器' },
      { id: 5, name: '高速冷冻离心机', model: 'Beckman Coulter Avanti J-26S', category: '生物分离工程', location: '化学生物楼 210 室', lab_room: 'Lab-210', map_x: 72, map_y: 72, safety_level: 1, hourly_rate: 30, status: 'available', approval_type: 'none', group_name: '生物催化与精准合成组', health_score: 98, support_batching: true, specs: '最高转速: 26,000 rpm; 温控范围: -10℃ 至 +40℃' }
    ]
  }
}

const openBookDialog = (eq) => {
  selectedEq.value = eq
  bookForm.value = {
    experiment_name: '',
    purpose: '',
    target_temp: eq.max_temp > 0 ? 800 : 0,
    vacuum_needed: eq.vacuum_required
  }
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours() + 1, 0)
  const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours() + 4, 0)
  timeRange.value = [start, end]
  batchAlert.value = { has_opportunity: false, discount_rate: 1.0 }
  bookDialogVisible.value = true
  checkBatchOpportunity()
}

const handleTimeChange = () => {
  checkBatchOpportunity()
}

// 检查防冲突与伴热拼单机会
const checkBatchOpportunity = async () => {
  if (!selectedEq.value || !timeRange.value || timeRange.value.length < 2) return
  try {
    const res = await request.post('/reservations/check', {
      equipment_id: selectedEq.value.id,
      start_time: timeRange.value[0],
      end_time: timeRange.value[1],
      target_temp: bookForm.value.target_temp,
      vacuum_needed: bookForm.value.vacuum_needed
    })
    if (res.data?.batching) {
      batchAlert.value = res.data.batching
    }
  } catch (err) {
    // 模拟拼单检测
    if (selectedEq.value.support_batching && bookForm.value.target_temp >= 500) {
      batchAlert.value = {
        has_opportunity: true,
        discount_rate: 0.85,
        reason: '检测到邻近高温相近实验，触发伴热拼单节能优惠(8.5折)！',
        adjacent_experiment: '二维过渡金属硫族化物CVD高温气相沉积',
        adjacent_temp: 1100
      }
    } else {
      batchAlert.value = { has_opportunity: false, discount_rate: 1.0 }
    }
  }
}

const estimatedCost = computed(() => {
  if (!selectedEq.value || !timeRange.value || timeRange.value.length < 2) return 0
  const hours = (new Date(timeRange.value[1]) - new Date(timeRange.value[0])) / 3600000
  const rate = selectedEq.value.hourly_rate || 0
  const discount = batchAlert.value.has_opportunity ? batchAlert.value.discount_rate : 1.0
  return hours * rate * discount
})

const submitBooking = async () => {
  if (!timeRange.value || timeRange.value.length < 2) {
    ElMessage.warning('请选择预约时间段')
    return
  }
  if (!bookForm.value.experiment_name) {
    ElMessage.warning('请输入实验课题名称')
    return
  }

  submitting.value = true
  try {
    const res = await request.post('/reservations', {
      equipment_id: selectedEq.value.id,
      start_time: timeRange.value[0],
      end_time: timeRange.value[1],
      experiment_name: bookForm.value.experiment_name,
      purpose: bookForm.value.purpose,
      target_temp: bookForm.value.target_temp,
      vacuum_needed: bookForm.value.vacuum_needed
    })
    ElMessage.success(res.message || '预约申请提交成功！')
    bookDialogVisible.value = false
    fetchEquipments()
  } catch (err) {
    // 错误已被拦截器提示
  } finally {
    submitting.value = false
  }
}

const getStatusText = (status) => {
  const map = { available: '空闲就绪', in_use: '实验运行中', reserved: '已被预约', maintenance: '维保检修', faulty: '故障停用' }
  return map[status] || status
}

const getStatusTag = (status) => {
  const map = { available: 'success', in_use: 'primary', reserved: 'warning', maintenance: 'danger', faulty: 'danger' }
  return map[status] || 'info'
}

const getHealthColor = (score) => {
  if (score >= 90) return '#67C23A'
  if (score >= 75) return '#E6A23C'
  return '#F56C6C'
}

onMounted(() => {
  fetchEquipments()
})
</script>

<style scoped>
.equipment-list-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card {
  border-radius: 8px;
}

.filter-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.right-filters {
  display: flex;
  gap: 10px;
  align-items: center;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
}

.eq-card {
  border-radius: 8px;
  display: flex;
  flex-direction: column;
}

:deep(.eq-card .el-card__body) {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.eq-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.eq-name {
  font-size: 15px;
  font-weight: bold;
  color: #1f2937;
  margin: 0 0 4px 0;
}

.eq-model {
  font-size: 12px;
  color: #6b7280;
}

.eq-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-line {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #4b5563;
}

.info-line .rate {
  font-weight: bold;
  color: #e65100;
}

.info-line .sub-text {
  font-size: 12px;
  color: #9ca3af;
}

.health-block {
  margin: 6px 0;
}

.health-lbl {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #6b7280;
  margin-bottom: 4px;
}

.tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 4px 0;
}

.specs-text {
  font-size: 12px;
  color: #6b7280;
  line-height: 1.5;
  background: #f8fafc;
  padding: 6px 8px;
  border-radius: 4px;
  margin: 6px 0;
}

.eq-footer {
  margin-top: 12px;
}

.unit-text {
  margin-left: 8px;
  font-size: 13px;
  color: #6b7280;
}

.cost-calc {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cost-num {
  font-size: 20px;
  font-weight: bold;
  color: #f56c6c;
}

.discount-badge {
  background: #f6ffed;
  border: 1px solid #b7eb8f;
  color: #52c41a;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
}

.cost-tip {
  font-size: 12px;
  color: #909399;
}
</style>
