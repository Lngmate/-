<template>
  <div class="user-manage-page">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <el-icon class="icon"><UserFilled /></el-icon>
            <span class="title">实验室用户权限中心与实名台账</span>
          </div>
          <div class="header-actions">
            <el-button type="primary" plain size="small" @click="fetchUsers">
              <el-icon><Refresh /></el-icon> 刷新用户列表
            </el-button>
          </div>
        </div>
      </template>

      <!-- 检索筛选栏 -->
      <div class="filters-row">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索学号/姓名/邮箱/手机号..."
          prefix-icon="Search"
          clearable
          style="width: 260px"
          @input="fetchUsers"
        />
        <el-select
          v-model="filters.role"
          placeholder="角色身份"
          clearable
          style="width: 160px"
          @change="fetchUsers"
        >
          <el-option label="普通学生" value="student" />
          <el-option label="导师/课题组长" value="advisor" />
          <el-option label="设备管理员" value="equipment_admin" />
          <el-option label="系统管理员" value="system_admin" />
        </el-select>

        <el-select
          v-model="filters.group_id"
          placeholder="归属课题组"
          clearable
          style="width: 220px"
          @change="fetchUsers"
        >
          <el-option
            v-for="grp in groups"
            :key="grp.id"
            :label="grp.name"
            :value="grp.id"
          />
        </el-select>

        <el-select
          v-model="filters.is_frozen"
          placeholder="账号状态"
          clearable
          style="width: 140px"
          @change="fetchUsers"
        >
          <el-option label="正常激活" value="false" />
          <el-option label="已冻结" value="true" />
        </el-select>
      </div>

      <!-- 用户数据表格 -->
      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        
        <el-table-column label="登录账号 / 学号" min-width="140">
          <template #default="{ row }">
            <div class="user-cell">
              <span class="uname">{{ row.username }}</span>
              <el-tag v-if="row.username === 'sysadmin'" size="small" type="danger" effect="plain">超管</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="real_name" label="真实姓名" min-width="130" />

        <el-table-column label="角色身份" width="140">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)">
              {{ getRoleName(row.role) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="group_name" label="归属科研课题组" min-width="180">
          <template #default="{ row }">
            <span v-if="row.group_name">{{ row.group_name }}</span>
            <span v-else class="text-muted">全校公共中心</span>
          </template>
        </el-table-column>

        <el-table-column label="信誉分" width="110">
          <template #default="{ row }">
            <el-tag :type="row.credit_score >= 90 ? 'success' : (row.credit_score >= 80 ? 'primary' : 'warning')">
              {{ row.credit_score }} 分
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="联系方式" min-width="180">
          <template #default="{ row }">
            <div class="contact-box">
              <div>{{ row.phone || '--' }}</div>
              <div class="email">{{ row.email || '--' }}</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_frozen ? 'danger' : 'success'">
              {{ row.is_frozen ? '已冻结' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="管理操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="openEditDialog(row)">
              编辑信息
            </el-button>
            <el-button type="warning" size="small" link @click="openCreditDialog(row)">
              调整信誉
            </el-button>
            <el-button
              :type="row.is_frozen ? 'success' : 'info'"
              size="small"
              link
              :disabled="row.username === 'sysadmin'"
              @click="toggleFreeze(row)"
            >
              {{ row.is_frozen ? '解冻' : '冻结' }}
            </el-button>
            <el-button
              type="danger"
              size="small"
              link
              :disabled="row.username === 'sysadmin'"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑资料对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑用户信息与权限" width="500px" destroy-on-close>
      <el-form :model="editForm" label-width="110px">
        <el-form-item label="登录账号">
          <el-input :model-value="selectedUser?.username" disabled />
        </el-form-item>
        <el-form-item label="真实姓名" required>
          <el-input v-model="editForm.real_name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="角色权限" required>
          <el-select v-model="editForm.role" style="width: 100%">
            <el-option label="普通学生" value="student" />
            <el-option label="导师/课题组长" value="advisor" />
            <el-option label="设备管理员" value="equipment_admin" />
            <el-option label="系统管理员" value="system_admin" />
          </el-select>
        </el-form-item>
        <el-form-item label="归属课题组">
          <el-select v-model="editForm.group_id" style="width: 100%" clearable placeholder="分配课题组">
            <el-option
              v-for="grp in groups"
              :key="grp.id"
              :label="grp.name"
              :value="grp.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="联系邮箱">
          <el-input v-model="editForm.email" placeholder="邮箱" />
        </el-form-item>
        <el-form-item label="手机号码">
          <el-input v-model="editForm.phone" placeholder="手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleUpdate">保存修改</el-button>
      </template>
    </el-dialog>

    <!-- 调整信誉积分对话框 -->
    <el-dialog v-model="creditDialogVisible" title="手动调节科研履约信誉积分" width="460px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="用户姓名">
          <el-input :model-value="`${selectedUser?.real_name} (${selectedUser?.username})`" disabled />
        </el-form-item>
        <el-form-item label="当前信誉分">
          <el-tag size="large">{{ selectedUser?.credit_score }} 分</el-tag>
        </el-form-item>
        <el-form-item label="积分调整增量" required>
          <el-input-number v-model="creditDelta" :min="-50" :max="50" style="width: 160px" />
          <span class="sub-hint"> (正数加分，负数扣分)</span>
        </el-form-item>
        <el-form-item label="调整原因说明" required>
          <el-input
            v-model="creditReason"
            type="textarea"
            rows="3"
            placeholder="如：表彰协助设备深度维护 / 违规带入外源危险化学品"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="creditDialogVisible = false">取消</el-button>
        <el-button type="warning" :loading="submitting" @click="handleAdjustCredit">确认调整</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import request from '../api/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UserFilled, Refresh } from '@element-plus/icons-vue'

const loading = ref(false)
const submitting = ref(false)
const users = ref([])
const groups = ref([])

const filters = reactive({
  keyword: '',
  role: '',
  group_id: '',
  is_frozen: ''
})

const selectedUser = ref(null)
const editDialogVisible = ref(false)
const editForm = reactive({
  real_name: '',
  role: '',
  group_id: null,
  email: '',
  phone: ''
})

const creditDialogVisible = ref(false)
const creditDelta = ref(5)
const creditReason = ref('')

const getRoleName = (role) => {
  const map = {
    student: '普通学生',
    advisor: '导师/组长',
    equipment_admin: '设备管理员',
    system_admin: '系统管理员'
  }
  return map[role] || role
}

const getRoleTagType = (role) => {
  const map = {
    student: 'info',
    advisor: 'primary',
    equipment_admin: 'warning',
    system_admin: 'danger'
  }
  return map[role] || 'info'
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await request.get('/users', { params: filters })
    if (res.data) {
      users.value = res.data
    }
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const fetchGroups = async () => {
  try {
    const res = await request.get('/groups')
    if (res.data) {
      groups.value = res.data
    }
  } catch (err) {
    console.error(err)
  }
}

const openEditDialog = (user) => {
  selectedUser.value = user
  editForm.real_name = user.real_name
  editForm.role = user.role
  editForm.group_id = user.group_id
  editForm.email = user.email
  editForm.phone = user.phone
  editDialogVisible.value = true
}

const handleUpdate = async () => {
  if (!selectedUser.value) return
  submitting.value = true
  try {
    await request.put(`/users/${selectedUser.value.id}`, editForm)
    ElMessage.success('用户信息已更新')
    editDialogVisible.value = false
    fetchUsers()
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
}

const openCreditDialog = (user) => {
  selectedUser.value = user
  creditDelta.value = 5
  creditReason.value = ''
  creditDialogVisible.value = true
}

const handleAdjustCredit = async () => {
  if (!selectedUser.value || !creditReason.value) {
    ElMessage.warning('请填写调整原因')
    return
  }
  submitting.value = true
  try {
    await request.post('/credits/adjust', {
      user_id: selectedUser.value.id,
      delta: creditDelta.value,
      reason: creditReason.value
    })
    ElMessage.success('信誉积分已成功调整')
    creditDialogVisible.value = false
    fetchUsers()
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
}

const toggleFreeze = async (user) => {
  const actionText = user.is_frozen ? '解冻' : '冻结'
  try {
    await ElMessageBox.confirm(`确定要${actionText}用户【${user.real_name}】的账户吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: user.is_frozen ? 'info' : 'warning'
    })
    await request.put(`/users/${user.id}/freeze`)
    ElMessage.success(`用户已成功${actionText}`)
    fetchUsers()
  } catch (e) {
    // cancelled
  }
}

const handleDelete = async (user) => {
  try {
    await ElMessageBox.confirm(`警告：确定要彻底删除用户【${user.real_name}】吗？此操作不可逆！`, '危险警告', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    await request.delete(`/users/${user.id}`)
    ElMessage.success('用户已删除')
    fetchUsers()
  } catch (e) {
    // cancelled
  }
}

onMounted(() => {
  fetchUsers()
  fetchGroups()
})
</script>

<style scoped>
.user-manage-page {
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
  gap: 8px;
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

.filters-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 20px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.user-cell .uname {
  font-weight: 600;
  color: #1e293b;
}

.contact-box {
  font-size: 12px;
}

.contact-box .email {
  color: #64748b;
  font-size: 11px;
}

.text-muted {
  color: #94a3b8;
}

.sub-hint {
  font-size: 12px;
  color: #64748b;
  margin-left: 8px;
}
</style>
