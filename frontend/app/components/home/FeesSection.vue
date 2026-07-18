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

  document.querySelectorAll('.fees .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section id="fees" class="fees">
    <div class="container reveal">
      <div class="eyebrow">Gas Fee — 004</div>
      <div class="section-head">
        <h2 class="section-head__title">One fee.<br>Two <em>directions</em>.</h2>
        <p class="section-head__sub">Mautrade doesn't charge a subscription or an entry fee. It charges a Gas Fee, and only when a Layer closes — 50% of the profit if you're up, 50% of the loss rebated back if you're down. Same formula, run to full decimal precision, every single time.</p>
      </div>

      <div class="fees__grid">
        <div class="fees__case">
          <div class="fees__case-label">Case 01 — Layer closes in profit</div>
          <h3 class="fees__case-title">Profit Share</h3>
          <p class="fees__case-desc">A Layer you held closes above its entry cost. <strong>The Gas Fee is 50% of that profit</strong> — the rest is yours. The platform only earns a Gas Fee when your Layer does too.</p>
          <div class="formula-block"><span class="label">Realized profit on the Layer</span>  <span class="val">50.00 USDT</span>
<span class="label">Gas Fee (50%)</span>               <span class="val">25.00 USDT</span>
<span class="label">Net profit credited to you</span>   <span class="result">+25.00 USDT</span>

<span class="note">// Calculated on the exact entry price recorded for this Layer</span>
<span class="note">// Written to the Gas Fee ledger the moment the Layer closes</span></div>
          <div class="fees__result">
            <div class="fees__result-item">
              <div class="fees__result-label">You receive</div>
              <div class="fees__result-value accent">+25.00 USDT</div>
            </div>
            <div class="fees__result-item">
              <div class="fees__result-label">Gas Fee (platform)</div>
              <div class="fees__result-value">+25.00 USDT</div>
            </div>
          </div>
        </div>

        <div class="fees__case">
          <div class="fees__case-label">Case 02 — Layer closes at a loss</div>
          <h3 class="fees__case-title">Loss Rebate</h3>
          <p class="fees__case-desc">A Layer closes below its entry cost. <strong>The platform absorbs 50% of that loss</strong> as a rebate credited straight to your account. It's part of the Gas Fee formula, not a promotional exception.</p>
          <div class="formula-block"><span class="label">Realized loss on the Layer</span>    <span class="val">50.00 USDT</span>
<span class="label">Loss rebate (50%)</span>            <span class="val">25.00 USDT</span>
<span class="label">Net loss you absorb</span>           <span class="result">-25.00 USDT</span>

<span class="note">// You never absorb more than half of any realized loss</span>
<span class="note">// Rebate applied to the same Layer, same moment it closes</span></div>
          <div class="fees__result">
            <div class="fees__result-item">
              <div class="fees__result-label">You absorb</div>
              <div class="fees__result-value accent">−25.00 USDT</div>
            </div>
            <div class="fees__result-item">
              <div class="fees__result-label">Platform absorbs</div>
              <div class="fees__result-value">−25.00 USDT</div>
            </div>
          </div>
        </div>
      </div>

      <div class="fees__precision reveal">
        <div class="fees__precision-icon">18</div>
        <div class="fees__precision-content">
          <h4>Eighteen decimal places. Every time.</h4>
          <p>
            Entry price, quantity, realized profit, Gas Fee — every one of those values is <strong>calculated and stored to eighteen decimal places</strong> inside the Layer Ledger, using fixed-point decimal math end to end. Nothing is rounded until it reaches your screen, and it's rounded the same way everywhere it appears. The number on your Layer detail page is the number that was actually computed — not an approximation, not "close enough."
          </p>
        </div>
      </div>
    </div>
  </section>
</template>
