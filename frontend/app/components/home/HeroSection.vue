<script setup>
onMounted(() => {
  const revealIO = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('is-in')
        revealIO.unobserve(entry.target)
      }
    })
  }, { threshold: 0.12, rootMargin: '0px 0px -60px 0px' })

  document.querySelectorAll('.hero .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section class="hero">
    <div class="hero__bg">
      <img src="https://picsum.photos/seed/mautrade-trading-floor-monitors-dark/1920/1200" alt="">
    </div>

    <div class="hero__topbar reveal">
      <div class="hero__topbar-item"><span class="pulse" /> Live — Signal #4,812 dispatching</div>
      <div class="hero__topbar-item">4 exchanges connected</div>
      <div class="hero__topbar-item">47,219 Layers tracked</div>
    </div>

    <div class="hero__grid">
      <div class="hero__main">
        <div class="eyebrow reveal">Spot Copy Trading — DCA Layering</div>
        <h1 class="hero__title reveal reveal--d1">
          <span>ONE SIGNAL.</span>
          <span>EVERY LAYER.</span>
          <span class="accent">zero rounding.</span>
        </h1>
        <p class="hero__sub reveal reveal--d2">
          Mautrade mirrors a professional trader's spot orders into your own exchange account — sized to your balance, not theirs. Every buy opens its own Layer, tracked separately down to the entry price and quantity. Sell one Layer and only that Layer closes; the rest stay exactly where they were. You keep 50% of every realized profit, and the platform absorbs 50% of every realized loss. Your API key can place orders — it can never withdraw.
        </p>
        <div class="hero__ctas reveal reveal--d3">
          <a href="#cta" class="btn-primary">Request early access</a>
          <a href="#explainer" class="btn-ghost">See how a Layer works</a>
        </div>
      </div>

      <aside class="hero__feed reveal reveal--d2">
        <div class="hero__feed-label">
          <span>Recent Layer activity</span>
          <span class="live">LIVE</span>
        </div>
        <a href="#how" class="hero__signal">
          <span class="hero__signal-type buy">BUY</span>
          <div class="hero__signal-body">
            <span class="hero__signal-pair">BTC / USDT</span>
            <span class="hero__signal-meta">10% of balance · 8,432 accounts</span>
          </div>
          <span class="hero__signal-time">12s</span>
        </a>
        <a href="#how" class="hero__signal">
          <span class="hero__signal-type sell">SELL</span>
          <div class="hero__signal-body">
            <span class="hero__signal-pair">ETH / USDT</span>
            <span class="hero__signal-meta">Layer #2 · 50% sold · 6,108 accounts</span>
          </div>
          <span class="hero__signal-time">4m</span>
        </a>
        <a href="#how" class="hero__signal">
          <span class="hero__signal-type buy">BUY</span>
          <div class="hero__signal-body">
            <span class="hero__signal-pair">SOL / USDT</span>
            <span class="hero__signal-meta">5% of balance · 7,891 accounts</span>
          </div>
          <span class="hero__signal-time">11m</span>
        </a>
        <a href="#how" class="hero__signal">
          <span class="hero__signal-type sell">SELL</span>
          <div class="hero__signal-body">
            <span class="hero__signal-pair">BTC / USDT</span>
            <span class="hero__signal-meta">Layer #3 · 100% sold · 8,401 accounts</span>
          </div>
          <span class="hero__signal-time">28m</span>
        </a>
        <a href="#how" class="hero__signal">
          <span class="hero__signal-type buy">BUY</span>
          <div class="hero__signal-body">
            <span class="hero__signal-pair">LINK / USDT</span>
            <span class="hero__signal-meta">8% of balance · 5,742 accounts</span>
          </div>
          <span class="hero__signal-time">41m</span>
        </a>
      </aside>
    </div>

    <div class="hero__stats reveal reveal--d4">
      <div class="hero__stat">
        <div class="hero__stat-num">&lt;5<span class="unit">sec</span></div>
        <div class="hero__stat-label">From signal to your account</div>
      </div>
      <div class="hero__stat">
        <div class="hero__stat-num">18<span class="unit">digit</span></div>
        <div class="hero__stat-label">Precision on every Gas Fee</div>
      </div>
      <div class="hero__stat">
        <div class="hero__stat-num">50<span class="unit">/50</span></div>
        <div class="hero__stat-label">Gas Fee split — profit &amp; loss</div>
      </div>
      <div class="hero__stat">
        <div class="hero__stat-num">0</div>
        <div class="hero__stat-label">Withdrawal rights on your API key</div>
      </div>
    </div>
  </section>
</template>
