import request from './request'

export const problemApi = {
  // 获取题目列表
  getList(params) {
    return request.get('/problem/list', { params })
  },

  // 获取题目详情
  getById(id) {
    return request.get(`/problem/${id}`)
  },

  // 创建题目（管理员）
  create(data) {
    return request.post('/problem', data)
  },

  // 更新题目（管理员）
  update(id, data) {
    return request.put(`/problem/${id}`, data)
  },

  // 删除题目（管理员）
  delete(id) {
    return request.delete(`/problem/${id}`)
  },

  // 上传测试用例（管理员）
  uploadTestcase(id, formData) {
    return request.post(`/problem/${id}/testcase`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  // 获取测试用例列表（管理员）
  getTestcases(id) {
    return request.get(`/problem/${id}/testcases`)
  },

  // 删除所有测试用例（管理员）
  deleteTestcases(id) {
    return request.delete(`/problem/${id}/testcases`)
  },
}
