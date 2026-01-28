import request from './request'

export const rankApi = {
  // 获取排行榜
  getList(params) {
    return request.get('/rank', { params })
  },
}
