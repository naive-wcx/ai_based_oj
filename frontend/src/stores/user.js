import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { userApi } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => ['admin', 'super_admin'].includes(user.value?.role))
  const isSuperAdmin = computed(() => user.value?.role === 'super_admin')
  const username = computed(() => user.value?.username || '')

  function setToken(newToken) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function setUser(newUser) {
    user.value = newUser
    localStorage.setItem('user', JSON.stringify(newUser))
  }

  async function login(credentials) {
    const res = await userApi.login(credentials)
    setToken(res.data.token)
    setUser(res.data.user)
    return res
  }

  async function fetchProfile() {
    if (!token.value) return
    try {
      const res = await userApi.getProfile()
      setUser(res.data)
    } catch (e) {
      logout()
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    token,
    user,
    isLoggedIn,
    isAdmin,
    isSuperAdmin,
    username,
    login,
    fetchProfile,
    logout,
  }
})
