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

  document.querySelectorAll('.stories .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section
    id="stories"
    class="stories"
  >
    <div class="container reveal">
      <div class="eyebrow">
        Client stories — 010
      </div>
      <div class="section-head">
        <h2 class="section-head__title">
          Three operators.<br>Three <em>verifiable</em> arcs.
        </h2>
        <p class="section-head__sub">
          Not before-and-after testimonials. These are three people who moved from manual trading or broken copy trading tools onto Mautrade, with results verified against their own trade history. Names anonymized at request. Numbers are real.
        </p>
      </div>
    </div>

    <div class="container">
      <div class="stories__list reveal">
        <!-- Story 1 -->
        <article class="story">
          <div class="story__profile">
            <div class="story__avatar">
              <img
                src="https://picsum.photos/seed/operator-shanghai-private-trader/200/200"
                alt=""
              >
            </div>
            <div class="story__num">
              CASE 01
            </div>
            <div class="story__name">
              Operator<br>K — Shanghai
            </div>
            <div class="story__meta">
              Signal channel operator · 6 months on Mautrade
            </div>
          </div>
          <div class="story__stats">
            <div class="story__metric">
              <span class="story__metric-label">Layers opened</span>
              <span class="story__metric-value">142</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Layers closed</span>
              <span class="story__metric-value">98 / 142</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Realized return</span>
              <span class="story__metric-value">+18.4%<span class="delta">net of Gas Fee</span></span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Gas Fee paid</span>
              <span class="story__metric-value">2,847 USDT</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Loss rebates received</span>
              <span class="story__metric-value">412 USDT</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Reconciliation holds</span>
              <span class="story__metric-value">0</span>
            </div>
          </div>
          <div class="story__body">
            <p class="story__quote">
              I ran a signal channel for two years on tools that averaged every follower's position into one number. The first time someone asked me to sell "just the March Layer," I couldn't answer that question. Mautrade tracks each buy on its own — that's not a nice-to-have, it's the whole product.
            </p>
            <div class="story__attribution">
              Operator K · Signal channel operator · 142 Layers · 6-month arc
            </div>
          </div>
        </article>

        <!-- Story 2 -->
        <article class="story">
          <div class="story__profile">
            <div class="story__avatar">
              <img
                src="https://picsum.photos/seed/operator-singapore-small-fund/200/200"
                alt=""
              >
            </div>
            <div class="story__num">
              CASE 02
            </div>
            <div class="story__name">
              Operator<br>L — Singapore
            </div>
            <div class="story__meta">
              Small fund · 14 months on Mautrade
            </div>
          </div>
          <div class="story__stats">
            <div class="story__metric">
              <span class="story__metric-label">Accounts following</span>
              <span class="story__metric-value">217</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Layers tracked</span>
              <span class="story__metric-value">3,841</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Average dispatch time</span>
              <span class="story__metric-value">3.2s<span class="delta">for 217 accounts</span></span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Reconciliation holds</span>
              <span class="story__metric-value">4 / 3,841</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Average resolution time</span>
              <span class="story__metric-value">&lt; 4 hours</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Net return (fund)</span>
              <span class="story__metric-value">+31.7%</span>
            </div>
          </div>
          <div class="story__body">
            <p class="story__quote">
              We tried building this in-house — six months, two engineers, one Gas Fee bug that cost us eight thousand dollars in a weekend. Mautrade's ledger runs on fixed-point decimals from day one. Every Layer, every fee, reconciles to the cent.
            </p>
            <div class="story__attribution">
              Operator L · Small fund manager · 217 accounts · 14-month arc
            </div>
          </div>
        </article>

        <!-- Story 3 -->
        <article class="story">
          <div class="story__profile">
            <div class="story__avatar">
              <img
                src="https://picsum.photos/seed/operator-dubai-dca-hodler/200/200"
                alt=""
              >
            </div>
            <div class="story__num">
              CASE 03
            </div>
            <div class="story__name">
              Operator<br>M — Dubai
            </div>
            <div class="story__meta">
              Long-term investor · 9 months on Mautrade
            </div>
          </div>
          <div class="story__stats">
            <div class="story__metric">
              <span class="story__metric-label">Bitcoin Layers</span>
              <span class="story__metric-value">38</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Ethereum Layers</span>
              <span class="story__metric-value">31</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Average entry (BTC)</span>
              <span class="story__metric-value">$58,420</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Realized return (closed)</span>
              <span class="story__metric-value">+12.1%<span class="delta">on closed Layers</span></span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Unrealized (open)</span>
              <span class="story__metric-value">+43.6%</span>
            </div>
            <div class="story__metric">
              <span class="story__metric-label">Time spent managing</span>
              <span class="story__metric-value">~12 min / week</span>
            </div>
          </div>
          <div class="story__body">
            <p class="story__quote">
              I don't want to watch the market. I want disciplined DCA into BTC and ETH, with a Gas Fee statement that adds up when I check it against my own exchange history. It always has.
            </p>
            <div class="story__attribution">
              Operator M · Long-term investor · 69 Layers · 9-month arc
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
