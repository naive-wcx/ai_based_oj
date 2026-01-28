<template>
  <div class="markdown-preview" v-html="renderedContent"></div>
</template>

<script setup>
import { computed } from 'vue'
import { marked } from 'marked'
import katex from 'katex'
import 'katex/dist/katex.min.css'

// 配置 marked
marked.setOptions({
  breaks: true,
  gfm: true,
})

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
})

// 渲染 LaTeX 公式
function renderLatex(text) {
  if (!text) return text
  
  // 处理行内公式 $...$
  text = text.replace(/\$([^\$\n]+?)\$/g, (match, formula) => {
    try {
      return katex.renderToString(formula, { 
        throwOnError: false,
        displayMode: false 
      })
    } catch (e) {
      return match
    }
  })
  
  // 处理块级公式 $$...$$
  text = text.replace(/\$\$([^\$]+?)\$\$/g, (match, formula) => {
    try {
      return '<div class="katex-block">' + katex.renderToString(formula, { 
        throwOnError: false,
        displayMode: true 
      }) + '</div>'
    } catch (e) {
      return match
    }
  })
  
  return text
}

const renderedContent = computed(() => {
  if (!props.content) return '<p style="color: #909399;">暂无内容</p>'
  
  // 先处理 LaTeX，再处理 Markdown
  let content = props.content
  
  // 保护代码块中的 $ 符号
  const codeBlocks = []
  content = content.replace(/```[\s\S]*?```/g, (match) => {
    codeBlocks.push(match)
    return `__CODE_BLOCK_${codeBlocks.length - 1}__`
  })
  content = content.replace(/`[^`]+`/g, (match) => {
    codeBlocks.push(match)
    return `__CODE_BLOCK_${codeBlocks.length - 1}__`
  })
  
  // 渲染 LaTeX
  content = renderLatex(content)
  
  // 恢复代码块
  content = content.replace(/__CODE_BLOCK_(\d+)__/g, (match, index) => {
    return codeBlocks[parseInt(index)]
  })
  
  // 渲染 Markdown
  return marked(content)
})
</script>

<style lang="scss" scoped>
.markdown-preview {
  line-height: 1.8;
  color: #303133;
  
  :deep(h1), :deep(h2), :deep(h3), :deep(h4), :deep(h5), :deep(h6) {
    margin: 16px 0 8px;
    font-weight: 600;
    color: #303133;
  }
  
  :deep(h1) { font-size: 24px; }
  :deep(h2) { font-size: 20px; }
  :deep(h3) { font-size: 18px; }
  :deep(h4) { font-size: 16px; }
  
  :deep(p) {
    margin: 8px 0;
  }
  
  :deep(ul), :deep(ol) {
    padding-left: 24px;
    margin: 8px 0;
  }
  
  :deep(li) {
    margin: 4px 0;
  }
  
  :deep(code) {
    background: #f5f7fa;
    padding: 2px 6px;
    border-radius: 4px;
    font-family: 'Fira Code', 'Monaco', monospace;
    font-size: 14px;
    color: #c7254e;
  }
  
  :deep(pre) {
    background: #f5f7fa;
    padding: 16px;
    border-radius: 8px;
    overflow-x: auto;
    margin: 12px 0;
    
    code {
      background: none;
      padding: 0;
      color: #303133;
    }
  }
  
  :deep(blockquote) {
    border-left: 4px solid #409eff;
    padding-left: 16px;
    margin: 12px 0;
    color: #606266;
    background: #f5f7fa;
    padding: 12px 16px;
    border-radius: 0 8px 8px 0;
  }
  
  :deep(table) {
    border-collapse: collapse;
    width: 100%;
    margin: 12px 0;
    
    th, td {
      border: 1px solid #e4e7ed;
      padding: 8px 12px;
      text-align: left;
    }
    
    th {
      background: #f5f7fa;
      font-weight: 600;
    }
    
    tr:nth-child(even) {
      background: #fafafa;
    }
  }
  
  :deep(a) {
    color: #409eff;
    text-decoration: none;
    
    &:hover {
      text-decoration: underline;
    }
  }
  
  :deep(img) {
    max-width: 100%;
    border-radius: 8px;
  }
  
  :deep(hr) {
    border: none;
    border-top: 1px solid #e4e7ed;
    margin: 16px 0;
  }
  
  // KaTeX 块级公式居中
  :deep(.katex-block) {
    text-align: center;
    margin: 16px 0;
    overflow-x: auto;
  }
}
</style>
