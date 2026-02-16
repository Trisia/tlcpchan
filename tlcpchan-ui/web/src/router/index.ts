import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: () => import('@/views/Dashboard.vue') },
    { path: '/instances', name: 'instances', component: () => import('@/views/Instances.vue') },
    { path: '/instances/:name', name: 'instance-detail', component: () => import('@/views/InstanceDetail.vue') },
    { path: '/keystores', name: 'keystores', component: () => import('@/views/KeyStores.vue') },
    { path: '/trusted', name: 'trusted', component: () => import('@/views/TrustedCertificates.vue') },
    { path: '/logs', name: 'logs', component: () => import('@/views/Logs.vue') },
    { path: '/settings', name: 'settings', component: () => import('@/views/Settings.vue') },
  ],
})

export default router
