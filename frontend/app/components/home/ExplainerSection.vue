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

  document.querySelectorAll('.explainer .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section
    id="explainer"
    class="explainer"
  >
    <div class="container reveal">
      <div class="eyebrow">
        What this is — 001
      </div>
      <div class="section-head">
        <h2 class="section-head__title">
          DCA Layering,<br>in plain <em>English</em>.
        </h2>
        <p class="section-head__sub">
          If you're new to Mautrade, here's the whole mechanism in four steps — no jargon left unexplained. If you already know how it works, scroll on to the methodology section.
        </p>
      </div>

      <div class="explainer__grid">
        <div class="explainer__step">
          <div class="explainer__step-num">
            STEP 01
          </div>
          <h3 class="explainer__step-title">
            The admin signals
          </h3>
          <p class="explainer__step-text">
            Inside the Mautrade dashboard, an admin buys or sells a spot pair — for example, <strong>10% of available USDT into BTC/USDT</strong>. That single action is the entire input. Nothing about your account changes until it fires.
          </p>
        </div>
        <div class="explainer__step">
          <div class="explainer__step-num">
            STEP 02
          </div>
          <h3 class="explainer__step-title">
            Sized to you, not them
          </h3>
          <p class="explainer__step-text">
            That 10% is calculated against <strong>your own spot balance</strong> at the moment the signal fires — not a fixed amount, not the admin's balance. $1,000 becomes a $100 order. $10,000 becomes $1,000. You choose whether to stay connected at all.
          </p>
        </div>
        <div class="explainer__step">
          <div class="explainer__step-num">
            STEP 03
          </div>
          <h3 class="explainer__step-title">
            A new Layer, every buy
          </h3>
          <p class="explainer__step-text">
            Each buy opens a <strong>Layer</strong> — a position tracked on its own, with its own entry price and quantity. Buy BTC/USDT three times and you hold three Layers. When the admin sells, they name the Layer — Layer 2, 100% — and only that one closes.
          </p>
        </div>
        <div class="explainer__step">
          <div class="explainer__step-num">
            STEP 04
          </div>
          <h3 class="explainer__step-title">
            Gas Fee, split in half
          </h3>
          <p class="explainer__step-text">
            No subscription. No fee upfront. When a Layer closes in profit, the <strong>Gas Fee is 50%</strong> of that profit — you keep the rest. When a Layer closes at a loss, the platform <strong>rebates 50%</strong> of it back to you. The fee only exists the moment a Layer closes.
          </p>
        </div>
      </div>

      <div class="explainer__analogy reveal">
        <div class="explainer__analogy-icon">
          "
        </div>
        <div>
          <div class="explainer__analogy-label">
            The simple version
          </div>
          <p class="explainer__analogy-text">
            Think of DCA Layering like a stack of separate purchase receipts for the same coin. Selling one receipt <em>never touches the others</em> — even though the coin itself sits in one wallet on the exchange.
          </p>
        </div>
      </div>
    </div>
  </section>
</template>
