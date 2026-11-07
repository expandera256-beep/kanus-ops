// FX08B-02: /api/pricing responses arriving out of order must not let an
// older language overwrite the newest one. No browser required.
import { readFileSync } from 'node:fs';
import { runInNewContext } from 'node:vm';
const html = readFileSync(new URL('./gui.html', import.meta.url), 'utf8');
const script = html.match(/<script>([\s\S]*?)<\/script>/)[1];
const start = script.indexOf('let pricingSeq = 0;');
const end = script.indexOf('function applyLocale(doc) {', start);
if (start < 0 || end < 0) throw new Error('pricing block not found');
const block = script.slice(start, end) + '\nglobalThis.refreshPricing = refreshPricing;';
const el = { textContent: 'TR' };
const pending = {};
const fetch = (url) => new Promise((resolve) => { const lang = url.match(/lang=(\w+)/)[1]; pending[lang] = () => resolve({ ok: true, json: () => Promise.resolve({ text: 'price-' + lang }) }); });
const ctx = { document: { querySelector: () => el }, fetch, encodeURIComponent, Promise };
runInNewContext(block, ctx);
const results = [];
const check = (name, ok, info) => { results.push(ok); console.log(`${ok ? 'PASS' : 'FAIL'} ${name} ${JSON.stringify(info)}`); };
// EN requested, then DE; DE answers first, EN (stale) answers last.
const pEn = ctx.refreshPricing('en'); const pDe = ctx.refreshPricing('de');
pending.de(); await pDe; pending.en(); await pEn;
check('stale EN response ignored, DE kept', el.textContent === 'price-de', el.textContent);
// In-order responses still apply.
const pRu = ctx.refreshPricing('ru'); pending.ru(); await pRu;
check('in-order response applied', el.textContent === 'price-ru', el.textContent);
if (results.includes(false)) { console.error('gui pricing order test FAILED'); process.exit(1); }
console.log('gui pricing order test passed');
