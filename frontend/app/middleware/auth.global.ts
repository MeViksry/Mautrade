export default defineNuxtRouteMiddleware(async (to) => {
  const tokenCookie = useCookie('auth_token')
  const { fetchUser, isAccountComplete } = useAuth()
  const isAuthRoute = to.path === '/signin' || to.path === '/signup'
  const isAdminRoute = to.path.startsWith('/admin')

  if (isAdminRoute) {
    return
  }

  await fetchUser()

  const isRoot = to.path === '/'
  const isOnboarding = to.path.startsWith('/onboarding')

  // Completely unauthenticated
  if (!tokenCookie.value) {
    if (isRoot) return navigateTo('/signup')
    if (isOnboarding) return navigateTo('/signup')
    if (!isAuthRoute && to.path.startsWith('/dashboard')) return navigateTo('/signin')
    return
  }

  // Authenticated
  if (isAccountComplete.value) {
    if (isRoot || isAuthRoute || isOnboarding) {
      return navigateTo('/dashboard')
    }
    return
  }

  // Authenticated but incomplete (e.g. verified OTP but hasn't finished onboarding)
  if (!isAccountComplete.value) {
    if (isRoot) {
      tokenCookie.value = null
      return navigateTo('/signup')
    }
    if (isAuthRoute) {
      tokenCookie.value = null
      return
    }
    if (!isOnboarding) {
      return navigateTo('/onboarding')
    }
  }
})
