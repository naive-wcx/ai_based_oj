<template>
  <el-tag :type="tagType" size="small">{{ label }}</el-tag>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  difficulty: {
    type: String,
    default: 'easy',
  },
})

const tagType = computed(() => {
  const types = {
    easy: 'success',
    medium: 'warning',
    hard: 'danger',
  }
  return types[props.difficulty] || 'info'
})

const label = computed(() => {
  const labels = {
    easy: '简单',
    medium: '中等',
    hard: '困难',
  }
  return labels[props.difficulty] || props.difficulty
})
</script>
