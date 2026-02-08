<template>
  <header class="navbar">
    <div class="navbar-container">
      <div class="navbar-brand">
        <router-link to="/" class="logo">
          <img class="logo-image" :src="logoUrl" alt="USTC OJ" />
          <span class="logo-text">USTC OJ</span>
        </router-link>
      </div>

      <nav class="navbar-menu">
        <router-link to="/problems" class="nav-link">题目</router-link>
        <router-link to="/submissions" class="nav-link">提交</router-link>
        <router-link to="/contests" class="nav-link">比赛</router-link>
        <router-link to="/rank" class="nav-link">排行榜</router-link>
      </nav>

      <div class="navbar-actions">
        <template v-if="userStore.isLoggedIn">
          <el-dropdown trigger="click" popper-class="minimal-dropdown">
            <span class="user-dropdown">
              <span class="username">{{ userStore.username }}</span>
              <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="$router.push('/profile')">个人中心</el-dropdown-item>
                <el-dropdown-item v-if="userStore.isAdmin" @click="$router.push('/admin')">管理后台</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <router-link to="/login" class="login-link">登录</router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'
import { message } from '@/utils/message'
import { ArrowDown } from '@element-plus/icons-vue'
import logoUrl from '@/assets/logo.png'

const userStore = useUserStore()
const router = useRouter()

function handleLogout() {
  userStore.logout()
  message.success('已退出登录')
  router.push('/')
}
</script>

<style lang="scss" scoped>
.navbar {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 100;
  transition: all 0.3s ease;
}

.navbar-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.navbar-brand {
  .logo {
    display: flex;
    align-items: center;
    gap: 12px;
    text-decoration: none;
    
    .logo-image {
      width: 32px;
      height: 32px;
      object-fit: contain;
    }

    .logo-text {
      font-family: var(--font-family-base);
      font-size: 18px;
      font-weight: 700;
      color: var(--color-text-primary);
      letter-spacing: -0.01em;
    }
  }
}

.navbar-menu {
  display: flex;
  gap: 40px;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  
  .nav-link {
    color: var(--color-text-secondary);
    font-size: 15px;
    font-weight: 500;
    text-decoration: none;
    transition: color 0.2s;
    letter-spacing: 0.02em;
    
    &:hover {
      color: var(--color-primary);
    }
    
    &.router-link-active {
      color: var(--color-text-primary);
      font-weight: 600;
    }
  }
}

.navbar-actions {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: var(--radius-md);
  transition: background 0.2s;

  &:hover {
    background: rgba(0, 0, 0, 0.03);
  }
  
  .username {
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-primary);
  }
  
  .dropdown-icon {
    font-size: 12px;
    color: var(--color-text-secondary);
  }
}

.login-link {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-primary);
  text-decoration: none;
  
  &:hover {
    text-decoration: underline;
  }
}

/* Responsive adjustment */
@media (max-width: 768px) {
  .navbar-menu {
    display: none;
  }
}
</style>
