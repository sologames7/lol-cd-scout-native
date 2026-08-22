const FALLBACK = '16.16.1';
const DD = 'https://ddragon.leagueoflegends.com';
const START = 17 * 60 + 35;

const fmt = (s) => {
  s = Math.max(0, Math.ceil(s));
  if (s >= 60) {
    const m = Math.floor(s / 60);
    const r = s % 60;
    return `${m}:${r.toString().padStart(2, '0')}`;
  }
  return String(s);
};

async function ddragonVer() {
  for (const url of [`${DD}/api/versions.json`, '/ddragon/api/versions.json']) {
    try {
      const r = await fetch(url);
      if (!r.ok) continue;
      const v = await r.json();
      if (Array.isArray(v) && v[0]) return v[0];
    } catch { /* suivant */ }
  }
  return FALLBACK;
}

function setIcons(ver) {
  const cdn = `${DD}/cdn/${ver}`;
  document.querySelectorAll('[data-dd]').forEach((img) => {
    const path = img.getAttribute('data-dd');
    if (!path) return;
    img.src = `${cdn}/${path}`;
  });
}

function tickMock(t0) {
  const now = START + (performance.now() - t0) / 1000;
  const clock = document.getElementById('mk-clock');
  if (clock) clock.textContent = fmt(now);

  document.querySelectorAll('[data-cd]').forEach((el) => {
    const base = +el.dataset.cd;
    const left = base - (performance.now() - t0) / 1000;
    const hv = el.querySelector('.hv');
    if (left <= 0) {
      el.classList.remove('mk-cd');
      el.classList.add('mk-up');
      if (hv) hv.textContent = '';
      el.style.removeProperty('--cd-p');
      return;
    }
    el.classList.add('mk-cd');
    el.classList.remove('mk-up');
    el.style.setProperty('--cd-p', String(Math.min(100, (left / base) * 100)));
    if (hv) hv.textContent = fmt(left);
  });

  document.querySelectorAll('[data-obj]').forEach((el) => {
    const left = +el.dataset.obj - (performance.now() - t0) / 1000;
    const row = el.closest('.mk-hobj');
    if (left <= 0) {
      el.textContent = 'en vie';
      if (row) { row.classList.remove('hot', 'warn'); row.classList.add('live'); }
      return;
    }
    el.textContent = fmt(left);
    if (row) {
      row.classList.toggle('hot', left <= 45);
      row.classList.toggle('warn', left > 45 && left <= 120);
    }
  });
}

async function boot() {
  const ver = await ddragonVer();
  setIcons(ver);
  const t0 = performance.now();
  tickMock(t0);
  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (reduced) return;
  setInterval(() => tickMock(t0), 300);
}

boot();
