<template>
  <div ref="editorRef" class="code-editor"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, shallowRef } from 'vue'
import * as monaco from 'monaco-editor'
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker'
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker'
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker'

// Monaco Editor workers
self.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === 'json') {
      return new jsonWorker()
    }
    if (label === 'css' || label === 'scss' || label === 'less') {
      return new cssWorker()
    }
    if (label === 'html' || label === 'handlebars' || label === 'razor') {
      return new htmlWorker()
    }
    if (label === 'typescript' || label === 'javascript') {
      return new tsWorker()
    }
    return new editorWorker()
  },
}

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  language: {
    type: String,
    default: 'cpp',
  },
  theme: {
    type: String,
    default: 'vs-dark', // vs, vs-dark, hc-black
  },
  readonly: {
    type: Boolean,
    default: false,
  },
  tabSize: {
    type: Number,
    default: 4,
  },
})

const emit = defineEmits(['update:modelValue'])

const editorRef = ref(null)
const editor = shallowRef(null)

const updateEditorOptions = () => {
  if (editor.value) {
    editor.value.getModel().updateOptions({
      tabSize: props.tabSize,
      insertSpaces: true,
    })
  }
}

onMounted(() => {
  if (editorRef.value) {
    editor.value = monaco.editor.create(editorRef.value, {
      value: props.modelValue,
      language: props.language,
      theme: props.theme,
      readOnly: props.readonly,
      automaticLayout: true,
      minimap: {
        enabled: !props.readonly, // Disable minimap in readonly mode for a cleaner look
      },
      fontSize: 14,
      fontFamily: "'Fira Code', 'Monaco', monospace",
      lineHeight: 21, // Adjust line height for a more compact feel
      scrollBeyondLastLine: false,
      roundedSelection: false,
      overviewRulerLanes: 0,
    })

    updateEditorOptions()

    editor.value.onDidChangeModelContent(() => {
      if (!props.readonly) {
        emit('update:modelValue', editor.value.getValue())
      }
    })
  }
})

// Update editor content when modelValue changes
watch(
  () => props.modelValue,
  (newValue) => {
    if (editor.value && newValue !== editor.value.getValue()) {
      editor.value.setValue(newValue)
    }
  }
)

// Update editor language when language prop changes
watch(
  () => props.language,
  (newLang) => {
    if (editor.value) {
      monaco.editor.setModelLanguage(editor.value.getModel(), newLang)
    }
  }
)

// Watch for readonly prop changes
watch(
  () => props.readonly,
  (isReadonly) => {
    if (editor.value) {
      editor.value.updateOptions({ readOnly: isReadonly })
    }
  }
)

// Watch for tabSize prop changes
watch(() => props.tabSize, updateEditorOptions)

onBeforeUnmount(() => {
  if (editor.value) {
    editor.value.dispose()
  }
})
</script>

<style scoped>
.code-editor {
  width: 100%;
  height: 100%;
  min-height: 400px; /* Or a specific height based on layout */
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  overflow: hidden;
}
</style>
