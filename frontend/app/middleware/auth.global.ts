export default defineNuxtRouteMiddleware(async (to) => {
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
