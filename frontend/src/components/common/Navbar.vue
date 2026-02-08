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
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
  position: sticky;
  top: 0;
  z-index: 1000;
  height: 60px;
  display: flex;
  align-items: center;
}

.navbar-container {
  max-width: var(--container-width);
  width: 100%;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.navbar-brand {
  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
    text-decoration: none;
    
    .logo-image {
      width: 24px;
      height: 24px;
      object-fit: contain;
    }

    .logo-text {
      font-family: var(--font-sans);
      font-size: 16px;
      font-weight: 700;
      color: var(--swiss-text-main);
      letter-spacing: -0.02em;
    }
  }
}

.navbar-menu {
  display: flex;
  gap: 32px;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  
  .nav-link {
    color: var(--swiss-text-secondary);
    font-size: 14px;
    font-weight: 500;
    text-decoration: none;
    transition: color 0.2s;
    
    &:hover {
      color: var(--swiss-primary);
    }
    
    &.router-link-active {
      color: var(--swiss-text-main);
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
  gap: 6px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: var(--radius-xs);
  transition: background 0.2s;

  &:hover {
    background: rgba(0, 0, 0, 0.03);
  }
  
  .username {
    font-size: 13px;
    font-weight: 500;
    color: var(--swiss-text-main);
  }
  
  .dropdown-icon {
    font-size: 10px;
    color: var(--swiss-text-secondary);
  }
}

.login-link {
  font-size: 13px;
  font-weight: 600;
  color: var(--swiss-primary);
  text-decoration: none;
  
  &:hover {
    opacity: 0.8;
  }
}

@media (max-width: 768px) {
  .navbar-menu {
    display: none;
  }
}
</style>
