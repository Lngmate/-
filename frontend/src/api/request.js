import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const res = response.data
    // 如果返回的 code 不是 200，说明业务异常
    if (res.code && res.code !== 200) {
      ElMessage.error(res.message || '系统请求发生异常')
      return Promise.reject(new Error(res.message || 'Error'))
    }
    return res
  },
  (error) => {
    if (error.response?.status === 401) {
      const currentPath = window.location.pathname
      if (currentPath !== '/login' && currentPath !== '/register') {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        ElMessage.warning('登录认证已失效，请重新登录')
        setTimeout(() => {
          window.location.href = '/login'
        }, 800)
      }
    } else {
      const msg = error.response?.data?.message || error.message || '网络通信异常，请检查后端服务'
      ElMessage.error(msg)
    }
    return Promise.reject(error)
  }
)

export default request
