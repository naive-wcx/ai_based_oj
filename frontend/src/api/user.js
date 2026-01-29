import request from './request'

export const userApi = {
  // 登录
  login(data) {
    return request.post('/user/login', data)
  },

  // 获取个人信息
  getProfile() {
    return request.get('/user/profile')
  },

  // 更新个人信息
  updateProfile(data) {
    return request.put('/user/profile', data)
  },
}
