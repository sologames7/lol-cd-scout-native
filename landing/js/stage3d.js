import * as THREE from 'three';

const HREF = 'https://github.com/sologames7/lol-cd-scout-native/releases/latest';

function texOpt(t) {
  if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
  t.anisotropy = 8;
  t.needsUpdate = true;
  return t;
}

function canvasTex(w, h, draw) {
  const c = document.createElement('canvas');
  c.width = w;
  c.height = h;
  const ctx = c.getContext('2d');
  draw(ctx, w, h);
  return texOpt(new THREE.CanvasTexture(c));
}

function punchBlack(img) {
  const c = document.createElement('canvas');
  c.width = img.naturalWidth || img.width;
  c.height = img.naturalHeight || img.height;
  const ctx = c.getContext('2d');
  ctx.drawImage(img, 0, 0);
  const data = ctx.getImageData(0, 0, c.width, c.height);
  const px = data.data;
  for (let i = 0; i < px.length; i += 4) {
    const r = px[i];
    const g = px[i + 1];
    const b = px[i + 2];
    if (r < 22 && g < 22 && b < 22) px[i + 3] = 0;
  }
  ctx.putImageData(data, 0, 0);
  return texOpt(new THREE.CanvasTexture(c));
}

function loadImage(url) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = reject;
    img.src = url;
  });
}

function std(color, metal, rough, extra = {}) {
  return new THREE.MeshStandardMaterial({
    color, metalness: metal, roughness: rough, ...extra,
  });
}

function makeCardBody(w, h, t) {
  const g = new THREE.Group();

  const frame = new THREE.Mesh(
    new THREE.BoxGeometry(w + 0.07, h + 0.07, t * 0.78),
    std(0xc9a24a, 1, 0.22, { emissive: 0x1a6a58, emissiveIntensity: 0.22 }),
  );
  frame.position.z = -0.006;
  g.add(frame);

  const body = new THREE.Mesh(
    new THREE.BoxGeometry(w, h, t),
    std(0x0a1216, 0.78, 0.26, { emissive: 0x072420, emissiveIntensity: 0.45 }),
  );
  g.add(body);

  const plate = new THREE.Mesh(
    new THREE.PlaneGeometry(w * 0.92, h * 0.92),
    std(0x071014, 0.35, 0.48, { emissive: 0x0a2e28, emissiveIntensity: 0.55 }),
  );
  plate.position.z = t / 2 + 0.002;
  g.add(plate);

  const rim = new THREE.Mesh(
    new THREE.PlaneGeometry(w * 0.94, h * 0.94),
    new THREE.MeshBasicMaterial({
      color: 0x3ee0c8, transparent: true, opacity: 0.14, depthWrite: false,
    }),
  );
  rim.position.z = t / 2 + 0.001;
  g.add(rim);

  const backMark = new THREE.Mesh(
    new THREE.CircleGeometry(0.38, 48),
    new THREE.MeshBasicMaterial({
      color: 0x3ee0c8, transparent: true, opacity: 0.22, side: THREE.BackSide,
    }),
  );
  backMark.position.z = -t / 2 - 0.002;
  g.add(backMark);

  return g;
}

export async function makeLandingStage(scene) {
  await (document.fonts ? document.fonts.ready : Promise.resolve());

  const [logoImg, btnImg] = await Promise.all([
    loadImage('assets/logo.png'),
    loadImage('assets/btn-download.png'),
  ]);
  const logoMap = punchBlack(logoImg);
  const btnMap = punchBlack(btnImg);

  const root = new THREE.Group();
  root.name = 'landing3d';
  scene.add(root);

  const W = 1.58;
  const H = 2.12;
  const T = 0.1;
  const card = makeCardBody(W, H, T);
  card.name = 'cd-card';
  root.add(card);

  const logo = new THREE.Mesh(
    new THREE.CircleGeometry(0.54, 64),
    new THREE.MeshBasicMaterial({
      map: logoMap, transparent: true, depthWrite: false, toneMapped: false,
    }),
  );
  logo.position.set(0, 0.34, T / 2 + 0.014);
  card.add(logo);

  const wordTex = canvasTex(1024, 320, (ctx, w, h) => {
    ctx.clearRect(0, 0, w, h);
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.shadowColor = '#3ee0c8';
    ctx.shadowBlur = 24;
    ctx.fillStyle = '#3ee0c8';
    ctx.font = '700 86px "Space Grotesk", sans-serif';
    ctx.fillText('CD SCOUT', w / 2, h * 0.36);
    ctx.shadowBlur = 0;
    ctx.fillStyle = '#eef6f4';
    ctx.font = '600 34px "Space Grotesk", sans-serif';
    ctx.fillText('Progresser. Apprendre.', w / 2, h * 0.72);
  });
  const word = new THREE.Mesh(
    new THREE.PlaneGeometry(1.28, 0.4),
    new THREE.MeshBasicMaterial({
      map: wordTex, transparent: true, depthWrite: false, toneMapped: false,
    }),
  );
  word.position.set(0, -0.34, T / 2 + 0.014);
  card.add(word);

  const btn = new THREE.Group();
  btn.name = 'cta';
  const btnW = 1.18;
  const btnH = 0.34;
  const btnT = 0.055;
  const btnBody = new THREE.Mesh(
    new THREE.BoxGeometry(btnW, btnH, btnT),
    std(0x123a34, 0.62, 0.28, { emissive: 0x1c6e5e, emissiveIntensity: 0.7 }),
  );
  const btnFace = new THREE.Mesh(
    new THREE.PlaneGeometry(btnW * 1.02, btnH * 1.08),
    new THREE.MeshBasicMaterial({
      map: btnMap, transparent: true, alphaTest: 0.08, toneMapped: false,
    }),
  );
  btnFace.position.z = btnT / 2 + 0.003;
  btn.add(btnBody, btnFace);
  btn.position.set(0, -0.82, T / 2 + 0.036);
  btn.userData.href = HREF;
  btn.userData.clickable = true;
  card.userData.href = HREF;
  card.userData.clickable = true;
  card.add(btn);

  const glow = new THREE.PointLight(0x3ee0c8, 0, 6, 1.45);
  glow.position.set(0, 0.12, 0.62);
  card.add(glow);

  const goldLight = new THREE.PointLight(0xffe2a0, 0, 7, 1.4);
  goldLight.position.set(0.35, 0.4, 0.45);
  card.add(goldLight);

  const left = makeCopyPlaque(['Commence enfin', 'à progresser.']);
  const right = makeCopyPlaque(['Et à apprendre', 'le jeu.']);
  root.add(left, right);

  root.visible = false;
  const origin = new THREE.Vector3(0, -0.55, -4.55);
  const rest = new THREE.Vector3(0, 0.22, -4.42);
  const leftRest = new THREE.Vector3(-1.78, 0.18, -4.12);
  const rightRest = new THREE.Vector3(1.78, 0.18, -4.12);

  return {
    root, cards: [left, right], keys: [], cta: card, copy: left,
    update(dt, reveal) {
      root.visible = reveal > 0.02;
      glow.intensity = reveal * 2.1;
      goldLight.intensity = reveal * 0.85;
      const pop = 1 - Math.pow(1 - reveal, 2.2);
      const t = performance.now();
      const bob = Math.sin(t * 0.0012) * 0.04 * pop;
      card.position.set(
        lerp(origin.x, rest.x, pop),
        lerp(origin.y, rest.y + bob, pop),
        lerp(origin.z, rest.z, pop),
      );
      card.rotation.y = (1 - pop) * 0.85 + Math.sin(t * 0.0007) * 0.08 * pop;
      card.rotation.x = Math.sin(t * 0.00055) * 0.03 * pop;
      card.scale.setScalar(0.18 + pop * 0.78);

      const sidePop = Math.max(0, (pop - 0.22) / 0.78);
      const sideE = 1 - Math.pow(1 - sidePop, 2);
      left.position.set(
        lerp(origin.x, leftRest.x, sideE),
        lerp(origin.y, leftRest.y + Math.sin(t * 0.00105) * 0.035 * sideE, sideE),
        lerp(origin.z, leftRest.z, sideE),
      );
      right.position.set(
        lerp(origin.x, rightRest.x, sideE),
        lerp(origin.y, rightRest.y + Math.sin(t * 0.00115 + 1.2) * 0.035 * sideE, sideE),
        lerp(origin.z, rightRest.z, sideE),
      );
      left.rotation.y = lerp(0.85, 0.38, sideE);
      right.rotation.y = lerp(-0.85, -0.38, sideE);
      left.scale.setScalar(0.12 + sideE * 0.88);
      right.scale.setScalar(0.12 + sideE * 0.88);
    },
  };
}

function makeCopyPlaque(lines) {
  const W = 1.28;
  const H = 0.92;
  const T = 0.08;
  const g = makeCardBody(W, H, T);
  const tex = canvasTex(1024, 640, (ctx, w, h) => {
    ctx.clearRect(0, 0, w, h);
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.shadowColor = '#3ee0c8';
    ctx.shadowBlur = 16;
    lines.forEach((ln, i) => {
      const last = i === lines.length - 1;
      ctx.fillStyle = last ? '#3ee0c8' : '#eef6f4';
      ctx.font = last
        ? '700 70px "Space Grotesk", sans-serif'
        : '700 52px "Space Grotesk", sans-serif';
      ctx.fillText(ln, w / 2, h * (0.38 + i * 0.24));
    });
  });
  const face = new THREE.Mesh(
    new THREE.PlaneGeometry(W * 0.9, H * 0.7),
    new THREE.MeshBasicMaterial({
      map: tex, transparent: true, depthWrite: false, toneMapped: false,
    }),
  );
  face.position.z = T / 2 + 0.012;
  g.add(face);
  return g;
}

function lerp(a, b, t) {
  return a + (b - a) * t;
}
