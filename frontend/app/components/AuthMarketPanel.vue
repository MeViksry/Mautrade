<template>
  <aside class="auth-market">
    <div class="auth-market__status">
      <span />
      Mautrade Account Flow
    </div>

    <div class="auth-market__metrics">
      <div class="metric-box">
        <strong>6</strong>
        <span>Overview Layers</span>
        <em>Active layer preview</em>
      </div>
      <div class="metric-box">
        <strong>4</strong>
        <span>Exchange Bindings</span>
        <em>Binance / OKX / Bybit / Tokocrypto</em>
      </div>
      <div class="metric-box metric-box--glow">
        <strong>500</strong>
        <span>Minimum Gas Fee</span>
        <em>USDT required before dashboard</em>
      </div>
      <div class="metric-box">
        <strong>OTP</strong>
        <span>Email Verification</span>
        <em>Required after register</em>
      </div>
    </div>

    <div class="auth-market__quote">
      <div class="quote-mark">
        MAUTRADE FLOW
      </div>
      <p>Register with email, verify OTP, complete onboarding, then enter the user dashboard with country, exchange, and gas fee data already prepared.</p>
      <div class="quote-user">
        <div>ID</div>
        <span>
          <strong>User Account</strong>
          <small>Ready for dashboard sync</small>
        </span>
      </div>
    </div>

    <div class="auth-market__headline">
      <h2>Mautrade dashboard.<br><span>Ready before trading.</span></h2>
      <p>Onboarding data feeds Active Layers, Exchange Bindings, Gas Fee, and Trading History so the account starts from the right profile.</p>
    </div>
  </aside>
</template>

<style scoped>
.auth-market {
  position: relative;
  overflow: hidden;
  min-height: 100vh;
  padding: 3rem;
  border-left: 1px solid var(--line);
  background:
    linear-gradient(180deg, rgba(255, 90, 0, 0.1), transparent 38%),
    var(--bg-elevated);
}

.auth-market::before {
  content: "";
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.035) 1px, transparent 1px);
  background-size: 52px 52px;
  mask-image: linear-gradient(to bottom, #000, transparent 88%);
  pointer-events: none;
}

.auth-market > * {
  position: relative;
}

.auth-market__status {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 4.5rem;
  font-family: var(--mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.14em;
  color: var(--accent);
  text-transform: uppercase;
}

.auth-market__status span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
}

.auth-market__metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.metric-box {
  min-height: 118px;
  padding: 1.4rem;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--charcoal);
}

.metric-box--glow {
  background: rgba(255, 90, 0, 0.12);
  border-color: rgba(255, 90, 0, 0.35);
}

.metric-box strong {
  display: block;
  font-family: 'Oswald', sans-serif;
  font-size: 2rem;
  line-height: 1;
  color: var(--text);
}

.metric-box span {
  display: block;
  margin-top: 0.5rem;
  color: var(--text-mute);
}

.metric-box em {
  display: block;
  margin-top: 0.35rem;
  font-family: var(--mono);
  font-size: 11px;
  font-style: normal;
  font-weight: 700;
  color: var(--accent);
}

.auth-market__quote {
  margin-top: 3rem;
  padding: 2rem;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--charcoal);
}

.quote-mark {
  margin-bottom: 1rem;
  color: var(--accent);
  font-family: var(--mono);
  font-weight: 700;
}

.auth-market__quote p {
  color: var(--text);
  font-size: 1.05rem;
  line-height: 1.7;
}

.quote-user {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 1.5rem;
}

.quote-user div {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: var(--accent);
  color: #00130b;
  font-family: var(--mono);
  font-weight: 700;
}

.quote-user span {
  display: flex;
  flex-direction: column;
}

.quote-user strong {
  color: var(--text);
}

.quote-user small {
  color: var(--text-mute);
}

.auth-market__headline {
  margin-top: 4rem;
}

.auth-market__headline h2 {
  margin: 0;
  color: var(--text);
  font-size: clamp(2.1rem, 4vw, 3.5rem);
  line-height: 1.05;
}

.auth-market__headline span {
  color: var(--accent);
}

.auth-market__headline p {
  max-width: 540px;
  margin-top: 1rem;
  color: var(--text-mute);
}

@media (max-width: 980px) {
  .auth-market {
    min-height: auto;
    padding: 2rem;
  }
}

@media (max-width: 640px) {
  .auth-market__metrics {
    grid-template-columns: 1fr;
  }
}
</style>
