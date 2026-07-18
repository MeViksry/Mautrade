<script setup>
const title = 'Mautrade'
const description = 'Mautrade is a crypto trading dashboard for active layers, exchange bindings, gas fee tracking, and trading history.'
const theme = useState('dashboard-theme', () => 'dark')

const applyTheme = (value) => {
  document.documentElement.dataset.theme = value
}

onMounted(() => {
  const savedTheme = localStorage.getItem('dashboard-theme')

  if (savedTheme === 'light' || savedTheme === 'dark') {
    theme.value = savedTheme
  }

  applyTheme(theme.value)

  watch(theme, (value) => {
    applyTheme(value)
    localStorage.setItem('dashboard-theme', value)
  })
})

useHead({
  htmlAttrs: { lang: 'en' },
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
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
