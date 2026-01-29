import request from './request'

export const statisticsApi = {
  getPublic() {
    return request.get('/statistics')
  },
}
