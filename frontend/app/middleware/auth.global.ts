export default defineNuxtRouteMiddleware((to) => {
  const tokenCookie = useCookie('auth_token')
  const isAuthRoute = to.path === '/signin' || to.path === '/signup'

  // If user is accessing protected routes without a token, redirect to signin
  if (!tokenCookie.value && !isAuthRoute) {
    // Only redirect if it's explicitly a protected route (dashboard, onboarding)
    if (to.path.startsWith('/dashboard') || to.path.startsWith('/onboarding')) {
      return navigateTo('/signin')
    }
  }

  // If user is authenticated and tries to access signin/signup, redirect to dashboard
  if (tokenCookie.value && isAuthRoute) {
    return navigateTo('/dashboard')
  }
})
