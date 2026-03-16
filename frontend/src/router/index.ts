import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import AdminView from '../views/AdminView.vue'
import CallLogView from '../views/CallLogView.vue'
import SettingsView from '../views/SettingsView.vue'
import DirectoryView from '../views/DirectoryView.vue'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', component: LoginView },
  { path: '/extensions', component: DashboardView, meta: { requiresAuth: true } },
  { path: '/directory', component: DirectoryView, meta: { requiresAuth: true } },
  { path: '/settings', component: SettingsView, meta: { requiresAuth: true } },
  { path: '/call-log', component: CallLogView, meta: { requiresAuth: true, requiresCallLog: true } },
  { path: '/admin', component: AdminView, meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/', redirect: '/extensions' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return '/login'
  }
  if (to.meta.requiresAdmin && !auth.isAdmin) {
    return '/extensions'
  }
  if (to.meta.requiresCallLog && !auth.hasCallLogAccess) {
    return '/extensions'
  }
})

export default router
