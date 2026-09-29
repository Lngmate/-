<template>
  <!-- 登录与注册页面使用沉浸式无边框视窗 -->
  <div v-if="isAuthPage" class="auth-wrapper">
    <router-view />
  </div>

  <!-- 主系统工作台布局 -->
  <div v-else class="app-layout">
    <!-- 顶部现代化导航栏 -->
    <header class="app-header">
      <div class="header-left">
        <div class="logo-icon">
          <el-icon><Cpu /></el-icon>
        </div>
        <div class="system-title">
          <span class="main-title">高校实验室大型精密仪器预约与共享平台</span>
          <span class="sub-title">LabShare Enterprise · 智慧科研资源池</span>
        </div>
      </div>

      <div class="header-right">
        <!-- 核心新功能：站内实时消息与醒目红点提示 (QQ/微信风格提醒) -->
        <el-popover
          placement="bottom-end"
          :width="400"
          trigger="click"
          popper-class="notif-popover"
        >
          <template #reference>
            <div class="header-notif-trigger">
              <el-badge
                :value="unreadNotifCount"
                :hidden="unreadNotifCount === 0"
                :max="99"
                class="header-notif-badge"
              >
                <div class="bell-btn-box" :class="{ 'has-unread': unreadNotifCount > 0 }">
                  <el-icon :size="20"><Bell /></el-icon>
                  <span v-if="unreadNotifCount > 0" class="pulse-ring"></span>
                </div>
              </el-badge>
            </div>
          </template>

          <div class="notif-panel">
            <div class="notif-header">
              <div class="nh-title">
                <span class="nh-text">站内待办与系统提醒</span>
                <el-tag v-if="unreadNotifCount > 0" type="danger" size="small" effect="dark" round>
                  {{ unreadNotifCount }} 条未读
                </el-tag>
                <el-tag v-else type="info" size="small" effect="plain" round>
                  全部已读
                </el-tag>
              </div>
              <el-button
                v-if="unreadNotifCount > 0"
                link
                type="primary"
                size="small"
                @click="handleReadAllNotifs"
              >
                全部标为已读
              </el-button>
            </div>

            <el-scrollbar max-height="380px" class="notif-scrollbar">
              <div v-if="notifList.length === 0" class="notif-empty">
                <el-empty description="暂无新消息通知" :image-size="70" />
              </div>
              <div v-else class="notif-items">
                <div
                  v-for="item in notifList"
                  :key="item.id"
                  class="notif-item"
                  :class="{ 'is-unread': !item.is_read }"
                  @click="handleNotifClick(item)"
                >
                  <div class="notif-dot-col">
                    <span
                      class="status-indicator"
                      :class="{
                        'dot-pending': item.type === 'reservation_pending',
                        'dot-approved': item.type === 'reservation_approved',
                        'dot-rejected': item.type === 'reservation_rejected',
                        'dot-system': item.type === 'system_notice' || item.type === 'reservation_cancel'
                      }"
                    ></span>
                  </div>
                  <div class="notif-content-col">
                    <div class="ni-header">
                      <span class="ni-title">{{ item.title }}</span>
                      <span class="ni-time">{{ formatTime(item.created_at) }}</span>
                    </div>
                    <div class="ni-desc">{{ item.content }}</div>
                    <div class="ni-action-bar">
                      <span v-if="!item.is_read" class="unread-tag">NEW</span>
                      <span class="jump-hint">
                        {{ item.type === 'reservation_pending' ? '去处理审批 ➔' : '查看工单详情 ➔' }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </el-scrollbar>
          </div>
        </el-popover>

        <!-- 信用积分徽章 -->
        <el-tag
          size="default"
          :type="authStore.creditScore >= 90 ? 'success' : (authStore.creditScore >= 80 ? 'primary' : 'warning')"
          effect="plain"
          class="credit-tag"
        >
          信誉分: {{ authStore.creditScore }}
        </el-tag>

        <!-- 用户中心下拉菜单 -->
        <el-dropdown trigger="click" @command="handleUserCommand">
          <div class="user-profile-btn">
            <el-avatar size="small" :style="{ backgroundColor: getRoleAvatarColor(authStore.role) }">
              {{ (authStore.user?.real_name || '研')[0] }}
            </el-avatar>
            <div class="user-info-text">
              <span class="uname">{{ authStore.user?.real_name || '未登录' }}</span>
              <span class="urole">{{ getRoleLabel(authStore.role) }}</span>
            </div>
            <el-icon class="arrow-icon"><CaretBottom /></el-icon>
          </div>

          <template #dropdown>
            <el-dropdown-menu class="user-dropdown-menu">
              <div class="dropdown-header">
                <div class="dh-name">{{ authStore.user?.real_name }}</div>
                <div class="dh-group">{{ authStore.user?.group_name || '全校公共实验室' }}</div>
              </div>
              <el-dropdown-item divided command="profile">
                <el-icon><User /></el-icon> 个人资料设置
              </el-dropdown-item>
              <el-dropdown-item command="password">
                <el-icon><Lock /></el-icon> 修改登录密码
              </el-dropdown-item>
              <el-dropdown-item v-if="authStore.isSystemAdmin" command="users">
                <el-icon><UserFilled /></el-icon> 用户与权限管理
              </el-dropdown-item>
              <el-dropdown-item divided command="logout" class="logout-item">
                <el-icon><SwitchButton /></el-icon> 退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <div class="main-container">
      <!-- 左侧边栏导航 -->
      <aside class="app-sidebar">
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          background-color="#ffffff"
          text-color="#334155"
          active-text-color="#1890ff"
        >
          <el-menu-item index="/dashboard">
            <el-icon><DataBoard /></el-icon>
            <span>决策与运行大屏</span>
          </el-menu-item>

          <el-menu-item index="/equipments">
            <el-icon><Location /></el-icon>
            <span>设备大厅 & 2D地图</span>
          </el-menu-item>

          <el-menu-item index="/calendar">
            <el-icon><Calendar /></el-icon>
            <span>可视化日历排期</span>
          </el-menu-item>

          <el-menu-item index="/my-reservations">
            <el-icon><Tickets /></el-icon>
            <div class="menu-item-content">
              <span>我的预约与现场核销</span>
              <el-badge
                v-if="studentUnreadCount > 0"
                :value="studentUnreadCount"
                class="menu-badge badge-green"
              />
            </div>
          </el-menu-item>

          <el-menu-item v-if="authStore.isAdvisor || authStore.isEquipmentAdmin || authStore.isSystemAdmin" index="/approvals">
            <el-icon><Stamp /></el-icon>
            <div class="menu-item-content">
              <span>多级审批工作台</span>
              <el-badge
                v-if="approvalUnreadCount > 0"
                :value="approvalUnreadCount"
                class="menu-badge badge-red"
              />
            </div>
          </el-menu-item>

          <el-menu-item index="/groups">
            <el-icon><Money /></el-icon>
            <span>课题组机时与结算</span>
          </el-menu-item>

          <el-menu-item index="/credits">
            <el-icon><Medal /></el-icon>
            <span>科研履约信用档案</span>
          </el-menu-item>

          <el-menu-item v-if="authStore.isEquipmentAdmin || authStore.isSystemAdmin" index="/manage-equipments">
            <el-icon><Setting /></el-icon>
            <span>设备台账与维保管理</span>
          </el-menu-item>

          <el-menu-item v-if="authStore.isSystemAdmin" index="/users">
            <el-icon><UserFilled /></el-icon>
            <span>用户与权限管理</span>
          </el-menu-item>
        </el-menu>
      </aside>

      <!-- 右侧主视窗内容 -->
      <main class="app-main">
        <router-view />
      </main>
    </div>

    <!-- 个人资料编辑弹窗 -->
    <el-dialog v-model="profileDialogVisible" title="个人资料设置" width="460px" destroy-on-close>
      <el-form :model="profileForm" label-width="90px">
        <el-form-item label="登录账号">
          <el-input :model-value="authStore.user?.username" disabled />
        </el-form-item>
        <el-form-item label="真实姓名">
          <el-input v-model="profileForm.real_name" placeholder="真实姓名" />
        </el-form-item>
        <el-form-item label="归属课题组">
          <el-input :model-value="authStore.user?.group_name || '全校实验室公共中心'" disabled />
        </el-form-item>
        <el-form-item label="联系邮箱">
          <el-input v-model="profileForm.email" placeholder="输入邮箱" />
        </el-form-item>
        <el-form-item label="手机号码">
          <el-input v-model="profileForm.phone" placeholder="输入手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="profileDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="updatingProfile" @click="handleSaveProfile">保存修改</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码弹窗 -->
    <el-dialog v-model="pwdDialogVisible" title="修改登录密码" width="440px" destroy-on-close>
      <el-form :model="pwdForm" label-width="90px">
        <el-form-item label="原密码" required>
          <el-input v-model="pwdForm.old_password" type="password" placeholder="请输入原密码" show-password />
        </el-form-item>
        <el-form-item label="新密码" required>
          <el-input v-model="pwdForm.new_password" type="password" placeholder="请输入新密码" show-password />
        </el-form-item>
        <el-form-item label="确认新密码" required>
          <el-input v-model="pwdForm.confirm_password" type="password" placeholder="请再次输入新密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="changingPwd" @click="handleSavePassword">确认修改</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import {
  Cpu,
  User,
  UserFilled,
  Lock,
  DataBoard,
  Location,
  Calendar,
  Tickets,
  Stamp,
  Money,
  Medal,
  Setting,
  CaretBottom,
  SwitchButton,
  Bell
} from '@element-plus/icons-vue'
import { getNotifications, markAsRead, markAllAsRead } from './api/notification'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isAuthPage = computed(() => {
  return route.path === '/login' || route.path === '/register'
})

const getRoleLabel = (role) => {
  const map = {
    student: '普通学生',
    advisor: '导师/组长',
    equipment_admin: '设备管理员',
    system_admin: '系统管理员'
  }
  return map[role] || '科研人员'
}

const getRoleAvatarColor = (role) => {
  const map = {
    student: '#1890ff',
    advisor: '#722ed1',
    equipment_admin: '#fa8c16',
    system_admin: '#f5222d'
  }
  return map[role] || '#1890ff'
}

// ---------------- 站内消息与红点通知逻辑 ----------------
const notifList = ref([])
const unreadNotifCount = ref(0)
let lastUnreadCount = 0
let notifTimer = null

// 计算审批工作台专属未读数（待审批提醒）
const approvalUnreadCount = computed(() => {
  return notifList.value.filter(n => !n.is_read && n.type === 'reservation_pending').length
})

// 计算学生我的预约专属未读数（审批通过/驳回提醒）
const studentUnreadCount = computed(() => {
  return notifList.value.filter(n => !n.is_read && (n.type === 'reservation_approved' || n.type === 'reservation_rejected')).length
})

const fetchNotifications = async (triggerToast = false) => {
  if (!authStore.isLoggedIn || isAuthPage.value) return
  try {
    const res = await getNotifications()
    if (res.code === 200 && res.data) {
      notifList.value = res.data.list || []
      const currentUnread = res.data.unread_count || 0

      // 如果有新增未读消息且要求弹窗提示（类似QQ/微信新消息弹窗）
      if (triggerToast && currentUnread > lastUnreadCount) {
        const latest = notifList.value.find(n => !n.is_read)
        if (latest) {
          ElNotification({
            title: latest.title,
            message: latest.content,
            type: latest.type === 'reservation_approved' ? 'success' : (latest.type === 'reservation_rejected' ? 'warning' : 'primary'),
            duration: 6000,
            position: 'top-right'
          })
        }
      }

      unreadNotifCount.value = currentUnread
      lastUnreadCount = currentUnread
    }
  } catch (err) {
    // 静默忽略轮询错误
  }
}

const handleNotifClick = async (item) => {
  if (!item.is_read) {
    try {
      await markAsRead(item.id)
      item.is_read = true
      if (unreadNotifCount.value > 0) {
        unreadNotifCount.value--
        lastUnreadCount = unreadNotifCount.value
      }
    } catch (e) {
      console.error(e)
    }
  }

  // 根据通知类型智能跳转到对应模块
  if (item.type === 'reservation_pending') {
    router.push('/approvals')
  } else if (item.type === 'reservation_approved' || item.type === 'reservation_rejected' || item.type === 'reservation_cancel') {
    router.push('/my-reservations')
  }
}

const handleReadAllNotifs = async () => {
  try {
    await markAllAsRead()
    notifList.value.forEach(item => item.is_read = true)
    unreadNotifCount.value = 0
    lastUnreadCount = 0
    ElMessage.success('全部消息已标为已读')
  } catch (err) {
    console.error(err)
  }
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  const now = new Date()
  const diffSec = Math.floor((now - d) / 1000)
  if (diffSec < 60) return '刚刚'
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `${diffMin}分钟前`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour}小时前`
  return `${d.getMonth() + 1}月${d.getDate()}日 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

// ---------------- 个人资料与修改密码 ----------------
const profileDialogVisible = ref(false)
const updatingProfile = ref(false)
const profileForm = reactive({
  real_name: '',
  email: '',
  phone: ''
})

const pwdDialogVisible = ref(false)
const changingPwd = ref(false)
const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const handleUserCommand = async (command) => {
  if (command === 'profile') {
    profileForm.real_name = authStore.user?.real_name || ''
    profileForm.email = authStore.user?.email || ''
    profileForm.phone = authStore.user?.phone || ''
    profileDialogVisible.value = true
  } else if (command === 'password') {
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm_password = ''
    pwdDialogVisible.value = true
  } else if (command === 'users') {
    router.push('/users')
  } else if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出当前系统登录吗？', '退出提示', {
        confirmButtonText: '确定退出',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await authStore.logout()
      ElMessage.success('已安全退出系统')
      router.push('/login')
    } catch (e) {
      // cancelled
    }
  }
}

const handleSaveProfile = async () => {
  updatingProfile.value = true
  try {
    await authStore.updateProfile(profileForm)
    ElMessage.success('个人资料保存成功')
    profileDialogVisible.value = false
  } catch (err) {
    console.error(err)
  } finally {
    updatingProfile.value = false
  }
}

const handleSavePassword = async () => {
  if (!pwdForm.old_password || !pwdForm.new_password) {
    ElMessage.warning('请输入原密码与新密码')
    return
  }
  if (pwdForm.new_password !== pwdForm.confirm_password) {
    ElMessage.error('两次输入的新密码不一致')
    return
  }
  changingPwd.value = true
  try {
    await authStore.changePassword(pwdForm.old_password, pwdForm.new_password)
    ElMessage.success('密码修改成功，请妥善保管')
    pwdDialogVisible.value = false
  } catch (err) {
    console.error(err)
  } finally {
    changingPwd.value = false
  }
}

// 路由变化时自动刷新通知状态
watch(() => route.path, () => {
  if (authStore.isLoggedIn && !isAuthPage.value) {
    fetchNotifications(false)
  }
})

onMounted(() => {
  if (authStore.isLoggedIn && !isAuthPage.value) {
    fetchNotifications(false)
  }
  // 每 5 秒轮询一次新通知，保持类似 QQ/微信的实时消息红点响应
  notifTimer = setInterval(() => {
    fetchNotifications(true)
  }, 5000)
})

onUnmounted(() => {
  if (notifTimer) {
    clearInterval(notifTimer)
    notifTimer = null
  }
})
</script>

<style>
/* 全局基础重置 */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f1f5f9;
  color: #1e293b;
}

.auth-wrapper {
  min-height: 100vh;
}

.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  height: 60px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #1890ff, #0050b3);
  color: #fff;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.system-title {
  display: flex;
  flex-direction: column;
}

.main-title {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.sub-title {
  font-size: 11px;
  color: #64748b;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

/* 消息通知铃铛与醒目红点 */
.header-notif-trigger {
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bell-btn-box {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  position: relative;
  transition: all 0.25s ease;
}

.bell-btn-box:hover {
  background: #eff6ff;
  color: #1890ff;
  border-color: #bfdbfe;
  transform: scale(1.05);
}

.bell-btn-box.has-unread {
  background: #fef2f2;
  color: #ef4444;
  border-color: #fecaca;
}

/* QQ/微信动态呼吸脉冲光晕效果 */
.pulse-ring {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: rgba(239, 68, 68, 0.4);
  animation: pulse-animation 2s infinite ease-out;
  pointer-events: none;
}

@keyframes pulse-animation {
  0% {
    transform: scale(0.8);
    opacity: 0.9;
  }
  70% {
    transform: scale(2.2);
    opacity: 0;
  }
  100% {
    transform: scale(2.5);
    opacity: 0;
  }
}

.header-notif-badge :deep(.el-badge__content) {
  font-weight: 700;
  background-color: #ef4444;
  box-shadow: 0 0 0 2px #fff;
}

/* 消息下拉面板 */
.notif-panel {
  padding: 4px 0;
}

.notif-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px 12px 16px;
  border-bottom: 1px solid #f1f5f9;
}

.nh-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nh-text {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.notif-empty {
  padding: 30px 0;
}

.notif-items {
  display: flex;
  flex-direction: column;
}

.notif-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid #f8fafc;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.notif-item:hover {
  background-color: #f8fafc;
}

.notif-item.is-unread {
  background-color: #f0fdf4;
}

.notif-dot-col {
  padding-top: 4px;
}

.status-indicator {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-pending {
  background-color: #eab308;
  box-shadow: 0 0 6px #facc15;
}

.dot-approved {
  background-color: #22c55e;
  box-shadow: 0 0 6px #4ade80;
}

.dot-rejected {
  background-color: #ef4444;
  box-shadow: 0 0 6px #f87171;
}

.dot-system {
  background-color: #3b82f6;
  box-shadow: 0 0 6px #60a5fa;
}

.notif-content-col {
  flex: 1;
  min-width: 0;
}

.ni-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 4px;
}

.ni-title {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 230px;
}

.ni-time {
  font-size: 11px;
  color: #94a3b8;
}

.ni-desc {
  font-size: 12px;
  color: #475569;
  line-height: 1.5;
  margin-bottom: 6px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.ni-action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.unread-tag {
  font-size: 10px;
  font-weight: 700;
  color: #ef4444;
  background: #fee2e2;
  padding: 1px 5px;
  border-radius: 4px;
}

.jump-hint {
  font-size: 11px;
  color: #1890ff;
  margin-left: auto;
}

/* 侧边栏红点与徽章 */
.menu-item-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.menu-badge :deep(.el-badge__content) {
  margin-right: 6px;
  font-weight: 700;
}

.badge-red :deep(.el-badge__content) {
  background-color: #ef4444;
}

.badge-green :deep(.el-badge__content) {
  background-color: #10b981;
}

.credit-tag {
  font-weight: 600;
}

.user-profile-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.user-profile-btn:hover {
  background-color: #f1f5f9;
}

.user-info-text {
  display: flex;
  flex-direction: column;
}

.uname {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.urole {
  font-size: 11px;
  color: #64748b;
}

.arrow-icon {
  font-size: 12px;
  color: #94a3b8;
}

.main-container {
  display: flex;
  flex: 1;
}

.app-sidebar {
  width: 220px;
  background: #ffffff;
  border-right: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.sidebar-menu {
  border-right: none;
  height: calc(100vh - 60px);
  position: sticky;
  top: 60px;
}

.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
  margin: 4px 8px;
  border-radius: 8px;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #e6f7ff !important;
  color: #1890ff !important;
  font-weight: 600;
}

.app-main {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.user-dropdown-menu .dropdown-header {
  padding: 10px 16px;
  background: #f8fafc;
  margin-bottom: 4px;
}

.dh-name {
  font-weight: 700;
  font-size: 14px;
  color: #0f172a;
}

.dh-group {
  font-size: 12px;
  color: #64748b;
  margin-top: 2px;
}

.logout-item {
  color: #f5222d !important;
}
</style>
