<template>
  <footer class="footer">
    <div class="footer-content">
      <p>USTC OJ 在线评测系统 - 支持 AI 智能判题</p>
      <p class="copyright">© {{ year }} All Rights Reserved</p>
    </div>
  </footer>
</template>

<script setup>
const year = new Date().getFullYear()
</script>

<style lang="scss" scoped>
.footer {
  background: #303133;
  color: #909399;
  padding: 24px 0;
  text-align: center;
  margin-top: auto;
}

.footer-content {
  p {
    margin: 4px 0;
    font-size: 14px;
  }
  
  .copyright {
    font-size: 12px;
  }
}
</style>
