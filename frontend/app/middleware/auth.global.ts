export default defineNuxtRouteMiddleware(async (to) => {
  const tokenCookie = useCookie('auth_token')
  const { fetchUser, isAccountComplete } = useAuth()
  const isAuthRoute = to.path === '/signin' || to.path === '/signup'
  const isAdminRoute = to.path.startsWith('/admin')

  if (isAdminRoute) {
    return
  }

  await fetchUser()

  if (!tokenCookie.value) {
    if (!isAuthRoute && (to.path.startsWith('/dashboard') || to.path.startsWith('/onboarding'))) {
      return navigateTo('/signin')
    }
    return
  }

  if (tokenCookie.value && isAuthRoute) {
    if (isAccountComplete.value) {
      return navigateTo('/dashboard')
    } else {
      tokenCookie.value = null
      return
    }
  }

  if (tokenCookie.value && !isAccountComplete.value) {
    if (to.path !== '/' && !to.path.startsWith('/signup') && !to.path.startsWith('/onboarding') && !to.path.startsWith('/signin')) {
      return navigateTo('/onboarding')
    }
  }
})
