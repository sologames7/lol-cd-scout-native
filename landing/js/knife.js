import * as THREE from 'three';

function std(color, metal, rough, extra = {}) {
  return new THREE.MeshStandardMaterial({
    color, metalness: metal, roughness: rough, ...extra,
  });
}

function scimitarGeo() {
  const s = new THREE.Shape();
  s.moveTo(0.0, 0.02);
  s.lineTo(0.055, 0.02);
  s.quadraticCurveTo(0.16, -0.28, 0.14, -0.72);
  s.quadraticCurveTo(0.08, -1.12, 0.0, -1.28);
  s.quadraticCurveTo(-0.02, -1.05, 0.012, -0.62);
  s.quadraticCurveTo(0.02, -0.28, 0.0, 0.02);
  return new THREE.ExtrudeGeometry(s, {
    depth: 0.016,
    bevelEnabled: true,
    bevelThickness: 0.004,
    bevelSize: 0.003,
    bevelSegments: 2,
    curveSegments: 18,
  });
}

/** Épée de Samira : cimeterre orné, lame vers -Z. */
export function createKnife() {
  const root = new THREE.Group();
  root.name = 'samira-blade';

  const gold = std(0xe0c35a, 1, 0.18);
  const goldHi = std(0xf3d78a, 1, 0.12, { emissive: 0x6a4810, emissiveIntensity: 0.15 });
  const steel = std(0xd8e0ea, 0.96, 0.16);
  const edge = std(0xf4f7ff, 1, 0.08, { emissive: 0x8aa0b8, emissiveIntensity: 0.25 });
  const wrap = std(0x6b1c28, 0.08, 0.62);
  const iron = std(0x1a1a22, 0.85, 0.38);
  const teal = std(0x3ee0c8, 0.45, 0.22, { emissive: 0x14c4a8, emissiveIntensity: 1.1 });

  const blade = new THREE.Mesh(scimitarGeo(), steel);
  blade.rotation.set(Math.PI / 2, 0, 0);
  blade.position.set(0, 0, -0.04);
  root.add(blade);

  const edgeMesh = new THREE.Mesh(scimitarGeo(), edge);
  edgeMesh.rotation.set(Math.PI / 2, 0, 0);
  edgeMesh.position.set(0.006, 0, -0.04);
  edgeMesh.scale.set(0.92, 1.01, 0.35);
  root.add(edgeMesh);

  const guard = new THREE.Mesh(new THREE.BoxGeometry(0.34, 0.07, 0.055), gold);
  guard.position.z = 0.02;
  root.add(guard);
  const ring = new THREE.Mesh(new THREE.TorusGeometry(0.055, 0.012, 8, 18), goldHi);
  ring.rotation.y = Math.PI / 2;
  ring.position.set(-0.16, 0, 0.02);
  root.add(ring);
  const gem = new THREE.Mesh(new THREE.OctahedronGeometry(0.028, 0), teal);
  gem.position.set(0, 0.045, 0.02);
  root.add(gem);

  const handle = new THREE.Mesh(new THREE.CylinderGeometry(0.038, 0.048, 0.38, 12), wrap);
  handle.rotation.x = Math.PI / 2;
  handle.position.z = 0.24;
  root.add(handle);
  for (let i = 0; i < 5; i++) {
    const band = new THREE.Mesh(new THREE.TorusGeometry(0.046, 0.008, 6, 12), gold);
    band.position.z = 0.1 + i * 0.07;
    root.add(band);
  }

  const pommel = new THREE.Mesh(new THREE.SphereGeometry(0.055, 12, 10), goldHi);
  pommel.position.z = 0.44;
  root.add(pommel);
  const cap = new THREE.Mesh(new THREE.OctahedronGeometry(0.03, 0), teal);
  cap.position.z = 0.5;
  root.add(cap);

  const trail = new THREE.Mesh(
    new THREE.TorusGeometry(0.95, 0.035, 6, 28, Math.PI * 0.85),
    new THREE.MeshBasicMaterial({
      color: 0xe8fff8, transparent: true, opacity: 0, depthWrite: false, side: THREE.DoubleSide,
    }),
  );
  trail.rotation.set(0.2, 1.15, 0.4);
  trail.position.set(-0.15, 0.05, -0.55);
  trail.name = 'slash';
  root.add(trail);

  return root;
}
