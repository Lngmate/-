<template>
  <div class="register-container">
    <div class="register-bg-overlay"></div>

    <div class="register-card">
      <div class="register-header">
        <div class="logo-box">
          <el-icon><Cpu /></el-icon>
        </div>
        <div class="header-text">
          <h2>注册高校实验室科研账户</h2>
          <p>加入全校大型仪器开放共享资源池，开启高效率科研实验预约</p>
        </div>
      </div>

      <el-form
        ref="regFormRef"
        :model="regForm"
        :rules="regRules"
        label-position="top"
        class="reg-form"
        @keyup.enter="handleRegister"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="登录账号 / 学号 / 工号" prop="username">
              <el-input
                v-model="regForm.username"
                placeholder="建议填写校园统一学号或工号"
                prefix-icon="User"
                clearable
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="真实姓名" prop="real_name">
              <el-input
                v-model="regForm.real_name"
                placeholder="实验室实名制安全准入要求"
                prefix-icon="Postcard"
                clearable
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="登录密码" prop="password">
              <el-input
                v-model="regForm.password"
                type="password"
                placeholder="至少 3 位密码"
                prefix-icon="Lock"
                show-password
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input
                v-model="regForm.confirmPassword"
                type="password"
                placeholder="请再次输入密码"
                prefix-icon="Lock"
                show-password
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="科研身份类型" prop="role">
              <el-select v-model="regForm.role" placeholder="请选择您的身份" style="width: 100%">
                <el-option label="研究生 / 本科生 (学生)" value="student" />
                <el-option label="课题组长 / 导师 (审核与经费划拨)" value="advisor" />
                <el-option label="实验技术员 / 设备管理员 (排期与维保)" value="equipment_admin" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="归属科研课题组" prop="group_id">
              <el-select
                v-model="regForm.group_id"
                placeholder="选择所在课题组 (用于经费自动核销)"
                style="width: 100%"
                filterable
              >
                <el-option
                  v-for="grp in groups"
                  :key="grp.id"
                  :label="`${grp.name} (${grp.department} - ${grp.leader_name})`"
                  :value="grp.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="电子邮箱" prop="email">
              <el-input
                v-model="regForm.email"
                placeholder="用于接收审批通过与上机通知"
                prefix-icon="Message"
                clearable
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号码" prop="phone">
              <el-input
                v-model="regForm.phone"
                placeholder="用于接收动态核销签到短信"
                prefix-icon="Phone"
                clearable
              />
            </el-form-item>
          </el-col>
        </el-row>

        <div class="form-tip">
          <el-icon><InfoFilled /></el-icon>
          <span>提示：注册成功后将自动分配 100 点初始科研信用基础分，享有标准预约与同温伴热拼单权限。</span>
        </div>

        <div class="btn-row">
          <el-button
            type="primary"
            size="large"
            class="submit-reg-btn"
            :loading="loading"
            @click="handleRegister"
          >
            立即注册并登录系统
          </el-button>
          
          <div class="login-link-box">
            已有账号？
            <router-link to="/login" class="to-login-link">返回账号登录</router-link>
          </div>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import request from '../api/request'
import { ElMessage } from 'element-plus'
import {
  Cpu,
  User,
  Postcard,
  Lock,
  Message,
  Phone,
  InfoFilled
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const regFormRef = ref(null)
const loading = ref(false)
const groups = ref([])

const regForm = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  real_name: '',
  role: 'student',
  group_id: null,
  email: '',
  phone: ''
})

const validateConfirmPwd = (rule, value, callback) => {
  if (value !== regForm.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const regRules = {
  username: [
    { required: true, message: '请输入账号', trigger: 'blur' },
    { min: 3, message: '账号长度不能少于3个字符', trigger: 'blur' }
  ],
  real_name: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 3, message: '密码不能少于3位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPwd, trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择科研身份', trigger: 'change' }
  ],
  group_id: [
    { required: true, message: '请选择所属课题组', trigger: 'change' }
  ]
}

const fetchGroups = async () => {
  try {
    const res = await request.get('/groups')
    if (res.data) {
      groups.value = res.data
      if (groups.value.length > 0 && !regForm.group_id) {
        regForm.group_id = groups.value[0].id
      }
    }
  } catch (e) {
    console.error('获取课题组列表失败:', e)
  }
}

const handleRegister = async () => {
  if (!regFormRef.value) return
  await regFormRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const data = await authStore.register({
        username: regForm.username,
        password: regForm.password,
        real_name: regForm.real_name,
        role: regForm.role,
        group_id: regForm.group_id,
        email: regForm.email,
        phone: regForm.phone
      })
      ElMessage.success(`恭喜 ${data.user.real_name}，科研账户注册成功！已自动为您登录。`)
      router.push('/dashboard')
    } catch (err) {
      console.error(err)
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  fetchGroups()
})
</script>

<style scoped>
.register-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at 10% 20%, #0d1b2a 0%, #1b263b 40%, #0a1128 100%);
  position: relative;
  padding: 32px 16px;
}

.register-bg-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: 
    radial-gradient(rgba(24, 144, 255, 0.12) 1px, transparent 1px),
    radial-gradient(rgba(0, 196, 140, 0.08) 1px, transparent 1px);
  background-size: 32px 32px;
  background-position: 0 0, 16px 16px;
  pointer-events: none;
}

.register-card {
  position: relative;
  z-index: 10;
  width: 740px;
  max-width: 100%;
  background: #ffffff;
  border-radius: 16px;
  padding: 36px 44px;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.4);
}

.register-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f1f5f9;
}

.logo-box {
  width: 46px;
  height: 46px;
  border-radius: 12px;
  background: linear-gradient(135deg, #1890ff, #0050b3);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  box-shadow: 0 8px 16px rgba(24, 144, 255, 0.25);
}

.header-text h2 {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
}

.header-text p {
  font-size: 13px;
  color: #64748b;
  margin-top: 4px;
}

:deep(.el-form-item) {
  margin-bottom: 18px;
}

:deep(.el-form-item__label) {
  font-weight: 600;
  font-size: 13px;
  color: #334155;
  padding-bottom: 4px;
}

.form-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 12px;
  margin: 12px 0 20px 0;
}

.btn-row {
  display: flex;
  flex-direction: column;
  gap: 14px;
  align-items: center;
}

.submit-reg-btn {
  width: 100%;
  font-weight: 600;
  letter-spacing: 2px;
  border-radius: 8px;
}

.login-link-box {
  font-size: 13px;
  color: #64748b;
}

.to-login-link {
  color: #1890ff;
  font-weight: 600;
  text-decoration: none;
}

.to-login-link:hover {
  text-decoration: underline;
}

@media (max-width: 640px) {
  .register-card {
    padding: 24px;
  }
}
</style>
