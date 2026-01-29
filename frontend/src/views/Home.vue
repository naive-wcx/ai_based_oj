<template>
  <div class="home">
    <section class="hero">
      <h1>USTC OJ 在线评测系统</h1>
      <p class="subtitle">支持 AI 智能判题的现代化编程练习平台</p>
      <div class="hero-actions">
        <el-button type="primary" size="large" @click="$router.push('/problems')">
          开始做题
        </el-button>
        <el-button size="large" @click="$router.push('/rank')">
          查看排行
        </el-button>
      </div>
    </section>

    <section class="features">
      <div class="feature-card">
        <div class="feature-icon">📝</div>
        <h3>丰富题库</h3>
        <p>涵盖多种算法和数据结构的编程题目</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">🤖</div>
        <h3>AI 智能判题</h3>
        <p>利用大模型分析代码，检测算法和语言规范</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">⚡</div>
        <h3>即时反馈</h3>
        <p>提交代码后快速获得评测结果</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">🏆</div>
        <h3>排行榜</h3>
        <p>与其他用户竞争，展示你的编程实力</p>
      </div>
    </section>

    <section class="stats" v-if="stats">
      <div class="stat-item">
        <div class="stat-value">{{ stats.problems || 0 }}</div>
        <div class="stat-label">题目数量</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{{ stats.users || 0 }}</div>
        <div class="stat-label">用户数</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{{ stats.submissions || 0 }}</div>
        <div class="stat-label">提交次数</div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { statisticsApi } from '@/api/statistics'

const stats = ref({
  problems: 0,
  users: 0,
  submissions: 0,
})

async function fetchStats() {
  try {
    const res = await statisticsApi.getPublic()
    stats.value = res.data || stats.value
  } catch (e) {
    console.error(e)
  }
}

onMounted(() => {
  fetchStats()
})
</script>

<style lang="scss" scoped>
.home {
  max-width: 1200px;
  margin: 0 auto;
}

.hero {
  text-align: center;
  padding: 80px 20px;
  
  h1 {
    font-size: 48px;
    font-weight: 700;
    color: #303133;
    margin-bottom: 16px;
  }
  
  .subtitle {
    font-size: 20px;
    color: #606266;
    margin-bottom: 32px;
  }
  
  .hero-actions {
    display: flex;
    justify-content: center;
    gap: 16px;
  }
}

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 24px;
  padding: 40px 0;
}

.feature-card {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  text-align: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s, box-shadow 0.2s;
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  }
  
  .feature-icon {
    font-size: 48px;
    margin-bottom: 16px;
  }
  
  h3 {
    font-size: 20px;
    color: #303133;
    margin-bottom: 8px;
  }
  
  p {
    font-size: 14px;
    color: #909399;
  }
}

.stats {
  display: flex;
  justify-content: center;
  gap: 80px;
  padding: 60px 0;
  
  .stat-item {
    text-align: center;
    
    .stat-value {
      font-size: 48px;
      font-weight: 700;
      color: #409eff;
    }
    
    .stat-label {
      font-size: 16px;
      color: #909399;
      margin-top: 8px;
    }
  }
}
</style>
