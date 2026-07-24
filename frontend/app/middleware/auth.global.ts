export default defineNuxtRouteMiddleware(async (to) => {
  const isAdminRoute = to.path.startsWith('/admin')
  const isAdminLogin = to.path === '/admin/login'

  if (isAdminRoute) {
    const adminToken = useCookie('admin_auth_token')

    // If not logged in and trying to access protected admin pages
    if (!adminToken.value && !isAdminLogin) {
      return navigateTo('/admin/login')
    }

    // If already logged in and trying to access admin login
    if (adminToken.value && isAdminLogin) {
      return navigateTo('/admin')
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
      return navigateTo('/signin')
    }
    return
  }

  if (tokenCookie.value && !isAccountComplete.value) {
    if (isDashboard) {
      await logout()
      return navigateTo('/signup')
    }
    if (!isOnboarding) {
      return navigateTo('/onboarding')
    }
  }

  if (tokenCookie.value && isAccountComplete.value && isOnboarding) {
    return navigateTo('/dashboard')
  }
})
