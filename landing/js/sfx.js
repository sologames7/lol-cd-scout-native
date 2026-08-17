let ctx = null;

function ac() {
  if (!ctx) ctx = new (window.AudioContext || window.webkitAudioContext)();
  if (ctx.state === 'suspended') ctx.resume();
  return ctx;
}

export function unlockSfx() {
  ac();
}

function env(gain, t, a, d, peak = 0.4) {
  gain.gain.cancelScheduledValues(t);
  gain.gain.setValueAtTime(0.0001, t);
  gain.gain.exponentialRampToValueAtTime(peak, t + a);
  gain.gain.exponentialRampToValueAtTime(0.0001, t + a + d);
}

export function playKnife() {
  const a = ac();
  const t = a.currentTime;
  const n = a.createBuffer(1, a.sampleRate * 0.18, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * (1 - i / d.length);
  const src = a.createBufferSource();
  src.buffer = n;
  const bp = a.createBiquadFilter();
  bp.type = 'bandpass';
  bp.frequency.value = 1800;
  const g = a.createGain();
  env(g, t, 0.01, 0.16, 0.35);
  src.connect(bp).connect(g).connect(a.destination);
  src.start(t);

  const osc = a.createOscillator();
  osc.type = 'sawtooth';
  osc.frequency.setValueAtTime(420, t);
  osc.frequency.exponentialRampToValueAtTime(140, t + 0.12);
  const g2 = a.createGain();
  env(g2, t, 0.005, 0.12, 0.12);
  osc.connect(g2).connect(a.destination);
  osc.start(t);
  osc.stop(t + 0.14);
}

export function playGun() {
  const a = ac();
  const t = a.currentTime;
  const n = a.createBuffer(1, a.sampleRate * 0.28, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / d.length, 1.4);
  const src = a.createBufferSource();
  src.buffer = n;
  const lp = a.createBiquadFilter();
  lp.type = 'lowpass';
  lp.frequency.value = 900;
  const g = a.createGain();
  env(g, t, 0.002, 0.22, 0.55);
  src.connect(lp).connect(g).connect(a.destination);
  src.start(t);

  const osc = a.createOscillator();
  osc.type = 'sine';
  osc.frequency.setValueAtTime(90, t);
  osc.frequency.exponentialRampToValueAtTime(38, t + 0.18);
  const g2 = a.createGain();
  env(g2, t, 0.002, 0.18, 0.45);
  osc.connect(g2).connect(a.destination);
  osc.start(t);
  osc.stop(t + 0.2);
}

export function playHit() {
  const a = ac();
  const t = a.currentTime;
  const osc = a.createOscillator();
  osc.type = 'triangle';
  osc.frequency.setValueAtTime(180, t);
  osc.frequency.exponentialRampToValueAtTime(70, t + 0.12);
  const g = a.createGain();
  env(g, t, 0.004, 0.14, 0.28);
  osc.connect(g).connect(a.destination);
  osc.start(t);
  osc.stop(t + 0.16);

  const n = a.createBuffer(1, a.sampleRate * 0.12, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * (1 - i / d.length);
  const src = a.createBufferSource();
  src.buffer = n;
  const bp = a.createBiquadFilter();
  bp.type = 'highpass';
  bp.frequency.value = 600;
  const g2 = a.createGain();
  env(g2, t, 0.003, 0.1, 0.2);
  src.connect(bp).connect(g2).connect(a.destination);
  src.start(t);
}

export function playBurst() {
  const a = ac();
  const t = a.currentTime;
  const n = a.createBuffer(1, a.sampleRate * 0.5, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / d.length, 0.7);
  const src = a.createBufferSource();
  src.buffer = n;
  const bp = a.createBiquadFilter();
  bp.type = 'bandpass';
  bp.frequency.value = 1200;
  const g = a.createGain();
  env(g, t, 0.01, 0.42, 0.4);
  src.connect(bp).connect(g).connect(a.destination);
  src.start(t);
}

let sparkle = null;

export function startSparkle() {
  stopSparkle();
  const a = ac();
  const master = a.createGain();
  master.gain.value = 0.05;
  master.connect(a.destination);

  const osc1 = a.createOscillator();
  osc1.type = 'sine';
  osc1.frequency.value = 1860;
  const osc2 = a.createOscillator();
  osc2.type = 'triangle';
  osc2.frequency.value = 2480;
  const g1 = a.createGain();
  g1.gain.value = 0.22;
  const g2 = a.createGain();
  g2.gain.value = 0.12;
  osc1.connect(g1).connect(master);
  osc2.connect(g2).connect(master);

  const lfo = a.createOscillator();
  lfo.type = 'sine';
  lfo.frequency.value = 6.5;
  const lfoG = a.createGain();
  lfoG.gain.value = 0.018;
  lfo.connect(lfoG).connect(master.gain);

  const n = a.createBuffer(1, a.sampleRate * 2, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * 0.4;
  const src = a.createBufferSource();
  src.buffer = n;
  src.loop = true;
  const bp = a.createBiquadFilter();
  bp.type = 'highpass';
  bp.frequency.value = 4200;
  const ng = a.createGain();
  ng.gain.value = 0.08;
  src.connect(bp).connect(ng).connect(master);

  osc1.start();
  osc2.start();
  lfo.start();
  src.start();
  sparkle = { master, osc1, osc2, lfo, src };
}

export function stopSparkle() {
  if (!sparkle) return;
  try {
    sparkle.osc1.stop();
    sparkle.osc2.stop();
    sparkle.lfo.stop();
    sparkle.src.stop();
    sparkle.master.disconnect();
  } catch {
    /* already stopped */
  }
  sparkle = null;
}

export function playIgnite() {
  const a = ac();
  const t = a.currentTime;
  const n = a.createBuffer(1, a.sampleRate * 0.35, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / d.length, 0.8);
  const src = a.createBufferSource();
  src.buffer = n;
  const bp = a.createBiquadFilter();
  bp.type = 'bandpass';
  bp.frequency.value = 2400;
  const g = a.createGain();
  env(g, t, 0.02, 0.28, 0.28);
  src.connect(bp).connect(g).connect(a.destination);
  src.start(t);

  const osc = a.createOscillator();
  osc.type = 'sawtooth';
  osc.frequency.setValueAtTime(280, t);
  osc.frequency.exponentialRampToValueAtTime(90, t + 0.2);
  const g2 = a.createGain();
  env(g2, t, 0.01, 0.18, 0.12);
  osc.connect(g2).connect(a.destination);
  osc.start(t);
  osc.stop(t + 0.22);
}

export function playKick() {
  const a = ac();
  const t = a.currentTime;
  const n = a.createBuffer(1, a.sampleRate * 0.38, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / d.length, 1.1);
  const src = a.createBufferSource();
  src.buffer = n;
  const lp = a.createBiquadFilter();
  lp.type = 'lowpass';
  lp.frequency.value = 420;
  const g = a.createGain();
  env(g, t, 0.004, 0.32, 0.58);
  src.connect(lp).connect(g).connect(a.destination);
  src.start(t);

  const osc = a.createOscillator();
  osc.type = 'triangle';
  osc.frequency.setValueAtTime(110, t);
  osc.frequency.exponentialRampToValueAtTime(42, t + 0.22);
  const g2 = a.createGain();
  env(g2, t, 0.003, 0.24, 0.38);
  osc.connect(g2).connect(a.destination);
  osc.start(t);
  osc.stop(t + 0.28);
}

export function playFlash() {
  const a = ac();
  const t = a.currentTime;
  const n = a.createBuffer(1, a.sampleRate * 0.55, a.sampleRate);
  const d = n.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / d.length, 0.45);
  const src = a.createBufferSource();
  src.buffer = n;
  const hp = a.createBiquadFilter();
  hp.type = 'highpass';
  hp.frequency.value = 1800;
  const g = a.createGain();
  env(g, t, 0.004, 0.48, 0.32);
  src.connect(hp).connect(g).connect(a.destination);
  src.start(t);
}
