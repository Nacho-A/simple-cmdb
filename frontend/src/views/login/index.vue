<template>
  <div class="login">
    <div class="head">
      <div class="name">Cursor CMDB</div>
      <div class="desc">企业级资产配置管理 · JWT + Casbin · 动态路由权限</div>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="form">
      <el-form-item label="用户名" prop="username">
        <el-input v-model="form.username" placeholder="请输入用户名" autocomplete="username" />
      </el-form-item>

      <el-form-item label="密码" prop="password">
        <el-input
          v-model="form.password"
          placeholder="请输入密码"
          autocomplete="current-password"
          show-password
          type="password"
        />
      </el-form-item>

      <el-form-item label="验证码" prop="captcha">
        <div class="captcha-row">
          <el-input v-model="form.captcha" placeholder="输入右侧数字" maxlength="4" />
          <el-button class="captcha-btn" @click="refreshCaptcha">{{ captcha }}</el-button>
        </div>
      </el-form-item>

      <div class="ops">
        <el-checkbox v-model="form.remember">记住密码</el-checkbox>
      </div>

      <el-button type="primary" class="submit" :loading="loading" @click="onSubmit">登录</el-button>

      <div class="hint">默认管理员：admin / admin123</div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const formRef = ref<FormInstance>()
const captcha = ref('0000')

const LS_KEY = 'cmdb_remember'

const form = reactive({
  username: '',
  password: '',
  captcha: '',
  remember: true,
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

function refreshCaptcha() {
  captcha.value = String(Math.floor(1000 + Math.random() * 9000))
}

function loadRemember() {
  try {
    const raw = localStorage.getItem(LS_KEY)
    if (!raw) return
    const data = JSON.parse(raw)
    form.username = data.username || ''
    form.password = data.password || ''
    form.remember = !!data.remember
  } catch {}
}

async function onSubmit() {
  await formRef.value?.validate()
  if (String(form.captcha).trim() !== captcha.value) {
    ElMessage.error('验证码错误')
    refreshCaptcha()
    return
  }

  loading.value = true
  try {
    await userStore.login(form.username.trim(), form.password)
    await userStore.fetchMe()

    if (form.remember) {
      localStorage.setItem(
        LS_KEY,
        JSON.stringify({ remember: true, username: form.username, password: form.password }),
      )
    } else {
      localStorage.removeItem(LS_KEY)
    }

    ElMessage.success('登录成功')
    router.replace('/dashboard')
  } finally {
    loading.value = false
  }
}

refreshCaptcha()
loadRemember()
</script>

<style scoped>
.login {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.head {
  margin-bottom: 6px;
}

.name {
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 0.2px;
}

.desc {
  opacity: 0.75;
  margin-top: 4px;
  font-size: 13px;
}

.form {
  margin-top: 6px;
}

.captcha-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
}

.captcha-btn {
  width: 112px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
    monospace;
  letter-spacing: 2px;
}

.ops {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 2px;
}

.submit {
  width: 100%;
  margin-top: 8px;
  border-radius: 12px;
}

.hint {
  text-align: center;
  opacity: 0.7;
  font-size: 12px;
  margin-top: 10px;
}
</style>

