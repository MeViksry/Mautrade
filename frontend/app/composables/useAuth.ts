export const useAuth = () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase
  const tokenCookie = useCookie<string | null>('auth_token', {
    maxAge: 30 * 24 * 60 * 60, // 30 days
    secure: !import.meta.dev,
    sameSite: 'lax'
  })

  // State to hold the current user data
  const user = useState('auth_user', () => null)

  // We can track the dev OTP to show it temporarily if needed
  const devOtp = useState('auth_dev_otp', () => '')

  const login = async (payload: { email: string, password: string }) => {
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const response = await $fetch<any>(`${apiBase}/auth/login`, {
        method: 'POST',
        body: payload
      })

      if (response.devOtp) {
        devOtp.value = response.devOtp
      }

      if (response.otpRequired) {
        return { otpRequired: true, expiresAt: response.otpExpiresAt }
      }

      if (response.session?.token) {
        tokenCookie.value = response.session.token
        user.value = response.user
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

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const register = async (payload: any) => {
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const response = await $fetch<any>(`${apiBase}/auth/register`, {
        method: 'POST',
        body: payload
      })

      if (response.devOtp) {
        devOtp.value = response.devOtp
      }

      return { otpRequired: response.otpRequired, expiresAt: response.otpExpiresAt }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      if (error.response?.status === 409) {
        throw new Error('Account already exists', { cause: error })
      }
      throw new Error(error.data?.error || error.message || 'Failed to register', { cause: error })
    }
  }

  const verifyOtp = async (payload: { email: string, code: string, purpose: string }) => {
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const response = await $fetch<any>(`${apiBase}/auth/verify-otp`, {
        method: 'POST',
        body: payload
      })

      if (response.session?.token) {
        tokenCookie.value = response.session.token
        user.value = response.user
        devOtp.value = '' // clear the dev OTP
      }
      return { success: true }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      if (error.response?.status === 401) {
        throw new Error('Invalid OTP code', { cause: error })
      }
      if (error.response?.status === 410) {
        throw new Error('OTP has expired', { cause: error })
      }
      throw new Error(error.data?.error || error.message || 'Failed to verify OTP', { cause: error })
    }
  }
  const logout = async () => {
    try {
      if (tokenCookie.value) {
        await $fetch(`${apiBase}/auth/logout`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${tokenCookie.value}`
          }
        })
      }
    } catch {
      // ignore logout errors on client
    } finally {
      tokenCookie.value = null
      user.value = null
      await navigateTo('/signin')
    }
  }

  const fetchUser = async () => {
    if (!tokenCookie.value) {
      user.value = null
      return null
    }

    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const response = await $fetch<any>(`${apiBase}/auth/me`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${tokenCookie.value}`
        }
      })
      user.value = response.user
      return response.user
    } catch {
      tokenCookie.value = null
      user.value = null
      return null
    }
  }

  return {
    user,
    tokenCookie,
    devOtp,
    login,
    register,
    verifyOtp,
    logout,
    fetchUser
  }
}
