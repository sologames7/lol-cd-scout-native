import * as THREE from 'three';

function std(color, metal, rough, extra = {}) {
  return new THREE.MeshStandardMaterial({
    color, metalness: metal, roughness: rough, ...extra,
  });
}

const pal = {
  gold: std(0xd4b24a, 1, 0.22),
  goldHi: std(0xf3d78a, 1, 0.12, { emissive: 0x6a4810, emissiveIntensity: 0.2 }),
  steel: std(0xc8d0d8, 0.95, 0.18),
  dark: std(0x1a1c22, 0.75, 0.4),
  iron: std(0x3a3a42, 0.88, 0.35),
  leather: std(0x5a3218, 0.08, 0.7),
  wood: std(0x6a4424, 0.08, 0.75),
  crimson: std(0x8a1820, 0.25, 0.45, { emissive: 0x4a0808, emissiveIntensity: 0.25 }),
  redG: std(0xcc3333, 0.2, 0.25, { transparent: true, opacity: 0.78, emissive: 0x661010, emissiveIntensity: 0.35 }),
  blueG: std(0x3a88cc, 0.2, 0.2, { transparent: true, opacity: 0.75, emissive: 0x104066, emissiveIntensity: 0.4 }),
  greenG: std(0x3aaa58, 0.2, 0.25, { transparent: true, opacity: 0.78, emissive: 0x145020, emissiveIntensity: 0.35 }),
  purpleG: std(0x7a3acc, 0.25, 0.2, { transparent: true, opacity: 0.72, emissive: 0x3a1066, emissiveIntensity: 0.45 }),
  ice: std(0xa8d8ff, 0.35, 0.12, { transparent: true, opacity: 0.7, emissive: 0x4a88cc, emissiveIntensity: 0.4 }),
  fire: std(0xff6a1a, 0.15, 0.3, { emissive: 0xff3300, emissiveIntensity: 0.85 }),
  voidC: std(0x4a1080, 0.4, 0.18, { emissive: 0x2a0848, emissiveIntensity: 0.7 }),
  bone: std(0xe8d8c0, 0.08, 0.55),
  cloth: std(0x4a3a58, 0.05, 0.8),
};

function put(g, mesh, x = 0, y = 0, z = 0, rx = 0, ry = 0, rz = 0) {
  mesh.position.set(x, y, z);
  mesh.rotation.set(rx, ry, rz);
  g.add(mesh);
  return mesh;
}

function potion(glass, cork = pal.wood) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.13, 12, 10), glass), 0, -0.02, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.1, 0.12, 0.16, 10), glass), 0, 0.1, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.04, 0.055, 0.08, 8), glass), 0, 0.22, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.045, 0.05, 0.05, 8), cork), 0, 0.28, 0);
  return g;
}

function elixir(glass) {
  const g = potion(glass, pal.gold);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.08, 0.012, 6, 14), pal.gold), 0, 0.14, 0, Math.PI / 2);
  return g;
}

function bootPair(foot, cuff, extra) {
  const g = new THREE.Group();
  const one = (x) => {
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.11, 0.09, 0.22), foot), x, -0.12, 0.04);
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.1, 0.18, 0.1), cuff), x, 0.02, -0.04);
  };
  one(-0.08);
  one(0.08);
  if (extra) extra(g);
  return g;
}

function sword({ blade = pal.steel, guard = pal.gold, hilt = pal.leather, len = 0.62, w = 0.07 } = {}) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.BoxGeometry(w, len, 0.018), blade), 0, 0.16, 0);
  put(g, new THREE.Mesh(new THREE.ConeGeometry(w * 0.55, 0.1, 4), blade), 0, 0.16 + len / 2 + 0.04, 0, 0, 0, Math.PI / 4);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.22, 0.035, 0.055), guard), 0, -0.16, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.028, 0.032, 0.16, 8), hilt), 0, -0.26, 0);
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.035, 8, 6), guard), 0, -0.35, 0);
  return g;
}

function axe(head, haft = pal.wood) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.025, 0.03, 0.7, 8), haft), 0, 0.05, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.28, 0.16, 0.05), head), 0.12, 0.28, 0);
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.1, 0.18, 4), head), 0.28, 0.28, 0, 0, 0, -Math.PI / 2);
  return g;
}

function staff(shaft, gem, gemY = 0.38) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.022, 0.028, 0.72, 8), shaft), 0, 0.05, 0);
  put(g, new THREE.Mesh(new THREE.OctahedronGeometry(0.1, 0), gem), 0, gemY, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.12, 0.012, 6, 16), pal.gold), 0, gemY, 0, Math.PI / 2);
  return g;
}

function book(cover) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.28, 0.06, 0.36), cover), 0, 0, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.24, 0.05, 0.32), pal.bone), 0, 0.02, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.07, 0.36), pal.gold), -0.12, 0, 0);
  return g;
}

function shield(face, boss = pal.gold) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.22, 0.24, 0.05, 8), face), 0, 0, 0, Math.PI / 2);
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.07, 10, 8), boss), 0, 0, 0.04);
  return g;
}

function hat(felt, band = pal.gold) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.16, 0.42, 10), felt), 0, 0.18, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.22, 0.22, 0.03, 12), felt), 0, -0.02, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.12, 0.02, 6, 14), band), 0, 0.04, 0, Math.PI / 2);
  put(g, new THREE.Mesh(new THREE.OctahedronGeometry(0.05, 0), pal.goldHi), 0, 0.38, 0);
  return g;
}

function hourglass() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.12, 0.04, 0.16, 8), pal.gold), 0, 0.12, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.04, 0.12, 0.16, 8), pal.gold), 0, -0.12, 0);
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.045, 8, 6), pal.fire), 0, 0.06, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.22, 0.03, 0.22), pal.gold), 0, 0.22, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.22, 0.03, 0.22), pal.gold), 0, -0.22, 0);
  return g;
}

function angel() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.07, 10, 8), pal.bone), 0, 0.22, 0);
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.1, 0.28, 8), pal.gold), 0, 0.02, 0);
  const wing = (x, ry) => {
    const w = new THREE.Mesh(new THREE.BoxGeometry(0.18, 0.04, 0.28), pal.bone);
    w.position.set(x, 0.12, -0.02);
    w.rotation.y = ry;
    g.add(w);
  };
  wing(-0.14, 0.5);
  wing(0.14, -0.5);
  return g;
}

function gauntlet() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.22, 0.28), pal.steel), 0, 0.02, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.18, 0.08, 0.12), pal.gold), 0, 0.14, 0.08);
  for (let i = 0; i < 4; i++) {
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.03, 0.04, 0.1), pal.steel), -0.06 + i * 0.04, -0.12, 0.1);
  }
  return g;
}

function spear(head = pal.steel) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.018, 0.022, 0.78, 8), pal.wood), 0, 0.02, 0);
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.055, 0.22, 5), head), 0, 0.48, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.02, 0.02), pal.gold), 0, 0.34, 0);
  return g;
}

function bow() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.26, 0.018, 6, 18, Math.PI), pal.wood), 0, 0, 0, 0, 0, Math.PI / 2);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.006, 0.006, 0.5, 6), pal.bone), 0, 0, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.008, 0.012, 0.42, 6), pal.steel), 0.08, 0.04, 0, 0, 0, -0.4);
  return g;
}

function cannon() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.07, 0.09, 0.42, 10), pal.iron), 0, 0.04, 0, 0, 0, Math.PI / 2);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.1, 0.22), pal.wood), 0, -0.12, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.09, 0.015, 6, 12), pal.gold), 0.2, 0.04, 0, 0, Math.PI / 2);
  return g;
}

function totem(crystal) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.025, 0.035, 0.42, 8), pal.wood), 0, -0.04, 0);
  put(g, new THREE.Mesh(new THREE.OctahedronGeometry(0.11, 0), crystal), 0, 0.22, 0);
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.04, 8, 6), pal.gold), 0, 0.22, 0.08);
  return g;
}

function ring() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.1, 0.028, 8, 18), pal.gold), 0, 0, 0, Math.PI / 2);
  put(g, new THREE.Mesh(new THREE.OctahedronGeometry(0.05, 0), pal.blueG), 0, 0.08, 0);
  return g;
}

function crystal(mat, s = 0.16) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.OctahedronGeometry(s, 0), mat));
  return g;
}

function cloak(col) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.18, 0.42, 8, 1, true), col), 0, 0.02, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.1, 0.02, 6, 12), pal.gold), 0, 0.2, 0, Math.PI / 2);
  return g;
}

function vest(mat) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.32, 0.36, 0.12), mat), 0, 0.04, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.34, 0.05, 0.14), pal.gold), 0, 0.14, 0);
  return g;
}

function pickaxe() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.022, 0.028, 0.55, 8), pal.wood), 0, 0, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.42, 0.08, 0.06), pal.iron), 0, 0.26, 0);
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.05, 0.14, 4), pal.iron), 0.22, 0.26, 0, 0, 0, -Math.PI / 2);
  return g;
}

function hammer() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.025, 0.03, 0.55, 8), pal.wood), 0, 0, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.14, 0.28), pal.iron), 0, 0.28, 0);
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.04, 0.3), pal.gold), 0, 0.28, 0);
  return g;
}

function scythe() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.02, 0.025, 0.7, 8), pal.dark), 0, 0.05, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.22, 0.03, 6, 14, Math.PI * 0.7), pal.crimson), 0.12, 0.38, 0, 0, 0, -0.4);
  return g;
}

function dagger(blade = pal.steel) {
  const g = sword({ blade, guard: pal.iron, hilt: pal.leather, len: 0.32, w: 0.045 });
  g.scale.setScalar(0.85);
  return g;
}

function medallion() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.14, 0.14, 0.03, 16), pal.goldHi), 0, 0.04, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.15, 0.018, 8, 18), pal.gold), 0, 0.04, 0, Math.PI / 2);
  put(g, new THREE.Mesh(new THREE.OctahedronGeometry(0.05, 0), pal.fire), 0, 0.06, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.04, 0.008, 6, 10), pal.gold), 0, 0.2, 0);
  return g;
}

function chalice() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.1, 0.04, 0.16, 10), pal.gold), 0, 0.1, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.03, 0.03, 0.12, 8), pal.gold), 0, -0.04, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.08, 0.08, 0.03, 10), pal.gold), 0, -0.12, 0);
  put(g, new THREE.Mesh(new THREE.CircleGeometry(0.08, 12), pal.blueG), 0, 0.18, 0, -Math.PI / 2);
  return g;
}

function mask(col) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.14, 12, 10, 0, Math.PI * 2, 0, Math.PI / 1.6), col), 0, 0.06, 0);
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.03, 8, 6), pal.goldHi), -0.05, 0.1, 0.1);
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.03, 8, 6), pal.goldHi), 0.05, 0.1, 0.1);
  return g;
}

function fang() {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.06, 0.42, 6), pal.bone), 0, 0.08, 0);
  put(g, new THREE.Mesh(new THREE.TorusGeometry(0.05, 0.015, 6, 10), pal.crimson), 0, -0.12, 0, Math.PI / 2);
  return g;
}

function heart(mat) {
  const g = new THREE.Group();
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.1, 10, 8), mat), -0.05, 0.04, 0);
  put(g, new THREE.Mesh(new THREE.SphereGeometry(0.1, 10, 8), mat), 0.05, 0.04, 0);
  put(g, new THREE.Mesh(new THREE.ConeGeometry(0.12, 0.18, 4), mat), 0, -0.1, 0, Math.PI, 0, Math.PI / 4);
  return g;
}

const BY_ID = {
  2003: () => potion(pal.redG),
  2031: () => potion(pal.greenG),
  2055: () => totem(pal.crimson),
  2138: () => elixir(std(0x8a8a90, 0.4, 0.25, { transparent: true, opacity: 0.8 })),
  2139: () => elixir(pal.blueG),
  2140: () => elixir(pal.redG),
  3340: () => totem(pal.greenG),
  3363: () => {
    const g = new THREE.Group();
    put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.08, 0.1, 0.12, 10), pal.gold), 0, 0, 0);
    put(g, new THREE.Mesh(new THREE.CircleGeometry(0.07, 16), pal.blueG), 0, 0, 0.07);
    put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.02, 0.02, 0.18, 6), pal.dark), 0, -0.14, 0);
    return g;
  },
  3364: () => {
    const g = totem(pal.redG);
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.08, 0.012, 6, 12), pal.gold), 0, 0.22, 0);
    return g;
  },
  1001: () => bootPair(pal.leather, pal.leather),
  3006: () => bootPair(pal.steel, pal.dark, (g) => {
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.12, 0.02), pal.steel), -0.08, 0.12, 0.02);
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.12, 0.02), pal.steel), 0.08, 0.12, 0.02);
  }),
  3009: () => bootPair(pal.leather, pal.gold, (g) => {
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.02, 0.22), pal.gold), -0.08, 0.12, -0.02, 0.4);
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.02, 0.22), pal.gold), 0.08, 0.12, -0.02, 0.4);
  }),
  3020: () => bootPair(pal.cloth, pal.purpleG, (g) => {
    put(g, new THREE.Mesh(new THREE.OctahedronGeometry(0.04, 0), pal.purpleG), 0, 0.16, 0);
  }),
  3047: () => bootPair(pal.steel, pal.iron, (g) => {
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.08, 0.12), pal.steel), -0.08, 0.08, 0.02);
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.08, 0.12), pal.steel), 0.08, 0.08, 0.02);
  }),
  3111: () => bootPair(pal.iron, pal.gold, (g) => {
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.06, 0.012, 6, 10), pal.gold), -0.08, 0.1, 0.06, Math.PI / 2);
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.06, 0.012, 6, 10), pal.gold), 0.08, 0.1, 0.06, Math.PI / 2);
  }),
  3117: () => bootPair(pal.leather, pal.dark),
  3158: () => bootPair(pal.leather, pal.goldHi),
  1036: () => sword({ len: 0.5, w: 0.055, guard: pal.iron }),
  1037: () => pickaxe(),
  1038: () => sword({ len: 0.78, w: 0.1, guard: pal.goldHi, blade: pal.steel }),
  1042: () => dagger(),
  1028: () => crystal(pal.crimson, 0.16),
  1029: () => vest(pal.leather),
  1031: () => vest(pal.steel),
  1033: () => cloak(pal.cloth),
  1018: () => cloak(std(0x6a88aa, 0.1, 0.7)),
  1026: () => staff(pal.wood, pal.purpleG),
  1052: () => book(pal.cloth),
  1058: () => staff(pal.dark, pal.goldHi, 0.42),
  1053: () => staff(pal.wood, pal.crimson),
  1054: () => shield(pal.wood, pal.iron),
  1055: () => sword({ len: 0.48, guard: pal.gold, hilt: pal.wood }),
  1056: () => ring(),
  3057: () => sword({ blade: std(0xe8f0ff, 0.4, 0.12, { emissive: 0xaad4ff, emissiveIntensity: 0.7 }), guard: pal.goldHi }),
  3067: () => heart(pal.gold),
  3108: () => book(pal.crimson),
  3133: () => hammer(),
  3134: () => dagger(pal.iron),
  3140: () => cloak(std(0xc8d0d8, 0.55, 0.3)),
  3031: () => {
    const g = sword({ blade: pal.goldHi, guard: pal.gold, len: 0.7, w: 0.08 });
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.28, 0.04, 0.12), pal.gold), 0, -0.12, 0, 0, 0, 0.4);
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.28, 0.04, 0.12), pal.gold), 0, -0.12, 0, 0, 0, -0.4);
    return g;
  },
  3089: () => hat(pal.voidC, pal.goldHi),
  3157: () => hourglass(),
  3026: () => angel(),
  3071: () => axe(pal.dark, pal.iron),
  3072: () => sword({ blade: pal.crimson, guard: pal.gold, len: 0.68 }),
  3074: () => {
    const g = new THREE.Group();
    put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.025, 0.03, 0.55, 8), pal.dark), 0, 0, 0);
    for (let i = 0; i < 3; i++) {
      const a = (i - 1) * 0.55;
      put(g, new THREE.Mesh(new THREE.ConeGeometry(0.06, 0.28, 5), pal.steel), Math.sin(a) * 0.08, 0.32, Math.cos(a) * 0.04, 0, 0, a);
    }
    return g;
  },
  3075: () => {
    const g = vest(pal.steel);
    for (let i = 0; i < 8; i++) {
      const a = (i / 8) * Math.PI * 2;
      put(g, new THREE.Mesh(new THREE.ConeGeometry(0.02, 0.1, 4), pal.iron), Math.cos(a) * 0.16, 0.04, Math.sin(a) * 0.08, 0, 0, a);
    }
    return g;
  },
  3078: () => {
    const g = new THREE.Group();
    put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.03, 0.03, 0.2, 8), pal.gold), 0, -0.1, 0);
    for (let i = 0; i < 3; i++) {
      const a = (i / 3) * Math.PI * 2;
      put(g, new THREE.Mesh(new THREE.BoxGeometry(0.05, 0.42, 0.02), pal.steel), Math.cos(a) * 0.06, 0.16, Math.sin(a) * 0.06, 0, a);
    }
    return g;
  },
  3085: () => bow(),
  3094: () => cannon(),
  3115: () => {
    const g = staff(pal.bone, pal.greenG);
    put(g, new THREE.Mesh(new THREE.ConeGeometry(0.05, 0.18, 5), pal.bone), 0.08, 0.32, 0, 0, 0, 1.1);
    return g;
  },
  3116: () => staff(pal.gold, pal.ice),
  3135: () => staff(pal.voidC, pal.purpleG),
  3142: () => sword({
    blade: std(0xa8c8e8, 0.3, 0.15, { transparent: true, opacity: 0.55, emissive: 0x4a88cc, emissiveIntensity: 0.5 }),
    guard: pal.dark,
  }),
  3153: () => sword({ blade: pal.greenG, guard: pal.gold, hilt: pal.leather, len: 0.6 }),
  3165: () => book(pal.voidC),
  3190: () => medallion(),
  3068: () => {
    const g = shield(pal.fire, pal.gold);
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.2, 0.02, 6, 16), pal.fire), 0, 0, 0.02);
    return g;
  },
  3742: () => vest(pal.iron),
  3053: () => gauntlet(),
  6672: () => {
    const g = spear(pal.gold);
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.08, 0.02, 6, 10), pal.crimson), 0.08, 0.2, 0, 0.6);
    return g;
  },
  3508: () => sword({ blade: pal.goldHi, guard: pal.voidC, len: 0.64 }),
  3102: () => mask(pal.purpleG),
  3110: () => heart(pal.ice),
  3143: () => {
    const g = shield(pal.steel, pal.iron);
    for (let i = 0; i < 6; i++) {
      const a = (i / 6) * Math.PI * 2;
      put(g, new THREE.Mesh(new THREE.ConeGeometry(0.025, 0.1, 4), pal.iron), Math.cos(a) * 0.2, Math.sin(a) * 0.2, 0.02, 0, 0, a);
    }
    return g;
  },
  3065: () => mask(pal.gold),
  3107: () => chalice(),
  6653: () => {
    const g = book(pal.crimson);
    put(g, new THREE.Mesh(new THREE.ConeGeometry(0.04, 0.14, 6), pal.fire), 0, 0.16, 0);
    return g;
  },
  6655: () => {
    const g = new THREE.Group();
    put(g, new THREE.Mesh(new THREE.SphereGeometry(0.14, 14, 12), pal.blueG));
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.18, 0.02, 8, 18), pal.gold), 0, 0, 0, 0.6);
    return g;
  },
  4645: () => crystal(pal.voidC, 0.18),
  4633: () => {
    const g = crystal(pal.purpleG, 0.14);
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.16, 0.015, 6, 16), pal.voidC), 0, 0, 0, 0.8);
    return g;
  },
  6692: () => sword({ blade: pal.dark, guard: pal.goldHi, len: 0.62 }),
  6694: () => spear(pal.ice),
  6695: () => fang(),
  6333: () => scythe(),
  3814: () => {
    const g = dagger(pal.dark);
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.18, 0.03), pal.voidC), 0.12, 0.02, 0);
    return g;
  },
  3139: () => sword({ blade: pal.steel, guard: pal.gold, len: 0.55, w: 0.06 }),
  3046: () => {
    const g = new THREE.Group();
    const a = dagger(pal.steel);
    a.position.x = -0.08;
    a.rotation.z = 0.3;
    const b = dagger(pal.steel);
    b.position.x = 0.08;
    b.rotation.z = -0.3;
    g.add(a, b);
    return g;
  },
  3036: () => spear(pal.steel),
  3033: () => axe(pal.crimson),
  3161: () => spear(pal.goldHi),
  3124: () => {
    const g = sword({ blade: pal.crimson, guard: pal.gold, len: 0.5 });
    put(g, new THREE.Mesh(new THREE.BoxGeometry(0.05, 0.4, 0.016), pal.steel), 0.06, 0.12, 0);
    return g;
  },
  4005: () => staff(pal.gold, pal.goldHi),
  4401: () => {
    const g = shield(pal.greenG, pal.wood);
    put(g, new THREE.Mesh(new THREE.TorusGeometry(0.18, 0.02, 6, 12), pal.greenG), 0, 0, 0.03);
    return g;
  },
  6617: () => crystal(pal.ice, 0.15),
  3091: () => {
    const g = new THREE.Group();
    put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.025, 0.03, 0.45, 8), pal.wood), 0, -0.05, 0);
    put(g, new THREE.Mesh(new THREE.IcosahedronGeometry(0.12, 0), pal.blueG), 0, 0.22, 0);
    return g;
  },
};

function fallback(kind) {
  if (kind.rarity === 'consumable') return potion(pal.redG);
  if (kind.rarity === 'trinket') return totem(pal.greenG);
  if (kind.rarity === 'boots') return bootPair(pal.leather, pal.leather);
  if ((kind.tags || []).includes('SpellDamage')) return staff(pal.wood, pal.purpleG);
  if ((kind.tags || []).includes('Armor')) return shield(pal.steel);
  if ((kind.tags || []).includes('Damage')) return sword();
  return crystal(pal.gold, 0.14);
}

export function createPriceTag(gold) {
  const g = new THREE.Group();
  g.name = 'price';
  put(g, new THREE.Mesh(new THREE.BoxGeometry(0.38, 0.12, 0.03), pal.wood), 0, 0, 0);
  put(g, new THREE.Mesh(new THREE.CylinderGeometry(0.045, 0.045, 0.012, 12), pal.goldHi), -0.12, 0, 0.02, Math.PI / 2);
  const c = document.createElement('canvas');
  c.width = 256;
  c.height = 64;
  const ctx = c.getContext('2d');
  ctx.clearRect(0, 0, 256, 64);
  ctx.fillStyle = '#f3d78a';
  ctx.font = '700 44px "Chakra Petch", sans-serif';
  ctx.textAlign = 'left';
  ctx.textBaseline = 'middle';
  ctx.fillText(gold > 0 ? String(gold) : '—', 8, 34);
  const map = new THREE.CanvasTexture(c);
  if (THREE.SRGBColorSpace) map.colorSpace = THREE.SRGBColorSpace;
  const label = new THREE.Mesh(
    new THREE.PlaneGeometry(0.22, 0.055),
    new THREE.MeshBasicMaterial({ map, transparent: true, depthWrite: false }),
  );
  label.position.set(0.08, 0, 0.02);
  g.add(label);
  return g;
}

export function createItemModel(kind) {
  const fn = BY_ID[kind.id] || BY_ID[String(kind.id)];
  const model = fn ? fn(kind) : fallback(kind);
  model.name = `item-${kind.id}`;
  return model;
}

export const SHELF_IDS = Object.keys(BY_ID);

/** 29 pièces : 5 sur le comptoir, 24 en étagères. */
export const ON_SHELF = [
  '3031', '3089', '3157', '3026', '3078',
  '2003', '2055', '3340', '3006', '3047',
  '1038', '1058', '3071', '3074', '3072',
  '3153', '3142', '3116', '3135', '3165',
  '3190', '3068', '6672', '3053', '6333',
  '6653', '3115', '3508', '6692',
];
