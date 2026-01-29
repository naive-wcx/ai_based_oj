<template>
  <div class="home-container">
    <section class="hero">
      <el-row justify="center">
        <el-col :xl="10" :lg="12" :md="16">
          <h1 class="hero-title">USTC OJ 在线评测系统</h1>
          <p class="hero-subtitle">
            一个为现代编程学习者设计的、支持 AI 智能分析的在线评测平台
          </p>
          <div class="hero-actions">
            <el-button
              type="primary"
              size="large"
              round
              @click="$router.push('/problems')"
              >开始挑战</el-button
            >
            <el-button size="large" round @click="$router.push('/rank')"
              >查看排行</el-button
            >
          </div>
        </el-col>
      </el-row>
    </section>

    <el-main class="main-content">
      <el-row justify="center">
        <el-col :span="22">
          <section class="features">
            <el-row :gutter="30">
              <el-col
                :lg="6"
                :md="12"
                v-for="feature in features"
                :key="feature.title"
              >
                <el-card shadow="hover" class="feature-card">
                  <div class="feature-icon">
                    <el-icon :size="40">
                      <component :is="feature.icon"></component>
                    </el-icon>
                  </div>
                  <h3 class="feature-title">{{ feature.title }}</h3>
                  <p class="feature-description">{{ feature.description }}</p>
                </el-card>
              </el-col>
            </el-row>
          </section>

          <el-divider />

          <section class="stats" v-if="stats">
            <el-row :gutter="40" justify="center">
              <el-col :md="8" :sm="12" :xs="24" class="stat-item">
                <el-statistic :value="stats.problems || 0">
                  <template #title>
                    <div class="stat-title">
                      <el-icon><Collection /></el-icon>
                      <span>总题目数</span>
                    </div>
                  </template>
                </el-statistic>
              </el-col>
              <el-col :md="8" :sm="12" :xs="24" class="stat-item">
                <el-statistic :value="stats.users || 0">
                  <template #title>
                    <div class="stat-title">
                      <el-icon><User /></el-icon>
                      <span>注册用户</span>
                    </div>
                  </template>
                </el-statistic>
              </el-col>
              <el-col :md="8" :sm="12" :xs="24" class="stat-item">
                <el-statistic :value="stats.submissions || 0">
                  <template #title>
                    <div class="stat-title">
                      <el-icon><DataLine /></el-icon>
                      <span>累计提交</span>
                    </div>
                  </template>
                </el-statistic>
              </el-col>
            </el-row>
          </section>
        </el-col>
      </el-row>
    </el-main>
  </div>
</template>

<script setup>
import { ref, onMounted, shallowRef } from 'vue'
import { statisticsApi } from '@/api/statistics'
import {
  Collection,
  Cpu,
  DataLine,
  Trophy,
  User,
} from '@element-plus/icons-vue'

const features = shallowRef([
  {
    icon: Collection,
    title: '丰富题库',
    description: '涵盖多种算法和数据结构的编程题目',
  },
  {
    icon: Cpu,
    title: 'AI 智能分析',
    description: '利用大模型分析代码，提供超越传统 OJ 的代码洞察',
  },
  {
    icon: DataLine,
    title: '即时反馈',
    description: '提交代码后快速获得评测结果与性能分析',
  },
  {
    icon: Trophy,
    title: '实力排行榜',
    description: '与其他用户同台竞技，见证你的成长',
  },
])

const stats = ref({
  problems: 0,
  users: 0,
  submissions: 0,
})

async function fetchStats() {
  try {
    const res = await statisticsApi.getPublic()
    if (res && res.data) {
      stats.value = res.data
    }
  } catch (e) {
    console.error('Failed to fetch statistics:', e)
  }
}

onMounted(() => {
  fetchStats()
})
</script>

<style lang="scss" scoped>
.home-container {
  background-color: #f7f8fa;
}

.hero {
  text-align: center;
  padding: 100px 20px;
  background-color: #ffffff;
  border-bottom: 1px solid #e4e7ed;

  .hero-title {
    font-size: 42px;
    font-weight: 700;
    color: #303133;
    margin-bottom: 24px;
  }

  .hero-subtitle {
    font-size: 18px;
    color: #606266;
    margin-bottom: 40px;
    line-height: 1.6;
    max-width: 600px;
    margin-left: auto;
    margin-right: auto;
  }

  .hero-actions {
    display: flex;
    justify-content: center;
    gap: 20px;
  }
}

.main-content {
  padding: 60px 20px;
}

.features {
  margin-bottom: 60px;
}

.feature-card {
  text-align: center;
  border-radius: 12px;
  padding: 24px;
  height: 100%;
  .feature-icon {
    color: #409eff;
    margin-bottom: 20px;
  }
  .feature-title {
    font-size: 18px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 12px;
  }
  .feature-description {
    font-size: 14px;
    color: #909399;
    line-height: 1.5;
  }
}

.stats {
  padding: 40px 20px;
  background-color: #ffffff;
  border-radius: 12px;
}

.stat-item {
  display: flex;
  justify-content: center;
}

.stat-title {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #606266;
  gap: 8px;
}

// Responsive adjustments
@media (max-width: 768px) {
  .hero {
    padding: 60px 20px;
    .hero-title {
      font-size: 32px;
    }
    .hero-subtitle {
      font-size: 16px;
    }
  }

  .feature-card {
    margin-bottom: 24px;
  }

  .stats {
    padding: 20px 10px;
  }
}
</style>
