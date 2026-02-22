<template>
  <el-container class="layout">
    <el-drawer v-model="mobileMenuOpen" direction="ltr" size="240px" class="mobile-drawer" :with-header="false">
      <div class="mobile-drawer-content">
        <div class="logo">
          <h1>TLCP Channel</h1>
        </div>
        <el-menu :default-active="route.path" router background-color="#1d1e1f" text-color="#bfcbd9"
          active-text-color="#409eff" @select="mobileMenuOpen = false">
          <el-menu-item index="/">
            <el-icon>
              <DataBoard />
            </el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          <el-menu-item index="/instances">
            <el-icon>
              <Connection />
            </el-icon>
            <span>实例管理</span>
          </el-menu-item>
          <el-menu-item index="/keystores">
            <el-icon>
              <Lock />
            </el-icon>
            <span>密钥管理</span>
          </el-menu-item>
          <el-menu-item index="/trusted">
            <el-icon>
              <Lock />
            </el-icon>
            <span>信任证书</span>
          </el-menu-item>
          <el-menu-item index="/logs">
            <el-icon>
              <Document />
            </el-icon>
            <span>日志查看</span>
          </el-menu-item>
          <el-menu-item index="/settings">
            <el-icon>
              <Setting />
            </el-icon>
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
              <el-icon>
                <Link />
              </el-icon>
              GitHub
            </a>
            <a href="https://github.com/Trisia/tlcpchan/tree/main/docs" target="_blank" class="link">
              <el-icon>
                <Reading />
              </el-icon>
              文档
            </a>
          </div>
        </div>
      </div>
    </el-drawer>

    <el-aside width="220px" class="aside desktop-aside">
      <div class="logo">
        <h1>TLCP Channel</h1>
      </div>
      <el-menu :default-active="route.path" router background-color="#1d1e1f" text-color="#bfcbd9"
        active-text-color="#409eff">
        <el-menu-item index="/">
          <el-icon>
            <DataBoard />
          </el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/instances">
          <el-icon>
            <Connection />
          </el-icon>
          <span>实例管理</span>
        </el-menu-item>
        <el-menu-item index="/keystores">
          <el-icon>
            <Lock />
          </el-icon>
          <span>密钥管理</span>
        </el-menu-item>
        <el-menu-item index="/trusted">
          <el-icon>
            <Lock />
          </el-icon>
          <span>信任证书</span>
        </el-menu-item>
        <el-menu-item index="/logs">
          <el-icon>
            <Document />
          </el-icon>
          <span>日志查看</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon>
            <Setting />
          </el-icon>
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
            <el-icon>
              <Link />
            </el-icon>
            GitHub
          </a>
          <a href="https://github.com/Trisia/tlcpchan/tree/main/docs" target="_blank" class="link">
            <el-icon>
              <Reading />
            </el-icon>
            文档
          </a>
        </div>
      </div>
    </el-aside>

    <el-container class="main-container">
      <el-header class="mobile-header">
        <div class="header-left">
          <el-button text @click="mobileMenuOpen = true" class="hamburger-btn">
            <el-icon size="24">
              <Menu />
            </el-icon>
          </el-button>
          <h1 class="mobile-title">TLCP Channel</h1>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Menu } from '@element-plus/icons-vue'
import { systemApi } from '@/api/index'
import axios from 'axios'

const route = useRoute()
const uiVersion = ref('dev')
const backendVersion = ref('')
const mobileMenuOpen = ref(false)

onMounted(() => {
  // 获取UI版本
  axios.get('./version.txt', { responseType: 'text' })
    .then((response) => {
      uiVersion.value = response.data.trim()
    })
    .catch(() => {
      uiVersion.value = 'dev'
    })

  // 获取后端版本
  systemApi.version()
    .then((response) => {
      backendVersion.value = response.version
    })
    .catch(() => {
      // ignore
    })
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

.main-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.main {
  background-color: #f5f7fa;
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.mobile-header {
  display: none;
  align-items: center;
  background-color: #1d1e1f;
  height: 60px;
  padding: 0 16px;
  border-bottom: 1px solid #3a3b3c;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hamburger-btn {
  color: #fff;
  padding: 8px;
}

.mobile-title {
  color: #fff;
  font-size: 18px;
  margin: 0;
  font-weight: 600;
}

.mobile-drawer-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #1d1e1f;
}

@media (max-width: 768px) {
  .desktop-aside {
    display: none;
  }

  .mobile-header {
    display: flex;
  }

  .main {
    padding: 16px;
  }
}
</style>
