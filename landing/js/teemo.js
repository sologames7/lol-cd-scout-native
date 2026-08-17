import * as THREE from 'three';

function std(color, metal, rough, extra = {}) {
  return new THREE.MeshStandardMaterial({
    color, metalness: metal, roughness: rough, ...extra,
  });
}

export function createPinataTeemo(faceMap) {
  const root = new THREE.Group();
  root.position.set(0, 0.15, -3.15);

  const paper = std(0x2f6b32, 0.05, 0.72);
  const tan = std(0xd2b48c, 0.08, 0.65);
  const dark = std(0x1a3d1c, 0.05, 0.7);
  const gold = std(0xd4b24a, 0.85, 0.3);
  const brown = std(0x5a3a1a, 0.1, 0.7);

  const body = new THREE.Group();
  body.name = 'body';
  const torso = new THREE.Mesh(new THREE.SphereGeometry(0.62, 18, 14), paper);
  torso.scale.set(1, 1.15, 0.9);
  body.add(torso);
  const belly = new THREE.Mesh(new THREE.SphereGeometry(0.42, 12, 10), tan);
  belly.position.set(0, -0.15, 0.28);
  body.add(belly);

  const head = new THREE.Mesh(new THREE.SphereGeometry(0.48, 18, 14), tan);
  head.position.y = 0.72;
  body.add(head);

  if (faceMap) {
    const face = new THREE.Mesh(
      new THREE.CircleGeometry(0.38, 24),
      new THREE.MeshStandardMaterial({ map: faceMap, roughness: 0.55, metalness: 0.05 }),
    );
    face.position.set(0, 0.74, 0.36);
    body.add(face);
  }

  const earL = new THREE.Mesh(new THREE.ConeGeometry(0.16, 0.38, 8), tan);
  earL.position.set(-0.28, 1.12, 0);
  earL.rotation.z = 0.45;
  const earR = earL.clone();
  earR.position.x = 0.28;
  earR.rotation.z = -0.45;
  body.add(earL, earR);

  const goggle = new THREE.Mesh(new THREE.TorusGeometry(0.16, 0.03, 8, 16), gold);
  goggle.position.set(0, 0.78, 0.4);
  body.add(goggle);

  const gun = new THREE.Mesh(new THREE.CylinderGeometry(0.035, 0.04, 0.7, 8), brown);
  gun.rotation.z = Math.PI / 2;
  gun.position.set(0.55, 0.15, 0.2);
  body.add(gun);

  const fringeColors = [0x2f6b32, 0xc45c1a, 0xd4b24a, 0x8b1e1e, 0x3ee0c8];
  const pieces = [];
  const addPiece = (mesh) => {
    body.add(mesh);
    pieces.push({
      m: mesh,
      origin: mesh.position.clone(),
      rot0: mesh.rotation.clone(),
      v: new THREE.Vector3(),
      spin: new THREE.Vector3(),
    });
  };
  addPiece(torso);
  addPiece(belly);
  addPiece(head);
  addPiece(earL);
  addPiece(earR);
  addPiece(goggle);
  addPiece(gun);
  body.children.forEach((ch) => {
    if (!pieces.some((p) => p.m === ch)) {
      pieces.push({
        m: ch,
        origin: ch.position.clone(),
        rot0: ch.rotation.clone(),
        v: new THREE.Vector3(),
        spin: new THREE.Vector3(),
      });
    }
  });

  for (let i = 0; i < 18; i++) {
    const strip = new THREE.Mesh(
      new THREE.PlaneGeometry(0.08, 0.32 + Math.random() * 0.2),
      std(fringeColors[i % fringeColors.length], 0.05, 0.8, { side: THREE.DoubleSide }),
    );
    const a = (i / 18) * Math.PI * 2;
    strip.position.set(Math.cos(a) * 0.5, -0.55, Math.sin(a) * 0.35);
    strip.rotation.y = a;
    addPiece(strip);
  }

  const rope = new THREE.Mesh(new THREE.CylinderGeometry(0.02, 0.02, 1.6, 6), brown);
  rope.position.y = 1.7;
  root.add(rope);
  const knot = new THREE.Mesh(new THREE.SphereGeometry(0.06, 8, 8), gold);
  knot.position.y = 0.95;
  root.add(knot);

  body.position.y = 0.05;
  root.add(body);

  const confetti = [];
  const cGroup = new THREE.Group();
  root.add(cGroup);
  for (let i = 0; i < 36; i++) {
    const m = new THREE.Mesh(
      new THREE.BoxGeometry(0.08, 0.08, 0.02),
      std(fringeColors[i % fringeColors.length], 0.1, 0.5),
    );
    m.visible = false;
    cGroup.add(m);
    confetti.push({ m, v: new THREE.Vector3(), spin: new THREE.Vector3() });
  }

  return {
    root, body, pieces, confetti,
    hit: 0,
    broken: false,
    t: 0,
    punch() {
      this.hit = 1;
    },
    burst() {
      if (this.broken) return;
      this.broken = true;
      this.hit = 1.4;
      for (const it of this.pieces) {
        const dir = it.origin.clone().normalize();
        if (dir.lengthSq() < 0.01) dir.set(Math.random() - 0.5, 1, Math.random() - 0.5).normalize();
        it.v.set(
          dir.x * (3 + Math.random() * 4),
          2.2 + Math.random() * 4,
          dir.z * (2 + Math.random() * 3) + 1.5,
        );
        it.spin.set((Math.random() - 0.5) * 10, (Math.random() - 0.5) * 10, (Math.random() - 0.5) * 8);
      }
      for (const it of this.confetti) {
        it.m.visible = true;
        it.m.position.set(0, 0.4, 0);
        it.v.set((Math.random() - 0.5) * 6, 2 + Math.random() * 5, (Math.random() - 0.5) * 5);
        it.spin.set(Math.random() * 8, Math.random() * 8, Math.random() * 8);
      }
    },
    reset() {
      this.broken = false;
      this.hit = 0;
      this.body.visible = true;
      this.body.rotation.set(0, 0, 0);
      this.body.position.set(0, 0.05, 0);
      for (const it of this.pieces) {
        it.m.position.copy(it.origin);
        it.m.rotation.copy(it.rot0);
        it.m.visible = true;
        it.v.set(0, 0, 0);
      }
      for (const it of this.confetti) {
        it.m.visible = false;
        it.m.position.set(0, 0, 0);
      }
    },
    update(dt) {
      this.t += dt;
      this.hit *= Math.pow(0.08, dt);
      if (!this.broken) {
        const sway = Math.sin(this.t * 1.6) * 0.12 + Math.sin(this.t * 28) * this.hit * 0.35;
        this.body.rotation.z = sway;
        this.body.rotation.x = Math.sin(this.t * 22) * this.hit * 0.25;
        this.body.position.x = Math.sin(this.t * 26) * this.hit * 0.18;
        this.body.position.y = 0.05 + Math.abs(Math.sin(this.t * 20)) * this.hit * 0.12;
      } else {
        for (const it of this.pieces) {
          it.m.position.addScaledVector(it.v, dt);
          it.v.y -= 7.5 * dt;
          it.m.rotation.x += it.spin.x * dt;
          it.m.rotation.y += it.spin.y * dt;
          if (it.m.position.y < -5) it.m.visible = false;
        }
        for (const it of this.confetti) {
          if (!it.m.visible) continue;
          it.m.position.addScaledVector(it.v, dt);
          it.v.y -= 8 * dt;
          it.m.rotation.x += it.spin.x * dt;
          it.m.rotation.z += it.spin.z * dt;
          if (it.m.position.y < -5) it.m.visible = false;
        }
      }
    },
  };
}
