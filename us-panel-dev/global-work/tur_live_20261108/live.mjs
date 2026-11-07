import { spawn } from 'node:child_process'; import fs from 'node:fs';
const DBG=9351, B='https://kanus.org';
const chrome=spawn('google-chrome',['--headless=new','--no-sandbox','--disable-gpu',`--remote-debugging-port=${DBG}`,'--user-data-dir=/tmp/kus-live','--window-size=1280,1600','about:blank'],{stdio:'ignore'});
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
let ws,id=0;const pending=new Map();let cons=[],bad=[],blocked=[],allpost=[];
const send=(m,p={})=>new Promise(r=>{const i=++id;pending.set(i,r);ws.send(JSON.stringify({id:i,method:m,params:p}))});
async function connect(){for(let i=0;i<50;i++){try{const j=await(await fetch(`http://127.0.0.1:${DBG}/json/list`)).json();const pg=j.find(t=>t.type==='page');if(pg){ws=new WebSocket(pg.webSocketDebuggerUrl);await new Promise(r=>ws.onopen=r);ws.onmessage=e=>{const m=JSON.parse(e.data);
 if(m.method==='Runtime.consoleAPICalled'&&m.params.type==='error')cons.push(m.params.args.map(a=>a.value??a.description).join(' '));
 if(m.method==='Runtime.exceptionThrown')cons.push('EXC '+(m.params.exceptionDetails.exception?.description||m.params.exceptionDetails.text));
 if(m.method==='Log.entryAdded'&&m.params.entry.level==='error')cons.push('LOG '+m.params.entry.text+' '+(m.params.entry.url||''));
 if(m.method==='Network.responseReceived'&&m.params.response.status>=400)bad.push(m.params.response.status+' '+m.params.response.url);
 if(m.method==='Fetch.requestPaused'){const q=m.params.request;if(q.method!=='GET')allpost.push(q.method+' '+q.url);if(q.method!=='GET'&&q.method!=='HEAD'&&!(q.method==='POST'&&q.url.split('?')[0]==='https://kanus.org/hcdn-cgi/jschallenge-validate')){blocked.push(q.method+' '+q.url);send('Fetch.failRequest',{requestId:m.params.requestId,errorReason:'Failed'})}else send('Fetch.continueRequest',{requestId:m.params.requestId})}
 if(m.id&&pending.has(m.id)){pending.get(m.id)(m);pending.delete(m.id)}};return}}catch{}await sleep(200)}throw new Error('chrome')}
const ev=async e=>(await send('Runtime.evaluate',{expression:e,awaitPromise:true,returnByValue:true})).result?.result?.value;
const go=async u=>{await send('Page.navigate',{url:u});await sleep(2500);for(let i=0;i<20;i++){const t=await ev('document.title');if(!/Checking your browser/.test(t||''))break;await sleep(1000)}await sleep(1500)};
const shot=async n=>{const r=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:true});fs.writeFileSync(n,Buffer.from(r.result.data,'base64'))};
const out={};
await connect();for(const d of['Page','Runtime','Network','Log'])await send(d+'.enable');await send('Fetch.enable',{patterns:[{urlPattern:'*'}]});
const info=`(()=>{const t=document.body.innerText;return{lang:document.documentElement.lang,title:document.title,len:t.length,text:t,iso:(t.match(/\\d{4}-\\d{2}-\\d{2}(T[\\d:.]+Z?)?/g)||[]),slogan:/Yerli Teknoloji, Küresel Vizyon|Built in Türkiye\\. Designed for the World\\./i.test(document.documentElement.outerHTML),switcher:[...document.querySelectorAll('button,a,select,option')].filter(e=>/^(TR|EN|DE|RU|Türkçe|English|Deutsch|Русский)$/i.test((e.innerText||e.textContent||'').trim())).map(e=>e.tagName+':'+(e.innerText||e.textContent).trim())}})()`;
const formTest=`(async()=>{const f=document.querySelector('form');if(!f)return{form:false};const set=(el,v)=>{const p=Object.getPrototypeOf(el);Object.getOwnPropertyDescriptor(p,'value').set.call(el,v);el.dispatchEvent(new Event('input',{bubbles:true}))};const ins=[...f.querySelectorAll('input:not([type=hidden]):not([type=checkbox])')];ins.forEach(i=>set(i,i.type==='email'?'x@example.com':'Test'));const ta=f.querySelector('textarea');if(ta)set(ta,'test');f.querySelectorAll('input[type=checkbox]').forEach(c=>{if(!c.checked)c.click()});(f.querySelector('button[type=submit],button')||{click(){}}).click();await new Promise(r=>setTimeout(r,2500));const n=document.querySelector('.notice,[role=alert],.error');return{form:true,notice:n?n.innerText:'',body:document.body.innerText.slice(0,0)}})()`;
for(const lang of['tr','en','de','ru'])for(const p of['/','/us','/contact','/news']){cons=[];bad=[];blocked=[];
 await go(B+p);await ev(`localStorage.setItem('kanus.lang','${lang}')`);await go(B+p);
 const r=await ev(info);await shot(`${lang}_${p==='/'?'home':p.slice(1)}.png`);
 let form=null;if(p==='/contact'){form=await ev(formTest);form.blocked=[...blocked]}
 out[`${lang}${p}`]={...r,text:r.text.slice(0,1500),cons:[...cons],bad:[...bad],form,stored:await ev(`localStorage.getItem('kanus.lang')`)};console.log(lang,p,r.lang,r.iso,cons.length,bad.length,JSON.stringify(form))}
// switcher click test on /us
out.click={};await go(B+'/us');await ev(`localStorage.setItem('kanus.lang','tr')`);await go(B+'/us');
for(const lang of['en','de','ru','tr']){const c=await ev(`(()=>{const L={tr:/^(TR|Türkçe)$/i,en:/^(EN|English)$/i,de:/^(DE|Deutsch)$/i,ru:/^(RU|Русский)$/i}['${lang}'];const s=[...document.querySelectorAll('select')].find(s=>[...s.options].some(o=>o.value==='${lang}'));if(s){s.value='${lang}';s.dispatchEvent(new Event('change',{bubbles:true}));return 'select'}const b=[...document.querySelectorAll('button,a,[role=menuitem],li')].find(e=>L.test((e.innerText||'').trim())||e.dataset?.lang==='${lang}');if(b){b.click();return b.tagName}return null})()`);await sleep(1500);
 const r=await ev(`({lang:document.documentElement.lang,ls:localStorage.getItem('kanus.lang'),t:document.body.innerText.slice(0,300)})`);out.click[lang]={how:c,...r};await shot(`click_${lang}_us.png`);console.log('click',lang,c,r.lang,r.ls)}
out.allpost=allpost;fs.writeFileSync('live_results.json',JSON.stringify(out,null,1));chrome.kill('SIGKILL');process.exit(0)
