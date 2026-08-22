import { ddragonVersion, ddragonCdn } from './ddragon.js';
import { createItemModel, createPriceTag, SHELF_IDS, ON_SHELF } from './itemModels.js';

const RARITY = {
  consumable: { metal: 0x4a8a3a, emissive: 0x1a4018, glow: 0.35, hex: '#6bc46b' },
  trinket: { metal: 0x3ee0c8, emissive: 0x0a4a40, glow: 0.55, hex: '#3ee0c8' },
  boots: { metal: 0x8a6424, emissive: 0x3a2808, glow: 0.25, hex: '#c4a06a' },
  basic: { metal: 0x8a9098, emissive: 0x2a3038, glow: 0.2, hex: '#b8c0c8' },
  epic: { metal: 0x4a88d4, emissive: 0x102848, glow: 0.45, hex: '#5aa0ff' },
  legendary: { metal: 0xd4b24a, emissive: 0x6a4810, glow: 0.7, hex: '#f3d78a' },
};

function has(tags, k) {
  return tags.includes(k);
}

function classify(it, tags) {
  if (it.requiredChampion || it.hideFromAll) return null;
  if (it.gold && it.gold.purchasable === false && !has(tags, 'Trinket') && !has(tags, 'Vision')) return null;
  if (has(tags, 'Consumable') || has(tags, 'Potion')) return 'consumable';
  if (has(tags, 'Trinket') || has(tags, 'Vision') || /ward|totem|oracle/i.test(it.name || '')) return 'trinket';
  if (has(tags, 'Boots')) return 'boots';
  const from = it.from || [];
  const into = it.into || [];
  const gold = (it.gold && it.gold.total) || 0;
  if (!from.length) return 'basic';
  if (into.length) return 'epic';
  if (gold >= 1600) return 'legendary';
  return 'epic';
}

export function createItemMesh(kind, opts = {}) {
  const g = createItemModel(kind);
  g.userData.rarity = kind.rarity;
  g.userData.gold = kind.gold;
  if (opts.price !== false) {
    const tag = createPriceTag(kind.gold || 0);
    tag.position.set(0, -0.46, 0.02);
    g.add(tag);
  }
  return g;
}

export async function loadShopCatalog() {
  const ver = await ddragonVersion();
  let data = {};
  try {
    const r = await fetch(ddragonCdn(ver, 'data/en_US/item.json'));
    if (!r.ok) throw new Error(String(r.status));
    const json = await r.json();
    data = json.data || {};
  } catch (err) {
    console.warn('item.json', err);
    return [];
  }
  const out = [];
  for (const id of SHELF_IDS) {
    const it = data[id];
    if (!it) continue;
    const tags = it.tags || [];
    const rarity = classify(it, tags) || 'basic';
    out.push({
      id,
      name: it.name,
      gold: (it.gold && it.gold.total) || 0,
      tags,
      rarity,
    });
  }
  return out;
}

export { RARITY, ON_SHELF };
