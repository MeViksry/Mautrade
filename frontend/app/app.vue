<script setup>
const title = 'Mautrade — Spot DCA Layering Copy Trading. Built to the Decimal.'
const description = 'Every Master Signal opens a new Layer. Every sell targets one Layer, and one Layer only. Every Gas Fee is split exactly 50/50. Copy trading built like financial infrastructure, not a trading bot.'

useHead({
  htmlAttrs: { lang: 'en' },
  link: [
    { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
    { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Oswald:wght@200;300;400;500;700&family=Roboto:wght@100;300;400;500&family=JetBrains+Mono:wght@400;500&display=swap' }
  ],
  script: [
    { src: 'https://cdn.jsdelivr.net/npm/gsap@3.12.5/dist/gsap.min.js', defer: true },
    { src: 'https://cdn.jsdelivr.net/npm/gsap@3.12.5/dist/ScrollTrigger.min.js', defer: true },
    { src: 'https://cdn.jsdelivr.net/npm/lenis@1.3.13/dist/lenis.min.js', defer: true }
  ]
})

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  twitterCard: 'summary_large_image'
})

onMounted(() => {
  const initGlobalScripts = () => {
    if (typeof window.Lenis === 'undefined' || typeof window.gsap === 'undefined' || typeof window.ScrollTrigger === 'undefined') {
      setTimeout(initGlobalScripts, 200)
      return
    }

    /* ============ LENIS SMOOTH SCROLL ============ */
    const lenis = new window.Lenis({
      duration: 1.1,
      easing: t => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
      smoothWheel: true,
      smoothTouch: false
    })

    function raf(time) {
      lenis.raf(time)
      requestAnimationFrame(raf)
    }
    requestAnimationFrame(raf)

    /* ============ GSAP / SCROLLTRIGGER ============ */
    const gsap = window.gsap
    const ScrollTrigger = window.ScrollTrigger

    gsap.registerPlugin(ScrollTrigger)
    lenis.on('scroll', ScrollTrigger.update)
    gsap.ticker.add((time) => {
      lenis.raf(time * 1000)
    })
    gsap.ticker.lagSmoothing(0)

    /* ============ SMOOTH ANCHOR LINKS ============ */
    document.querySelectorAll('a[href^="#"]').forEach((link) => {
      link.addEventListener('click', (e) => {
        const target = document.querySelector(link.getAttribute('href'))
        if (target) {
          e.preventDefault()
          lenis.scrollTo(target, { offset: -40, duration: 1.4 })
        }
      })
    })

    /* ============ RESIZE HANDLER ============ */
    let resizeTimer
    window.addEventListener('resize', () => {
      clearTimeout(resizeTimer)
      resizeTimer = setTimeout(() => {
        ScrollTrigger.refresh()
      }, 200)
    })

    window.lenis = lenis
  }

  initGlobalScripts()
})
</script>

<template>
  <div>
    <div class="noise-overlay" />
    <Navbar />
    <NuxtPage />
    <Footer />
  </div>
</template>
