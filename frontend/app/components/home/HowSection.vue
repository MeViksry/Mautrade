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

  document.querySelectorAll('.how .reveal').forEach(el => revealIO.observe(el))

  // Wait for GSAP to be available (loaded via CDN)
  const initCards = () => {
    if (typeof window.gsap === 'undefined' || typeof window.ScrollTrigger === 'undefined') {
      setTimeout(initCards, 200)
      return
    }

    const gsap = window.gsap
    const ScrollTrigger = window.ScrollTrigger
    gsap.registerPlugin(ScrollTrigger)

    const cards = gsap.utils.toArray('.cardstack__card')
    const N = cards.length

    if (cards.length > 0 && !window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      cards.forEach((card, i) => {
        gsap.set(card, {
          y: i === 0 ? 0 : window.innerHeight * 0.6,
          opacity: i === 0 ? 1 : 0,
          scale: 1,
          zIndex: i + 1,
          filter: 'brightness(1)'
        })
      })

      ScrollTrigger.create({
        trigger: '#cardstack',
        start: 'top top',
        end: `+=${N * 100}%`,
        pin: '#cardstack-pin',
        pinSpacing: true,
        scrub: 1,
        invalidateOnRefresh: true,
        onUpdate: (self) => {
          const seg = 1 / N
          cards.forEach((card, i) => {
            const entryProgress = gsap.utils.clamp(0, 1, (self.progress - (i - 1) * seg) / seg)
            const exitProgress = i < N - 1 ? gsap.utils.clamp(0, 1, (self.progress - i * seg) / seg) : 0

            if (i === 0) {
              gsap.set(card, {
                y: 0,
                opacity: 1 - exitProgress * 0.4,
                scale: 1 - exitProgress * 0.04,
                filter: `brightness(${1 - exitProgress * 0.35})`,
                zIndex: i + 1
              })
            } else {
              gsap.set(card, {
                y: (1 - entryProgress) * window.innerHeight * 0.6,
                opacity: entryProgress * (1 - exitProgress * 0.4),
                scale: 1 - exitProgress * 0.04,
                filter: `brightness(${1 - exitProgress * 0.35})`,
                zIndex: i + 1
              })
            }
          })
        },
        onLeaveBack: () => {
          cards.forEach((card, i) => {
            gsap.set(card, {
              y: i === 0 ? 0 : window.innerHeight * 0.6,
              opacity: i === 0 ? 1 : 0,
              scale: 1,
              filter: 'brightness(1)'
            })
          })
        }
      })
    }
  }

  initCards()
})
</script>

<template>
  <section id="how" class="how">
    <div class="container how__intro reveal">
      <div class="eyebrow">How it works — 003</div>
      <div class="section-head">
        <h2 class="section-head__title">From one signal<br>to a <em>closed</em> Layer.</h2>
        <p class="section-head__sub">Scroll to walk through the four stages of a single Master Signal: how it reaches every connected account, how each position is tracked separately as a Layer, how the system verifies your real balance before touching it, and how the Gas Fee is calculated. Every stage is built around one principle — your money is calculated exactly and never touched by human hands.</p>
      </div>
    </div>

    <div id="cardstack" class="cardstack">
      <div id="cardstack-pin" class="cardstack__pin">
        <!-- Card 1 -->
        <article class="cardstack__card">
          <div class="cardstack__card-bg">
            <div class="cardstack__card-bg-label">Stage 01 — The Signal</div>
            <img src="https://picsum.photos/seed/mautrade-signal-fanout-dashboard/1200/900" alt="">
          </div>
          <div class="cardstack__card-content">
            <div>
              <div class="cardstack__card-num">STAGE 01 — ONE DASHBOARD, ONE DECISION</div>
              <h3 class="cardstack__card-title">The admin<br>makes <em>one</em> move.</h3>
              <p class="cardstack__card-desc">An admin creates a Master Signal on the Mautrade dashboard — buy or sell, one symbol, one percentage. From that single action, the system builds a personal order for every connected account and dispatches them in parallel.</p>
              <ol class="cardstack__card-steps">
                <li>The admin issues a signal: <strong>"Buy BTC/USDT, 10% of available USDT"</strong></li>
                <li>Mautrade pulls every account with an active, verified Exchange Binding</li>
                <li>For each account, the 10% is calculated against <strong>that account's own spot balance</strong></li>
                <li>$1,000 becomes a $100 order. $10,000 becomes a $1,000 order — same signal, different size</li>
                <li>Every order reaches its exchange in parallel — not queued one after another</li>
                <li>Dispatch to every eligible account completes in under five seconds</li>
              </ol>
            </div>
            <div class="cardstack__card-meta">
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Speed</div>
                <div class="cardstack__card-meta-value accent">Under 5 seconds</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Sizing</div>
                <div class="cardstack__card-meta-value">Your own balance</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Your control</div>
                <div class="cardstack__card-meta-value">Stop anytime</div>
              </div>
            </div>
          </div>
        </article>

        <!-- Card 2 -->
        <article class="cardstack__card">
          <div class="cardstack__card-bg">
            <div class="cardstack__card-bg-label">Stage 02 — The Layer</div>
            <img src="https://picsum.photos/seed/mautrade-layer-isolation-ledger/1200/900" alt="">
          </div>
          <div class="cardstack__card-content">
            <div>
              <div class="cardstack__card-num">STAGE 02 — INDEPENDENT LAYERS</div>
              <h3 class="cardstack__card-title">Every buy opens<br>its <em>own</em> Layer.</h3>
              <p class="cardstack__card-desc">Mautrade doesn't track "your BTC" as one blended number. Every buy signal opens a distinct Layer — its own entry price, its own quantity, its own status — inside a ledger the exchange never sees.</p>
              <ol class="cardstack__card-steps">
                <li>Each buy signal creates a new Layer, numbered in the order it opened</li>
                <li>Layer 1, Layer 2, Layer 3 — same coin, three separately tracked positions</li>
                <li>A sell signal names the Layer explicitly: <strong>"Sell Layer 2, 100%"</strong></li>
                <li>Only Layer 2 closes. Layer 1 and Layer 3 are untouched — no auto-merge, no cross-Layer FIFO</li>
                <li>Sell only part of a Layer and the remainder keeps its original entry price</li>
                <li>This is specific-lot tracking — the same principle real accounting uses for cost basis</li>
              </ol>
            </div>
            <div class="cardstack__card-meta">
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Tracking</div>
                <div class="cardstack__card-meta-value accent">Per Layer</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Method</div>
                <div class="cardstack__card-meta-value">Specific-lot ID</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Cross-Layer effect</div>
                <div class="cardstack__card-meta-value">None</div>
              </div>
            </div>
          </div>
        </article>

        <!-- Card 3 -->
        <article class="cardstack__card">
          <div class="cardstack__card-bg">
            <div class="cardstack__card-bg-label">Stage 03 — The Verification</div>
            <img src="https://picsum.photos/seed/mautrade-reconciliation-check-engine/1200/900" alt="">
          </div>
          <div class="cardstack__card-content">
            <div>
              <div class="cardstack__card-num">STAGE 03 — RECONCILIATION FIRST</div>
              <h3 class="cardstack__card-title">Checked first.<br><em>Then</em> executed.</h3>
              <p class="cardstack__card-desc">Before a single order reaches your exchange, Mautrade reads your actual balance and compares it to what the Layer Ledger expects. Withdraw funds manually, rotate your API key, switch exchanges — the system catches it before it trades on stale information.</p>
              <ol class="cardstack__card-steps">
                <li>The system reads your <strong>live exchange balance</strong> right before the trade</li>
                <li>That balance is compared against the quantity the Layer Ledger has on record</li>
                <li>Match → the order executes exactly as sized</li>
                <li>Mismatch → the order is <strong>held, not guessed</strong>, and you're notified with the reason</li>
                <li>A changed API key or a new exchange connection triggers the same check automatically</li>
                <li>No trade ever executes against a Layer that no longer matches reality</li>
              </ol>
            </div>
            <div class="cardstack__card-meta">
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">When</div>
                <div class="cardstack__card-meta-value accent">Before every trade</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">On mismatch</div>
                <div class="cardstack__card-meta-value">Hold + notify</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Wrong-Layer sells</div>
                <div class="cardstack__card-meta-value">Prevented by design</div>
              </div>
            </div>
          </div>
        </article>

        <!-- Card 4 -->
        <article class="cardstack__card">
          <div class="cardstack__card-bg">
            <div class="cardstack__card-bg-label">Stage 04 — The Fee</div>
            <img src="https://picsum.photos/seed/mautrade-gas-fee-calculation-engine/1200/900" alt="">
          </div>
          <div class="cardstack__card-content">
            <div>
              <div class="cardstack__card-num">STAGE 04 — GAS FEE, TO THE DECIMAL</div>
              <h3 class="cardstack__card-title">Half of every<br><em>outcome</em>, exactly.</h3>
              <p class="cardstack__card-desc">Closing a Layer triggers one calculation: the Gas Fee. Profit is split 50/50 in your favor after the fee; loss is split 50/50 with the platform absorbing half. No subscription, no entry cost — the fee only exists at the moment a Layer closes.</p>
              <ol class="cardstack__card-steps">
                <li>When a Layer closes, realized profit or loss is calculated first</li>
                <li>Profit: <strong>Gas Fee = 50% of the profit</strong>, credited to the platform; you keep the rest</li>
                <li>Loss: <strong>the platform rebates 50% of the loss</strong> back to you</li>
                <li>Every step of that calculation runs on fixed-point decimals — no float, no silent rounding</li>
                <li>The result is written to your Gas Fee ledger immediately and permanently</li>
                <li>You can review the exact math behind every single Gas Fee, on every single Layer</li>
              </ol>
            </div>
            <div class="cardstack__card-meta">
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">On profit</div>
                <div class="cardstack__card-meta-value accent">Gas Fee = 50%</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">On loss</div>
                <div class="cardstack__card-meta-value">50% rebated</div>
              </div>
              <div class="cardstack__card-meta-item">
                <div class="cardstack__card-meta-label">Calculation</div>
                <div class="cardstack__card-meta-value">Fixed-point decimal</div>
              </div>
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
