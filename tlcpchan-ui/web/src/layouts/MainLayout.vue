<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="logo">
        <h1>TLCP Channel</h1>
      </div>
      <el-menu :default-active="route.path" router background-color="#1d1e1f" text-color="#bfcbd9" active-text-color="#409eff">
        <el-menu-item index="/">
          <el-icon><DataBoard /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/instances">
          <el-icon><Connection /></el-icon>
          <span>实例管理</span>
        </el-menu-item>
        <el-menu-item index="/certificates">
          <el-icon><Key /></el-icon>
          <span>证书管理</span>
        </el-menu-item>
        <el-menu-item index="/logs">
          <el-icon><Document /></el-icon>
          <span>日志查看</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
      <div class="sidebar-footer">
        <div class="version-info">
          <span>UI: v{{ uiVersion }}</span>
          <span>后端: v{{ backendVersion || '-' }}</span>
        </div>
        <div class="links">
          <a href="https://github.com/Trisia/tlcpchan" target="_blank" class="link">
            <el-icon><Link /></el-icon>
            GitHub
          </a>
          <a href="https://github.com/Trisia/tlcpchan/tree/main/docs" target="_blank" class="link">
            <el-icon><Reading /></el-icon>
            文档
          </a>
        </div>
      </div>
    </el-aside>
    <el-main class="main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { uiApi, systemApi } from '@/api'

const route = useRoute()
const uiVersion = ref('dev')
const backendVersion = ref('')

onMounted(async () => {
  uiVersion.value = await uiApi.fetchStaticVersion()
  try {
    const info = await systemApi.version()
    backendVersion.value = info.version
  } catch {
    // ignore
  }
})
</script>

<style scoped>
.layout {
  height: 100vh;
}
.aside {
  background-color: #1d1e1f;
  display: flex;
  flex-direction: column;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #3a3b3c;
}
.logo h1 {
  color: #fff;
  font-size: 18px;
  margin: 0;
  font-weight: 600;
}
.el-menu {
  border-right: none;
  flex: 1;
}
.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #3a3b3c;
}
.version-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
  font-size: 12px;
  color: #909399;
}
.links {
  display: flex;
  gap: 16px;
}
.link {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  text-decoration: none;
  font-size: 13px;
  transition: color 0.2s;
}
.link:hover {
  color: #409eff;
}
.main {
  background-color: #f5f7fa;
  padding: 20px;
  overflow-y: auto;
}
</style>
