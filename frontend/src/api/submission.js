import request from './request'

export const submissionApi = {
  // 提交代码
  submit(data) {
    return request.post('/submission', data)
  },

  // 获取提交详情
  getById(id) {
    return request.get(`/submission/${id}`)
  },

  // 获取提交列表
  getList(params) {
    return request.get('/submission/list', { params })
  },

  // 获取我的提交
  getMySubmissions(params) {
    return request.get('/submission/my', { params })
  },
}
