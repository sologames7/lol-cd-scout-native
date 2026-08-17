import * as THREE from 'three';

const FALLBACK = '16.16.1';

let cached = '';

export async function ddragonVersion() {
  if (cached) return cached;
  try {
    const r = await fetch('/ddragon/api/versions.json');
    if (!r.ok) throw new Error(String(r.status));
    const v = await r.json();
    cached = (Array.isArray(v) && v[0]) || FALLBACK;
  } catch {
    cached = FALLBACK;
  }
  return cached;
}

export function ddragonCdn(ver, path) {
  return `/ddragon/cdn/${ver}/${path}`;
}

export function loadTex(url) {
  return new Promise((resolve, reject) => {
    const loader = new THREE.TextureLoader();
    loader.setCrossOrigin('anonymous');
    loader.load(url, (t) => {
      if (THREE.SRGBColorSpace) t.colorSpace = THREE.SRGBColorSpace;
      t.anisotropy = 8;
      resolve(t);
    }, undefined, reject);
  });
}
