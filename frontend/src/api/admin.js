import request from './request'

export const adminApi = {
  // 获取用户列表
  getUserList(params) {
    return request.get('/admin/users', { params })
  },

  // 创建用户（管理员分配账号）
  createUser(data) {
    return request.post('/admin/users', data)
  },

  // 批量创建用户
  createUsersBatch(data) {
    return request.post('/admin/users/batch', data)
  },

  // 更新用户信息
  updateUser(id, data) {
    return request.put(`/admin/users/${id}`, data)
  },

  // 设置用户角色
  setUserRole(id, role) {
    return request.put(`/admin/users/${id}/role`, { role })
  },

  // 创建比赛
  createContest(data) {
    return request.post('/admin/contests', data)
  },

  // 更新比赛
  updateContest(id, data) {
    return request.put(`/admin/contests/${id}`, data)
  },

  // 删除比赛
  deleteContest(id) {
    return request.delete(`/admin/contests/${id}`)
  },

  // 获取比赛排行榜
  getContestLeaderboard(id, params = {}) {
    return request.get(`/admin/contests/${id}/leaderboard`, { params })
  },

  // 导出比赛成绩
  exportContestLeaderboard(id, params = {}) {
    return request.get(`/admin/contests/${id}/export`, { params, responseType: 'blob' })
  },

  // 重置用户窗口期比赛开赛状态
  resetContestUserStart(contestId, userId) {
    return request.post(`/admin/contests/${contestId}/users/${userId}/reset-start`)
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
