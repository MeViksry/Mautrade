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

  document.querySelectorAll('.performance .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section class="performance">
    <div class="container reveal">
      <div class="eyebrow">What we commit to — 008</div>
      <div class="section-head">
        <h2 class="section-head__title">Numbers<br>we <em>stand behind</em>.</h2>
        <p class="section-head__sub">Not aspirational marketing copy. These are the standards the system is built and tested against every week. If we miss one in production, we treat it as a defect to fix — not a target to renegotiate.</p>
      </div>

      <div class="performance__grid">
        <div class="perf-stat">
          <div class="perf-stat__num">&lt;5<span class="unit">s</span></div>
          <div class="perf-stat__label">Signal to execution</div>
          <p class="perf-stat__desc">From an admin's confirmed Master Signal to the last eligible account's order reaching its exchange — under five seconds, tested weekly at full account load.</p>
        </div>
        <div class="perf-stat">
          <div class="perf-stat__num">18<span class="unit">dp</span></div>
          <div class="perf-stat__label">Decimal precision</div>
          <p class="perf-stat__desc">Entry price, quantity, realized profit, Gas Fee — every value calculated and stored to eighteen decimal places, with fixed-point math from the Layer Ledger to your screen.</p>
        </div>
        <div class="perf-stat">
          <div class="perf-stat__num">100<span class="unit">%</span></div>
          <div class="perf-stat__label">Duplicate-trade prevention</div>
          <p class="perf-stat__desc">Every order carries a unique identifier tied to its signal, account, and Layer. Retries and network glitches are rejected before they can execute the same trade twice.</p>
        </div>
        <div class="perf-stat">
          <div class="perf-stat__num">0</div>
          <div class="perf-stat__label">Withdrawal permission</div>
          <p class="perf-stat__desc">Every connected API key is checked for withdrawal rights before it's accepted. None get through — not by mistake, not by request.</p>
        </div>
      </div>
    </div>
  </section>
</template>
