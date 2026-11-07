const fs=require('fs'),vm=require('vm'),assert=require('assert');
const html=fs.readFileSync('/tmp/global-review.o9517a1v/term/gui.html','utf8');const func=html.slice(html.indexOf('function applyLocale(doc)'),html.indexOf('function loadLocale(lang)'));
const hostile='<script>globalThis.PWNED=true</script><img src=x onerror="PWNED=true">';
const el={dataset:{i18n:'k'}},ph={dataset:{i18nPh:'p'}};const opts=[];const select={replaceChildren(...o){opts.push(...o)}};
for(const x of [el,ph])Object.defineProperty(x,'innerHTML',{set(){throw Error('unsafe innerHTML')}});
const document={documentElement:{},querySelectorAll(s){return s==='[data-i18n]'?[el]:[ph]},querySelector(){return select},createElement(){const o={};Object.defineProperty(o,'innerHTML',{set(){throw Error('unsafe option HTML')}});return o}};
const context={document,i18nState:{lang:'tr'}};vm.createContext(context);vm.runInContext(func,context);context.applyLocale({messages:{k:hostile,p:'" onfocus="PWNED=true'},lang:'en',dir:'rtl" onload="PWNED=true',languages:[{code:'en',name:hostile}]});
assert.equal(el.textContent,hostile);assert.equal(ph.placeholder,'" onfocus="PWNED=true');assert.equal(opts[0].textContent,hostile);assert.equal(document.documentElement.dir,'ltr');assert.equal(context.PWNED,undefined);
context.applyLocale({messages:{},dir:'rtl'});assert.equal(document.documentElement.dir,'rtl');console.log('PASS: translation text, placeholder and option labels use inert DOM properties; malicious dir forced ltr; valid rtl accepted. DOM mock, not a real browser.');
