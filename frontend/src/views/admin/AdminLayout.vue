<template>
  <div class="admin-layout">
    <aside class="admin-sidebar">
      <div class="sidebar-header">
        <h3>管理后台</h3>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/admin/problems">
          <el-icon><Document /></el-icon>
          <span>题目管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/contests">
          <el-icon><Trophy /></el-icon>
          <span>比赛管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </aside>
    
    <main class="admin-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Document, User, Setting, Trophy } from '@element-plus/icons-vue'

const route = useRoute()

const activeMenu = computed(() => route.path)
</script>

<style lang="scss" scoped>
.admin-layout {
  display: flex;
  min-height: calc(100vh - 60px);
  margin: -20px;
}

.admin-sidebar {
  width: 220px;
  background: #fff;
  border-right: 1px solid #e6e6e6;
  
  .sidebar-header {
    padding: 20px;
    border-bottom: 1px solid #e6e6e6;
    
    h3 {
      margin: 0;
      font-size: 18px;
      color: #303133;
    }
  }
  
  .sidebar-menu {
    border-right: none;
  }
}

.admin-main {
  flex: 1;
  padding: 20px;
  background: #f5f7fa;
  overflow-y: auto;
}
</style>
