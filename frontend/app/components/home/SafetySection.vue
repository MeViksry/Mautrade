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

  document.querySelectorAll('.safety .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section id="safety" class="safety">
    <div class="container reveal">
      <div class="eyebrow">Safety — 005</div>
      <div class="section-head">
        <h2 class="section-head__title">An API key<br>that can <em>trade</em>. Nothing else.</h2>
        <p class="section-head__sub">Connecting Mautrade to your exchange account means handing over an API key. We built the platform around one rule for that key: it can place orders, and it can do nothing else. Here's exactly what that means in practice — and what we've deliberately left out, even when it would be convenient.</p>
      </div>

      <div class="safety__grid">
        <div class="safety__col">
          <h3>What we do</h3>
          <p>Every API key is encrypted the moment it reaches Mautrade, decrypted only in memory for the instant it takes to place an order, and never written to a log, an error message, or a support ticket.</p>
          <ul class="safety__list">
            <li><strong>Trade-only, verified at connection</strong>Every key you connect is checked for withdrawal permission before it's accepted. A key that can withdraw is rejected outright — not flagged and allowed through anyway.</li>
            <li><strong>Two-factor on every admin account</strong>Confirming a Master Signal requires 2FA. Admin sessions expire after inactivity, and a login from a new device forces reauthentication and an alert.</li>
            <li><strong>A Layer Ledger that can't be edited quietly</strong>Every trade, every Gas Fee, every status change on a Layer is written once and kept permanently. There's no update path that erases what happened before.</li>
            <li><strong>One idempotency key per trade</strong>Every order carries a unique identifier tied to its signal, account, and Layer. If a retry sends the same order twice, the duplicate is rejected before it reaches the exchange.</li>
          </ul>
        </div>

        <div class="safety__col">
          <h3>What we refuse to do</h3>
          <p>Some of these would make the platform look more capable. We left them out anyway, because each one asks you to trust something we'd rather not depend on. These aren't settings — they're decisions built into the architecture.</p>
          <ul class="safety__list">
            <li><strong>No withdrawal access, under any setting</strong>There is no configuration, no support override, no future update that accepts a key with withdrawal rights. It's rejected at connection, permanently.</li>
            <li><strong>No custody of your funds</strong>Orders execute directly on your own exchange account, through your own key. Mautrade holds no wallet and no balance of yours — there's nothing here for us to lose or freeze.</li>
            <li><strong>No signal without a person behind it</strong>Every Master Signal is a deliberate decision made by an admin, not an algorithm reacting to price. If no one decided to trade, nothing trades.</li>
            <li><strong>No futures, no margin, no leverage</strong>Spot only. Leverage can produce losses the 50% rebate wasn't built to absorb, so it isn't offered.</li>
          </ul>
        </div>
      </div>
    </div>
  </section>
</template>
