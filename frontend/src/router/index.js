import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { title: '用户登录 · 高校实验室设备共享系统', public: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { title: '账户注册 · 高校实验室设备共享系统', public: true }
  },
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { title: '决策与运行大屏' }
  },
  {
    path: '/equipments',
    name: 'EquipmentList',
    component: () => import('../views/EquipmentList.vue'),
    meta: { title: '设备大厅 & 2D地图' }
  },
  {
    path: '/calendar',
    name: 'CalendarSchedule',
    component: () => import('../views/CalendarSchedule.vue'),
    meta: { title: '可视化日历排期' }
  },
  {
    path: '/my-reservations',
    name: 'MyReservations',
    component: () => import('../views/MyReservations.vue'),
    meta: { title: '我的预约与现场核销' }
  },
  {
    path: '/approvals',
    name: 'ApprovalWorkflow',
    component: () => import('../views/ApprovalWorkflow.vue'),
    meta: { title: '多级审批工作台' }
  },
  {
    path: '/groups',
    name: 'GroupSettlement',
    component: () => import('../views/GroupSettlement.vue'),
    meta: { title: '课题组机时与结算' }
  },
  {
    path: '/credits',
    name: 'CreditCenter',
    component: () => import('../views/CreditCenter.vue'),
    meta: { title: '科研履约信用档案' }
  },
  {
    path: '/manage-equipments',
    name: 'EquipmentManage',
    component: () => import('../views/EquipmentManage.vue'),
    meta: { title: '设备台账与健康度管理' }
  },
  {
    path: '/users',
    name: 'UserManage',
    component: () => import('../views/UserManage.vue'),
    meta: { title: '用户权限管理中心' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 全局路由导航守卫
router.beforeEach((to, from, next) => {
  if (to.meta.title) {
    document.title = `${to.meta.title} - LabShare`
  }

  const token = localStorage.getItem('token')

  // 1. 访问公开页面 (Login / Register)
  if (to.meta.public) {
    if (token) {
      next('/dashboard')
    } else {
      next()
    }
    return
  }

  // 2. 访问受保护页面：未登录则重定向到登录页
  if (!token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }

  next()
})

export default router
