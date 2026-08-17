import * as THREE from 'three';

const KINDS = [
  { id: 'flash', file: 'assets/summoners/flash.png', metal: 0xf2d35e, emissive: 0x9a7a18, glow: 0.55 },
  { id: 'teleport', file: 'assets/summoners/teleport.png', metal: 0x4aa8ff, emissive: 0x163a8a, glow: 0.65 },
  { id: 'ignite', file: 'assets/summoners/ignite.png', metal: 0xff5a22, emissive: 0x8a1800, glow: 0.7 },
  { id: 'barrier', file: 'assets/summoners/barrier.png', metal: 0xffe566, emissive: 0x6a5810, glow: 0.5 },
];

export async function loadSummonerTextures() {
  const loader = new THREE.TextureLoader();
  const out = [];
  for (const k of KINDS) {
    const map = await new Promise((resolve, reject) => {
      loader.load(k.file, resolve, undefined, reject);
    });
    if (THREE.SRGBColorSpace) map.colorSpace = THREE.SRGBColorSpace;
    map.anisotropy = 8;
    out.push({ ...k, map });
  }
  return out;
}

/** Jeton 3D rayon 1, même pivot que la pièce, pour garder le spin. */
export function createSpellToken(kind) {
  const g = new THREE.Group();
  g.name = kind.id;

  const rim = new THREE.MeshStandardMaterial({
    color: kind.metal, metalness: 0.92, roughness: 0.28,
    emissive: kind.emissive, emissiveIntensity: kind.glow * 0.35,
  });
  const face = new THREE.MeshStandardMaterial({
    map: kind.map, metalness: 0.35, roughness: 0.4,
    emissive: kind.emissive, emissiveIntensity: kind.glow * 0.22,
  });
  const back = new THREE.MeshStandardMaterial({
    color: kind.metal, metalness: 0.85, roughness: 0.32,
    map: kind.map, emissive: kind.emissive, emissiveIntensity: 0.12,
  });

  const front = new THREE.Mesh(new THREE.CircleGeometry(0.92, 48), face);
  front.position.z = 0.055;
  const rear = new THREE.Mesh(new THREE.CircleGeometry(0.92, 48), back);
  rear.position.z = -0.055;
  rear.rotation.y = Math.PI;
  const body = new THREE.Mesh(new THREE.CylinderGeometry(1, 1, 0.1, 48), rim);
  body.rotation.x = Math.PI / 2;
  const torus = new THREE.Mesh(new THREE.TorusGeometry(1.0, 0.07, 10, 48), rim);
  g.add(body, torus, front, rear);

  if (kind.id === 'teleport') {
    const ring = new THREE.Mesh(
      new THREE.TorusGeometry(0.62, 0.045, 8, 32),
      new THREE.MeshStandardMaterial({
        color: 0x7fd4ff, metalness: 0.4, roughness: 0.2,
        emissive: 0x2a88ff, emissiveIntensity: 0.9,
      }),
    );
    ring.name = 'extra';
    g.add(ring);
  }
  if (kind.id === 'ignite') {
    const flame = new THREE.Group();
    flame.name = 'extra';
    const fire = new THREE.MeshStandardMaterial({
      color: 0xff6a1a, emissive: 0xff3300, emissiveIntensity: 1.1,
      metalness: 0.05, roughness: 0.45, transparent: true, opacity: 0.92,
    });
    for (let i = 0; i < 3; i++) {
      const cone = new THREE.Mesh(new THREE.ConeGeometry(0.16 - i * 0.03, 0.42 + i * 0.08, 7), fire);
      cone.position.set((i - 1) * 0.12, 0.55 + i * 0.04, 0);
      flame.add(cone);
    }
    g.add(flame);
  }
  if (kind.id === 'barrier') {
    const bubble = new THREE.Mesh(
      new THREE.SphereGeometry(1.08, 24, 16, 0, Math.PI * 2, 0, Math.PI * 0.55),
      new THREE.MeshStandardMaterial({
        color: 0xffe566, metalness: 0.15, roughness: 0.15,
        transparent: true, opacity: 0.28,
        emissive: 0xc4a020, emissiveIntensity: 0.4, side: THREE.DoubleSide,
      }),
    );
    bubble.rotation.x = -0.4;
    bubble.name = 'extra';
    g.add(bubble);
  }
  if (kind.id === 'flash') {
    const bolt = new THREE.Mesh(
      new THREE.OctahedronGeometry(0.22, 0),
      new THREE.MeshStandardMaterial({
        color: 0xfff4a8, emissive: 0xffe066, emissiveIntensity: 1.2, metalness: 0.2, roughness: 0.25,
      }),
    );
    bolt.position.set(0, 0, 0.22);
    bolt.name = 'extra';
    g.add(bolt);
  }
  return g;
}
