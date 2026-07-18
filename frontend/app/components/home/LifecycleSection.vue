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

  document.querySelectorAll('.lifecycle .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section class="lifecycle">
    <div class="container reveal">
      <div class="eyebrow">The lifecycle of a trade — 007</div>
      <div class="section-head">
        <h2 class="section-head__title">What happens<br>from <em>signal</em> to notification.</h2>
        <p class="section-head__sub">A Master Signal is one action by one admin. Here's everything that happens between that action and the notification landing on your phone. Every step is automated, observable, and protected against failure.</p>
      </div>

      <div class="lifecycle__diagram">
        <div class="lifecycle__title">The lifecycle of a single trade</div>
        <div class="lifecycle__flow">
          <div class="lifecycle__node">
            <div class="lifecycle__node-num">01</div>
            <div class="lifecycle__node-label">The signal</div>
            <div class="lifecycle__node-action">An admin creates a Master Signal on the dashboard — for example, <strong>"Buy BTC/USDT, 10% of available USDT"</strong> — and confirms it with 2FA before it goes anywhere.</div>
          </div>
          <div class="lifecycle__arrow">↓</div>
          <div class="lifecycle__node">
            <div class="lifecycle__node-num">02</div>
            <div class="lifecycle__node-label">Eligibility check</div>
            <div class="lifecycle__node-action">The system pulls every account with an <strong>active, verified Exchange Binding</strong> and enough balance to participate. Anyone who doesn't qualify is skipped, with the reason recorded.</div>
          </div>
          <div class="lifecycle__arrow">↓</div>
          <div class="lifecycle__node">
            <div class="lifecycle__node-num">03</div>
            <div class="lifecycle__node-label">Sizing</div>
            <div class="lifecycle__node-action">Each eligible account gets its own order size, calculated against <strong>that account's own spot balance</strong> at that moment — not a fixed amount, not the admin's balance.</div>
          </div>
          <div class="lifecycle__arrow">↓</div>
          <div class="lifecycle__node">
            <div class="lifecycle__node-num">04</div>
            <div class="lifecycle__node-label">Reconciliation</div>
            <div class="lifecycle__node-action">Before dispatch, the account's live exchange balance is checked against what the Layer Ledger expects. A mismatch — a withdrawal, a new API key — <strong>holds that one order</strong> and notifies the account.</div>
          </div>
          <div class="lifecycle__arrow">↓</div>
          <div class="lifecycle__node">
            <div class="lifecycle__node-num">05</div>
            <div class="lifecycle__node-label">Execution</div>
            <div class="lifecycle__node-action">The order reaches the exchange through the account's own encrypted API key. The system waits for confirmation and records the <strong>actual fill price and quantity</strong> — not an estimate.</div>
          </div>
          <div class="lifecycle__arrow">↓</div>
          <div class="lifecycle__node">
            <div class="lifecycle__node-num">06</div>
            <div class="lifecycle__node-label">Recording</div>
            <div class="lifecycle__node-action">The confirmed fill opens or closes a Layer in the ledger. If it's a sell, the Gas Fee is calculated in the same step — <strong>entirely to eighteen decimal places.</strong></div>
          </div>
          <div class="lifecycle__arrow">↓</div>
          <div class="lifecycle__node">
            <div class="lifecycle__node-num">07</div>
            <div class="lifecycle__node-label">Notification</div>
            <div class="lifecycle__node-action">You receive a push notification: <strong>"Layer 4 BTC/USDT opened at $67,432.18 · quantity 0.0148 · 10% of your balance was used."</strong> From the admin's confirmation to this notification: under five seconds.</div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
