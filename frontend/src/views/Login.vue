<template>
  <div class="login-container">
    <div class="login-bg-overlay"></div>

    <div class="login-card-wrapper">
      <!-- 左侧科研系统品牌介绍卡片 -->
      <div class="brand-panel">
        <div class="brand-header">
          <div class="brand-logo">
            <el-icon><Cpu /></el-icon>
          </div>
          <div class="brand-text">
            <h2>高校大型仪器设备共享系统</h2>
            <p class="en-tag">LabShare Enterprise · Precision Instrument Sharing</p>
          </div>
        </div>

        <div class="feature-highlights">
          <div class="feature-item">
            <div class="f-icon green"><el-icon><Opportunity /></el-icon></div>
            <div class="f-content">
              <h4>绿色节能“伴热拼单”引擎</h4>
              <p>智能识别前后60分钟同温同工况实验，享受8.5折机时优惠</p>
            </div>
          </div>

          <div class="feature-item">
            <div class="f-icon blue"><el-icon><Location /></el-icon></div>
            <div class="f-content">
              <h4>2D 数字孪生实验室机位交互地图</h4>
              <p>全景空间动态色彩映射：空闲、运行倒计时、维保自适应锁定</p>
            </div>
          </div>

          <div class="feature-item">
            <div class="f-icon orange"><el-icon><Medal /></el-icon></div>
            <div class="f-content">
              <h4>科研履约弹性信誉积分模型</h4>
              <p>正向履约激励、高分免审直通、防占座超时动态核销释放</p>
            </div>
          </div>
        </div>

        <div class="system-footer-note">
          <span>教育部重点实验室 / 国家重点实验室联盟支撑平台</span>
        </div>
      </div>

      <!-- 右侧登录操作表单 -->
      <div class="form-panel">
        <div class="form-header">
          <h3>科研人员与管理员登录</h3>
          <p>请输入校园统一身份账号或使用下方预置演示角色体验</p>
        </div>

        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="学号 / 教工号 / 登录账号"
              size="large"
              prefix-icon="User"
              clearable
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="登录密码 (预置账号统一为 123)"
              size="large"
              prefix-icon="Lock"
              show-password
            />
          </el-form-item>

          <div class="form-options">
            <el-checkbox v-model="rememberMe">记住登录状态</el-checkbox>
            <router-link to="/register" class="register-link">
              注册新用户 / 绑定课题组 <el-icon><ArrowRight /></el-icon>
            </router-link>
          </div>

          <el-button
            type="primary"
            size="large"
            class="submit-btn"
            :loading="loading"
            @click="handleLogin"
          >
            登 录 系 统
          </el-button>
        </el-form>

        <div class="login-footer-hint">
          <el-alert
            title="测试提示：全套测试账号及初始密码请查阅根目录下【测试账号密码清单.txt】（初始密码统一为 123）"
            type="info"
            :closable="false"
            show-icon
            class="account-alert"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { ElMessage } from 'element-plus'
import {
  Cpu,
  User,
  Lock,
  Opportunity,
  Location,
  Medal,
  ArrowRight
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loginFormRef = ref(null)
const loading = ref(false)
const rememberMe = ref(true)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [{ required: true, message: '请输入登录账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入登录密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const data = await authStore.login(loginForm.username, loginForm.password)
      ElMessage.success(`欢迎回来，${data.user.real_name}！`)
      const redirect = route.query.redirect || '/dashboard'
      router.push(redirect)
    } catch (err) {
      console.error(err)
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  if (authStore.isLoggedIn) {
    router.push('/dashboard')
  }
})
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at 10% 20%, #0d1b2a 0%, #1b263b 40%, #0a1128 100%);
  position: relative;
  padding: 24px;
}

.login-bg-overlay {
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

.login-card-wrapper {
  position: relative;
  z-index: 10;
  display: flex;
  width: 980px;
  max-width: 100%;
  min-height: 600px;
  background: #ffffff;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.4);
}

/* 左侧品牌面板 */
.brand-panel {
  flex: 1.1;
  background: linear-gradient(145deg, #0b1d3a, #112d5e);
  padding: 44px;
  color: #ffffff;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  position: relative;
}

.brand-panel::after {
  content: '';
  position: absolute;
  top: -80px;
  right: -80px;
  width: 260px;
  height: 260px;
  background: radial-gradient(circle, rgba(24, 144, 255, 0.25), transparent 70%);
  border-radius: 50%;
}

.brand-header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.brand-logo {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #1890ff, #0050b3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  box-shadow: 0 8px 16px rgba(24, 144, 255, 0.3);
}

.brand-text h2 {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.en-tag {
  font-size: 11px;
  color: #93c5fd;
  margin-top: 4px;
}

.feature-highlights {
  display: flex;
  flex-direction: column;
  gap: 24px;
  margin: 40px 0;
}

.feature-item {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.f-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.f-icon.green {
  background: rgba(16, 185, 129, 0.18);
  color: #10b981;
}

.f-icon.blue {
  background: rgba(59, 130, 246, 0.18);
  color: #3b82f6;
}

.f-icon.orange {
  background: rgba(245, 158, 11, 0.18);
  color: #f59e0b;
}

.f-content h4 {
  font-size: 14px;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 4px;
}

.f-content p {
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.5;
}

.system-footer-note {
  font-size: 11px;
  color: #64748b;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 16px;
}

/* 右侧表单面板 */
.form-panel {
  flex: 1.2;
  padding: 44px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  background: #ffffff;
}

.form-header h3 {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
}

.form-header p {
  font-size: 13px;
  color: #64748b;
  margin-top: 6px;
  margin-bottom: 24px;
}

.login-form {
  margin-bottom: 12px;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.register-link {
  font-size: 13px;
  color: #1890ff;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
}

.register-link:hover {
  text-decoration: underline;
}

.submit-btn {
  width: 100%;
  font-weight: 600;
  letter-spacing: 2px;
  border-radius: 8px;
}

/* 快速角色选择 */
.demo-accounts-section {
  margin-top: 18px;
}

.divider {
  position: relative;
  text-align: center;
  margin-bottom: 14px;
}

.divider::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  width: 100%;
  height: 1px;
  background: #e2e8f0;
}

.divider span {
  position: relative;
  background: #ffffff;
  padding: 0 12px;
  font-size: 11px;
  color: #94a3b8;
  font-weight: 500;
}

.quick-roles-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.role-pill {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  cursor: pointer;
  transition: all 0.2s ease;
}

.role-pill:hover {
  border-color: #1890ff;
  background: #e6f7ff;
  transform: translateY(-1px);
}

.pill-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2px;
}

.role-badge {
  font-size: 11px;
  font-weight: 600;
  color: #1890ff;
}

.real-name {
  font-size: 12px;
  font-weight: 600;
  color: #1e293b;
}

.pill-desc {
  font-size: 10px;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 860px) {
  .brand-panel {
    display: none;
  }
  .form-panel {
    padding: 30px;
  }
}
</style>
