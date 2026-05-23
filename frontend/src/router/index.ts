import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/Login.vue'),
        meta: { guest: true },
    },
    {
        path: '/register',
        name: 'Register',
        component: () => import('@/views/Register.vue'),
        meta: { guest: true },
    },
    {
        path: '/',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
    },
    {
        path: '/feed',
        name: 'Feed',
        component: () => import('@/views/Feed.vue'),
        meta: { requiresAuth: true },
    },
    {
        path: '/article/:id',
        name: 'ArticleDetail',
        component: () => import('@/views/ArticleDetail.vue'),
    },
    {
        path: '/editor',
        name: 'EditorCreate',
        component: () => import('@/views/Editor.vue'),
        meta: { requiresAuth: true },
    },
    {
        path: '/editor/:id',
        name: 'EditorEdit',
        component: () => import('@/views/Editor.vue'),
        meta: { requiresAuth: true },
    },
    {
        path: '/profile/:id',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
    },
    {
        path: '/settings',
        name: 'Settings',
        component: () => import('@/views/Settings.vue'),
        meta: { requiresAuth: true },
    },
    {
        path: '/chat',
        name: 'Chat',
        component: () => import('@/views/Chat.vue'),
        meta: { requiresAuth: true },
    },
    {
        path: '/:pathMatch(.*)*',
        name: 'NotFound',
        component: () => import('@/views/NotFound.vue'),
    },
]

const router = createRouter({
    history: createWebHistory(),
    routes,
})

// 全局前置守卫
router.beforeEach(async (to, _from, next) => {
    const authStore = useAuthStore()

    // 如果有 Token 但没有用户信息，尝试获取用户
    if (authStore.accessToken && !authStore.user) {
        try {
            await authStore.fetchCurrentUser()
        } catch {
            // Token 无效，忽略
        }
    }

    // 需要登录的页面
    if (to.meta.requiresAuth && !authStore.isLoggedIn) {
        return next({ name: 'Login', query: { redirect: to.fullPath } })
    }

    // 访客页面（已登录用户不应访问登录/注册）
    if (to.meta.guest && authStore.isLoggedIn) {
        return next({ name: 'Home' })
    }

    next()
})

export default router
