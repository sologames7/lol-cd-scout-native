import * as THREE from 'three';
import { createItemMesh, ON_SHELF } from './items.js';

function std(color, metal, rough, extra = {}) {
  return new THREE.MeshStandardMaterial({
    color, metalness: metal, roughness: rough, ...extra,
  });
}

function woodCanvas(w, h, a, b) {
  const c = document.createElement('canvas');
  c.width = w;
  c.height = h;
  const g = c.getContext('2d');
  g.fillStyle = a;
  g.fillRect(0, 0, w, h);
  for (let x = 0; x < w; x += 7) {
    g.fillStyle = `rgba(20,10,4,${0.08 + Math.random() * 0.16})`;
    g.fillRect(x, 0, 2 + Math.random() * 5, h);
    if (x % 28 < 4) {
      g.fillStyle = b;
      g.fillRect(x, 0, 3, h);
    }
  }
  const t = new THREE.CanvasTexture(c);
  t.wrapS = t.wrapT = THREE.RepeatWrapping;
  t.anisotropy = 8;
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  return t;
}

function cobbleCanvas() {
  const c = document.createElement('canvas');
  c.width = 256;
  c.height = 256;
  const g = c.getContext('2d');
  g.fillStyle = '#5a584c';
  g.fillRect(0, 0, 256, 256);
  for (let y = 0; y < 256; y += 18) {
    for (let x = 0; x < 256; x += 22) {
      const ox = (Math.floor(y / 18) % 2) * 11;
      g.fillStyle = `rgb(${90 + Math.random() * 35},${86 + Math.random() * 28},${72 + Math.random() * 20})`;
      g.fillRect(x + ox + 1, y + 1, 18, 14);
      g.strokeStyle = 'rgba(30,24,16,0.35)';
      g.strokeRect(x + ox + 1, y + 1, 18, 14);
    }
  }
  const t = new THREE.CanvasTexture(c);
  t.wrapS = t.wrapT = THREE.RepeatWrapping;
  t.repeat.set(10, 8);
  t.anisotropy = 8;
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  return t;
}

function plankMat(tex, color, rx, ry) {
  const map = tex.clone();
  map.repeat.set(rx, ry);
  map.needsUpdate = true;
  return new THREE.MeshStandardMaterial({ map, color, metalness: 0.04, roughness: 0.84 });
}

function canvasStripe() {
  const c = document.createElement('canvas');
  c.width = 256;
  c.height = 128;
  const g = c.getContext('2d');
  for (let i = 0; i < 12; i++) {
    g.fillStyle = i % 2 ? '#6a3a1c' : '#c4a06a';
    g.fillRect(0, i * 11, 256, 11);
  }
  const t = new THREE.CanvasTexture(c);
  t.wrapS = t.wrapT = THREE.RepeatWrapping;
  t.repeat.set(2, 1);
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  return t;
}

function signTex() {
  const c = document.createElement('canvas');
  c.width = 512;
  c.height = 256;
  const g = c.getContext('2d');
  g.fillStyle = '#6a4424';
  g.fillRect(0, 0, 512, 256);
  g.strokeStyle = '#2a1810';
  g.lineWidth = 12;
  g.strokeRect(10, 10, 492, 236);
  g.fillStyle = '#f0d8a8';
  g.textAlign = 'center';
  g.font = '700 42px Georgia, serif';
  g.fillText('MARCHAND AMBULANT', 256, 110);
  g.font = 'italic 26px Georgia, serif';
  g.fillStyle = '#d4b24a';
  g.fillText('Fontaine  ·  Faille de l\'invocateur', 256, 168);
  const t = new THREE.CanvasTexture(c);
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  return t;
}

function barrel(wood, iron) {
  const g = new THREE.Group();
  const body = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.3, 0.62, 12), wood);
  const hoop = (y) => {
    const h = new THREE.Mesh(new THREE.TorusGeometry(0.3, 0.018, 6, 16), iron);
    h.rotation.x = Math.PI / 2;
    h.position.y = y;
    return h;
  };
  g.add(body, hoop(0.22), hoop(-0.22), hoop(0));
  return g;
}

function crate(wood, iron) {
  const g = new THREE.Group();
  g.add(new THREE.Mesh(new THREE.BoxGeometry(0.55, 0.42, 0.55), wood));
  const band = new THREE.Mesh(new THREE.BoxGeometry(0.58, 0.05, 0.58), iron);
  band.position.y = 0.12;
  g.add(band);
  return g;
}

function candle(x, y, z) {
  const g = new THREE.Group();
  const wax = new THREE.Mesh(new THREE.CylinderGeometry(0.028, 0.032, 0.14, 8), std(0xe8d8b8, 0.05, 0.7));
  const flame = new THREE.Mesh(
    new THREE.ConeGeometry(0.018, 0.05, 6),
    new THREE.MeshBasicMaterial({ color: 0xffe08a }),
  );
  flame.position.y = 0.1;
  const l = new THREE.PointLight(0xffc48a, 0.5, 2.6, 1.8);
  l.position.y = 0.12;
  g.add(wax, flame, l);
  g.position.set(x, y, z);
  return g;
}

function wheel(wood, iron) {
  const g = new THREE.Group();
  g.add(new THREE.Mesh(new THREE.TorusGeometry(0.42, 0.055, 8, 18), wood));
  const hub = new THREE.Mesh(new THREE.CylinderGeometry(0.08, 0.08, 0.1, 10), iron);
  hub.rotation.z = Math.PI / 2;
  g.add(hub);
  for (let i = 0; i < 8; i++) {
    const spoke = new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.78, 0.03), wood);
    spoke.rotation.z = (i / 8) * Math.PI;
    g.add(spoke);
  }
  return g;
}

function yordleMerchant(cloth, skin, wood) {
  const g = new THREE.Group();
  const body = new THREE.Mesh(new THREE.SphereGeometry(0.22, 12, 10), cloth);
  body.scale.set(1, 1.15, 0.9);
  body.position.y = 0.28;
  const head = new THREE.Mesh(new THREE.SphereGeometry(0.16, 12, 10), skin);
  head.position.y = 0.58;
  const hat = new THREE.Mesh(new THREE.ConeGeometry(0.18, 0.22, 8), std(0x4a2a14, 0.05, 0.8));
  hat.position.y = 0.74;
  hat.rotation.z = 0.15;
  const ear = (x) => {
    const e = new THREE.Mesh(new THREE.ConeGeometry(0.05, 0.16, 6), skin);
    e.position.set(x, 0.68, 0);
    e.rotation.z = x > 0 ? -0.5 : 0.5;
    return e;
  };
  const nose = new THREE.Mesh(new THREE.SphereGeometry(0.04, 8, 6), skin);
  nose.position.set(0, 0.55, 0.14);
  g.add(body, head, hat, ear(-0.14), ear(0.14), nose);
  g.position.set(-1.05, -1.22, -0.35);
  g.rotation.y = 0.35;
  return g;
}

function fountain(stone, waterM) {
  const g = new THREE.Group();
  const base = new THREE.Mesh(new THREE.CylinderGeometry(1.15, 1.35, 0.35, 16), stone);
  const bowl = new THREE.Mesh(new THREE.CylinderGeometry(1.05, 0.95, 0.22, 16), stone);
  bowl.position.y = 0.28;
  const water = new THREE.Mesh(new THREE.CircleGeometry(0.92, 20), waterM);
  water.rotation.x = -Math.PI / 2;
  water.position.y = 0.4;
  const pillar = new THREE.Mesh(new THREE.CylinderGeometry(0.12, 0.16, 0.7, 10), stone);
  pillar.position.y = 0.7;
  const top = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.22, 0.12, 12), stone);
  top.position.y = 1.08;
  g.add(base, bowl, water, pillar, top);
  const moon = new THREE.PointLight(0xb8c8d8, 0.45, 8, 1.6);
  moon.position.set(0, 1.4, 0);
  g.add(moon);
  g.position.set(2.8, -1.22, -6.6);
  return g;
}

function makeChest(wood, gold, iron) {
  const root = new THREE.Group();
  root.name = 'chest';
  const body = new THREE.Group();
  const box = new THREE.Mesh(new THREE.BoxGeometry(0.95, 0.52, 0.62), wood);
  box.position.y = 0.26;
  body.add(box);
  const bandL = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.54, 0.64), gold);
  bandL.position.set(-0.32, 0.26, 0);
  const bandR = bandL.clone();
  bandR.position.x = 0.32;
  const bandM = new THREE.Mesh(new THREE.BoxGeometry(0.97, 0.06, 0.64), gold);
  bandM.position.y = 0.12;
  const lock = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.16, 0.08), gold);
  lock.position.set(0, 0.38, 0.34);
  const hasp = new THREE.Mesh(new THREE.CylinderGeometry(0.04, 0.04, 0.1, 8), iron);
  hasp.rotation.x = Math.PI / 2;
  hasp.position.set(0, 0.48, 0.34);
  body.add(bandL, bandR, bandM, lock, hasp);
  root.add(body);

  const lid = new THREE.Group();
  lid.position.set(0, 0.52, -0.28);
  const lidBox = new THREE.Mesh(new THREE.BoxGeometry(0.97, 0.16, 0.64), wood);
  lidBox.position.set(0, 0.08, 0.28);
  const lidBand = new THREE.Mesh(new THREE.BoxGeometry(0.99, 0.05, 0.66), gold);
  lidBand.position.set(0, 0.08, 0.28);
  const arch = new THREE.Mesh(
    new THREE.CylinderGeometry(0.31, 0.31, 0.95, 12, 1, false, 0, Math.PI),
    wood,
  );
  arch.rotation.z = Math.PI / 2;
  arch.position.set(0, 0.16, 0.28);
  lid.add(lidBox, lidBand, arch);
  lid.rotation.x = -0.32;
  root.add(lid);

  const loot = [];
  for (let i = 0; i < 10; i++) {
    const coin = new THREE.Mesh(
      new THREE.CylinderGeometry(0.07, 0.07, 0.014, 12),
      std(0xd4b24a, 1, 0.28),
    );
    coin.rotation.x = Math.PI / 2;
    coin.position.set((Math.random() - 0.5) * 0.45, 0.56 + Math.random() * 0.08, (Math.random() - 0.5) * 0.25);
    coin.visible = false;
    root.add(coin);
    loot.push(coin);
  }

  const glowCore = new THREE.Mesh(
    new THREE.SphereGeometry(0.2, 12, 10),
    new THREE.MeshBasicMaterial({ color: 0xffe08a, transparent: true, opacity: 0.7, depthWrite: false }),
  );
  glowCore.position.y = 0.28;
  const glowHalo = new THREE.Mesh(
    new THREE.SphereGeometry(0.38, 12, 10),
    new THREE.MeshBasicMaterial({
      color: 0xffc14a, transparent: true, opacity: 0.16, depthWrite: false, side: THREE.BackSide,
    }),
  );
  glowHalo.position.y = 0.3;
  const innerGold = new THREE.PointLight(0xffc14a, 2.4, 3.8, 1.35);
  innerGold.position.set(0, 0.32, 0.08);
  const innerCurse = new THREE.PointLight(0xe8b060, 0.8, 3.2, 1.5);
  innerCurse.position.set(0, 0.36, -0.05);
  root.add(glowCore, glowHalo, innerGold, innerCurse);

  const motes = [];
  for (let i = 0; i < 14; i++) {
    const m = new THREE.Mesh(
      new THREE.SphereGeometry(0.018, 6, 6),
      new THREE.MeshBasicMaterial({ color: i % 2 ? 0xffe08a : 0xffd978 }),
    );
    root.add(m);
    motes.push({ m, a: Math.random() * Math.PI * 2, s: 0.4 + Math.random() * 0.8 });
  }

  const bits = [];
  const stash = (mesh) => {
    bits.push({
      m: mesh,
      origin: mesh.position.clone(),
      rot0: mesh.rotation.clone(),
      v: new THREE.Vector3(),
      spin: new THREE.Vector3(),
    });
  };
  [box, bandL, bandR, bandM, lock, hasp, lidBox, lidBand, arch].forEach(stash);
  return { root, body, lid, bits, loot, glowCore, glowHalo, innerGold, innerCurse, motes, lidOpen: -0.32 };
}

function stockWagon(root, catalog, wood, gold) {
  if (!catalog.length) return;
  const slots = [];
  for (let k = 0; k < 3; k++) {
    const board = new THREE.Mesh(new THREE.BoxGeometry(2.05, 0.05, 0.36), wood);
    board.position.set(-1.52, -0.18 + k * 0.62, -3.55);
    board.rotation.y = Math.PI / 2;
    root.add(board);
    for (let i = 0; i < 4; i++) {
      slots.push({ x: -1.36, y: 0.08 + k * 0.62, z: -4.3 + i * 0.48, ry: Math.PI / 2 });
    }
  }
  for (let k = 0; k < 3; k++) {
    const board = new THREE.Mesh(new THREE.BoxGeometry(2.05, 0.05, 0.36), wood);
    board.position.set(1.52, -0.18 + k * 0.62, -3.55);
    board.rotation.y = -Math.PI / 2;
    root.add(board);
    for (let i = 0; i < 4; i++) {
      slots.push({ x: 1.36, y: 0.08 + k * 0.62, z: -4.3 + i * 0.48, ry: -Math.PI / 2 });
    }
  }
  for (let i = 0; i < 5; i++) {
    slots.push({ x: -0.8 + i * 0.4, y: -0.22, z: -0.18, ry: 0, counter: true });
  }
  const ordered = [];
  for (const id of ON_SHELF) {
    const it = catalog.find((c) => String(c.id) === String(id));
    if (it) ordered.push(it);
  }
  ordered.forEach((it, i) => {
    if (i >= slots.length) return;
    const slot = slots[i];
    const mesh = createItemMesh(it, { price: true });
    mesh.scale.setScalar(slot.counter ? 0.38 : 0.34);
    mesh.position.set(slot.x, slot.y, slot.z);
    mesh.rotation.y = slot.ry + 0.15;
    root.add(mesh);
  });
}

export function createHold(catalog = []) {
  const root = new THREE.Group();
  root.name = 'fountain-shop';

  const plank = woodCanvas(256, 256, '#6a4324', '#3a2414');
  const plankDark = woodCanvas(256, 256, '#4a301c', '#2a180e');
  const woodM = plankMat(plank, 0xb07a44, 2, 2);
  const crateM = plankMat(plank, 0x9a6a3a, 1, 1);
  const barrelM = plankMat(plank, 0x8a5528, 1, 1);
  const darkW = plankMat(plankDark, 0x6a4428, 2, 2);
  const iron = std(0x2a2a2e, 0.85, 0.42);
  const gold = std(0xd4b24a, 1, 0.28);
  const rust = std(0x5a3a22, 0.55, 0.55);
  const stone = std(0x8a8478, 0.12, 0.78);
  const grass = std(0x3a5a28, 0.05, 0.9);
  const cloth = std(0x5a2a18, 0.05, 0.8);
  const skin = std(0xc4a070, 0.05, 0.65);
  const waterM = new THREE.MeshStandardMaterial({
    color: 0x4a88a8, metalness: 0.35, roughness: 0.15, transparent: true, opacity: 0.72,
  });

  const ground = new THREE.Mesh(new THREE.BoxGeometry(16, 0.12, 16), new THREE.MeshStandardMaterial({
    map: cobbleCanvas(), color: 0xc8c0b0, metalness: 0.08, roughness: 0.86,
  }));
  ground.position.set(0, -1.28, -3.2);
  root.add(ground);

  const lawn = new THREE.Mesh(new THREE.BoxGeometry(16, 0.04, 4.2), grass);
  lawn.position.set(0, -1.2, -8.4);
  root.add(lawn);
  const lawnF = lawn.clone();
  lawnF.position.set(0, -1.2, 2.4);
  root.add(lawnF);

  const awning = new THREE.Mesh(
    new THREE.BoxGeometry(4.4, 0.04, 3.6),
    new THREE.MeshStandardMaterial({ map: canvasStripe(), roughness: 0.85, metalness: 0.02 }),
  );
  awning.position.set(0, 1.35, -1.15);
  awning.rotation.x = 0.12;
  root.add(awning);
  [-1.9, 1.9].forEach((x) => {
    const pole = new THREE.Mesh(new THREE.CylinderGeometry(0.05, 0.06, 2.6, 8), darkW);
    pole.position.set(x, 0.05, 0.45);
    root.add(pole);
  });

  const counter = new THREE.Mesh(new THREE.BoxGeometry(3.4, 0.7, 0.55), woodM);
  counter.position.set(0, -0.88, -0.35);
  const top = new THREE.Mesh(new THREE.BoxGeometry(3.5, 0.06, 0.62), crateM);
  top.position.set(0, -0.5, -0.35);
  root.add(counter, top);

  const sign = new THREE.Mesh(
    new THREE.PlaneGeometry(1.7, 0.85),
    new THREE.MeshStandardMaterial({ map: signTex(), roughness: 0.8 }),
  );
  sign.position.set(0, 0.95, 0.48);
  root.add(sign);

  const wallL = new THREE.Mesh(new THREE.BoxGeometry(0.12, 2.2, 3.4), darkW);
  wallL.position.set(-1.85, -0.12, -3.55);
  const wallR = wallL.clone();
  wallR.position.x = 1.85;
  const wallBack = new THREE.Mesh(new THREE.BoxGeometry(3.82, 2.2, 0.12), darkW);
  wallBack.position.set(0, -0.12, -5.22);
  const floorW = new THREE.Mesh(new THREE.BoxGeometry(3.7, 0.1, 3.3), woodM);
  floorW.position.set(0, -1.18, -3.55);
  root.add(wallL, wallR, wallBack, floorW);

  const roof = new THREE.Mesh(new THREE.BoxGeometry(4.0, 0.1, 3.5), darkW);
  roof.position.set(0, 1.02, -3.55);
  root.add(roof);

  const wL = wheel(crateM, iron);
  wL.position.set(-1.95, -0.85, -2.4);
  wL.rotation.y = Math.PI / 2;
  const wR = wheel(crateM, iron);
  wR.position.set(1.95, -0.85, -2.4);
  wR.rotation.y = Math.PI / 2;
  const wL2 = wheel(crateM, iron);
  wL2.position.set(-1.95, -0.85, -4.7);
  wL2.rotation.y = Math.PI / 2;
  const wR2 = wheel(crateM, iron);
  wR2.position.set(1.95, -0.85, -4.7);
  wR2.rotation.y = Math.PI / 2;
  root.add(wL, wR, wL2, wR2);

  const partitionZ = -2.18;
  const openingW = 1.55;
  const openingH = 1.85;
  const sideW = (3.7 - openingW) / 2;
  const partL = new THREE.Mesh(new THREE.BoxGeometry(sideW, 2.05, 0.1), darkW);
  partL.position.set(-(openingW / 2 + sideW / 2), -0.18, partitionZ);
  const partR = partL.clone();
  partR.position.x = openingW / 2 + sideW / 2;
  const lintel = new THREE.Mesh(new THREE.BoxGeometry(openingW + 0.16, 0.16, 0.12), darkW);
  lintel.position.set(0, -1.22 + openingH + 0.08, partitionZ);
  root.add(partL, partR, lintel);

  const hingeX = -openingW / 2 + 0.04;
  const doorPivot = new THREE.Group();
  doorPivot.name = 'door';
  doorPivot.position.set(hingeX, -1.22 + openingH / 2, partitionZ + 0.02);
  const door = new THREE.Mesh(new THREE.BoxGeometry(openingW - 0.06, openingH - 0.08, 0.08), woodM);
  door.position.x = (openingW - 0.06) / 2;
  doorPivot.add(door);
  for (let i = 0; i < 3; i++) {
    const strap = new THREE.Mesh(new THREE.BoxGeometry(openingW - 0.1, 0.06, 0.1), rust);
    strap.position.set((openingW - 0.06) / 2, -0.55 + i * 0.55, 0);
    doorPivot.add(strap);
  }
  const handle = new THREE.Mesh(new THREE.TorusGeometry(0.06, 0.016, 6, 10), iron);
  handle.position.set(openingW - 0.28, 0, 0.06);
  doorPivot.add(handle);
  root.add(doorPivot);

  const chestBuilt = makeChest(crateM, gold, iron);
  chestBuilt.root.position.set(0, -1.18, -4.45);
  root.add(chestBuilt.root);

  const b1 = barrel(barrelM, rust);
  b1.position.set(-2.55, -0.95, -0.15);
  const b2 = barrel(barrelM, rust);
  b2.position.set(-2.75, -0.95, 0.45);
  b2.rotation.z = 1.55;
  const c1 = crate(crateM, rust);
  c1.position.set(2.45, -1.05, 0.15);
  c1.rotation.y = 0.25;
  root.add(b1, b2, c1);

  const sackM = std(0x8a7040, 0.04, 0.85);
  const sack = (x, z) => {
    const s = new THREE.Mesh(new THREE.SphereGeometry(0.22, 10, 8), sackM);
    s.scale.set(1.1, 0.7, 0.9);
    s.position.set(x, -1.08, z);
    root.add(s);
  };
  sack(2.7, -1.1);
  sack(2.45, -1.45);

  const bottleM = std(0x1a4a38, 0.3, 0.25, { transparent: true, opacity: 0.75 });
  for (let i = 0; i < 5; i++) {
    const bot = new THREE.Mesh(new THREE.CylinderGeometry(0.035, 0.04, 0.16, 8), bottleM);
    bot.position.set(-0.7 + i * 0.16, -0.38, -0.28);
    root.add(bot);
  }

  for (let i = 0; i < 10; i++) {
    const coin = new THREE.Mesh(new THREE.CylinderGeometry(0.045, 0.045, 0.012, 10), gold);
    coin.rotation.x = Math.PI / 2;
    coin.position.set(0.55 + (Math.random() - 0.5) * 0.4, -0.45 + Math.random() * 0.03, -0.22 + (Math.random() - 0.5) * 0.18);
    root.add(coin);
  }

  root.add(yordleMerchant(cloth, skin, woodM));
  root.add(fountain(stone, waterM));
  root.add(candle(-1.5, -0.42, -0.2));
  root.add(candle(1.35, -0.42, -0.22));
  root.add(candle(-1.55, 0.22, -3.4));
  root.add(candle(1.55, 0.22, -4.6));

  stockWagon(root, catalog, crateM, gold);

  const lantern = (x, y, z) => {
    const g = new THREE.Group();
    const cage = new THREE.Mesh(new THREE.CylinderGeometry(0.09, 0.1, 0.22, 8), iron);
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.07, 10, 8),
      new THREE.MeshStandardMaterial({
        color: 0xffe2a0, emissive: 0xffc14a, emissiveIntensity: 1.4, roughness: 0.4,
      }),
    );
    const chain = new THREE.Mesh(new THREE.CylinderGeometry(0.012, 0.012, 0.4, 6), iron);
    chain.position.y = 0.3;
    g.add(cage, glow, chain);
    const light = new THREE.PointLight(0xffc48a, 2.2, 8, 1.25);
    g.add(light);
    g.position.set(x, y, z);
    root.add(g);
    return light;
  };
  const lights = [
    lantern(-1.15, 1.05, -0.4),
    lantern(1.2, 1.05, -0.5),
    lantern(0, 0.85, -3.9),
  ];
  lights.forEach((l, i) => { l.userData.phase = i * 1.7; });

  const fill = new THREE.PointLight(0xc4a070, 0.7, 14, 1.25);
  fill.position.set(0, 0.9, -1.2);
  root.add(fill);
  const chestGlow = new THREE.PointLight(0xffd978, 0, 5, 1.6);
  chestGlow.position.set(0, -0.4, -4.3);
  root.add(chestGlow);

  const doorRest = doorPivot.position.clone();
  const doorRot0 = doorPivot.rotation.clone();
  const chestPos = chestBuilt.root.position.clone();

  return {
    root,
    door: doorPivot,
    chest: chestBuilt.root,
    chestPos,
    lights,
    chestGlow,
    kicked: false,
    burst: false,
    awake: false,
    hit: 0,
    doorV: new THREE.Vector3(),
    doorSpin: new THREE.Vector3(),
    t: 0,
    kick() {
      if (this.kicked) return;
      this.kicked = true;
      this.doorV.set(1.35, 1.9, -4.4);
      this.doorSpin.set(1.1, -5.2, 0.9);
    },
    punchChest() { this.hit = 1; },
    burstChest() {
      if (this.burst) return;
      this.burst = true;
      this.hit = 1.5;
      this.chestGlow.intensity = 10;
      chestBuilt.glowCore.visible = false;
      chestBuilt.glowHalo.visible = false;
      chestBuilt.innerGold.intensity = 0;
      chestBuilt.innerCurse.intensity = 0;
      for (const it of chestBuilt.motes) it.m.visible = false;
      for (const it of chestBuilt.bits) {
        it.v.set(
          (Math.random() - 0.5) * 5,
          2.4 + Math.random() * 3.2,
          -0.4 + Math.random() * 2.2,
        );
        it.spin.set((Math.random() - 0.5) * 8, (Math.random() - 0.5) * 10, (Math.random() - 0.5) * 8);
      }
      for (const coin of chestBuilt.loot) {
        coin.visible = true;
        coin.userData.v = new THREE.Vector3(
          (Math.random() - 0.5) * 4, 2 + Math.random() * 3, (Math.random() - 0.5) * 3,
        );
        coin.userData.spin = (Math.random() - 0.5) * 10;
      }
      chestBuilt.lid.rotation.x = -1.6;
    },
    reset() {
      this.kicked = false;
      this.burst = false;
      this.awake = false;
      this.hit = 0;
      this.doorV.set(0, 0, 0);
      this.doorSpin.set(0, 0, 0);
      doorPivot.position.copy(doorRest);
      doorPivot.rotation.copy(doorRot0);
      doorPivot.visible = true;
      chestBuilt.root.position.copy(chestPos);
      chestBuilt.root.rotation.set(0, 0, 0);
      chestBuilt.lid.rotation.set(chestBuilt.lidOpen, 0, 0);
      this.chestGlow.intensity = 0;
      chestBuilt.glowCore.visible = true;
      chestBuilt.glowHalo.visible = true;
      chestBuilt.innerGold.intensity = 2.4;
      chestBuilt.innerCurse.intensity = 0.8;
      for (const it of chestBuilt.motes) it.m.visible = true;
      for (const it of chestBuilt.bits) {
        it.m.position.copy(it.origin);
        it.m.rotation.copy(it.rot0);
        it.m.visible = true;
      }
      for (const coin of chestBuilt.loot) coin.visible = false;
    },
    update(dt) {
      this.t += dt;
      this.hit *= Math.pow(0.06, dt);
      this.chestGlow.intensity *= Math.pow(0.18, dt);
      const dim = this.awake ? 1 : 0.18;
      for (const l of this.lights) {
        l.intensity = (2.15 * dim) + Math.sin(this.t * 3.1 + (l.userData.phase || 0)) * 0.22 * dim;
      }
      if (!this.burst) {
        const pulse = 0.82 + Math.sin(this.t * 3.4) * 0.18;
        chestBuilt.innerGold.intensity = 2.4 * pulse;
        chestBuilt.innerCurse.intensity = 0.8 * pulse;
        chestBuilt.glowCore.scale.setScalar(0.9 + Math.sin(this.t * 4.2) * 0.12);
        chestBuilt.glowHalo.material.opacity = 0.12 + Math.sin(this.t * 2.6) * 0.05;
        for (const it of chestBuilt.motes) {
          it.a += dt * it.s;
          it.m.position.y = 0.42 + Math.sin(it.a) * 0.16;
          it.m.position.x = Math.cos(it.a * 0.7) * 0.22;
          it.m.position.z = Math.sin(it.a * 0.9) * 0.16;
        }
      }
      if (!this.kicked) {
        doorPivot.rotation.y = Math.sin(this.t * 0.7) * 0.012;
      } else {
        doorPivot.position.addScaledVector(this.doorV, dt);
        this.doorV.y -= 11 * dt;
        doorPivot.rotation.x += this.doorSpin.x * dt;
        doorPivot.rotation.y += this.doorSpin.y * dt;
        doorPivot.rotation.z += this.doorSpin.z * dt;
        if (doorPivot.position.y < -1.05) {
          doorPivot.position.y = -1.05;
          this.doorV.y *= -0.22;
          this.doorV.x *= 0.55;
          this.doorV.z *= 0.55;
          this.doorSpin.multiplyScalar(0.55);
          if (Math.abs(this.doorV.y) < 0.35) this.doorV.set(0, 0, 0);
        }
      }
      if (!this.burst) {
        chestBuilt.root.position.x = Math.sin(this.t * 28) * this.hit * 0.08;
        chestBuilt.root.rotation.z = Math.sin(this.t * 24) * this.hit * 0.12;
        chestBuilt.lid.rotation.x = chestBuilt.lidOpen - this.hit * 0.22;
      } else {
        for (const it of chestBuilt.bits) {
          it.m.position.addScaledVector(it.v, dt);
          it.v.y -= 8.5 * dt;
          it.m.rotation.x += it.spin.x * dt;
          it.m.rotation.y += it.spin.y * dt;
          if (it.m.position.y < -4) it.m.visible = false;
        }
        for (const coin of chestBuilt.loot) {
          if (!coin.visible) continue;
          coin.position.addScaledVector(coin.userData.v, dt);
          coin.userData.v.y -= 8 * dt;
          coin.rotation.z += coin.userData.spin * dt;
          if (coin.position.y < -4) coin.visible = false;
        }
      }
    },
  };
}

export function createHandLantern() {
  const root = new THREE.Group();
  root.name = 'hand-lantern';
  const brass = std(0xc4a056, 0.92, 0.28);
  const brassDark = std(0x8a6424, 0.88, 0.38);
  const iron = std(0x2a2a2e, 0.82, 0.42);
  const glass = new THREE.MeshStandardMaterial({
    color: 0x4a4030,
    emissive: 0xffc14a,
    emissiveIntensity: 0.04,
    transparent: true,
    opacity: 0.45,
    roughness: 0.35,
    metalness: 0.05,
  });
  const base = new THREE.Mesh(new THREE.CylinderGeometry(0.07, 0.08, 0.045, 12), brassDark);
  root.add(base);
  const cap = new THREE.Mesh(new THREE.CylinderGeometry(0.055, 0.07, 0.04, 12), brass);
  cap.position.y = 0.2;
  root.add(cap);
  const hat = new THREE.Mesh(new THREE.ConeGeometry(0.08, 0.06, 10), brass);
  hat.position.y = 0.24;
  root.add(hat);
  const pane = new THREE.Mesh(new THREE.CylinderGeometry(0.055, 0.06, 0.16, 10, 1, true), glass);
  pane.position.y = 0.11;
  root.add(pane);
  for (let i = 0; i < 4; i++) {
    const a = (i / 4) * Math.PI * 2;
    const bar = new THREE.Mesh(new THREE.BoxGeometry(0.012, 0.17, 0.012), iron);
    bar.position.set(Math.cos(a) * 0.058, 0.11, Math.sin(a) * 0.058);
    root.add(bar);
  }
  const flame = new THREE.Mesh(
    new THREE.SphereGeometry(0.032, 10, 8),
    new THREE.MeshBasicMaterial({ color: 0xfff4c8 }),
  );
  flame.position.y = 0.1;
  flame.name = 'flame';
  flame.visible = false;
  root.add(flame);
  const handle = new THREE.Mesh(new THREE.TorusGeometry(0.045, 0.008, 6, 14, Math.PI), brass);
  handle.rotation.x = Math.PI;
  handle.position.y = 0.28;
  root.add(handle);
  const grip = new THREE.Mesh(new THREE.CylinderGeometry(0.016, 0.018, 0.12, 8), std(0x3a2214, 0.1, 0.65));
  grip.rotation.z = 0.9;
  grip.position.set(-0.07, -0.02, 0.02);
  root.add(grip);
  const light = new THREE.PointLight(0xffd6a0, 0, 14, 1.05);
  light.position.set(0, 0.1, 0.02);
  root.add(light);
  const spot = new THREE.SpotLight(0xffe4b8, 0, 16, 0.78, 0.42, 1.1);
  spot.position.set(0, 0.1, 0.02);
  const aim = new THREE.Object3D();
  aim.position.set(0.15, 0.08, -3.4);
  root.add(aim);
  spot.target = aim;
  root.add(spot);
  return { root, light, spot, flame, glass, on: false };
}
