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
      <MarkdownPreview :content="contestRulesMarkdown" />
    </el-card>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">评测环境参数</div>
      </template>
      <MarkdownPreview :content="environmentMarkdown" />
    </el-card>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">语言与编译/运行命令</div>
      </template>
      <MarkdownPreview :content="languageMarkdown" />
    </el-card>

    <el-card class="card" shadow="never">
      <template #header>
        <div class="card-title">判题细则</div>
      </template>
      <MarkdownPreview :content="judgeRulesMarkdown" />
    </el-card>
  </div>
</template>

<script setup>
import MarkdownPreview from '@/components/common/MarkdownPreview.vue'

const contestRulesMarkdown = `
- **OI 赛制**：比赛进行中普通用户仅显示 \`Submitted\`，详细结果赛后查看。
- **OI 总分可见时机**：固定起止模式在比赛结束后可见；窗口模式在你的个人时长结束后可见。
- **IOI 赛制**：按题目得分计分，赛时可查看分数与通过情况。
- **固定起止（fixed）**：所有人使用同一比赛开始和结束时间。
- **窗口模式（window）**：在窗口期内点击“开始比赛”后，进入个人计时。
- **窗口模式提示**：没点“开始比赛”前，不会进入个人计时。
- **提交次数上限**：每场比赛每位用户最多提交 99 次。
`

const environmentMarkdown = `
- **操作系统**：Debian 12（bookworm）x86_64
- **判题沙箱**：simple sandbox
- **栈空间设置**：Linux 下运行前会按 \`memory_limit * 1024\`（KB）设置 \`ulimit -s\`
- **虚拟内存限制**：Linux 下运行前会按 \`memory_limit * 1024\`（KB）设置 \`ulimit -v\`
- **内存统计**：按程序运行期间虚拟内存峰值（VmPeak）统计，显示单位为 KB
- **编译超时**：30 秒
`

const languageMarkdown = `
| 语言 | 源文件名 | 编译命令 | 运行命令 | 编译器/运行时版本（当前环境） |
| --- | --- | --- | --- | --- |
| C | \`main.c\` | \`gcc -o main main.c -O2 -Wall -lm -std=c11\` | \`./main\` | gcc 12.2.0（Debian 12） |
| C++ | \`main.cpp\` | \`g++ -o main main.cpp -O2 -Wall -std=c++17\` | \`./main\` | g++ 12.2.0（Debian 12） |
| Python | \`main.py\` | - | \`python3 main.py\` | Python 3.11.x（Debian 12） |
| Java | \`Main.java\` | \`javac Main.java\` | \`java Main\` | OpenJDK 17 |
| Go | \`main.go\` | \`go build -o main main.go\` | \`./main\` | Go 1.21.x |

版本以评测机实际安装环境为准。
`

const judgeRulesMarkdown = `
- 时间限制：按题目 \`time_limit\`（ms）判定，运行超时返回 \`TLE\`。
- 内存限制：按题目 \`memory_limit\`（MB）限制虚拟内存，超过限制返回 \`Memory Limit Exceeded\`。
- 内存统计：\`memory_used\` 显示程序运行期间虚拟内存峰值（VmPeak，KB）。
- 栈空间：运行前会执行 \`ulimit -s = memory_limit * 1024\`（KB）。
- 虚拟内存限制：运行前会执行 \`ulimit -v = memory_limit * 1024\`（KB）。
- 编译超时：编译阶段超时上限为 30 秒。
`
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
</style>
