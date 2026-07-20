/* eslint-disable */
// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2026-06-30',

  css: ['~/assets/css/main.css'],

  modules: [
    '@nuxt/eslint',
    '@nuxt/ui'
  ],

  devtools: {
    enabled: true
  },

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  },

  routeRules: {
    '/': { redirect: { to: '/dashboard', statusCode: 302 } }
  },

  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8080/api/v1'
    }
  }
})
