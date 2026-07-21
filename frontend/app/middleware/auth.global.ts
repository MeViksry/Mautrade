export default defineNuxtRouteMiddleware(async (to) => {
  const tokenCookie = useCookie('auth_token')
  const { fetchUser, isAccountComplete } = useAuth()
  const isAuthRoute = to.path === '/signin' || to.path === '/signup'

  await fetchUser()

  if (!tokenCookie.value) {
    if (!isAuthRoute && (to.path.startsWith('/dashboard') || to.path.startsWith('/onboarding'))) {
      return navigateTo('/signup')
    }
    return
  }

  if (tokenCookie.value && isAuthRoute) {
    if (isAccountComplete.value) {
      return navigateTo('/dashboard')
    } else {
      return navigateTo('/onboarding')
    }
  }

  if (tokenCookie.value && !isAccountComplete.value) {
    if (!to.path.startsWith('/signup') && !to.path.startsWith('/onboarding')) {
      return navigateTo('/signup')
    }
  }
})
