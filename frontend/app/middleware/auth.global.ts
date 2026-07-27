export default defineNuxtRouteMiddleware(async (to) => {
  const nuxtApp = useNuxtApp()
  const isAdminRoute = to.path.startsWith('/admin')
  const isAdminLogin = to.path === '/admin/login'

  if (isAdminRoute) {
    const adminToken = useCookie('admin_auth_token')

    // If not logged in and trying to access protected admin pages
    if (!adminToken.value && !isAdminLogin) {
      return nuxtApp.runWithContext(() => navigateTo('/admin/login'))
    }

    // If already logged in and trying to access admin login
    if (adminToken.value && isAdminLogin) {
      return nuxtApp.runWithContext(() => navigateTo('/admin'))
    }

    // Do not run the regular user auth checks for admin routes
    return
  }

  const tokenCookie = useCookie('auth_token')
  const { isAccountComplete, logout } = useAuth()
  const isOnboarding = to.path.startsWith('/onboarding')
  const isDashboard = to.path.startsWith('/dashboard')

  if (!tokenCookie.value) {
    if (isOnboarding || isDashboard) {
      return nuxtApp.runWithContext(() => navigateTo('/signin'))
    }
    return
  }

  if (tokenCookie.value && !isAccountComplete.value) {
    if (isDashboard) {
      await logout({ redirectTo: false })
      return nuxtApp.runWithContext(() => navigateTo('/signup'))
    }
    if (!isOnboarding) {
      return nuxtApp.runWithContext(() => navigateTo('/onboarding'))
    }
  }

  if (tokenCookie.value && isAccountComplete.value && isOnboarding) {
    return nuxtApp.runWithContext(() => navigateTo('/dashboard'))
  }
})
