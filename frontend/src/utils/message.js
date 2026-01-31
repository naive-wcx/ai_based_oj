import { ElMessage } from 'element-plus'

const DEFAULTS = {
  duration: 1500,
  offset: 72,
  showClose: false,
}

function normalize(type, input, options) {
  if (typeof input === 'string') {
    return { type, message: input, ...DEFAULTS, ...(options || {}) }
  }
  return { type, ...DEFAULTS, ...(input || {}) }
}

export const message = {
  success(input, options) {
    return ElMessage(normalize('success', input, options))
  },
  error(input, options) {
    return ElMessage(normalize('error', input, options))
  },
  warning(input, options) {
    return ElMessage(normalize('warning', input, options))
  },
  info(input, options) {
    return ElMessage(normalize('info', input, options))
  },
}
