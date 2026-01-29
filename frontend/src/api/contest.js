import request from './request'

export const contestApi = {
  getList(params) {
    return request.get('/contest/list', { params })
  },

  getById(id) {
    return request.get(`/contest/${id}`)
  },
}
