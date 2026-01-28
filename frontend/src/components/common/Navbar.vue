<template>
  <header class="navbar">
    <div class="navbar-container">
      <div class="navbar-brand">
        <router-link to="/" class="logo">
          <span class="logo-icon">⚡</span>
          <span class="logo-text">OJ</span>
        </router-link>
      </div>

      <nav class="navbar-menu">
        <router-link to="/problems" class="nav-link">题目</router-link>
        <router-link to="/submissions" class="nav-link">提交</router-link>
        <router-link to="/rank" class="nav-link">排行榜</router-link>
      </nav>

      <div class="navbar-actions">
        <template v-if="userStore.isLoggedIn">
          <el-dropdown trigger="click">
            <span class="user-dropdown">
              <el-avatar :size="32">{{ userStore.username[0]?.toUpperCase() }}</el-avatar>
              <span class="username">{{ userStore.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="$router.push('/profile')">
                  <el-icon><User /></el-icon> 个人中心
                </el-dropdown-item>
                <el-dropdown-item v-if="userStore.isAdmin" @click="$router.push('/admin')">
                  <el-icon><Setting /></el-icon> 管理后台
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon> 退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <el-button type="primary" plain @click="$router.push('/login')">登录</el-button>
          <el-button type="primary" @click="$router.push('/register')">注册</el-button>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowDown, User, Setting, SwitchButton } from '@element-plus/icons-vue'

const userStore = useUserStore()
const router = useRouter()

function handleLogout() {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}
</script>

<style lang="scss" scoped>
.navbar {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.navbar-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.navbar-brand {
  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 24px;
    font-weight: 700;
    color: #409eff;
    text-decoration: none;
    
    .logo-icon {
      font-size: 28px;
    }
  }
}

.navbar-menu {
  display: flex;
  gap: 32px;
  
  .nav-link {
    color: #606266;
    font-size: 15px;
    font-weight: 500;
    text-decoration: none;
    transition: color 0.2s;
    
    &:hover,
    &.router-link-active {
      color: #409eff;
    }
  }
}

.navbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  
  .username {
    font-size: 14px;
    color: #606266;
  }
}
</style>
