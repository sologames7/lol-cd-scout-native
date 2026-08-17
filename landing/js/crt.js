import * as THREE from 'three';

const vert = /* glsl */ `
varying vec2 vUv;
void main() {
  vUv = uv;
  gl_Position = vec4(position.xy, 0.0, 1.0);
}
`;

const frag = /* glsl */ `
precision highp float;
uniform sampler2D tDiffuse;
uniform vec2 uRes;
uniform float uTime;
uniform float uStrength;
varying vec2 vUv;

vec2 barrel(vec2 uv, float k) {
  vec2 c = uv * 2.0 - 1.0;
  float r2 = dot(c, c);
  c *= 1.0 + k * r2;
  return c * 0.5 + 0.5;
}

void main() {
  vec2 uv = vUv;
  float k = 0.18 * uStrength;
  vec2 b = barrel(uv, k);

  float aberr = 0.0045 * uStrength * (0.35 + dot(b - 0.5, b - 0.5));
  vec2 dir = (b - 0.5);
  float r = texture2D(tDiffuse, b + dir * aberr).r;
  float g = texture2D(tDiffuse, b).g;
  float bl = texture2D(tDiffuse, b - dir * aberr).b;
  vec3 col = vec3(r, g, bl);

  float scan = 0.985 + 0.015 * sin((b.y * uRes.y + uTime * 6.0) * 3.14159);
  col *= scan;

  float vig = smoothstep(1.55, 0.85, length((b - 0.5) * vec2(1.02, 1.08)));
  col *= mix(1.0, vig, 0.18 * uStrength);

  float mask = 0.99 + 0.01 * sin(b.x * uRes.x * 3.14159);
  col *= mask;

  col = mix(col, vec3(dot(col, vec3(0.299, 0.587, 0.114))), 0.08);
  col = col * 1.04 + 0.02;

  float ca = step(-0.04, b.x) * step(b.x, 1.04) * step(-0.04, b.y) * step(b.y, 1.04);
  col *= ca;
  if (ca < 0.5) col = vec3(0.10, 0.08, 0.07);

  gl_FragColor = vec4(col, 1.0);
}
`;

export function createCRT(renderer) {
  const scene = new THREE.Scene();
  const cam = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
  const geo = new THREE.PlaneGeometry(2, 2);
  const uniforms = {
    tDiffuse: { value: null },
    uRes: { value: new THREE.Vector2(1, 1) },
    uTime: { value: 0 },
    uStrength: { value: 1 },
  };
  const mat = new THREE.ShaderMaterial({
    uniforms, vertexShader: vert, fragmentShader: frag,
    depthTest: false, depthWrite: false, toneMapped: false,
  });
  scene.add(new THREE.Mesh(geo, mat));

  const rt = new THREE.WebGLRenderTarget(2, 2, {
    minFilter: THREE.LinearFilter,
    magFilter: THREE.LinearFilter,
    format: THREE.RGBAFormat,
  });

  return {
    resize(w, h) {
      const pr = renderer.getPixelRatio();
      rt.setSize(Math.max(1, Math.floor(w * pr)), Math.max(1, Math.floor(h * pr)));
      uniforms.uRes.value.set(w, h);
    },
    render(worldScene, worldCam, time, strength = 1) {
      uniforms.uStrength.value = strength;
      renderer.setRenderTarget(rt);
      renderer.render(worldScene, worldCam);
      renderer.setRenderTarget(null);
      uniforms.tDiffuse.value = rt.texture;
      uniforms.uTime.value = time;
      renderer.render(scene, cam);
    },
  };
}
