import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
  },
  {
    path: '/problems',
    name: 'Problems',
    component: () => import('@/views/problem/ProblemList.vue'),
  },
  {
    path: '/problem/:id',
    name: 'ProblemDetail',
    component: () => import('@/views/problem/ProblemDetail.vue'),
  },
  {
    path: '/submissions',
    name: 'Submissions',
    component: () => import('@/views/submission/SubmissionList.vue'),
  },
  {
    path: '/contests',
    name: 'Contests',
    component: () => import('@/views/contest/ContestList.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/contest/:id',
    name: 'ContestDetail',
    component: () => import('@/views/contest/ContestDetail.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/submission/:id',
    name: 'SubmissionDetail',
    component: () => import('@/views/submission/SubmissionDetail.vue'),
  },
  {
    path: '/rank',
    name: 'Rank',
    component: () => import('@/views/Rank.vue'),
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/user/Profile.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        redirect: '/admin/problems',
      },
      {
        path: 'problems',
        name: 'AdminProblems',
        component: () => import('@/views/admin/ProblemManage.vue'),
      },
      {
        path: 'contests',
        name: 'AdminContests',
        component: () => import('@/views/admin/ContestManage.vue'),
      },
      {
        path: 'contest/create',
        name: 'AdminContestCreate',
        component: () => import('@/views/admin/ContestEdit.vue'),
      },
      {
        path: 'contest/:id/edit',
        name: 'AdminContestEdit',
        component: () => import('@/views/admin/ContestEdit.vue'),
      },
      {
        path: 'problem/create',
        name: 'AdminProblemCreate',
        component: () => import('@/views/admin/ProblemEdit.vue'),
      },
      {
        path: 'problem/:id/edit',
        name: 'AdminProblemEdit',
        component: () => import('@/views/admin/ProblemEdit.vue'),
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/UserManage.vue'),
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/Settings.vue'),
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }
  
  if (to.meta.requiresAdmin && !userStore.isAdmin) {
    next({ name: 'Home' })
    return
  }
  
  next()
})

export default router
