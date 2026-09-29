<template>
  <div class="lab-map-container">
    <div class="map-header">
      <div class="title-wrap">
        <el-icon class="icon"><Location /></el-icon>
        <span class="title">实验室 2D 数字孪生交互平面图</span>
        <el-tag size="small" type="success" effect="plain">4大实验平台 · 16台高精尖机位工况实时同步</el-tag>
      </div>
      <div class="legend">
        <span class="legend-item"><span class="dot available"></span>空闲就绪</span>
        <span class="legend-item"><span class="dot in-use"></span>实验运行中</span>
        <span class="legend-item"><span class="dot maintenance"></span>维保/检修</span>
      </div>
    </div>

    <!-- 2D 实验室平面画布 -->
    <div class="map-canvas">
      <!-- 中心通道 -->
      <div class="central-corridor horizontal">
        <span>主实验楼中央实验廊道 (East-West Main Corridor)</span>
      </div>
      <div class="central-corridor vertical"></div>

      <!-- 四大功能实验区 -->
      <div class="room-zone zone-a">
        <div class="zone-badge">A区 · 微观形貌与电子显微平台</div>
        <div class="sub-room sr-1">Lab-301 FE-SEM</div>
        <div class="sub-room sr-2">Lab-302 HR-TEM</div>
        <div class="sub-room sr-3">Lab-201 CLSM</div>
        <div class="sub-room sr-4">Lab-101 AFM</div>
      </div>

      <div class="room-zone zone-b">
        <div class="zone-badge">B区 · 晶体结构与物性表征平台</div>
        <div class="sub-room sr-1">Lab-305 XRD</div>
        <div class="sub-room sr-2">Lab-501 XPS</div>
        <div class="sub-room sr-3">Lab-402 PPMS-9T</div>
        <div class="sub-room sr-4">Lab-205 FLS1000</div>
      </div>

      <div class="room-zone zone-c">
        <div class="zone-badge">C区 · 高温合成与薄膜制备平台</div>
        <div class="sub-room sr-1">Lab-104 管式炉</div>
        <div class="sub-room sr-2">Lab-106 磁控溅射</div>
        <div class="sub-room sr-3">Lab-108 PECVD</div>
        <div class="sub-room sr-4">Lab-209 冻干机</div>
      </div>

      <div class="room-zone zone-d">
        <div class="zone-badge">D区 · 生化分离与色谱质谱平台</div>
        <div class="sub-room sr-1">Lab-208 SFC-MS</div>
        <div class="sub-room sr-2">Lab-101B 600M NMR</div>
        <div class="sub-room sr-3">Lab-210 高速离心机</div>
        <div class="sub-room sr-4">Lab-303 ICP-MS</div>
      </div>

      <!-- 渲染 16 台设备机位点标 -->
      <div
        v-for="eq in equipments"
        :key="eq.id"
        class="device-node"
        :class="eq.status"
        :style="{ left: eq.map_x + '%', top: eq.map_y + '%' }"
        @click="selectDevice(eq)"
      >
        <div class="node-pulse"></div>
        <div class="node-icon">
          <el-icon><Cpu /></el-icon>
        </div>
        <div class="node-tooltip">
          <div class="node-name">{{ eq.name }}</div>
          <div class="node-room">{{ eq.lab_room }} · ¥{{ eq.hourly_rate }}/h</div>
          <div class="node-status">{{ getStatusText(eq.status) }}</div>
        </div>
      </div>
    </div>

    <!-- 选定设备详情弹窗 -->
    <el-dialog v-model="dialogVisible" title="机位实时运行工况与健康档案" width="520px" destroy-on-close>
      <div v-if="selectedEq" class="device-modal-detail">
        <div class="device-title-row">
          <h4>{{ selectedEq.name }}</h4>
          <el-tag :type="getStatusTag(selectedEq.status)">{{ getStatusText(selectedEq.status) }}</el-tag>
        </div>
        <el-descriptions :column="1" border size="small" class="mt-3">
          <el-descriptions-item label="规格型号">{{ selectedEq.model }}</el-descriptions-item>
          <el-descriptions-item label="设备分类">{{ selectedEq.category }}</el-descriptions-item>
          <el-descriptions-item label="安放位置">{{ selectedEq.location }} ({{ selectedEq.lab_room }})</el-descriptions-item>
          <el-descriptions-item label="小时费率">¥ {{ selectedEq.hourly_rate }} / 小时</el-descriptions-item>
          <el-descriptions-item label="安全等级">
            <el-tag size="small" :type="selectedEq.safety_level >= 3 ? 'danger' : (selectedEq.safety_level === 2 ? 'warning' : 'info')">
              安全 {{ selectedEq.safety_level }} 级 ({{ selectedEq.safety_level >= 3 ? '高危特种' : (selectedEq.safety_level === 2 ? '大型精密' : '常规基础') }})
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="健康指数 (Health)">
            <div class="health-wrap">
              <span :style="{ color: getHealthColor(selectedEq.health_score), fontWeight: 'bold' }">{{ selectedEq.health_score }} 分</span>
              <el-progress :percentage="selectedEq.health_score" :color="getHealthColor(selectedEq.health_score)" />
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="伴热拼单">
            <el-tag v-if="selectedEq.support_batching" size="small" type="success">支持节能拼单 (8.5折)</el-tag>
            <span v-else class="text-muted">常规排期</span>
          </el-descriptions-item>
          <el-descriptions-item label="技术指标">{{ selectedEq.specs || '暂无详细参数' }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button
          type="primary"
          :disabled="selectedEq?.status === 'maintenance' || selectedEq?.status === 'faulty'"
          @click="handleBook"
        >
          立即预约此设备
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Location, Cpu } from '@element-plus/icons-vue'

const props = defineProps({
  equipments: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['book'])

const selectedEq = ref(null)
const dialogVisible = ref(false)

const selectDevice = (eq) => {
  selectedEq.value = eq
  dialogVisible.value = true
}

const handleBook = () => {
  if (!selectedEq.value) return
  dialogVisible.value = false
  emit('book', selectedEq.value)
}

const getStatusText = (status) => {
  const map = {
    available: '空闲就绪',
    in_use: '实验运行中',
    reserved: '已被预约',
    maintenance: '维保检修',
    faulty: '故障停用',
    scrapped: '报废'
  }
  return map[status] || status
}

const getStatusTag = (status) => {
  const map = {
    available: 'success',
    in_use: 'primary',
    reserved: 'warning',
    maintenance: 'danger',
    faulty: 'danger'
  }
  return map[status] || 'info'
}

const getHealthColor = (score) => {
  if (score >= 85) return '#67c23a'
  if (score >= 70) return '#e6a23c'
  return '#f56c6c'
}
</script>

<style scoped>
.lab-map-container {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  padding: 20px;
  margin-top: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.map-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
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
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.legend {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: #64748b;
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

.dot.available {
  background-color: #52c41a;
  box-shadow: 0 0 6px rgba(82, 196, 26, 0.6);
}

.dot.in-use {
  background-color: #1890ff;
  box-shadow: 0 0 6px rgba(24, 144, 255, 0.6);
}

.dot.maintenance {
  background-color: #ff4d4f;
  box-shadow: 0 0 6px rgba(255, 77, 79, 0.6);
}

/* 2D 画布 */
.map-canvas {
  position: relative;
  width: 100%;
  height: 580px;
  background: #0f172a;
  border-radius: 10px;
  border: 2px solid #1e293b;
  overflow: hidden;
  background-image: 
    linear-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.04) 1px, transparent 1px);
  background-size: 30px 30px;
}

/* 走廊 */
.central-corridor.horizontal {
  position: absolute;
  left: 0;
  right: 0;
  top: 48%;
  height: 4%;
  background: rgba(30, 41, 59, 0.85);
  border-top: 1px dashed rgba(148, 163, 184, 0.3);
  border-bottom: 1px dashed rgba(148, 163, 184, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: #64748b;
  letter-spacing: 1px;
  pointer-events: none;
  z-index: 1;
}

.central-corridor.vertical {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 48%;
  width: 4%;
  background: rgba(30, 41, 59, 0.85);
  border-left: 1px dashed rgba(148, 163, 184, 0.3);
  border-right: 1px dashed rgba(148, 163, 184, 0.3);
  pointer-events: none;
  z-index: 1;
}

/* 房间区域 */
.room-zone {
  position: absolute;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 6px;
  padding: 10px;
  background: rgba(30, 41, 59, 0.35);
  pointer-events: none;
}

.zone-badge {
  position: absolute;
  top: 8px;
  left: 12px;
  font-size: 12px;
  font-weight: 700;
  color: #93c5fd;
  background: rgba(15, 23, 42, 0.7);
  padding: 2px 8px;
  border-radius: 4px;
}

.sub-room {
  position: absolute;
  font-size: 10px;
  color: #94a3b8;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(15, 23, 42, 0.4);
}

.zone-a {
  left: 2%;
  top: 2%;
  width: 45%;
  height: 44%;
}
.zone-a .sr-1 { left: 8%; top: 38%; }
.zone-a .sr-2 { left: 54%; top: 38%; }
.zone-a .sr-3 { left: 8%; top: 76%; }
.zone-a .sr-4 { left: 54%; top: 76%; }

.zone-b {
  right: 2%;
  top: 2%;
  width: 45%;
  height: 44%;
}
.zone-b .sr-1 { left: 8%; top: 38%; }
.zone-b .sr-2 { left: 54%; top: 38%; }
.zone-b .sr-3 { left: 8%; top: 76%; }
.zone-b .sr-4 { left: 54%; top: 76%; }

.zone-c {
  left: 2%;
  bottom: 2%;
  width: 45%;
  height: 44%;
}
.zone-c .sr-1 { left: 8%; top: 38%; }
.zone-c .sr-2 { left: 54%; top: 38%; }
.zone-c .sr-3 { left: 8%; top: 76%; }
.zone-c .sr-4 { left: 54%; top: 76%; }

.zone-d {
  right: 2%;
  bottom: 2%;
  width: 45%;
  height: 44%;
}
.zone-d .sr-1 { left: 8%; top: 38%; }
.zone-d .sr-2 { left: 54%; top: 38%; }
.zone-d .sr-3 { left: 8%; top: 76%; }
.zone-d .sr-4 { left: 54%; top: 76%; }

/* 设备机位点标 */
.device-node {
  position: absolute;
  transform: translate(-50%, -50%);
  cursor: pointer;
  z-index: 10;
  transition: transform 0.2s ease;
}

.device-node:hover {
  transform: translate(-50%, -50%) scale(1.25);
  z-index: 50;
}

.node-pulse {
  position: absolute;
  width: 38px;
  height: 38px;
  left: -5px;
  top: -5px;
  border-radius: 50%;
  animation: pulse 2s infinite ease-in-out;
  opacity: 0.6;
}

.node-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
}

.device-node.available .node-icon { background: #52c41a; }
.device-node.available .node-pulse { border: 2px solid #52c41a; }

.device-node.in_use .node-icon { background: #1890ff; }
.device-node.in_use .node-pulse { border: 2px solid #1890ff; animation-duration: 1.2s; }

.device-node.maintenance .node-icon { background: #ff4d4f; }
.device-node.maintenance .node-pulse { border: 2px solid #ff4d4f; }

@keyframes pulse {
  0% { transform: scale(0.85); opacity: 0.8; }
  50% { transform: scale(1.3); opacity: 0.2; }
  100% { transform: scale(0.85); opacity: 0.8; }
}

.node-tooltip {
  display: none;
  position: absolute;
  bottom: 34px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  padding: 6px 12px;
  border-radius: 6px;
  white-space: nowrap;
  font-size: 11px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  pointer-events: none;
}

.device-node:hover .node-tooltip {
  display: block;
}

.node-tooltip .node-name {
  font-weight: 700;
  color: #ffffff;
}

.node-tooltip .node-room {
  color: #93c5fd;
  font-size: 10px;
  margin: 2px 0;
}

.node-tooltip .node-status {
  color: #a7f3d0;
  font-size: 10px;
}

.device-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.device-title-row h4 {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.health-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}

.health-wrap .el-progress {
  flex: 1;
}

.text-muted {
  color: #94a3b8;
}
</style>
