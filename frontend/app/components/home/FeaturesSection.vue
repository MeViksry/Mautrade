<script setup>
onMounted(() => {
  const revealIO = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add('is-in')
        revealIO.unobserve(entry.target)
      }
    })
  }, { threshold: 0.12, rootMargin: '0px 0px -60px 0px' })

  document.querySelectorAll('.features > .container.reveal').forEach(el => revealIO.observe(el))

  // Staggered grid reveal
  const staggerGrids = document.querySelectorAll('.features__grid')
  staggerGrids.forEach((grid) => {
    const io = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('is-in')
          io.unobserve(entry.target)
        }
      })
    }, { threshold: 0.1 })
    io.observe(grid)
  })
})
</script>

<template>
  <section class="features">
    <div class="container reveal">
      <div class="eyebrow">
        What you get — 006
      </div>
      <div class="section-head">
        <h2 class="section-head__title">
          Nine decisions,<br>not a <em>checklist</em>.
        </h2>
        <p class="section-head__sub">
          Not a feature list padded to look complete. Nine specific problems we solved because we hit them ourselves or watched them break someone else's product.
        </p>
      </div>

      <div class="features__grid">
        <div
          class="feature"
          style="--row:0; --col:0"
        >
          <div class="feature__num">
            01
          </div>
          <div class="feature__name">
            Separate Layer Tracking
          </div>
          <div class="feature__tag">
            Every Layer, its own entry price
          </div>
          <p class="feature__desc">
            Every buy opens a Layer with its own price and quantity, tracked independently in the Layer Ledger. Selling one Layer <strong>never touches another</strong> — a real separate position, not a blended average.
          </p>
        </div>
        <div
          class="feature"
          style="--row:0; --col:1"
        >
          <div class="feature__num">
            02
          </div>
          <div class="feature__name">
            Parallel Signal Dispatch
          </div>
          <div class="feature__tag">
            Under 5 seconds, any account count
          </div>
          <p class="feature__desc">
            A Master Signal reaches every eligible account at once, sized to that account's own balance, respecting each exchange's own rate limits. <strong>10,000 accounts, same five seconds.</strong>
          </p>
        </div>
        <div
          class="feature"
          style="--row:0; --col:2"
        >
          <div class="feature__num">
            03
          </div>
          <div class="feature__name">
            Pre-Trade Reconciliation
          </div>
          <div class="feature__tag">
            Balance checked before every order
          </div>
          <p class="feature__desc">
            Before execution, Mautrade compares your live exchange balance to the Layer Ledger. A manual withdrawal or a rotated key <strong>pauses the trade instead of guessing</strong>.
          </p>
        </div>
        <div
          class="feature"
          style="--row:1; --col:0"
        >
          <div class="feature__num">
            04
          </div>
          <div class="feature__name">
            Idempotent Execution
          </div>
          <div class="feature__tag">
            One trade, one identifier
          </div>
          <p class="feature__desc">
            Every order carries a unique key tied to its signal, account, and Layer. A retried or duplicated network call is <strong>rejected before it ever reaches the exchange twice</strong>.
          </p>
        </div>
        <div
          class="feature"
          style="--row:1; --col:1"
        >
          <div class="feature__num">
            05
          </div>
          <div class="feature__name">
            Multi-Exchange Binding
          </div>
          <div class="feature__tag">
            Binance · OKX · Bybit · Gate
          </div>
          <p class="feature__desc">
            Connect the exchange you already use. Each one runs through its own adapter, so a new exchange integration <strong>never touches how existing accounts trade</strong>.
          </p>
        </div>
        <div
          class="feature"
          style="--row:1; --col:2"
        >
          <div class="feature__num">
            06
          </div>
          <div class="feature__name">
            Real-Time Admin Dashboard
          </div>
          <div class="feature__tag">
            Nuxt-based, live signal monitor
          </div>
          <p class="feature__desc">
            Every Master Signal shows its dispatch progress live — accounts succeeded, skipped, or failed — with a one-click retry for anything that didn't go through. <strong>No refresh needed.</strong>
          </p>
        </div>
        <div
          class="feature"
          style="--row:2; --col:0"
        >
          <div class="feature__num">
            07
          </div>
          <div class="feature__name">
            Mobile App for Followers
          </div>
          <div class="feature__tag">
            iOS + Android · biometric login
          </div>
          <p class="feature__desc">
            Layer positions, Gas Fee history, and trade records — from your phone. <strong>Your API key never touches the device</strong>; it stays encrypted on the server. Dark mode and push notifications included.
          </p>
        </div>
        <div
          class="feature"
          style="--row:2; --col:1"
        >
          <div class="feature__num">
            08
          </div>
          <div class="feature__name">
            Immutable Audit Trail
          </div>
          <div class="feature__tag">
            Every trade, every Gas Fee, logged
          </div>
          <p class="feature__desc">
            Every action on a Layer — opened, partially closed, closed, or flagged — is recorded permanently with who triggered it and when. <strong>Nothing can be deleted or silently changed.</strong>
          </p>
        </div>
        <div
          class="feature"
          style="--row:2; --col:2"
        >
          <div class="feature__num">
            09
          </div>
          <div class="feature__name">
            Trade-Only Enforcement
          </div>
          <div class="feature__tag">
            Withdrawal rights rejected at setup
          </div>
          <p class="feature__desc">
            Every API key is checked for withdrawal permission the moment it's connected. Keys with withdrawal rights are <strong>refused, no exceptions</strong>.
          </p>
        </div>
      </div>
    </div>
  </section>
</template>
