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

  document.querySelectorAll('.thesis .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section class="thesis">
    <div class="container reveal">
      <div class="eyebrow">Why we built this — 002</div>
      <div class="section-head">
        <h2 class="section-head__title">Copy trading breaks<br>at the <em>decimal point</em>.</h2>
        <p class="section-head__sub">We rebuilt copy trading from the ledger up, because every tool we tried made the same three mistakes: it rounded money to fit a float, it averaged your positions instead of isolating them, and it broke the moment your account changed. Mautrade fixes all three by design, not by patch.</p>
      </div>

      <div class="thesis__grid">
        <div class="thesis__col">
          <div class="thesis__col-num">01 / Accuracy</div>
          <h3>Numbers that reconcile. Always.</h3>
          <p>A standard floating-point number can't represent 0.1 exactly — that's not a bug, it's how the format works. Run that through thousands of Gas Fee calculations and the errors compound quietly, in someone's favor. Usually not yours.</p>
          <p>Mautrade's ledger runs on fixed-point decimal arithmetic end to end — Golang and Rust, no floating-point math anywhere near your money. Every entry price, quantity, and Gas Fee is stored to the full precision it was calculated at. What you see is what was actually computed.</p>
        </div>
        <div class="thesis__col">
          <div class="thesis__col-num">02 / Isolation</div>
          <h3>One buy, one Layer. Never a blend.</h3>
          <p>On the exchange, your BTC is one number in one wallet. Most copy trading tools inherit that and average your cost basis across every buy — which makes "sell 50% of Layer 2" a question they can't answer correctly.</p>
          <p>Mautrade keeps its own Layer Ledger, independent of the exchange. Layer 3 has its own entry price and quantity from the moment it opens until the moment it closes — sold in full, sold in part, or left untouched.</p>
        </div>
        <div class="thesis__col">
          <div class="thesis__col-num">03 / Resilience</div>
          <h3>Built for the account that changes.</h3>
          <p>You rotate an API key. You withdraw USDT to cover an expense. You move from one exchange to another mid-week. Every one of those is normal — and every one of those breaks a system that assumes your balance never moves outside it.</p>
          <p>Before every single trade, Mautrade checks your actual exchange balance against what the Layer Ledger expects. A mismatch pauses that trade and notifies you. It never guesses, and it never sells the wrong Layer.</p>
        </div>
      </div>

      <blockquote class="thesis__quote reveal">
        "Profit is not a promise. It is a number — <em>calculated to the decimal, reconciled before every trade, executed in seconds.</em> Everything else is marketing."
      </blockquote>
    </div>
  </section>
</template>
