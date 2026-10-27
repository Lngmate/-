<template>
  <div class="task-page">
    <div class="page-heading">
      <div><h1>实验任务</h1><p>将多台设备的使用安排连接成一次完整实验。</p></div>
      <el-button :loading="listLoading" @click="loadTasks">刷新任务</el-button>
    </div>

    <div class="planner-layout">
      <el-card shadow="never" class="task-form-card">
        <template #header><div class="section-heading"><el-icon><Connection /></el-icon><span>安排新实验</span><el-tag size="small" effect="plain">2–3 道工序</el-tag></div></template>
        <el-form label-position="top" :disabled="submitting">
          <el-form-item label="任务名称" required><el-input v-model="form.name" maxlength="100" placeholder="例如：薄膜材料制备与结构表征" /></el-form-item>
          <div class="window-fields">
            <el-form-item label="最早开始时间" required><el-date-picker v-model="form.earliest_start" type="datetime" placeholder="选择开始时间" /></el-form-item>
            <el-form-item label="最晚完成时间" required><el-date-picker v-model="form.deadline" type="datetime" placeholder="选择完成时间" /></el-form-item>
          </div>
          <div v-for="(step, index) in form.steps" :key="step.key" class="process-step">
            <div class="step-heading"><span class="step-number">{{ index + 1 }}</span><strong>工序 {{ index + 1 }}</strong><div class="step-actions">
              <el-button text size="small" :disabled="index === 0" @click="moveStep(index, -1)">上移</el-button>
              <el-button text size="small" :disabled="index === form.steps.length - 1" @click="moveStep(index, 1)">下移</el-button>
              <el-button text size="small" type="danger" :disabled="form.steps.length <= 2" @click="removeStep(index)">删除</el-button>
            </div></div>
            <el-form-item label="工序名称" required><el-input v-model="step.name" maxlength="80" placeholder="例如：高温热处理" /></el-form-item>
            <el-form-item label="使用设备" required>
              <el-select v-model="step.equipment_id" filterable placeholder="选择设备" :loading="equipmentLoading">
                <el-option v-for="eq in equipments" :key="eq.id" :label="`${eq.name} · ¥${eq.hourly_rate}/小时`" :value="eq.id" :disabled="['maintenance','faulty','scrapped'].includes(eq.status)" />
              </el-select>
            </el-form-item>
            <div class="window-fields">
              <el-form-item label="使用时长（分钟）"><el-input-number v-model="step.duration_minutes" :min="30" :max="1440" :step="30" step-strictly /></el-form-item>
              <el-form-item v-if="index > 0" label="与前一道工序最短间隔（分钟）"><el-input-number v-model="step.gap_minutes" :min="0" :max="10080" :step="15" step-strictly /></el-form-item>
            </div>
          </div>
          <el-button v-if="form.steps.length < 3" plain class="add-step" @click="addStep"><el-icon><Plus /></el-icon>添加工序</el-button>
          <el-form-item label="实验用途"><el-input v-model="form.purpose" type="textarea" :rows="2" maxlength="1000" /></el-form-item>
          <el-button type="primary" :loading="planning" :disabled="equipmentLoading" @click="generatePlan">生成排期方案</el-button>
        </el-form>
      </el-card>

      <div class="plan-column">
        <el-card shadow="never" class="plan-card" v-loading="planning">
          <template #header><div class="section-heading"><el-icon><Calendar /></el-icon><span>排期方案</span></div></template>
          <el-alert v-if="planError" :title="planError" type="error" :closable="false" show-icon class="plan-error" />
          <el-empty v-if="!plan" description="填写实验工序，查看可用时间安排" :image-size="90" />
          <template v-else>
            <div class="plan-overview"><div><span>预计完成</span><strong>{{ dateTime(plan.end_time) }}</strong></div><div><span>预计机时费</span><strong>¥{{ money(plan.total_cost) }}</strong></div></div>
            <ol class="plan-timeline">
              <li v-for="(step, index) in plan.steps" :key="index">
                <div v-if="index > 0" class="gap-note">间隔至少 {{ step.gap_minutes }} 分钟<span v-if="step.wait_minutes"> · 等待设备空档 {{ step.wait_minutes }} 分钟</span></div>
                <div v-else-if="step.wait_minutes" class="gap-note">等待设备空档 {{ step.wait_minutes }} 分钟</div>
                <div class="timeline-item"><span class="step-number">{{ index + 1 }}</span><div><strong>{{ step.name }}</strong><p>{{ step.equipment_name }}</p><time>{{ dateTime(step.start_time) }} — {{ dateTime(step.end_time) }}</time><div class="step-meta"><el-tag size="small" :type="step.approval_type === 'none' ? 'success' : 'warning'">{{ approvalText(step.approval_type) }}</el-tag><span>{{ step.duration_minutes }} 分钟 · ¥{{ money(step.total_cost) }}</span></div></div></div>
              </li>
            </ol>
            <p class="plan-help">方案尚未占用设备。确认后，各工序按设备规则审批，完成使用后分别结算。</p>
            <el-button type="primary" class="confirm-plan" :loading="submitting" :disabled="planning" @click="submitPlan">确认并预约全部工序</el-button>
          </template>
        </el-card>
        <el-alert title="冷却与转运时间可填写为工序间隔；如果期间仍需占用上一台设备，请计入该设备使用时长。" type="info" :closable="false" show-icon />
      </div>
    </div>

    <el-card shadow="never" class="history-card" v-loading="listLoading">
      <template #header><div class="section-heading"><el-icon><Document /></el-icon><span>我的实验任务</span><el-tag size="small" type="info">{{ tasks.length }}</el-tag></div></template>
      <el-alert v-if="listError" :title="listError" type="error" :closable="false" />
      <el-empty v-else-if="!tasks.length" description="还没有实验任务" :image-size="70" />
      <div v-for="task in tasks" :key="task.id" class="saved-task">
        <div class="saved-heading"><div><span class="task-ref">#{{ task.id }}</span><strong>{{ task.name }}</strong><p>提交于 {{ dateTime(task.created_at) }}</p></div><el-button plain @click="router.push('/my-reservations')">签到与履约</el-button></div>
        <el-alert v-if="task.reservations.some(r => ['cancelled','rejected','no_show'].includes(r.status))" title="部分工序未能继续，请确认剩余预约是否仍然需要，并及时取消不需要的工序。" type="warning" :closable="false" />
        <el-table :data="task.reservations" class="task-reservations">
          <el-table-column prop="task_step" label="顺序" width="65" />
          <el-table-column prop="experiment_name" label="工序" min-width="120" />
          <el-table-column prop="equipment_name" label="设备" min-width="180" />
          <el-table-column label="使用时间" min-width="230"><template #default="{ row }">{{ dateTime(row.start_time) }}<br />至 {{ dateTime(row.end_time) }}</template></el-table-column>
          <el-table-column label="状态" min-width="120"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="机时费" width="110"><template #default="{ row }">¥{{ money(row.total_cost) }}</template></el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '../api/request'

const router = useRouter()
let stepKey = 0
const newStep = () => ({ key: ++stepKey, name: '', equipment_id: null, duration_minutes: 60, gap_minutes: 0 })
const nextHour = new Date(Math.ceil((Date.now() + 3600000) / 3600000) * 3600000)
const form = reactive({ name: '', purpose: '', earliest_start: nextHour, deadline: new Date(nextHour.getTime() + 3 * 86400000), steps: [newStep(), newStep()] })
const equipments = ref([])
const tasks = ref([])
const plan = ref(null)
const planRequest = ref(null)
const planning = ref(false)
const submitting = ref(false)
const equipmentLoading = ref(false)
const listLoading = ref(false)
const planError = ref('')
const listError = ref('')
let revision = 0

watch(form, () => { revision++; plan.value = null; planRequest.value = null; planError.value = '' }, { deep: true, flush: 'sync' })

function addStep() { if (form.steps.length < 3) form.steps.push(newStep()) }
function removeStep(index) { form.steps.splice(index, 1); form.steps[0].gap_minutes = 0 }
function moveStep(index, direction) {
  const target = index + direction
  if (target < 0 || target >= form.steps.length) return
  const step = form.steps.splice(index, 1)[0]
  form.steps.splice(target, 0, step)
  form.steps[0].gap_minutes = 0
}
const errorText = error => error.response?.data?.message || error.message || '请求失败，请重试'
const money = value => Number(value || 0).toFixed(2)
const dateTime = value => new Date(value).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false })
const approvalText = policy => ({ none: '免审批', single: '设备管理员审批', dual: '导师与管理员审批' }[policy] || '设备管理员审批')
const statusText = status => ({ pending_advisor: '待导师审批', pending_admin: '待管理员审批', approved: '待签到', active: '使用中', completed: '已完成', rejected: '已驳回', cancelled: '已取消', no_show: '已爽约' }[status] || status)
const statusType = status => ['approved', 'completed'].includes(status) ? 'success' : ['rejected', 'cancelled', 'no_show'].includes(status) ? 'danger' : status === 'active' ? 'primary' : 'warning'

async function generatePlan() {
  plan.value = null
  planRequest.value = null
  planError.value = ''
  if (!form.name.trim() || !form.earliest_start || !form.deadline || form.steps.some(step => !step.name.trim() || !step.equipment_id || !Number.isInteger(step.duration_minutes) || !Number.isInteger(step.gap_minutes))) {
    planError.value = '请填写任务名称、时间窗口以及每道工序的名称、设备和时长。'
    return
  }
  const currentRevision = revision
  const payload = { name: form.name.trim(), purpose: form.purpose, earliest_start: new Date(form.earliest_start).toISOString(), deadline: new Date(form.deadline).toISOString(), steps: form.steps.map(({ key, ...step }, index) => ({ ...step, gap_minutes: index === 0 ? 0 : step.gap_minutes })) }
  planning.value = true
  try {
    const response = await request.post('/tasks/preview', payload)
    if (currentRevision !== revision) return
    plan.value = response.data
    planRequest.value = payload
  } catch (error) { if (currentRevision === revision) planError.value = errorText(error) }
  finally { planning.value = false }
}

async function submitPlan() {
  if (!plan.value || !planRequest.value || submitting.value) return
  submitting.value = true
  try {
    const response = await request.post('/tasks', { ...planRequest.value, start_times: plan.value.steps.map(step => step.start_time) })
    plan.value = null
    planRequest.value = null
    ElMessage.success(response.message)
    await loadTasks()
  } catch (error) {
    planError.value = errorText(error)
    plan.value = null
    planRequest.value = null
  } finally { submitting.value = false }
}

async function loadTasks() {
  listLoading.value = true
  listError.value = ''
  try { tasks.value = (await request.get('/tasks')).data || [] }
  catch (error) { listError.value = errorText(error) }
  finally { listLoading.value = false }
}

onMounted(async () => {
  equipmentLoading.value = true
  await Promise.all([
    loadTasks(),
    request.get('/equipments').then(response => { equipments.value = response.data || [] }).catch(error => { planError.value = errorText(error) }).finally(() => { equipmentLoading.value = false })
  ])
})
</script>

<style scoped>
.task-page { max-width: 1440px; margin: 0 auto; color: #1e293b; }
.page-heading, .saved-heading { display: flex; justify-content: space-between; align-items: center; gap: 16px; margin-bottom: 22px; }
.page-heading h1 { font-size: 26px; margin: 0 0 8px; }
.page-heading p, .saved-heading p { color: #64748b; font-size: 14px; margin: 0; }
.planner-layout { display: grid; grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr); gap: 24px; align-items: start; }
.section-heading { display: flex; align-items: center; gap: 10px; font-weight: 600; font-size: 16px; }
.section-heading .el-icon { color: #1890ff; }
.section-heading .el-tag { margin-left: auto; }
.window-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.task-page :deep(.el-select), .task-page :deep(.el-date-editor), .task-page :deep(.el-input-number) { width: 100%; }
.process-step { border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin-bottom: 16px; }
.step-heading { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; }
.step-number { display: inline-flex; flex-shrink: 0; width: 28px; height: 28px; border-radius: 50%; align-items: center; justify-content: center; background: #e6f7ff; color: #096dd9; font-size: 14px; font-weight: 600; }
.step-actions { margin-left: auto; display: flex; }
.step-actions .el-button + .el-button { margin-left: 0; }
.add-step { width: 100%; margin-bottom: 20px; border-style: dashed; }
.plan-column { display: grid; gap: 16px; }
.plan-overview { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; border-bottom: 1px solid #e2e8f0; padding-bottom: 22px; }
.plan-overview span { display: block; font-size: 13px; color: #64748b; margin-bottom: 8px; }
.plan-overview strong { font-size: 18px; }
.plan-timeline { list-style: none; margin: 24px 0; padding: 0; }
.plan-timeline li + li { margin-top: 20px; }
.timeline-item { display: flex; gap: 14px; }
.timeline-item p { color: #475569; margin: 8px 0; font-size: 14px; }
.timeline-item time { font-size: 13px; line-height: 1.8; }
.step-meta { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; font-size: 13px; color: #64748b; margin-top: 10px; }
.gap-note { font-size: 12px; color: #64748b; padding: 0 0 14px 42px; }
.plan-help { color: #64748b; font-size: 13px; line-height: 1.8; }
.confirm-plan { width: 100%; }
.plan-error { margin-bottom: 16px; }
.history-card { margin-top: 24px; }
.saved-task + .saved-task { margin-top: 24px; border-top: 1px solid #e2e8f0; padding-top: 24px; }
.saved-heading { margin-bottom: 14px; }
.saved-heading p { margin-top: 8px; font-size: 12px; }
.task-ref { color: #64748b; font-size: 13px; margin-right: 10px; }
.task-reservations { margin-top: 12px; }
@media (max-width: 1100px) { .planner-layout { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .window-fields, .plan-overview { grid-template-columns: 1fr; } .page-heading { align-items: flex-start; } .step-heading { flex-wrap: wrap; } }
</style>
