import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: () => import('@/views/Dashboard.vue') },
    { path: '/instances', name: 'instances', component: () => import('@/views/Instances.vue') },
    { path: '/instances/create', name: 'instance-create', component: () => import('@/views/CreateInstance.vue') },
    { path: '/instances/:name', name: 'instance-detail', component: () => import('@/views/InstanceDetail.vue') },
    { path: '/instances/:name/edit', name: 'instance-edit', component: () => import('@/views/EditInstance.vue') },
    { path: '/keystores', name: 'keystores', component: () => import('@/views/KeyStores.vue') },
    { path: '/keystores/create', name: 'keystores-create', component: () => import('@/views/CreateKeyStore.vue') },
    { path: '/keystores/generate', name: 'keystores-generate', component: () => import('@/views/GenerateKeyStore.vue') },
    { path: '/keystores/:name/update', name: 'keystores-update', component: () => import('@/views/UpdateCertificate.vue') },
    { path: '/keystores/:name/export-csr', name: 'keystores-export-csr', component: () => import('@/views/ExportCSR.vue') },
    { path: '/trusted', name: 'trusted', component: () => import('@/views/TrustedCertificates.vue') },
    { path: '/logs', name: 'logs', component: () => import('@/views/Logs.vue') },
    { path: '/settings', name: 'settings', component: () => import('@/views/Settings.vue') },
  ],
})

export default router
