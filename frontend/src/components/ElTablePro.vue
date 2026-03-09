<template>
  <div class="glass wrap">
    <div class="toolbar">
      <slot name="toolbar" />
    </div>

    <div class="table">
      <slot />
    </div>

    <div class="pager">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        :page-sizes="[10, 20, 50, 100]"
        @current-change="(v) => { $emit('update:page', v); $emit('change') }"
        @size-change="(v) => { $emit('update:pageSize', v); $emit('update:page', 1); $emit('change') }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  total: number
  page: number
  pageSize: number
}>()

defineEmits<{
  (e: 'update:page', v: number): void
  (e: 'update:pageSize', v: number): void
  (e: 'change'): void
}>()
</script>

<style scoped>
.wrap {
  padding: 12px;
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 10px;
  flex-shrink: 0;
}
.table {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
  flex-shrink: 0;
}
</style>

