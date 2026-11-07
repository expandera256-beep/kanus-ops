// Real headless Chrome (CDP) test of the release_us bundle against a local stub server.
import { spawn } from 'node:child_process';
const PORT = Number(process.env.PORT||8791), DBG = 9339, BASE = `http://127.0.0.1:${PORT}`;
const chrome = spawn('google-chrome', ['--headless=new', '--no-sandbox', '--disable-gpu', `--remote-debugging-port=${DBG}`, '--user-data-dir=/tmp/kus-chrome', 'about:blank'], { stdio: 'ignore' });
const sleep = ms => new Promise(r => setTimeout(r, ms));
let ws, id = 0; const pending = new Map();
async function connect() {
  for (let i = 0; i < 50; i++) { try { const j = await (await fetch(`http://127.0.0.1:${DBG}/json/list`)).json(); const pg = j.find(t => t.type === 'page'); if (pg) { ws = new WebSocket(pg.webSocketDebuggerUrl); await new Promise(r => ws.onopen = r); ws.onmessage = e => { const m = JSON.parse(e.data); if (m.id && pending.has(m.id)) { pending.get(m.id)(m); pending.delete(m.id); } }; return; } } catch {} await sleep(200); }
  throw new Error('chrome not reachable');
}
const send = (method, params = {}) => new Promise(r => { const i = ++id; pending.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
const ev = async expr => { const r = await send('Runtime.evaluate', { expression: expr, awaitPromise: true, returnByValue: true }); return r.result?.result?.value; };
const go = async url => { await send('Page.navigate', { url }); await sleep(1500); };
const results = []; const check = (name, ok, info) => { results.push({ name, ok, info }); console.log(`${ok ? 'PASS' : 'FAIL'} ${name} ${JSON.stringify(info)}`); };
const exp = {
  tr: { usP: 'Shopify üzerinden', err: 'Kaydedilemedi', date: '6 Kas 2026' },
  en: { usP: 'via Shopify', err: 'Could not save', date: 'Nov 6, 2026' },
  de: { usP: 'über Shopify', err: 'Speichern fehlgeschlagen', date: '06.11.2026' },
  ru: { usP: 'через Shopify', err: 'Не удалось сохранить', date: '6 нояб. 2026' },
};
try {
  await connect(); await send('Page.enable'); await send('Runtime.enable');
  for (const lang of ['tr', 'en', 'de', 'ru']) {
    await go(`${BASE}/us`); await ev(`localStorage.setItem('kanus.lang','${lang}')`); await go(`${BASE}/us`);
    const us = await ev(`({lang:document.documentElement.lang, text:document.body.innerText})`);
    check(`${lang} /us page translated`, us.lang === lang && us.text.includes(exp[lang].usP), { lang: us.lang, sample: us.text.slice(0, 120) });
    await go(`${BASE}/news`);
    const news = await ev(`document.body.innerText`);
    check(`${lang} date localized`, news.includes(exp[lang].date) && !news.includes('2026-11-06T'), { found: (news.match(/\n[^\n]*2026[^\n]*/) || [''])[0] });
    await go(`${BASE}/contact`);
    await ev(`(()=>{const set=(el,v)=>{const p=Object.getPrototypeOf(el);Object.getOwnPropertyDescriptor(p,'value').set.call(el,v);el.dispatchEvent(new Event('input',{bubbles:true}))};const f=document.querySelector('form');const [a,b]=f.querySelectorAll('input');set(a,'X');set(b,'x@example.com');set(f.querySelector('textarea'),'hi');f.querySelector('button').click()})()`);
    await sleep(800);
    const n = await ev(`(document.querySelector('.notice')||{}).innerText||''`);
    check(`${lang} submit error localized`, n.includes(exp[lang].err), { notice: n });
  }
} catch (e) { check('harness', false, String(e)); }
chrome.kill('SIGKILL');
const fail = results.filter(r => !r.ok).length;
console.log(`TOTAL ${results.length - fail} PASS / ${fail} FAIL`);
process.exit(fail ? 1 : 0);
