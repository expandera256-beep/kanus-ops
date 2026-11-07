// Real headless Chrome (CDP): release_us_fx FX component, 4 langs + missing/corrupt/stale fx.json.
import { spawn } from 'node:child_process';
const DBG = 9341, P = { ok: 8801, missing: 8802, corrupt: 8803, stale: 8804, future: 8805 };
const chrome = spawn('google-chrome', ['--headless=new', '--no-sandbox', '--disable-gpu', `--remote-debugging-port=${DBG}`, '--user-data-dir=/tmp/kus-chrome-fx', 'about:blank'], { stdio: 'ignore' });
const sleep = ms => new Promise(r => setTimeout(r, ms));
let ws, id = 0; const pending = new Map(), reqs = [];
async function connect() { for (let i = 0; i < 50; i++) { try { const j = await (await fetch(`http://127.0.0.1:${DBG}/json/list`)).json(); const pg = j.find(t => t.type === 'page'); if (pg) { ws = new WebSocket(pg.webSocketDebuggerUrl); await new Promise(r => ws.onopen = r); ws.onmessage = e => { const m = JSON.parse(e.data); if (m.method === 'Network.requestWillBeSent') reqs.push(m.params.request.url); if (m.id && pending.has(m.id)) { pending.get(m.id)(m); pending.delete(m.id); } }; return; } } catch {} await sleep(200); } throw new Error('chrome'); }
const send = (method, params = {}) => new Promise(r => { const i = ++id; pending.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
const ev = async e => (await send('Runtime.evaluate', { expression: e, awaitPromise: true, returnByValue: true })).result?.result?.value;
const go = async u => { await send('Page.navigate', { url: u }); await sleep(1500); };
const res = []; const check = (n, ok, info) => { res.push(ok); console.log(`${ok ? 'PASS' : 'FAIL'} ${n} ${JSON.stringify(info)}`); };
const probe = `(async()=>{const m=document.querySelector('main')||document.body;const p=document.createElement('p');p.id='fxp';p.setAttribute('data-kanus-try','18900');m.prepend(p);await new Promise(r=>setTimeout(r,400));const n=p.nextElementSibling;return {t:p.textContent,n:n&&n.classList.contains('kanus-fx-note')&&!n.hidden?n.textContent:''}})()`;
const exp = { tr: ['Yaklaşık tutar; ödeme TL ile alınır', '6 Kas 2026', '18.900'], en: ['payment is charged in Turkish lira', 'Nov 6, 2026', '18,900'], de: ['bezahlt wird in türkischen Lira', '06.11.2026', '18.900'], ru: ['оплата производится в турецких лирах', '6 нояб. 2026', '18 900'] };
try {
  await connect(); await send('Page.enable'); await send('Runtime.enable'); await send('Network.enable');
  for (const lang of Object.keys(exp)) {
    const B = `http://127.0.0.1:${P.ok}`; await go(`${B}/us`); await ev(`localStorage.setItem('kanus.lang','${lang}')`); await go(`${B}/us`);
    const r = await ev(probe); const [note, date, tl] = exp[lang];
    check(`${lang} FX shown (TL+EUR+USD)`, r.t.replace(/\s/g," ").includes(tl) && /337[.,]74/.test(r.t) && /377[.,]45/.test(r.t) && r.t.includes('≈') && r.t.includes('€') && r.t.includes('$'), r);
    check(`${lang} note + TCMB date localized`, r.n.includes(note) && r.n.includes(date) && !r.n.includes('..'), r.n);
    const us = await ev(`document.body.innerText`); check(`${lang} /us page still translated`, us.includes({ tr: 'Shopify üzerinden', en: 'via Shopify', de: 'über Shopify', ru: 'через Shopify' }[lang]), us.slice(0, 80));
  }
  // live language switch re-renders
  await ev(`localStorage.setItem('kanus.lang','de');document.documentElement.setAttribute('lang','de')`); await sleep(300);
  const sw = await ev(`document.querySelector('.kanus-fx-note').textContent`); check('lang switch re-renders note', sw.includes('türkischen Lira'), sw);
  for (const k of ['missing', 'corrupt', 'stale', 'future']) {
    const B = `http://127.0.0.1:${P[k]}`; await go(`${B}/us`); await ev(`localStorage.setItem('kanus.lang','tr')`); await go(`${B}/us`);
    const r = await ev(probe); check(`fx.json ${k} -> TL only, no FX/note`, r.t.includes('18.900') && !r.t.includes('≈') && !r.t.includes('€') && r.n === '', r);
  }
  check('no browser request to tcmb.gov.tr', !reqs.some(u => u.includes('tcmb')), reqs.filter(u => /fx|tcmb/.test(u)));
  check('browser fetched /fx.json', reqs.some(u => u.endsWith('/fx.json')), '');
} catch (e) { check('harness', false, String(e)); }
chrome.kill('SIGKILL'); const f = res.filter(x => !x).length; console.log(`TOTAL ${res.length - f} PASS / ${f} FAIL`); process.exit(f ? 1 : 0);
