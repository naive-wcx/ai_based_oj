import request from './request'

export const adminApi = {
  // 获取用户列表
  getUserList(params) {
    return request.get('/admin/users', { params })
  },

  // 设置用户角色
  setUserRole(id, role) {
    return request.put(`/admin/users/${id}/role`, { role })
  },

  // 获取 AI 设置
  getAISettings() {
    return request.get('/admin/settings/ai')
  },

  // 更新 AI 设置
  updateAISettings(data) {
    return request.put('/admin/settings/ai', data)
  },

  // 测试 AI 连接
  testAIConnection() {
    return request.post('/admin/settings/ai/test')
  },
}
