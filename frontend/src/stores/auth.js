import { defineStore } from 'pinia'
import request from '../api/request'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null')
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    role: (state) => state.user?.role || '',
    creditScore: (state) => state.user?.credit_score ?? 100,
    isStudent: (state) => state.user?.role === 'student',
    isAdvisor: (state) => state.user?.role === 'advisor',
    isEquipmentAdmin: (state) => state.user?.role === 'equipment_admin',
    isSystemAdmin: (state) => state.user?.role === 'system_admin',
    isAdmin: (state) => state.user?.role === 'equipment_admin' || state.user?.role === 'system_admin'
  },
  actions: {
    setUser(user, token) {
      this.user = user
      if (token) {
        this.token = token
        localStorage.setItem('token', token)
      }
      localStorage.setItem('user', JSON.stringify(user))
    },

    async login(username, password) {
      const res = await request.post('/auth/login', { username, password })
      if (res.data) {
        this.setUser(res.data.user, res.data.token)
        return res.data
      }
    },

    async register(payload) {
      const res = await request.post('/auth/register', payload)
      if (res.data) {
        this.setUser(res.data.user, res.data.token)
        return res.data
      }
    },

    async fetchProfile() {
      try {
        const res = await request.get('/auth/profile')
        if (res.data) {
          this.user = res.data
          localStorage.setItem('user', JSON.stringify(res.data))
          return res.data
        }
      } catch (e) {
        console.error('Fetch profile failed:', e)
      }
    },

    async updateProfile(payload) {
      const res = await request.put('/auth/profile', payload)
      if (res.data) {
        this.user = res.data
        localStorage.setItem('user', JSON.stringify(res.data))
        return res.data
      }
    },

    async changePassword(oldPassword, newPassword) {
      return await request.put('/auth/password', {
        old_password: oldPassword,
        new_password: newPassword
      })
    },

    async quickSwitchRole(targetRole) {
      try {
        const res = await request.post('/auth/quick-switch', { role: targetRole })
        if (res.data) {
          this.setUser(res.data.user, res.data.token)
          return res.data.user
        }
      } catch (err) {
        // 本地降级容错
        const mockUsers = {
          student: { id: 1, username: 'student', real_name: '张耀辉(学生)', role: 'student', group_id: 1, group_name: '先进纳米材料课题组', credit_score: 98 },
          advisor: { id: 4, username: 'advisor', real_name: '李教授(导师/组长)', role: 'advisor', group_id: 1, group_name: '先进纳米材料课题组', credit_score: 100 },
          equipment_admin: { id: 6, username: 'admin', real_name: '王佳豪(设备管理员)', role: 'equipment_admin', credit_score: 100 },
          system_admin: { id: 7, username: 'sysadmin', real_name: '李俊达(系统管理员)', role: 'system_admin', credit_score: 100 }
        }
        if (mockUsers[targetRole]) {
          this.setUser(mockUsers[targetRole], 'mock-token-' + targetRole)
          return mockUsers[targetRole]
        }
      }
    },

    async logout() {
      try {
        await request.post('/auth/logout')
      } catch (e) {
        // ignore error during logout
      } finally {
        this.token = ''
        this.user = null
        localStorage.removeItem('token')
        localStorage.removeItem('user')
      }
    }
  }
})
