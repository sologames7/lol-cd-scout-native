import * as THREE from 'three';

function std(color, metal, rough, extra = {}) {
  return new THREE.MeshStandardMaterial({
    color, metalness: metal, roughness: rough, ...extra,
  });
}

/** Cimeterre pirate, lame vers -Z (POV, vers le mur). */
export function createKnife() {
  const root = new THREE.Group();
  root.name = 'knife';
  const gold = std(0xd4b24a, 1, 0.22);
  const steel = std(0xc5d0dc, 0.95, 0.18);
  const wood = std(0x3a2214, 0.12, 0.62);
  const iron = std(0x1a1a1e, 0.88, 0.38);

  const handle = new THREE.Mesh(new THREE.CylinderGeometry(0.045, 0.055, 0.42, 10), wood);
  handle.rotation.x = Math.PI / 2;
  handle.position.z = 0.12;
  root.add(handle);

  const pommel = new THREE.Mesh(new THREE.SphereGeometry(0.06, 10, 8), gold);
  pommel.position.z = 0.34;
  root.add(pommel);

  const guard = new THREE.Mesh(new THREE.BoxGeometry(0.28, 0.06, 0.05), gold);
  guard.position.z = -0.1;
  root.add(guard);

  const blade = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.02, 0.95), steel);
  blade.position.set(0.04, 0, -0.58);
  blade.rotation.y = -0.12;
  root.add(blade);
  const edge = new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.008, 0.9), iron);
  edge.position.set(0.09, 0, -0.58);
  edge.rotation.y = -0.18;
  root.add(edge);
  const curve = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.018, 0.28), steel);
  curve.position.set(0.1, 0, -1.02);
  curve.rotation.y = -0.45;
  root.add(curve);

  return root;
}
