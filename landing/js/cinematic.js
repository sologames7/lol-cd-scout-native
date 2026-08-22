import * as THREE from 'three';
import { GLTFLoader } from '../vendor/loaders/GLTFLoader.js';
import { createRevolver } from './revolver.js';
import { createKnife } from './knife.js';
import { createSpellToken, loadSummonerTextures } from './summoners.js';
import { loadIconicSpells } from './spells.js';
import { createHold, createHandLantern } from './hold.js';
import { loadShopCatalog, createItemMesh } from './items.js';
import { makeLandingStage } from './stage3d.js';
import { createCRT } from './crt.js';
import {
  unlockSfx, playKnife, playGun, playHit, playBurst, playKick, playFlash,
  playIgnite, startSparkle, stopSparkle,
} from './sfx.js';

const clamp = (v, a, b) => Math.max(a, Math.min(b, v));
const lerp = (a, b, t) => a + (b - a) * t;
const smooth = (t) => t * t * (3 - 2 * t);

const Phase = { DARK: 0, LIT: 1, KICK: 2, AIM: 3, BOOM: 4, SHOW: 5 };

const hud = {
  hint: () => document.getElementById('scroll-hint'),
  nav: () => document.querySelector('.nav'),
  bang: () => document.getElementById('flashbang'),
};

function setHint(html) {
  const el = hud.hint();
  if (el) el.innerHTML = html;
}

function lockScroll(on) {
  document.documentElement.style.overflow = on ? 'hidden' : '';
  document.body.style.overflow = on ? 'hidden' : '';
}

function autoScrollLanding() {
  lockScroll(false);
  const dest = document.getElementById('produit');
  if (!dest) return;
  const from = window.scrollY;
  const top = Math.max(0, dest.getBoundingClientRect().top + window.scrollY - 24);
  const t0 = performance.now();
  const dur = 2200;
  const step = (now) => {
    const t = Math.min(1, (now - t0) / dur);
    const e = t * t * (3 - 2 * t);
    window.scrollTo(0, from + (top - from) * e);
    if (t < 1) requestAnimationFrame(step);
  };
  requestAnimationFrame(step);
}

async function boot() {
  const canvas = document.getElementById('stage');
  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  const renderer = new THREE.WebGLRenderer({
    canvas, antialias: true, alpha: false, powerPreference: 'high-performance',
  });
  renderer.setPixelRatio(Math.min(2, window.devicePixelRatio || 1));
  renderer.setClearColor(0x121018, 1);
  renderer.toneMapping = THREE.ACESFilmicToneMapping;
  renderer.toneMappingExposure = 0.78;
  if (THREE.SRGBColorSpace) renderer.outputColorSpace = THREE.SRGBColorSpace;

  const scene = new THREE.Scene();
  scene.fog = new THREE.Fog(0x1a1814, 9, 18);
  const camera = new THREE.PerspectiveCamera(64, 1, 0.06, 40);
  camera.position.set(0, 0.08, 0.2);

  const amb = new THREE.AmbientLight(0x5a4a3c, 0.16);
  scene.add(amb);
  const hemi = new THREE.HemisphereLight(0x8a7060, 0x2a2418, 0.18);
  scene.add(hemi);
  const key = new THREE.DirectionalLight(0xffe0b0, 0.12);
  key.position.set(-0.6, 2.2, 1.4);
  scene.add(key);

  const muzzle = new THREE.PointLight(0xffe08a, 0, 5, 2);
  scene.add(muzzle);

  let catalog = [];
  try {
    catalog = await loadShopCatalog();
  } catch (err) {
    console.warn('shop catalog', err);
  }
  const hold = createHold(catalog);
  scene.add(hold.root);

  const smokeTex = await loadTex('assets/smoke.png');

  const knife = createKnife();
  knife.scale.setScalar(1.05);
  camera.add(knife);

  const lantern = createHandLantern();
  lantern.root.scale.setScalar(0.95);
  camera.add(lantern.root);

  const gun = createRevolver();
  gun.scale.setScalar(0.42);
  gun.visible = false;
  camera.add(gun);
  scene.add(camera);

  const flashMesh = new THREE.Mesh(
    new THREE.CircleGeometry(0.18, 14),
    new THREE.MeshBasicMaterial({
      color: 0xfff3c0, transparent: true, opacity: 0, side: THREE.DoubleSide, depthWrite: false,
    }),
  );
  scene.add(flashMesh);

  const shards = makeShards(scene);
  const smoke = makeSmoke(scene, smokeTex);
  const barrelSmoke = makeBarrelSmoke(scene, smokeTex);
  const coins = await makeCoins(scene, catalog);
  const landing = await makeLandingStage(scene);
  const crt = createCRT(renderer);

  const st = {
    phase: Phase.DARK,
    lit: 0,
    kickT: 0,
    gunT: 0,
    shots: 0,
    recoil: 0,
    reveal: 0,
    qHeld: false,
    flash: 0,
    sparkle: false,
    t: 0,
  };

  const raycaster = new THREE.Raycaster();
  const pointer = new THREE.Vector2();
  canvas.style.pointerEvents = 'auto';

  const lightLantern = () => {
    if (st.phase !== Phase.DARK) return;
    unlockSfx();
    st.phase = Phase.LIT;
    hold.awake = true;
    lantern.on = true;
    lantern.flame.visible = true;
    lantern.glass.color.set(0xffe8b8);
    lantern.glass.opacity = 0.62;
    playIgnite();
    setHint('<kbd>Q</kbd> + clic gauche — Fendre le hayon');
  };

  const smashDoor = () => {
    if (st.phase !== Phase.LIT) return;
    unlockSfx();
    st.phase = Phase.KICK;
    st.kickT = 0;
    playKnife();
    playKick();
    playHit();
    hold.kick();
    setHint('');
  };

  const shoot = () => {
    if (st.phase !== Phase.AIM || st.shots >= 3) return;
    unlockSfx();
    st.shots += 1;
    st.recoil = 1;
    playGun();
    barrelSmoke.emit(flashMesh.position);
    if (st.shots < 3) {
      hold.punchChest();
      playHit();
    } else {
      explode();
    }
  };

  const explode = () => {
    if (st.phase !== Phase.AIM) return;
    st.phase = Phase.BOOM;
    stopSparkle();
    st.sparkle = false;
    hold.burstChest();
    burstShards(shards, hold.chestPos);
    playBurst();
    playFlash();
    st.flash = 1;
    setHint('');
    window.setTimeout(() => {
      st.phase = Phase.SHOW;
      autoScrollLanding();
    }, 700);
  };

  if (reduced) {
    st.phase = Phase.SHOW;
    st.lit = 1;
    st.reveal = 1;
    hold.awake = true;
    lantern.on = true;
    lantern.flame.visible = true;
    lockScroll(false);
  } else {
    lockScroll(true);
    setHint('<kbd>E</kbd> Allumer la lanterne');
  }

  canvas.addEventListener('pointerdown', (ev) => {
    unlockSfx();
    if (ev.button !== 0) return;
    if (st.phase === Phase.SHOW) {
      const r = canvas.getBoundingClientRect();
      pointer.x = ((ev.clientX - r.left) / r.width) * 2 - 1;
      pointer.y = -((ev.clientY - r.top) / r.height) * 2 + 1;
      raycaster.setFromCamera(pointer, camera);
      const hits = raycaster.intersectObject(landing.cta, true);
      if (hits.length && landing.cta.userData.href) {
        window.open(landing.cta.userData.href, '_blank', 'noopener');
      }
      return;
    }
    if (st.phase === Phase.LIT) smashDoor();
    else if (st.phase === Phase.AIM) shoot();
  });

  window.addEventListener('keydown', (ev) => {
    unlockSfx();
    if (ev.code === 'KeyE') {
      ev.preventDefault();
      lightLantern();
    }
    if (ev.code === 'KeyQ') {
      ev.preventDefault();
      st.qHeld = true;
      smashDoor();
    }
  });
  window.addEventListener('keyup', (ev) => {
    if (ev.code === 'KeyQ') st.qHeld = false;
  });

  const resize = () => {
    const w = window.innerWidth;
    const h = window.innerHeight;
    renderer.setSize(w, h, false);
    camera.aspect = w / h;
    camera.updateProjectionMatrix();
    crt.resize(w, h);
  };
  resize();
  window.addEventListener('resize', resize);

  let t0 = performance.now();
  const tick = (now) => {
    const dt = Math.min(0.05, (now - t0) / 1000);
    t0 = now;
    st.t += dt;
    apply(dt, {
      gun, knife, lantern, camera, hold, shards, smoke, barrelSmoke, coins, landing,
      muzzle, flashMesh, amb, hemi, key, renderer, st,
    });
    const bangEl = hud.bang();
    if (bangEl) bangEl.style.opacity = String(st.flash);
    renderer.toneMappingExposure = lerp(0.72, 1.12, st.lit) + st.flash * 1.45;
    crt.render(scene, camera, now * 0.001, lerp(0.7, 0.18, st.reveal));
    requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
}

function apply(dt, ctx) {
  const st = ctx.st;
  if (st.phase >= Phase.LIT) st.lit = Math.min(1, st.lit + dt * 2.4);
  if (st.phase === Phase.KICK) {
    st.kickT = Math.min(1, st.kickT + dt * 1.85);
    if (st.kickT >= 1) {
      st.phase = Phase.AIM;
      setHint('Clic gauche — Tirer sur le coffre-fort');
      if (!st.sparkle) {
        startSparkle();
        st.sparkle = true;
      }
    }
  }
  if (st.phase >= Phase.AIM) st.gunT = Math.min(1, st.gunT + dt * 2.2);
  if (st.phase >= Phase.BOOM) st.reveal = Math.min(1, st.reveal + dt * 0.85);
  st.recoil = Math.max(0, st.recoil - dt * 4.5);
  st.flash *= Math.pow(0.12, dt);

  ctx.amb.intensity = lerp(0.16, 0.7, st.lit);
  ctx.hemi.intensity = lerp(0.18, 0.55, st.lit);
  ctx.key.intensity = lerp(0.12, 0.35, st.lit);

  const kick = smooth(st.kickT);
  const stab = kick < 0.55 ? kick / 0.55 : 1;
  const retract = kick > 0.55 ? (kick - 0.55) / 0.45 : 0;
  const through = Math.max(kick, st.gunT * 0.4, st.reveal);
  const recoil = st.recoil;
  const camLock = st.phase >= Phase.BOOM ? Math.min(1, st.reveal * 2.4) : 0;
  const shake = (1 - camLock) * (stab * (1 - retract) * 0.05 + recoil * 0.05);

  ctx.camera.position.z = lerp(0.18, lerp(-0.85, -1.45, st.reveal), through);
  ctx.camera.position.x = shake * Math.sin(st.t * 18);
  ctx.camera.position.y = 0.08 + (1 - camLock) * (stab * (1 - retract) * 0.03 + recoil * 0.03);
  ctx.camera.rotation.z = (1 - camLock) * ((1 - stab) * -0.04 + stab * (1 - retract) * 0.18 - recoil * 0.03);
  ctx.camera.rotation.x = lerp(0.02, -0.04, through);

  const kOut = retract;
  const slash = stab * stab * (3 - 2 * stab);
  ctx.knife.visible = st.phase < Phase.AIM;
  ctx.knife.position.set(
    lerp(0.72, lerp(-0.55, 0.85, kOut), slash),
    lerp(0.38, lerp(-0.42, -0.55, kOut), slash),
    lerp(-0.42, lerp(-1.55, -0.18, kOut), slash),
  );
  ctx.knife.rotation.set(
    lerp(-0.55, lerp(0.95, 0.35, kOut), slash),
    lerp(0.55, lerp(-0.35, 0.15, kOut), slash),
    lerp(-1.35, lerp(1.55, 0.4, kOut), slash),
  );
  const trail = ctx.knife.getObjectByName('slash');
  if (trail) {
    const cut = slash * (1 - kOut);
    trail.material.opacity = cut > 0.12 && cut < 0.95 ? 0.55 * Math.sin(cut * Math.PI) : 0;
    trail.scale.setScalar(0.7 + cut * 0.9);
  }

  const bob = Math.sin(st.t * 2.2) * 0.01;
  ctx.lantern.root.position.set(-0.42 + recoil * 0.02, -0.34 + bob + stab * 0.04, -0.55);
  ctx.lantern.root.rotation.set(0.18 + stab * 0.12, 0.42, -0.18);
  const flicker = st.lit * (8.2 + Math.sin(st.t * 11) * 0.55 + Math.sin(st.t * 23) * 0.25);
  ctx.lantern.light.intensity = flicker;
  ctx.lantern.spot.intensity = st.lit * (5.2 + Math.sin(st.t * 11) * 0.35);
  ctx.lantern.glass.emissiveIntensity = lerp(0.04, 2.6, st.lit);
  if (ctx.lantern.flame) {
    ctx.lantern.flame.visible = st.lit > 0.15;
    ctx.lantern.flame.scale.setScalar(0.92 + Math.sin(st.t * 16) * 0.12);
  }

  const gunUp = smooth(st.gunT);
  ctx.gun.visible = gunUp > 0.04 && st.reveal < 0.92;
  ctx.gun.position.set(
    0.34,
    lerp(-1.35, -0.30, gunUp) + recoil * 0.04,
    lerp(-0.2, -0.62, gunUp) + recoil * 0.16,
  );
  ctx.gun.rotation.set(
    lerp(0.55, 0.06, gunUp) - recoil * 0.18,
    Math.PI + 0.08,
    0.05,
  );
  const drum = ctx.gun.getObjectByName('cylinder');
  if (drum) drum.rotation.z = st.shots * 1.05 + recoil * 0.4;

  const tip = ctx.gun.getObjectByName('muzzle');
  if (tip) {
    tip.getWorldPosition(ctx.flashMesh.position);
    ctx.flashMesh.lookAt(ctx.camera.position);
  }
  ctx.flashMesh.material.opacity = recoil * 0.9;
  ctx.flashMesh.scale.setScalar(0.35 + recoil * 1.8);
  ctx.muzzle.position.copy(ctx.flashMesh.position);
  ctx.muzzle.intensity = recoil * 16;

  ctx.hold.update(dt);
  ctx.barrelSmoke.update(dt);
  ctx.smoke.update(dt, st.reveal * 0.55, st.reveal);
  ctx.coins.update(dt, st.reveal, st.reveal > 0.65 ? (st.reveal - 0.65) / 0.35 : 0, ctx.hold.chestPos);
  ctx.landing.update(dt, st.reveal);

  const nav = hud.nav();
  if (nav) nav.style.opacity = String(0.18 + st.reveal * 0.82);
  const hint = hud.hint();
  if (hint && st.phase >= Phase.BOOM) hint.style.opacity = '0';
  else if (hint) hint.style.opacity = '1';
}

function makeShards(scene) {
  const group = new THREE.Group();
  scene.add(group);
  const cols = [0xd4b24a, 0x8a5528, 0xc45c1a, 0x3a2818, 0xf3d78a, 0x5a3a22];
  const items = [];
  for (let i = 0; i < 36; i++) {
    const m = new THREE.Mesh(
      new THREE.BoxGeometry(0.1, 0.08, 0.03),
      new THREE.MeshStandardMaterial({ color: cols[i % cols.length], metalness: 0.35, roughness: 0.5 }),
    );
    m.visible = false;
    group.add(m);
    items.push({ m, v: new THREE.Vector3(), spin: new THREE.Vector3() });
  }
  return items;
}

function burstShards(items, origin) {
  const o = origin || new THREE.Vector3(0, -0.6, -4.8);
  for (const it of items) {
    it.m.visible = true;
    it.m.position.copy(o);
    it.m.position.x += (Math.random() - 0.5) * 0.3;
    it.m.position.y += Math.random() * 0.2;
    it.v.set((Math.random() - 0.5) * 5, 2 + Math.random() * 4, (Math.random() - 0.5) * 3 + 1.2);
    it.spin.set(Math.random() * 8, Math.random() * 8, Math.random() * 8);
    it.m.material.opacity = 1;
    it.m.material.transparent = true;
  }
  const step = (dt) => {
    let alive = false;
    for (const it of items) {
      if (!it.m.visible) continue;
      it.m.position.addScaledVector(it.v, dt);
      it.v.y -= 3.2 * dt;
      it.m.rotation.x += it.spin.x * dt;
      it.m.rotation.y += it.spin.y * dt;
      it.m.material.opacity *= 0.985;
      if (it.m.material.opacity < 0.04) it.m.visible = false;
      else alive = true;
    }
    if (alive) requestAnimationFrame(() => step(0.016));
  };
  step(0.016);
}

function makeSmoke(scene, tex) {
  const sprites = [];
  for (let i = 0; i < 28; i++) {
    const mat = new THREE.SpriteMaterial({
      map: tex, transparent: true, depthWrite: false, opacity: 0,
      color: i % 3 === 0 ? 0xd8c8a8 : 0xb8b0a0,
    });
    const s = new THREE.Sprite(mat);
    s.scale.setScalar(1.4 + Math.random() * 1.8);
    scene.add(s);
    sprites.push({
      s,
      a: Math.random() * Math.PI * 2,
      r: 0.15 + Math.random() * 1.1,
      y: (Math.random() - 0.5) * 0.5,
      spin: (Math.random() - 0.5) * 0.4,
      grow: 1.2 + Math.random() * 1.4,
    });
  }
  return {
    update(dt, amount, reveal) {
      for (const it of sprites) {
        it.a += it.spin * dt;
        it.y += 0.18 * dt * amount;
        const x = Math.cos(it.a) * it.r * (0.4 + amount);
        const z = -4.7 + Math.sin(it.a * 0.7) * it.r;
        it.s.position.set(x, it.y - 0.2, z);
        it.s.scale.setScalar(it.grow * (0.35 + amount));
        it.s.material.opacity = amount * 0.38 * (1 - reveal * 0.6);
      }
    },
  };
}

function makeBarrelSmoke(scene, tex) {
  const puffs = [];
  for (let i = 0; i < 18; i++) {
    const mat = new THREE.SpriteMaterial({
      map: tex, transparent: true, depthWrite: false, depthTest: false, opacity: 0, color: 0xc8c2b4,
    });
    const s = new THREE.Sprite(mat);
    s.visible = false;
    scene.add(s);
    puffs.push({ s, life: 0, v: new THREE.Vector3() });
  }
  let cursor = 0;
  return {
    emit(pos) {
      for (let n = 0; n < 5; n++) {
        const it = puffs[cursor++ % puffs.length];
        it.s.position.copy(pos);
        it.s.position.x += (Math.random() - 0.5) * 0.05;
        it.s.position.y += 0.02 + (Math.random() - 0.5) * 0.04;
        it.s.position.z -= 0.06;
        it.life = 1;
        it.v.set((Math.random() - 0.5) * 0.18, 0.22 + Math.random() * 0.28, (Math.random() - 0.5) * 0.12);
        it.s.visible = true;
        it.s.scale.setScalar(0.08);
      }
    },
    update(dt) {
      for (const it of puffs) {
        if (it.life <= 0) {
          it.s.visible = false;
          continue;
        }
        it.life -= dt * 0.48;
        it.s.position.addScaledVector(it.v, dt);
        it.v.y += 0.18 * dt;
        it.s.material.opacity = Math.max(0, it.life) * 0.78;
        it.s.scale.setScalar(0.16 + (1 - it.life) * 0.72);
      }
    },
  };
}

async function makeCoins(scene, catalog = []) {
  const loader = new GLTFLoader();
  let proto = null;
  try {
    const gltf = await new Promise((resolve, reject) => {
      loader.load('assets/coin.glb', resolve, undefined, reject);
    });
    proto = gltf.scene;
  } catch {
    proto = fallbackCoin();
  }
  let spells = [];
  try {
    spells = await loadSummonerTextures();
  } catch (err) {
    console.warn('summoner textures', err);
  }
  let iconic = [];
  try {
    iconic = await loadIconicSpells();
  } catch (err) {
    console.warn('iconic spells', err);
  }

  const plan = [];
  for (let i = 0; i < 8; i++) plan.push({ type: 'coin' });
  for (const s of spells) {
    plan.push({ type: 'spell', kind: s });
    plan.push({ type: 'spell', kind: s });
  }
  for (const s of iconic) {
    plan.push({ type: 'spell', kind: s });
  }
  const floatItems = [
    ...catalog.filter((i) => i.rarity === 'legendary').slice(0, 10),
    ...catalog.filter((i) => i.rarity === 'epic').slice(0, 4),
    ...catalog.filter((i) => i.rarity === 'boots').slice(0, 2),
    ...catalog.filter((i) => i.rarity === 'consumable').slice(0, 3),
    ...catalog.filter((i) => i.rarity === 'trinket').slice(0, 2),
    ...catalog.filter((i) => i.rarity === 'basic').slice(0, 3),
  ];
  for (const it of floatItems) {
    plan.push({ type: 'item', kind: it });
  }

  const items = [];
  const spawn = (spec, band) => {
    let c;
    if (spec.type === 'spell') {
      c = createSpellToken(spec.kind);
    } else if (spec.type === 'item') {
      c = createItemMesh(spec.kind, { price: false });
    } else {
      c = proto.clone(true);
      c.traverse((o) => {
        if (o.isMesh) {
          o.castShadow = false;
          if (o.material) o.material = o.material.clone();
        }
      });
    }
    const extra = c.getObjectByName('extra');
    const base = spec.type === 'spell' ? 0.2 + Math.random() * 0.05
      : spec.type === 'item' ? 0.18 + Math.random() * 0.06
        : 0.16 + Math.random() * 0.06;
    c.scale.setScalar(base);
    scene.add(c);
    const yRest = band === 'high' ? 1.22 + Math.random() * 0.38 : -0.52 - Math.random() * 0.18;
    items.push({
      c, extra, base, band,
      a: Math.random() * Math.PI * 2,
      r: 1.1 + Math.random() * 2.2,
      yRest,
      y0: band === 'high' ? 0.2 : -0.4,
      spin: new THREE.Vector3(1 + Math.random(), 2 + Math.random() * 2, 0.4),
      extraSpeed: 1.6 + Math.random() * 2.2,
      speed: 0.35 + Math.random() * 0.5,
    });
  };
  for (const spec of plan) spawn(spec, 'low');
  for (const spec of plan) spawn(spec, 'high');
  return {
    update(dt, amount, settle, origin) {
      const o = origin || new THREE.Vector3(0, -0.55, -4.85);
      for (const it of items) {
        it.a += it.speed * dt;
        const floatX = Math.cos(it.a) * (1.15 + it.r) * amount;
        const floatY = lerp(it.y0, it.yRest + Math.sin(it.a) * 0.12, amount);
        const floatZ = -4.35 + Math.sin(it.a * 0.8) * 0.7;
        it.c.position.set(
          lerp(o.x, floatX, amount),
          lerp(o.y, floatY, amount),
          lerp(o.z, floatZ, amount),
        );
        it.c.rotation.x += it.spin.x * dt;
        it.c.rotation.y += it.spin.y * dt;
        if (it.extra) it.extra.rotation.z += it.extraSpeed * dt;
        it.c.visible = amount > 0.04;
        const park = settle * settle;
        it.c.position.y = lerp(it.c.position.y, it.yRest, park * 0.35);
        it.c.scale.setScalar(lerp(it.base, it.base * 0.78, park));
      }
    },
  };
}

function fallbackCoin() {
  const g = new THREE.Group();
  const m = new THREE.Mesh(
    new THREE.CylinderGeometry(1, 1, 0.12, 48),
    new THREE.MeshStandardMaterial({ color: 0xd4b24a, metalness: 1, roughness: 0.28 }),
  );
  m.rotation.x = Math.PI / 2;
  g.add(m);
  return g;
}

function loadTex(url) {
  return new Promise((resolve, reject) => {
    new THREE.TextureLoader().load(url, (t) => {
      t.colorSpace = THREE.SRGBColorSpace || t.colorSpace;
      resolve(t);
    }, undefined, reject);
  });
}

boot().catch((err) => {
  console.error(err);
  document.body.classList.add('no-webgl');
});
