// R2-01 UI: a failed /api/lang save must revert the selector, keep the
// previous language and show a localized error. No browser required.
import { readFileSync } from 'node:fs';
import { runInNewContext } from 'node:vm';

const html = readFileSync(new URL('./gui.html', import.meta.url), 'utf8');
const script = html.match(/<script>([\s\S]*?)<\/script>/)[1];
const start = script.indexOf('// i18n: dynamic language switching');
const end = script.indexOf('load();', start);
if (start < 0 || end < 0) throw new Error('i18n block not found');
const block = script.slice(start, end);

const MSG = {
  tr: { 'gui.lang.save_failed': 'Dil tercihi kaydedilemedi; önceki dil korunuyor.' },
  en: { 'gui.lang.save_failed': 'Could not save the language preference; keeping the previous language.' },
};

async function scenario(name, apiLang, expect) {
  const els = {};
  const el = (id) => (els[id] ||= { id, value: '', textContent: '', hidden: true, listeners: {}, dataset: {},
    addEventListener(t, f) { this.listeners[t] = f; }, replaceChildren(...c) { this.children = c; } });
  const calls = [];
  const document = {
    documentElement: { lang: 'tr', dir: 'ltr' },
    querySelector: (sel) => el(sel.replace('#', '')),
    querySelectorAll: () => [],
    createElement: () => ({}),
  };
  const fetch = (url, opts = {}) => {
    calls.push([url, opts.method || 'GET']);
    if (url.startsWith('/api/lang')) return apiLang();
    const lang = (url.match(/lang=(\w+)/) || [, 'tr'])[1];
    const doc = { lang, dir: 'ltr', languages: [{ code: 'tr', name: 'Türkçe' }, { code: 'en', name: 'English' }], messages: MSG[lang] };
    return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve(doc) });
  };
  runInNewContext(block, { document, fetch, encodeURIComponent, Array, Promise, console });
  await new Promise((r) => setTimeout(r, 20));
  const sel = el('langsel');
  sel.value = 'en';
  await sel.listeners.change({ target: sel });
  await new Promise((r) => setTimeout(r, 20));
  const loadedEn = calls.some(([u]) => u === '/api/i18n?lang=en');
  const status = el('langstatus');
  const got = { value: sel.value, loadedEn, htmlLang: document.documentElement.lang, statusShown: status.hidden === false, statusText: status.textContent };
  const ok = Object.entries(expect).every(([k, v]) => got[k] === v);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name} ${JSON.stringify(got)}`);
  return ok;
}

const results = [
  await scenario('500 reverts and shows TR error', () => Promise.resolve({ ok: false, status: 500 }),
    { value: 'tr', loadedEn: false, htmlLang: 'tr', statusShown: true, statusText: MSG.tr['gui.lang.save_failed'] }),
  await scenario('503 not-persisted reverts', () => Promise.resolve({ ok: false, status: 503 }),
    { value: 'tr', loadedEn: false, statusShown: true }),
  await scenario('network error reverts', () => Promise.reject(new Error('down')),
    { value: 'tr', loadedEn: false, statusShown: true }),
  await scenario('200 switches to en', () => Promise.resolve({ ok: true, status: 200 }),
    { value: 'en', loadedEn: true, htmlLang: 'en', statusShown: false }),
];
if (results.includes(false)) { console.error('gui lang UI test FAILED'); process.exit(1); }
console.log('gui lang UI test passed');
