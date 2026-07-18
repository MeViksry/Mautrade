<script setup>
const isScrolled = ref(false)

onMounted(() => {
  const handleScroll = () => {
    isScrolled.value = window.scrollY > 80
  }
  window.addEventListener('scroll', handleScroll, { passive: true })

  onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll)
  })
})
</script>

<template>
  <nav
    id="nav"
    class="nav"
    :class="{ 'nav--scrolled': isScrolled }"
  >
    <a
      href="#"
      class="nav__logo"
    >MAU<span class="dot" />TRADE</a>
    <ul class="nav__links">
      <li><a href="#explainer">What this is</a></li>
      <li><a href="#how">How it works</a></li>
      <li><a href="#fees">Gas Fee</a></li>
      <li><a href="#safety">Safety</a></li>
      <li><a href="#stories">Clients</a></li>
    </ul>
    <a
      href="#cta"
      class="nav__cta"
    >Request access</a>
  </nav>
</template>
