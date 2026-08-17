import * as THREE from 'three';

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
    const shade = Math.random() * 0.18;
    g.fillStyle = `rgba(20,10,4,${0.08 + shade})`;
    g.fillRect(x, 0, 2 + Math.random() * 5, h);
    if (x % 28 < 4) {
      g.fillStyle = b;
      g.fillRect(x, 0, 3, h);
    }
  }
  for (let i = 0; i < 18; i++) {
    g.fillStyle = `rgba(40,22,10,${0.12 + Math.random() * 0.2})`;
    g.beginPath();
    g.ellipse(Math.random() * w, Math.random() * h, 6 + Math.random() * 10, 3, Math.random(), 0, Math.PI * 2);
    g.fill();
  }
  const t = new THREE.CanvasTexture(c);
  t.wrapS = t.wrapT = THREE.RepeatWrapping;
  t.anisotropy = 8;
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  return t;
}

function plankMat(tex, color, repeatX, repeatY) {
  const map = tex.clone();
  map.repeat.set(repeatX, repeatY);
  map.needsUpdate = true;
  return new THREE.MeshStandardMaterial({
    map, color, metalness: 0.04, roughness: 0.82,
  });
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

function cursedPosterTex() {
  const c = document.createElement('canvas');
  c.width = 512;
  c.height = 720;
  const g = c.getContext('2d');
  g.fillStyle = '#c4a06a';
  g.fillRect(0, 0, 512, 720);
  for (let i = 0; i < 40; i++) {
    g.fillStyle = `rgba(80,40,16,${0.04 + Math.random() * 0.08})`;
    g.fillRect(Math.random() * 512, Math.random() * 720, 40 + Math.random() * 80, 8);
  }
  g.strokeStyle = '#3a2214';
  g.lineWidth = 16;
  g.strokeRect(22, 22, 468, 676);
  g.strokeStyle = '#6a3a18';
  g.lineWidth = 4;
  g.strokeRect(38, 38, 436, 644);

  g.fillStyle = '#1a120c';
  g.beginPath();
  g.ellipse(256, 198, 72, 78, 0, 0, Math.PI * 2);
  g.fill();
  g.fillStyle = '#c4a06a';
  g.beginPath();
  g.ellipse(232, 188, 16, 20, 0, 0, Math.PI * 2);
  g.ellipse(280, 188, 16, 20, 0, 0, Math.PI * 2);
  g.fill();
  g.fillStyle = '#1a120c';
  g.beginPath();
  g.arc(232, 190, 7, 0, Math.PI * 2);
  g.arc(280, 190, 7, 0, Math.PI * 2);
  g.fill();
  g.fillRect(248, 228, 16, 28);
  g.beginPath();
  g.moveTo(220, 268);
  g.lineTo(256, 248);
  g.lineTo(292, 268);
  g.lineTo(256, 258);
  g.closePath();
  g.fill();
  g.strokeStyle = '#1a120c';
  g.lineWidth = 8;
  g.beginPath();
  g.moveTo(186, 250);
  g.lineTo(140, 310);
  g.moveTo(326, 250);
  g.lineTo(372, 310);
  g.stroke();

  g.fillStyle = '#2a1810';
  g.textAlign = 'center';
  g.font = '700 28px Georgia, serif';
  g.fillText('AVIS AUX FLOUS', 256, 360);
  g.font = '700 42px Georgia, serif';
  g.fillText('ÉQUIPEMENT', 256, 430);
  g.fillText('DE RARETÉ', 256, 486);
  g.fillStyle = '#6b1c1c';
  g.font = '700 48px Georgia, serif';
  g.fillText('MAUDITE', 256, 552);
  g.fillStyle = '#3a2214';
  g.font = 'italic 22px Georgia, serif';
  g.fillText('qui touche  ·  qui meurt', 256, 610);
  g.beginPath();
  g.arc(256, 655, 18, 0, Math.PI * 2);
  g.fillStyle = '#7a1a1a';
  g.fill();
  g.fillStyle = '#d4b24a';
  g.font = '700 16px Georgia, serif';
  g.fillText('☠', 256, 661);

  const t = new THREE.CanvasTexture(c);
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  t.anisotropy = 8;
  return t;
}

function wallMapTex() {
  const c = document.createElement('canvas');
  c.width = 512;
  c.height = 384;
  const g = c.getContext('2d');
  g.fillStyle = '#b89468';
  g.fillRect(0, 0, 512, 384);
  g.strokeStyle = '#3a2414';
  g.lineWidth = 10;
  g.strokeRect(8, 8, 496, 368);
  g.strokeStyle = '#5a3a20';
  g.lineWidth = 2;
  g.beginPath();
  g.moveTo(40, 200);
  g.bezierCurveTo(120, 80, 220, 300, 400, 140);
  g.stroke();
  g.fillStyle = '#6b1c1c';
  g.beginPath();
  g.arc(268, 188, 8, 0, Math.PI * 2);
  g.fill();
  g.font = '700 22px Georgia, serif';
  g.fillStyle = '#2a1810';
  g.fillText('ILES MAUDITES', 150, 50);
  const t = new THREE.CanvasTexture(c);
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  return t;
}

function candle(x, y, z, lit = true) {
  const g = new THREE.Group();
  const wax = new THREE.Mesh(new THREE.CylinderGeometry(0.028, 0.032, 0.14, 8), std(0xe8d8b8, 0.05, 0.7));
  const flame = new THREE.Mesh(
    new THREE.ConeGeometry(0.018, 0.05, 6),
    new THREE.MeshBasicMaterial({ color: 0xffe08a }),
  );
  flame.position.y = 0.1;
  g.add(wax, flame);
  if (lit) {
    const l = new THREE.PointLight(0xffc48a, 0.55, 2.8, 1.8);
    l.position.y = 0.12;
    g.add(l);
  }
  g.position.set(x, y, z);
  return g;
}

function skull(iron, bone) {
  const g = new THREE.Group();
  const head = new THREE.Mesh(new THREE.SphereGeometry(0.11, 10, 8), bone);
  head.scale.set(1, 1.1, 0.9);
  const jaw = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.04, 0.08), bone);
  jaw.position.set(0, -0.1, 0.02);
  g.add(head, jaw);
  return g;
}

function dressHold(root, { crateM, barrelM, rust, iron, gold, ropeM }) {
  const bone = std(0xe8d8c0, 0.05, 0.65);
  const sackM = std(0x8a7040, 0.04, 0.85);
  const bottleM = std(0x1a4a38, 0.3, 0.25, { transparent: true, opacity: 0.75 });

  const extra = [
    [-2.4, -0.95, 0.35], [-2.85, -0.95, -1.7], [2.75, -0.95, -0.15],
    [2.4, -0.95, -5.4], [-2.6, -0.95, -5.6],
  ];
  extra.forEach(([x, y, z], i) => {
    const b = barrel(barrelM, rust);
    b.position.set(x, y, z);
    if (i % 2) b.rotation.z = 1.55;
    root.add(b);
  });

  for (let i = 0; i < 8; i++) {
    const ball = new THREE.Mesh(new THREE.SphereGeometry(0.09, 10, 8), iron);
    ball.position.set(-2.35 + (i % 3) * 0.16, -1.12 + Math.floor(i / 3) * 0.14, -2.55 + (i % 2) * 0.14);
    root.add(ball);
  }

  const sack = (x, z) => {
    const s = new THREE.Mesh(new THREE.SphereGeometry(0.22, 10, 8), sackM);
    s.scale.set(1.1, 0.7, 0.9);
    s.position.set(x, -1.08, z);
    root.add(s);
  };
  sack(2.7, -2.4);
  sack(2.45, -2.7);
  sack(-2.8, -4.4);

  const sk = skull(iron, bone);
  sk.position.set(-2.5, -0.78, -3.7);
  sk.rotation.y = 0.6;
  root.add(sk);

  root.add(candle(-2.48, -0.78, -3.55));
  root.add(candle(2.42, -0.78, -1.2));
  root.add(candle(-1.8, -1.08, -4.9));
  root.add(candle(1.7, -1.08, -4.95));

  const table = new THREE.Mesh(new THREE.BoxGeometry(0.9, 0.08, 0.5), crateM);
  table.position.set(-2.35, -0.82, -0.2);
  root.add(table);
  for (let i = 0; i < 3; i++) {
    const bot = new THREE.Mesh(new THREE.CylinderGeometry(0.035, 0.04, 0.16, 8), bottleM);
    bot.position.set(-2.5 + i * 0.14, -0.7, -0.15);
    root.add(bot);
  }

  const chain = (x, z) => {
    for (let i = 0; i < 6; i++) {
      const link = new THREE.Mesh(new THREE.TorusGeometry(0.04, 0.01, 6, 8), iron);
      link.position.set(x, 1.35 - i * 0.1, z);
      link.rotation.y = i * 0.7;
      root.add(link);
    }
  };
  chain(-3.15, -2.4);
  chain(3.15, -4.2);

  const hammock = new THREE.Mesh(new THREE.BoxGeometry(1.6, 0.04, 0.55), ropeM);
  hammock.position.set(2.2, 0.15, -6.6);
  hammock.rotation.z = 0.12;
  root.add(hammock);

  const map = new THREE.Mesh(
    new THREE.PlaneGeometry(1.15, 0.85),
    new THREE.MeshStandardMaterial({ map: wallMapTex(), roughness: 0.8 }),
  );
  map.position.set(-3.34, 0.45, -0.6);
  map.rotation.y = Math.PI / 2;
  root.add(map);

  const banner = new THREE.Mesh(
    new THREE.PlaneGeometry(0.7, 1.3),
    new THREE.MeshStandardMaterial({ color: 0x6b1c1c, roughness: 0.75, side: THREE.DoubleSide }),
  );
  banner.position.set(3.34, 0.35, -5.0);
  banner.rotation.y = -Math.PI / 2;
  root.add(banner);

  const porthole = new THREE.Mesh(new THREE.TorusGeometry(0.28, 0.04, 8, 18), rust);
  porthole.position.set(-3.34, 0.55, -6.2);
  porthole.rotation.y = Math.PI / 2;
  const moon = new THREE.Mesh(
    new THREE.CircleGeometry(0.26, 16),
    new THREE.MeshBasicMaterial({ color: 0xa8c8d8 }),
  );
  moon.position.set(-3.33, 0.55, -6.2);
  moon.rotation.y = Math.PI / 2;
  const moonL = new THREE.PointLight(0xb8d4e8, 0.7, 5, 1.6);
  moonL.position.set(-2.8, 0.55, -6.2);
  root.add(porthole, moon, moonL);

  const wheel = new THREE.Mesh(new THREE.TorusGeometry(0.32, 0.04, 8, 16), crateM);
  wheel.position.set(3.2, 0.2, -2.8);
  wheel.rotation.y = Math.PI / 2;
  root.add(wheel);
  for (let i = 0; i < 6; i++) {
    const spoke = new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.62, 0.03), crateM);
    spoke.position.copy(wheel.position);
    spoke.rotation.z = (i / 6) * Math.PI;
    spoke.rotation.y = Math.PI / 2;
    root.add(spoke);
  }

  const sconce = (x, z, rotY) => {
    const g = new THREE.Group();
    const arm = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.04, 0.18), rust);
    const cup = candle(0, 0.08, -0.08);
    g.add(arm, cup);
    g.position.set(x, 0.55, z);
    g.rotation.y = rotY;
    root.add(g);
  };
  sconce(-3.28, -3.3, Math.PI / 2);
  sconce(3.28, -1.8, -Math.PI / 2);
  sconce(-3.28, 0.5, Math.PI / 2);

  const coil = new THREE.Mesh(new THREE.TorusGeometry(0.18, 0.05, 8, 16), ropeM);
  coil.rotation.x = Math.PI / 2;
  coil.position.set(-2.5, -1.12, 0.55);
  root.add(coil);

  const goldPile = new THREE.Group();
  for (let i = 0; i < 12; i++) {
    const c = new THREE.Mesh(new THREE.CylinderGeometry(0.05, 0.05, 0.012, 10), gold);
    c.rotation.x = Math.PI / 2;
    c.position.set((Math.random() - 0.5) * 0.35, 0.04 + Math.random() * 0.04, (Math.random() - 0.5) * 0.2);
    goldPile.add(c);
  }
  goldPile.position.set(0.55, -1.1, -4.35);
  root.add(goldPile);
}

function makeChest(wood, gold, iron, dark) {
  const root = new THREE.Group();
  root.name = 'chest';

  const body = new THREE.Group();
  body.name = 'chest-body';
  const box = new THREE.Mesh(new THREE.BoxGeometry(0.95, 0.52, 0.62), wood);
  box.position.y = 0.26;
  body.add(box);
  const bandL = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.54, 0.64), gold);
  bandL.position.set(-0.32, 0.26, 0);
  const bandR = bandL.clone();
  bandR.position.x = 0.32;
  const bandM = new THREE.Mesh(new THREE.BoxGeometry(0.97, 0.06, 0.64), gold);
  bandM.position.y = 0.12;
  body.add(bandL, bandR, bandM);
  const lock = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.16, 0.08), gold);
  lock.position.set(0, 0.38, 0.34);
  body.add(lock);
  const hasp = new THREE.Mesh(new THREE.CylinderGeometry(0.04, 0.04, 0.1, 8), iron);
  hasp.rotation.x = Math.PI / 2;
  hasp.position.set(0, 0.48, 0.34);
  body.add(hasp);
  root.add(body);

  const lid = new THREE.Group();
  lid.name = 'chest-lid';
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
    new THREE.MeshBasicMaterial({
      color: 0xffe08a, transparent: true, opacity: 0.7, depthWrite: false,
    }),
  );
  glowCore.position.y = 0.28;
  root.add(glowCore);
  const glowHalo = new THREE.Mesh(
    new THREE.SphereGeometry(0.38, 12, 10),
    new THREE.MeshBasicMaterial({
      color: 0x4dffc8, transparent: true, opacity: 0.18, depthWrite: false, side: THREE.BackSide,
    }),
  );
  glowHalo.position.y = 0.3;
  root.add(glowHalo);
  const innerGold = new THREE.PointLight(0xffc14a, 2.4, 3.8, 1.35);
  innerGold.position.set(0, 0.32, 0.08);
  root.add(innerGold);
  const innerCurse = new THREE.PointLight(0x3ee0c8, 1.4, 3.2, 1.5);
  innerCurse.position.set(0, 0.36, -0.05);
  root.add(innerCurse);

  const motes = [];
  for (let i = 0; i < 14; i++) {
    const m = new THREE.Mesh(
      new THREE.SphereGeometry(0.018, 6, 6),
      new THREE.MeshBasicMaterial({ color: i % 2 ? 0xffe08a : 0x7fffd4 }),
    );
    m.position.set((Math.random() - 0.5) * 0.5, 0.4 + Math.random() * 0.25, (Math.random() - 0.5) * 0.3);
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
  stash(box);
  stash(bandL);
  stash(bandR);
  stash(bandM);
  stash(lock);
  stash(hasp);
  stash(lidBox);
  stash(lidBand);
  stash(arch);

  return { root, body, lid, bits, loot, glowCore, glowHalo, innerGold, innerCurse, motes, lidOpen: -0.32 };
}

export function createHold() {
  const root = new THREE.Group();
  root.name = 'hold';

  const plank = woodCanvas(256, 256, '#6a4324', '#3a2414');
  const plankDark = woodCanvas(256, 256, '#4a301c', '#2a180e');
  const floorM = plankMat(plank, 0xc4a078, 8, 5);
  const wallM = plankMat(plankDark, 0xa47848, 4, 3);
  const ceilM = plankMat(plankDark, 0x8a6240, 6, 3);
  const doorM = plankMat(plank, 0xb07a44, 2, 3);
  const crateM = plankMat(plank, 0x9a6a3a, 1, 1);
  const barrelM = plankMat(plank, 0x8a5528, 1, 1);
  const iron = std(0x2a2a2e, 0.85, 0.4);
  const gold = std(0xd4b24a, 1, 0.28);
  const rust = std(0x5a3a22, 0.55, 0.55);
  const ropeM = std(0x8a7048, 0.05, 0.78);

  const floor = new THREE.Mesh(new THREE.BoxGeometry(7.2, 0.12, 10.4), floorM);
  floor.position.set(0, -1.28, -3.1);
  floor.receiveShadow = true;
  root.add(floor);

  for (let i = 0; i < 11; i++) {
    const seam = new THREE.Mesh(new THREE.BoxGeometry(7.0, 0.01, 0.03), std(0x2a1810, 0.05, 0.9));
    seam.position.set(0, -1.215, 1.4 - i * 0.82);
    root.add(seam);
  }

  const wallL = new THREE.Mesh(new THREE.BoxGeometry(0.18, 3.1, 10.4), wallM);
  wallL.position.set(-3.45, 0.22, -3.1);
  const wallR = wallL.clone();
  wallR.position.x = 3.45;
  const wallBack = new THREE.Mesh(new THREE.BoxGeometry(7.2, 3.1, 0.18), wallM);
  wallBack.position.set(0, 0.22, -8.2);
  const wallFront = new THREE.Mesh(new THREE.BoxGeometry(7.2, 3.1, 0.18), wallM);
  wallFront.position.set(0, 0.22, 1.85);
  root.add(wallL, wallR, wallBack, wallFront);

  const ceil = new THREE.Mesh(new THREE.BoxGeometry(7.2, 0.14, 10.4), ceilM);
  ceil.position.set(0, 1.72, -3.1);
  root.add(ceil);

  for (let i = 0; i < 6; i++) {
    const beam = new THREE.Mesh(new THREE.BoxGeometry(7.1, 0.16, 0.22), plankMat(plankDark, 0x5a3a22, 4, 1));
    beam.position.set(0, 1.58, 1.2 - i * 1.7);
    root.add(beam);
    const postL = new THREE.Mesh(new THREE.BoxGeometry(0.16, 2.7, 0.16), plankMat(plankDark, 0x5a3a22, 1, 2));
    postL.position.set(-3.28, 0.2, 1.2 - i * 1.7);
    const postR = postL.clone();
    postR.position.x = 3.28;
    root.add(postL, postR);
  }

  const partitionZ = -2.18;
  const openingW = 1.72;
  const openingH = 2.22;
  const wallH = 3.1;
  const sideW = (7.2 - openingW) / 2;
  const partL = new THREE.Mesh(new THREE.BoxGeometry(sideW, wallH, 0.16), wallM);
  partL.position.set(-(openingW / 2 + sideW / 2), 0.22, partitionZ);
  const partR = partL.clone();
  partR.position.x = openingW / 2 + sideW / 2;
  const openBottom = -1.22;
  const openTop = openBottom + openingH;
  const wallTop = 0.22 + wallH / 2;
  const lintelH = Math.max(0.16, wallTop - openTop);
  const lintel = new THREE.Mesh(new THREE.BoxGeometry(openingW + 0.24, lintelH, 0.2), wallM);
  lintel.position.set(0, openTop + lintelH / 2, partitionZ);
  root.add(partL, partR, lintel);

  const frame = std(0x3a2818, 0.08, 0.7);
  const jambL = new THREE.Mesh(new THREE.BoxGeometry(0.1, openingH, 0.22), frame);
  jambL.position.set(-openingW / 2, -1.22 + openingH / 2, partitionZ);
  const jambR = jambL.clone();
  jambR.position.x = openingW / 2;
  const jambT = new THREE.Mesh(new THREE.BoxGeometry(openingW + 0.12, 0.1, 0.22), frame);
  jambT.position.set(0, -1.22 + openingH, partitionZ);
  root.add(jambL, jambR, jambT);

  const hingeX = -openingW / 2 + 0.04;
  const doorPivot = new THREE.Group();
  doorPivot.name = 'door';
  doorPivot.position.set(hingeX, -1.22 + openingH / 2, partitionZ + 0.02);
  const door = new THREE.Mesh(new THREE.BoxGeometry(openingW - 0.08, openingH - 0.08, 0.09), doorM);
  door.position.x = (openingW - 0.08) / 2;
  doorPivot.add(door);
  for (let i = 0; i < 3; i++) {
    const strap = new THREE.Mesh(new THREE.BoxGeometry(openingW - 0.12, 0.07, 0.11), rust);
    strap.position.set((openingW - 0.08) / 2, -0.7 + i * 0.7, 0);
    doorPivot.add(strap);
    const rivet = new THREE.Mesh(new THREE.SphereGeometry(0.035, 8, 8), iron);
    rivet.position.set(0.12, -0.7 + i * 0.7, 0.06);
    doorPivot.add(rivet);
  }
  const handle = new THREE.Mesh(new THREE.TorusGeometry(0.07, 0.018, 6, 12), iron);
  handle.position.set(openingW - 0.32, 0, 0.07);
  doorPivot.add(handle);
  const brace = new THREE.Mesh(new THREE.BoxGeometry(0.08, openingH - 0.2, 0.04), rust);
  brace.position.set(openingW * 0.55, 0, -0.04);
  doorPivot.add(brace);
  const poster = new THREE.Mesh(
    new THREE.PlaneGeometry(0.78, 1.12),
    new THREE.MeshStandardMaterial({ map: cursedPosterTex(), roughness: 0.82, metalness: 0.05 }),
  );
  poster.position.set((openingW - 0.08) / 2, 0.18, 0.055);
  doorPivot.add(poster);
  root.add(doorPivot);

  const chestBuilt = makeChest(crateM, gold, iron, std(0x2a1810, 0.1, 0.7));
  chestBuilt.root.position.set(0, -1.22, -4.55);
  root.add(chestBuilt.root);

  const dais = new THREE.Mesh(new THREE.BoxGeometry(1.5, 0.12, 1.0), crateM);
  dais.position.set(0, -1.22, -4.55);
  root.add(dais);

  const b1 = barrel(barrelM, rust);
  b1.position.set(-2.55, -0.95, -1.1);
  const b2 = barrel(barrelM, rust);
  b2.position.set(-2.7, -0.95, -0.45);
  b2.rotation.z = 0.08;
  const b3 = barrel(barrelM, rust);
  b3.position.set(2.6, -0.95, -3.6);
  const c1 = crate(crateM, rust);
  c1.position.set(-2.55, -1.05, -3.9);
  const c2 = crate(crateM, rust);
  c2.position.set(2.45, -1.05, -1.35);
  c2.rotation.y = 0.3;
  root.add(b1, b2, b3, c1, c2);

  const rope = new THREE.Mesh(new THREE.TorusGeometry(0.22, 0.05, 8, 16), ropeM);
  rope.rotation.x = Math.PI / 2;
  rope.position.set(2.55, -1.12, -0.7);
  root.add(rope);

  dressHold(root, { crateM, barrelM, rust, iron, gold, ropeM });

  const lantern = (x, y, z) => {
    const g = new THREE.Group();
    const cage = new THREE.Mesh(new THREE.CylinderGeometry(0.09, 0.1, 0.22, 8), iron);
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.07, 10, 8),
      new THREE.MeshStandardMaterial({
        color: 0xffe2a0, emissive: 0xffc14a, emissiveIntensity: 1.4, roughness: 0.4,
      }),
    );
    glow.position.y = 0.02;
    const chain = new THREE.Mesh(new THREE.CylinderGeometry(0.012, 0.012, 0.55, 6), iron);
    chain.position.y = 0.38;
    g.add(cage, glow, chain);
    const light = new THREE.PointLight(0xffc48a, 2.4, 9, 1.25);
    light.position.y = 0.02;
    g.add(light);
    g.position.set(x, y, z);
    root.add(g);
    return light;
  };
  const lights = [
    lantern(-1.15, 1.15, -1.2),
    lantern(1.35, 1.15, -3.9),
    lantern(-0.2, 1.2, -6.4),
  ];
  lights.forEach((l, i) => { l.userData.phase = i * 1.7; });

  const fill = new THREE.PointLight(0xc4a070, 0.85, 14, 1.25);
  fill.position.set(0, 0.8, -1.5);
  root.add(fill);
  const chestGlow = new THREE.PointLight(0xffd978, 0, 5, 1.6);
  chestGlow.position.set(0, -0.4, -4.4);
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
    punchChest() {
      this.hit = 1;
    },
    burstChest() {
      if (this.burst) return;
      this.burst = true;
      this.hit = 1.5;
      this.chestGlow.intensity = 10;
      if (chestBuilt.glowCore) chestBuilt.glowCore.visible = false;
      if (chestBuilt.glowHalo) chestBuilt.glowHalo.visible = false;
      chestBuilt.innerGold.intensity = 0;
      chestBuilt.innerCurse.intensity = 0;
      for (const it of chestBuilt.motes) it.m.visible = false;
      for (const it of chestBuilt.bits) {
        const dir = it.origin.clone().add(new THREE.Vector3(
          (Math.random() - 0.5) * 0.4,
          0.3,
          (Math.random() - 0.5) * 0.3,
        ));
        it.v.set(
          (dir.x || (Math.random() - 0.5)) * (2.5 + Math.random() * 3),
          2.4 + Math.random() * 3.2,
          -0.4 + Math.random() * 2.2,
        );
        it.spin.set((Math.random() - 0.5) * 8, (Math.random() - 0.5) * 10, (Math.random() - 0.5) * 8);
      }
      for (const coin of chestBuilt.loot) {
        coin.visible = true;
        coin.userData.v = new THREE.Vector3(
          (Math.random() - 0.5) * 4,
          2 + Math.random() * 3,
          (Math.random() - 0.5) * 3,
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
      chestBuilt.innerCurse.intensity = 1.4;
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
        l.intensity = (2.35 * dim) + Math.sin(this.t * 3.1 + (l.userData.phase || 0)) * 0.22 * dim;
      }
      if (!this.burst) {
        const pulse = 0.82 + Math.sin(this.t * 3.4) * 0.18;
        chestBuilt.innerGold.intensity = 2.4 * pulse;
        chestBuilt.innerCurse.intensity = 1.4 * pulse;
        chestBuilt.glowCore.scale.setScalar(0.9 + Math.sin(this.t * 4.2) * 0.12);
        chestBuilt.glowHalo.material.opacity = 0.14 + Math.sin(this.t * 2.6) * 0.06;
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

/** Lanterne de main gauche, lumière vers -Z (POV). */
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
