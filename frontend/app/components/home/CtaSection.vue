<script setup>
const selectedVolume = ref('lt100k')

function selectRadio(value) {
  selectedVolume.value = value
}

function handleSubmit() {
  alert('Thank you. A detailed walkthrough of Mautrade\'s Layer and Gas Fee mechanics is on its way to your inbox. We\'ll reach out within 72 hours if your use case fits the next cohort.')
}

onMounted(() => {
  const revealIO = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('is-in')
        revealIO.unobserve(entry.target)
      }
    })
  }, { threshold: 0.12, rootMargin: '0px 0px -60px 0px' })

  document.querySelectorAll('.cta .reveal').forEach(el => revealIO.observe(el))
})
</script>

<template>
  <section id="cta" class="cta">
    <div class="container reveal">
      <div class="eyebrow">Begin — 011</div>
      <div class="section-head">
        <h2 class="section-head__title">Request<br><em>early access</em>.</h2>
        <p class="section-head__sub">Mautrade is onboarding in closed cohorts — 220 active accounts today, growing in batches of 25 roughly every three weeks. Submit the form and we'll send a full walkthrough of the Layer and Gas Fee mechanics immediately, then follow up within 72 hours if the next cohort fits your case.</p>
      </div>

      <div class="cta__grid">
        <form class="cta__form" @submit.prevent="handleSubmit">
          <div class="form-row">
            <div class="form-field">
              <label for="name">Full name</label>
              <input id="name" type="text" name="name" placeholder="As it appears on your exchange account" required>
            </div>
            <div class="form-field">
              <label for="email">Email address</label>
              <input id="email" type="email" name="email" placeholder="you@domain.com" required>
            </div>
          </div>

          <div class="form-row">
            <div class="form-field">
              <label for="role">I am a</label>
              <select id="role" name="role">
                <option>Select one</option>
                <option>Follower — copying an admin's signals</option>
                <option>Admin — issuing signals for followers</option>
                <option>Small fund — fewer than 250 accounts</option>
                <option>Institutional — 250+ accounts</option>
                <option>Family office</option>
              </select>
            </div>
            <div class="form-field">
              <label for="exchange">Primary exchange</label>
              <select id="exchange" name="exchange">
                <option>Select one</option>
                <option>Binance</option>
                <option>OKX</option>
                <option>Bybit</option>
                <option>Gate</option>
                <option>Multiple exchanges</option>
              </select>
            </div>
          </div>

          <div class="form-field">
            <label>Approximate monthly volume (USDT)</label>
            <div class="form-radios">
              <label class="form-radio" :class="{ selected: selectedVolume === 'lt100k' }" @click="selectRadio('lt100k')">
                <input type="radio" name="volume" value="lt100k" :checked="selectedVolume === 'lt100k'">
                <span class="dot" />
                <span>Under 100k</span>
              </label>
              <label class="form-radio" :class="{ selected: selectedVolume === '100k-1m' }" @click="selectRadio('100k-1m')">
                <input type="radio" name="volume" value="100k-1m" :checked="selectedVolume === '100k-1m'">
                <span class="dot" />
                <span>100k – 1M</span>
              </label>
              <label class="form-radio" :class="{ selected: selectedVolume === '1m-10m' }" @click="selectRadio('1m-10m')">
                <input type="radio" name="volume" value="1m-10m" :checked="selectedVolume === '1m-10m'">
                <span class="dot" />
                <span>1M – 10M</span>
              </label>
              <label class="form-radio" :class="{ selected: selectedVolume === 'gt10m' }" @click="selectRadio('gt10m')">
                <input type="radio" name="volume" value="gt10m" :checked="selectedVolume === 'gt10m'">
                <span class="dot" />
                <span>Over 10M</span>
              </label>
            </div>
          </div>

          <div class="form-field">
            <label for="notes">What are you trying to do?</label>
            <textarea id="notes" name="notes" rows="3" placeholder="e.g. &quot;I want to mirror a trader's spot signals without giving anyone withdrawal access to my funds&quot; or &quot;I run a signal channel and need every follower's Layer tracked and reconciled correctly&quot;" />
          </div>

          <button type="submit" class="cta__submit">Send walkthrough + request access</button>
        </form>

        <aside class="cta__aside">
          <div class="cta__aside-block">
            <div class="cta__aside-label">What you receive</div>
            <div class="cta__aside-content">
              <strong>A detailed walkthrough of the mechanics</strong>
              Sent to your inbox immediately on submission — covers how a Master Signal becomes a Layer, how the Gas Fee is calculated, how reconciliation protects your account, and what onboarding looks like. About a 10-minute read.
            </div>
          </div>

          <div class="cta__aside-block">
            <div class="cta__aside-label">Current status</div>
            <ul class="cta__list">
              <li>220 active accounts currently on the platform</li>
              <li>Next onboarding: 25 accounts · approximately 3 weeks</li>
              <li>Review time: 72 hours from submission</li>
              <li>No commitment until you connect an exchange</li>
              <li>No subscription — the Gas Fee only applies when a Layer closes</li>
            </ul>
          </div>

          <div class="cta__aside-block">
            <div class="cta__aside-label">Prefer to talk to a person?</div>
            <div class="cta__aside-content">
              <strong>ops@mautrade.io</strong>
              For technical questions about the Layer Ledger, Gas Fee edge cases, or partnership inquiries — or if you'd rather skip the form. We read every message; typical response within 24 hours on business days.
            </div>
          </div>

          <div class="cta__aside-block">
            <div class="cta__aside-label">Not a fit?</div>
            <div class="cta__aside-content">
              If you're looking for automated algorithmic trading, futures copy trading, or a platform that holds your funds, Mautrade isn't the right product — it's spot-only, admin-driven, and never takes custody. We're glad to point you elsewhere. Honesty is cheaper than a bad fit.
            </div>
          </div>
        </aside>
      </div>
    </div>
  </section>
</template>
