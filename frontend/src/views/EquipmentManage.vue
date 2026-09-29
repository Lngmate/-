<template>
  <div class="equipment-manage-page">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <el-icon class="icon"><Setting /></el-icon>
            <span class="title">实验室设备全生命周期台账与维保管控</span>
          </div>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon> 登记新增设备
          </el-button>
        </div>
      </template>

      <el-table :data="equipments" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        
        <el-table-column prop="name" label="设备名称" min-width="180">
          <template #default="{ row }">
            <div class="name-cell">
              <span class="title">{{ row.name }}</span>
              <span class="sub">{{ row.model }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="category" label="分类" width="140" />

        <el-table-column label="安放位置" min-width="160">
          <template #default="{ row }">
            {{ row.location }} ({{ row.lab_room }})
          </template>
        </el-table-column>

        <el-table-column label="小时费率" width="110">
          <template #default="{ row }">
            ¥ {{ row.hourly_rate }} / h
          </template>
        </el-table-column>

        <!-- 创新点：健康度预测指标与预警 -->
        <el-table-column label="健康指数 (Health Score)" width="180">
          <template #default="{ row }">
            <div class="health-cell">
              <div class="score-line">
                <span>{{ row.health_score }} 分</span>
                <el-tag size="small" :type="row.health_score >= 85 ? 'success' : (row.health_score >= 70 ? 'warning' : 'danger')">
                  {{ row.health_score >= 85 ? '良好' : (row.health_score >= 70 ? '建议巡检' : '强制维保') }}
                </el-tag>
              </div>
              <el-progress :percentage="row.health_score" :color="getHealthColor(row.health_score)" :show-text="false" />
            </div>
          </template>
        </el-table-column>

        <el-table-column label="运行状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="管理操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button
              type="warning"
              size="small"
              plain
              @click="openMaintenanceDialog(row)"
            >
              预防性维保
            </el-button>
            <el-button
              type="primary"
              size="small"
              link
              @click="openEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              size="small"
              link
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 预防性维保对话框（创新点） -->
    <el-dialog v-model="maintDialogVisible" title="触发设备预防性维保与标定" width="480px" destroy-on-close>
      <el-alert
        type="info"
        title="维保说明：完成预防性维护或标定后，系统将自动把该设备健康指数恢复至 100 分，并记录维保审计台账。"
        show-icon
        :closable="false"
        style="margin-bottom: 16px"
      />
      <el-form label-width="100px">
        <el-form-item label="维保设备">
          <el-input :model-value="selectedEq?.name" disabled />
        </el-form-item>
        <el-form-item label="维保类型" required>
          <el-select v-model="maintForm.type" style="width: 100%">
            <el-option label="日常精密巡检与清洁" value="routine" />
            <el-option label="定期计量与标定校准" value="calibration" />
            <el-option label="易损耗材更换与检修" value="repair" />
          </el-select>
        </el-form-item>
        <el-form-item label="维保说明">
          <el-input v-model="maintForm.description" type="textarea" placeholder="填写维护说明，如分子泵换油、光路校准等" />
        </el-form-item>
        <el-form-item label="维保成本(元)">
          <el-input-number v-model="maintForm.cost" :min="0" :step="100" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="maintDialogVisible = false">取消</el-button>
        <el-button type="success" :loading="submitting" @click="handleMaintenance">
          确认完成维保并恢复健康度
        </el-button>
      </template>
    </el-dialog>

    <!-- 新增/编辑设备对话框 -->
    <el-dialog v-model="editDialogVisible" :title="isEdit ? '编辑设备' : '登记新增设备'" width="580px" destroy-on-close>
      <el-form :model="editForm" label-width="110px">
        <el-form-item label="设备名称" required>
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="规格型号">
          <el-input v-model="editForm.model" />
        </el-form-item>
        <el-form-item label="所属分类">
          <el-select v-model="editForm.category" style="width: 100%">
            <el-option label="显微镜与微观表征" value="显微镜与微观表征" />
            <el-option label="晶体结构与物性表征" value="晶体结构与物性表征" />
            <el-option label="高温合成与薄膜制备" value="高温合成与薄膜制备" />
            <el-option label="生化分离与质谱色谱" value="生化分离与质谱色谱" />
          </el-select>
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="14">
            <el-form-item label="物理位置" required>
              <el-input v-model="editForm.location" placeholder="如：理科实验大楼 301 室" />
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item label="房间编号">
              <el-input v-model="editForm.lab_room" placeholder="Lab-301" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="小时单价">
              <el-input-number v-model="editForm.hourly_rate" :min="0" :step="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="安全等级">
              <el-select v-model="editForm.safety_level" style="width: 100%">
                <el-option label="1 级 (基础常规)" :value="1" />
                <el-option label="2 级 (大型精密)" :value="2" />
                <el-option label="3 级 (危险特种)" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="审批策略">
          <el-select v-model="editForm.approval_type" style="width: 100%">
            <el-option label="免审直通（学生预约立即生效）" value="none" />
            <el-option label="单审（设备管理员排期审核）" value="single" />
            <el-option label="双审（导师审经费资质 + 管理员审机时）" value="dual" />
          </el-select>
        </el-form-item>
        <el-form-item label="技术指标">
          <el-input v-model="editForm.specs" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSaveEquipment">
          保存设备
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../api/request'
import { Setting, Plus } from '@element-plus/icons-vue'

const equipments = ref([])
const loading = ref(false)
const submitting = ref(false)

const maintDialogVisible = ref(false)
const editDialogVisible = ref(false)
const isEdit = ref(false)
const selectedEq = ref(null)

const maintForm = ref({ type: 'routine', description: '完成设备定期清洁、光路微调与状态标定', cost: 200 })
const editForm = ref({
  name: '',
  model: '',
  category: '显微镜与微观表征',
  location: '',
  lab_room: '',
  hourly_rate: 100,
  safety_level: 1,
  approval_type: 'single',
  specs: ''
})

const fetchEquipments = async () => {
  loading.value = true
  try {
    const res = await request.get('/equipments')
    if (res.data) equipments.value = res.data
  } catch (err) {
  } finally {
    loading.value = false
  }
}

const openMaintenanceDialog = (row) => {
  selectedEq.value = row
  maintDialogVisible.value = true
}

const handleMaintenance = async () => {
  submitting.value = true
  try {
    const res = await request.post(`/equipments/${selectedEq.value.id}/maintenance`, maintForm.value)
    ElMessage.success(res.message || '维保完成！健康度已重置为100')
    maintDialogVisible.value = false
    fetchEquipments()
  } catch (err) {
  } finally {
    submitting.value = false
  }
}

const openCreateDialog = () => {
  isEdit.value = false
  editForm.value = {
    name: '',
    model: '',
    category: '显微镜与微观表征',
    location: '',
    lab_room: '',
    hourly_rate: 100,
    safety_level: 1,
    approval_type: 'single',
    specs: ''
  }
  editDialogVisible.value = true
}

const openEditDialog = (row) => {
  isEdit.value = true
  selectedEq.value = row
  editForm.value = { ...row }
  editDialogVisible.value = true
}

const handleSaveEquipment = async () => {
  if (!editForm.value.name) {
    ElMessage.warning('请输入设备名称')
    return
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      await request.put(`/equipments/${selectedEq.value.id}`, editForm.value)
      ElMessage.success('更新成功！')
    } else {
      await request.post('/equipments', editForm.value)
      ElMessage.success('新增设备成功！')
    }
    editDialogVisible.value = false
    fetchEquipments()
  } catch (err) {
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要彻底删除设备【${row.name}】吗？若存在进行中或审批中的预约将拦截删除。`, '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await request.delete(`/equipments/${row.id}`)
    ElMessage.success('设备已从台账中成功删除')
    fetchEquipments()
  } catch (e) {
    // cancelled
  }
}

const getHealthColor = (score) => {
  if (score >= 90) return '#67C23A'
  if (score >= 75) return '#E6A23C'
  return '#F56C6C'
}

const getStatusText = (status) => {
  const map = { available: '空闲就绪', in_use: '实验运行中', reserved: '已被预约', maintenance: '维保检修', faulty: '故障停用' }
  return map[status] || status
}

const getStatusTag = (status) => {
  const map = { available: 'success', in_use: 'primary', reserved: 'warning', maintenance: 'danger', faulty: 'danger' }
  return map[status] || 'info'
}

onMounted(() => {
  fetchEquipments()
})
</script>

<style scoped>
.equipment-manage-page {
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

.name-cell {
  display: flex;
  flex-direction: column;
}

.name-cell .title {
  font-weight: bold;
  color: #1e293b;
}

.name-cell .sub {
  font-size: 12px;
  color: #64748b;
}

.health-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.score-line {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  font-weight: bold;
}
</style>
