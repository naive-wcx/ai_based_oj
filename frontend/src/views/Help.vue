<template>
  <div class="help-page">
    <section class="hero">
      <h1>评测环境帮助</h1>
      <p>这里是做题和比赛时会用到的规则说明。</p>
    </section>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">比赛怎么判</div>
      </template>
      <ul class="rules">
        <li><strong>OI 赛制</strong>：比赛进行中普通用户仅显示 `Submitted`，详细结果赛后查看。</li>
        <li><strong>OI 总分可见时机</strong>：固定起止模式在比赛结束后可见；窗口模式在你的个人时长结束后可见。</li>
        <li><strong>IOI 赛制</strong>：按题目得分计分，赛时可查看分数与通过情况。</li>
        <li><strong>固定起止（fixed）</strong>：所有人使用同一比赛开始和结束时间。</li>
        <li><strong>窗口模式（window）</strong>：在窗口期内点击“开始比赛”后，进入个人计时。</li>
        <li><strong>窗口模式提示</strong>：没点“开始比赛”前，不会进入个人计时。</li>
        <li><strong>提交次数上限</strong>：每场比赛每位用户最多提交 99 次。</li>
      </ul>
    </el-card>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">评测环境参数</div>
      </template>
      <div class="kv-grid">
        <div class="kv-item">
          <div class="k">操作系统</div>
          <div class="v">Debian 12（bookworm）x86_64</div>
        </div>
        <div class="kv-item">
          <div class="k">判题沙箱</div>
          <div class="v">simple sandbox</div>
        </div>
        <div class="kv-item">
          <div class="k">栈空间设置</div>
          <div class="v">Linux 下运行前会按 `memory_limit * 1024`（KB）设置 `ulimit -s`</div>
        </div>
        <div class="kv-item">
          <div class="k">虚拟内存限制</div>
          <div class="v">Linux 下运行前会按 `memory_limit * 1024`（KB）设置 `ulimit -v`</div>
        </div>
        <div class="kv-item">
          <div class="k">内存统计</div>
          <div class="v">按程序运行期间虚拟内存峰值（VmPeak）统计，显示单位为 KB</div>
        </div>
        <div class="kv-item">
          <div class="k">编译超时</div>
          <div class="v">30 秒</div>
        </div>
      </div>
    </el-card>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">语言与编译/运行命令</div>
      </template>
      <el-table :data="languages" border>
        <el-table-column prop="language" label="语言" width="100" />
        <el-table-column prop="source" label="源文件名" width="120" />
        <el-table-column prop="compile" label="编译命令" min-width="340" />
        <el-table-column prop="run" label="运行命令" min-width="220" />
        <el-table-column prop="version" label="编译器/运行时版本（当前环境）" min-width="240" />
      </el-table>
      <p class="tip">
        版本以评测机实际安装环境为准。
      </p>
    </el-card>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">判题细则</div>
      </template>
      <ul class="rules">
        <li>时间限制：按题目 `time_limit`（ms）判定，运行超时返回 `TLE`。</li>
        <li>内存限制：按题目 `memory_limit`（MB）限制虚拟内存，超过限制返回 `Memory Limit Exceeded`。</li>
        <li>内存统计：`memory_used` 显示程序运行期间虚拟内存峰值（VmPeak，KB）。</li>
        <li>栈空间：运行前会执行 `ulimit -s = memory_limit * 1024`（KB）。</li>
        <li>虚拟内存限制：运行前会执行 `ulimit -v = memory_limit * 1024`（KB）。</li>
        <li>编译超时：编译阶段超时上限为 30 秒。</li>
      </ul>
    </el-card>
  </div>
</template>

<script setup>
const languages = [
  {
    language: 'C',
    source: 'main.c',
    compile: 'gcc -o main main.c -O2 -Wall -lm -std=c11',
    run: './main',
    version: 'gcc 12.2.0（Debian 12）',
  },
  {
    language: 'C++',
    source: 'main.cpp',
    compile: 'g++ -o main main.cpp -O2 -Wall -std=c++17',
    run: './main',
    version: 'g++ 12.2.0（Debian 12）',
  },
  {
    language: 'Python',
    source: 'main.py',
    compile: '-',
    run: 'python3 main.py',
    version: 'Python 3.11.x（Debian 12）',
  },
  {
    language: 'Java',
    source: 'Main.java',
    compile: 'javac Main.java',
    run: 'java Main',
    version: 'OpenJDK 17',
  },
  {
    language: 'Go',
    source: 'main.go',
    compile: 'go build -o main main.go',
    run: './main',
    version: 'Go 1.21.x',
  },
]
</script>

<style scoped lang="scss">
.help-page {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.hero {
  background: linear-gradient(120deg, #f8fbff, #ffffff);
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 24px;

  h1 {
    margin: 0 0 8px 0;
    font-size: 28px;
  }

  p {
    margin: 0;
    color: var(--swiss-text-secondary);
  }
}

.card {
  :deep(.el-card__body) {
    padding-top: 16px !important;
  }
}

.card-title {
  font-weight: 600;
  color: var(--swiss-text-main);
}

.kv-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(260px, 1fr));
  gap: 14px;
}

.kv-item {
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 12px;
  background: #fcfcfd;

  .k {
    font-size: 12px;
    color: var(--swiss-text-secondary);
    margin-bottom: 4px;
  }

  .v {
    font-weight: 500;
    word-break: break-word;
  }
}

.tip {
  margin: 12px 0 0 0;
  color: var(--swiss-text-secondary);
  font-size: 13px;
}

.rules {
  margin: 14px 0 0 0;
  padding-left: 18px;
  line-height: 1.8;
}

@media (max-width: 900px) {
  .kv-grid {
    grid-template-columns: 1fr;
  }
}
</style>
