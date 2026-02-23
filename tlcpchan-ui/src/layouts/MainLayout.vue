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
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  Menu,
  DataBoard,
  Connection,
  Lock,
  Document,
  Setting
} from '@element-plus/icons-vue'

const route = useRoute()
const mobileMenuOpen = ref(false)
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
