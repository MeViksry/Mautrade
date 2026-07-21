export const useAdminAuth = () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase
  const tokenCookie = useCookie<string | null>('admin_auth_token', {
    maxAge: 30 * 24 * 60 * 60, // 30 days
    secure: !import.meta.dev,
    sameSite: 'lax'
  })

  // State to hold the current admin user data
  const adminUser = useState('admin_auth_user', () => null)

  const loginAdmin = async (payload: { email: string, password: string }) => {
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const response = await $fetch<any>(`${apiBase}/admin/auth/login`, {
        method: 'POST',
        body: payload
      })

      if (response.session?.token) {
        tokenCookie.value = response.session.token
        adminUser.value = response.admin
      }
      return { success: true }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      if (error.response?.status === 401) {
        throw new Error('Invalid email or password', { cause: error })
      }
      throw new Error(error.data?.error || error.message || 'Failed to login', { cause: error })
    }
  }

  const fetchAdminUser = async () => {
    if (!tokenCookie.value) return null

    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const response = await $fetch<any>(`${apiBase}/admin/auth/me`, {
        headers: {
          Authorization: `Bearer ${tokenCookie.value}`
        }
      })
      adminUser.value = response.admin
      return response.admin
    } catch {
      tokenCookie.value = null
      adminUser.value = null
      return null
    }
  }

  const logoutAdmin = async () => {
    try {
      if (tokenCookie.value) {
        await $fetch(`${apiBase}/admin/auth/logout`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${tokenCookie.value}`
          }
        })
      }
    } catch (e) {
      console.warn('Logout request failed', e)
    } finally {
      tokenCookie.value = null
      adminUser.value = null
    }
  }

  return {
    adminUser,
    tokenCookie,
    loginAdmin,
    fetchAdminUser,
    logoutAdmin
  }
}
