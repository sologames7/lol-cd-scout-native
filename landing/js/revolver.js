import * as THREE from 'three';

function std(color, metal, rough, extra = {}) {
  return new THREE.MeshStandardMaterial({
    color, metalness: metal, roughness: rough, ...extra,
  });
}

/** Revolver pirate façon Miss Fortune : canon vers +Z (caméra). */
export function createRevolver() {
  const root = new THREE.Group();
  root.name = 'mf-revolver';

  const gold = std(0xd4b24a, 1, 0.22);
  const goldHi = std(0xf3d78a, 1, 0.12);
  const brass = std(0x8a6424, 1, 0.34);
  const wood = std(0x3a2214, 0.12, 0.62);
  const iron = std(0x16161a, 0.88, 0.38);
  const teal = std(0x2ee6c5, 0.45, 0.22, {
    emissive: 0x14c4a8,
    emissiveIntensity: 1.35,
  });

  const add = (mesh, x, y, z) => {
    mesh.position.set(x, y, z);
    mesh.castShadow = false;
    mesh.receiveShadow = false;
    root.add(mesh);
    return mesh;
  };

  const cyl = (rt, rb, h, mat, seg = 14) =>
    new THREE.Mesh(new THREE.CylinderGeometry(rt, rb, h, seg), mat);
  const box = (w, h, d, mat) => new THREE.Mesh(new THREE.BoxGeometry(w, h, d), mat);

  const alongZ = (mesh) => {
    mesh.rotation.x = Math.PI / 2;
    return mesh;
  };

  add(alongZ(cyl(0.078, 0.09, 2.2, gold, 18)), 0, 0.02, 0.95);
  add(alongZ(cyl(0.055, 0.062, 1.75, brass, 12)), 0, -0.12, 0.72);
  add(alongZ(cyl(0.1, 0.1, 0.14, goldHi, 18)), 0, 0.02, 2.08);
  add(alongZ(cyl(0.028, 0.04, 0.1, teal, 10)), 0, 0.02, 2.16);

  const drum = add(alongZ(cyl(0.23, 0.23, 0.4, gold, 22)), 0, 0.0, -0.18);
  drum.name = 'cylinder';
  for (let i = 0; i < 6; i++) {
    const a = (i / 6) * Math.PI * 2;
    add(alongZ(cyl(0.058, 0.058, 0.44, iron, 8)), Math.cos(a) * 0.135, Math.sin(a) * 0.135, -0.18);
  }

  add(box(0.24, 0.34, 0.78, gold), 0, -0.03, -0.5);
  add(box(0.05, 0.22, 0.4, goldHi), 0.125, 0.04, -0.42);
  add(box(0.05, 0.22, 0.4, goldHi), -0.125, 0.04, -0.42);

  const hammer = add(box(0.07, 0.18, 0.2, brass), 0, 0.24, -0.78);
  hammer.rotation.x = -0.45;
  hammer.name = 'hammer';

  add(box(0.04, 0.13, 0.07, iron), 0, -0.24, -0.58);
  const guard = new THREE.Mesh(new THREE.TorusGeometry(0.13, 0.02, 8, 18, Math.PI), gold);
  guard.rotation.set(Math.PI / 2, 0, Math.PI / 2);
  add(guard, 0, -0.26, -0.55);

  const grip = add(box(0.17, 0.48, 0.3, wood), 0, -0.42, -0.95);
  grip.rotation.x = 0.48;
  add(box(0.19, 0.08, 0.32, gold), 0, -0.2, -0.82);

  add(box(0.035, 0.09, 0.09, gold), 0, 0.13, 1.72);
  add(alongZ(cyl(0.12, 0.12, 0.06, teal, 16)), 0, 0.02, -0.38);

  const bore = alongZ(cyl(0.042, 0.018, 0.55, iron, 12));
  add(bore, 0, 0.02, 1.92);
  const tip = new THREE.Object3D();
  tip.name = 'muzzle';
  tip.position.set(0, 0.02, 2.22);
  root.add(tip);

  return root;
}
