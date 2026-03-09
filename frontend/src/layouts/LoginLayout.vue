<template>
  <div class="login-wrap">
    <div class="login-card glass">
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.login-wrap {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.login-card {
  width: min(420px, 92vw);
  padding: 22px 22px 18px;
}
</style>

