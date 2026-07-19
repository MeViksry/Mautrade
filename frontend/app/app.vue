<script setup>
const title = 'Mautrade'
const description = 'Mautrade is a crypto trading dashboard for active layers, exchange bindings, gas fee tracking, and trading history.'
const theme = useState('dashboard-theme', () => 'dark')
const themeBootstrapScript = `;(function(){try{var saved=localStorage.getItem('dashboard-theme');var system=window.matchMedia&&window.matchMedia('(prefers-color-scheme: light)').matches?'light':'dark';var theme=(saved==='light'||saved==='dark')?saved:system;document.documentElement.dataset.theme=theme;document.documentElement.style.colorScheme=theme;}catch(error){document.documentElement.dataset.theme='dark';document.documentElement.style.colorScheme='dark';}})();`

const applyTheme = (value) => {
  document.documentElement.dataset.theme = value
  document.documentElement.style.colorScheme = value
}

const getSystemTheme = () => {
  return window.matchMedia?.('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
}

const getSavedTheme = () => {
  const savedTheme = localStorage.getItem('dashboard-theme')
  return savedTheme === 'light' || savedTheme === 'dark' ? savedTheme : null
}

onMounted(() => {
  const currentTheme = document.documentElement.dataset.theme
  theme.value = currentTheme === 'light' || currentTheme === 'dark'
    ? currentTheme
    : getSavedTheme() || getSystemTheme()

  applyTheme(theme.value)

  watch(theme, (value) => {
    applyTheme(value)
    localStorage.setItem('dashboard-theme', value)
  })

  const systemTheme = window.matchMedia?.('(prefers-color-scheme: light)')
  systemTheme?.addEventListener('change', () => {
    if (getSavedTheme()) return
    theme.value = getSystemTheme()
  })
})

useHead({
  htmlAttrs: { lang: 'en' },
  script: [
    {
      innerHTML: themeBootstrapScript,
      tagPosition: 'head'
    }
  ],
  link: [
    { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
    { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Oswald:wght@200;300;400;500;700&family=Roboto:wght@100;300;400;500&family=JetBrains+Mono:wght@400;500&display=swap' }
  ]
})

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  twitterTitle: title,
  twitterDescription: description,
  twitterCard: 'summary_large_image'
})
</script>

<template>
  <div>
    <div class="noise-overlay" />
    <NuxtLayout>
      <NuxtPage :page-key="route => route.fullPath" />
    </NuxtLayout>
  </div>
</template>
