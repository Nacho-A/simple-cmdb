<template>
  <div class="labels">
    <div v-for="(row, idx) in rows" :key="idx" class="row">
      <el-input v-model="row.key" placeholder="key" class="k" />
      <el-input v-model="row.value" placeholder="value" class="v" />
      <el-button text type="danger" class="x" @click="remove(idx)">
        <el-icon><Delete /></el-icon>
      </el-button>
    </div>
    <el-button plain size="small" @click="add">+ 添加标签</el-button>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

const props = defineProps<{
  modelValue: Record<string, string>
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', v: Record<string, string>): void
}>()

const rows = reactive<{ key: string; value: string }[]>([])

const normalized = computed(() => props.modelValue || {})

watch(
  normalized,
  (v) => {
    rows.splice(0, rows.length)
    for (const [k, val] of Object.entries(v)) rows.push({ key: k, value: String(val ?? '') })
    if (!rows.length) rows.push({ key: '', value: '' })
  },
  { immediate: true },
)

function sync() {
  const out: Record<string, string> = {}
  for (const r of rows) {
    const k = r.key.trim()
    if (!k) continue
    out[k] = r.value
  }
  emit('update:modelValue', out)
}

watch(rows, sync, { deep: true })

function add() {
  rows.push({ key: '', value: '' })
}

function remove(i: number) {
  rows.splice(i, 1)
  if (!rows.length) rows.push({ key: '', value: '' })
}
</script>

<style scoped>
.labels {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.row {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 10px;
  align-items: center;
}
.x {
  padding: 8px 10px;
}
</style>

