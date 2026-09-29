import request from './request'

// 获取当前用户的通知列表及未读数
export function getNotifications() {
  return request({
    url: '/notifications',
    method: 'get',
  })
}

// 快速获取未读消息数
export function getUnreadCount() {
  return request({
    url: '/notifications/unread-count',
    method: 'get',
  })
}

// 标记单条通知为已读
export function markAsRead(id) {
  return request({
    url: `/notifications/${id}/read`,
    method: 'put',
  })
}

// 标记所有通知为已读
export function markAllAsRead() {
  return request({
    url: '/notifications/read-all',
    method: 'put',
  })
}
