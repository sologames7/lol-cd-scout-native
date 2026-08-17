import { createSpellToken } from './summoners.js';
import { ddragonVersion, ddragonCdn, loadTex } from './ddragon.js';

const ICONIC = [
  { id: 'malphite-r', file: 'UFSlash.png', metal: 0x9aacbc, emissive: 0x3a5068, glow: 0.7 },
  { id: 'lux-r', file: 'LuxR.png', metal: 0xe8d48a, emissive: 0x8a7018, glow: 0.75 },
  { id: 'ashe-r', file: 'EnchantedCrystalArrow.png', metal: 0x7ec8ff, emissive: 0x164a8a, glow: 0.7 },
  { id: 'ezreal-r', file: 'EzrealR.png', metal: 0xffe566, emissive: 0x8a7010, glow: 0.65 },
  { id: 'jinx-r', file: 'JinxR.png', metal: 0xff6ad5, emissive: 0x8a1860, glow: 0.7 },
  { id: 'yasuo-r', file: 'YasuoR.png', metal: 0xa8c8e8, emissive: 0x2a5080, glow: 0.6 },
  { id: 'zed-r', file: 'ZedR.png', metal: 0x6a2a8a, emissive: 0x3a1050, glow: 0.75 },
  { id: 'thresh-q', file: 'ThreshQ.png', metal: 0xc4e07a, emissive: 0x4a6810, glow: 0.65 },
  { id: 'blitz-q', file: 'RocketGrab.png', metal: 0xffa030, emissive: 0x8a4010, glow: 0.6 },
  { id: 'mf-r', file: 'MissFortuneBulletTime.png', metal: 0xd4b24a, emissive: 0x8a5010, glow: 0.65 },
  { id: 'veigar-r', file: 'VeigarR.png', metal: 0x9a6aff, emissive: 0x3a1868, glow: 0.75 },
  { id: 'amumu-r', file: 'CurseoftheSadMummy.png', metal: 0x8aaa58, emissive: 0x304818, glow: 0.55 },
  { id: 'orianna-r', file: 'OrianaDetonateCommand.png', metal: 0xd4b24a, emissive: 0x6a4810, glow: 0.7 },
  { id: 'leesin-r', file: 'LeeSinR.png', metal: 0xff7a3a, emissive: 0x8a2808, glow: 0.65 },
];

export async function loadIconicSpells() {
  const ver = await ddragonVersion();
  const out = [];
  for (const k of ICONIC) {
    try {
      const map = await loadTex(ddragonCdn(ver, `img/spell/${k.file}`));
      out.push({ ...k, map });
    } catch (err) {
      console.warn('spell', k.id, err);
    }
  }
  return out;
}

export function createIconicToken(kind) {
  return createSpellToken(kind);
}
