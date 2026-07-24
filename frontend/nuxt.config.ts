export default defineNuxtConfig({

  modules: [
    '@nuxt/eslint',
    '@nuxt/ui'
  ],

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE || 'http://localhost:8080/api/v1',
      gasFeeDepositAddress: process.env.GAS_FEE_DEPOSIT_ADDRESS
    }
  },
  srcDir: 'app/',

  routeRules: {
    '/': { redirect: { to: '/dashboard', statusCode: 302 } }
  },
  compatibilityDate: '2024-04-03',

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  }
})
