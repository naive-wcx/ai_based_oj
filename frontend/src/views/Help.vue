<template>
  <div class="help-page">
    <section class="hero">
      <h1>评测环境帮助</h1>
      <p>本页用于公示 OJ 判题环境与资源限制规则，避免“本地能过、线上不过”的环境差异问题。</p>
    </section>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">系统环境</div>
      </template>
      <div class="kv-grid">
        <div class="kv-item">
          <div class="k">操作系统</div>
          <div class="v">Debian 12（bookworm）x86_64</div>
        </div>
        <div class="kv-item">
          <div class="k">判题沙箱</div>
          <div class="v">simple sandbox（当前后端实现）</div>
        </div>
        <div class="kv-item">
          <div class="k">工作目录</div>
          <div class="v">`./data/sandbox/{submission_id}`</div>
        </div>
        <div class="kv-item">
          <div class="k">输出比较规则</div>
          <div class="v">忽略行尾空格与首尾空白，统一换行符后比较</div>
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
        说明：命令与 `backend/internal/judge/sandbox/sandbox.go` 一致；版本以线上评测机实际安装为准。
      </p>
    </el-card>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">时间、内存与栈空间规则</div>
      </template>
      <el-alert type="warning" :closable="false" show-icon>
        当前 `simple sandbox` 不做真实内存统计，`memory_used` 固定为 0（KB）；但会在 Linux 下设置栈上限。
      </el-alert>
      <ul class="rules">
        <li>时间限制：按题目 `time_limit`（ms）判定，运行超时返回 `TLE`。</li>
        <li>内存限制：当前实现未对进程总内存做硬限制；但 `memory_limit` 会用于设置栈空间上限。</li>
        <li>栈空间：Linux 评测机会在运行前执行 `ulimit -s`，设置为 `memory_limit * 1024`（KB），与题目空间限制同量级。</li>
        <li>编译超时：编译阶段超时上限为 30 秒。</li>
      </ul>
      <p class="tip">
        如后续切换到 isolate/docker 沙箱，上述内存与栈规则会变化，届时会同步更新本页。
      </p>
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
