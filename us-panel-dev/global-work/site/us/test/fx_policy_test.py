"""Shared FX policy vectors across Python (fx_snapshot), JS (fx_component in node vm) and PHP (fx_cron.php)."""
import datetime as dt, json, os, subprocess, sys, tempfile, shutil
H = os.path.dirname(os.path.abspath(__file__)); US = os.path.join(H, '..'); sys.path.insert(0, US); import fx_snapshot as fx
V = json.load(open(os.path.join(H, 'fx_vectors.json'))); PHP = os.environ.get('PHP', '/workspace/tools/php/php')
res = []
def rec(impl, name, ok): res.append(ok); print(('PASS' if ok else 'FAIL'), impl, name)
def doc(date, eur=55.0, usd=50.0): return {'schema': 'kanus.fx/1', 'base': 'TRY', 'date': date, 'rates': {'EUR': eur, 'USD': usd}}
iso = lambda s: dt.datetime.fromisoformat(s.replace('Z', '+00:00'))
for name, date, now, acc in V['date']:
    try:
        d = dt.date.fromisoformat(date); fx.validate(doc(date), fx.today_trt(iso(now))); got = True
    except (fx.Bad, ValueError): got = False
    rec('py', name, got == acc)
for name, e, u, acc in V['rates']:
    try: fx.validate(doc('2026-11-08', e, u), dt.date(2026, 11, 8)); got = True
    except (fx.Bad, TypeError, ZeroDivisionError): got = False
    rec('py', name, got == acc)
# JS
js = r'''
const fs=require('fs'),vm=require('vm');const V=JSON.parse(fs.readFileSync(process.argv[1]));const src=fs.readFileSync(process.argv[2],'utf8');
let NOW=0;const ctx={window:{},localStorage:{getItem:()=>"en"},document:{querySelectorAll:()=>[],documentElement:{}},fetch:()=>new Promise(()=>{}),MutationObserver:class{observe(){}},Intl,BigInt,Date:class extends Date{static now(){return NOW}},Math,Number,String,Object,isFinite,JSON};vm.createContext(ctx);vm.runInContext(src,ctx);const T=ctx.window.kanusFx._t;const out=[];
for(const [n,date,now,acc] of V.date){NOW=Date.parse(now);out.push([n,!!T.ok({schema:"kanus.fx/1",base:"TRY",date,rates:{EUR:55,USD:50}})===acc])}
NOW=Date.parse("2026-11-08T12:00:00Z");for(const [n,e,u,acc] of V.rates){out.push([n,!!T.ok({schema:"kanus.fx/1",base:"TRY",date:"2026-11-08",rates:{EUR:e,USD:u}})===acc])}
for(const [n,k,r,w] of V.round){out.push([n,T.conv(k,r)===w])}
console.log(JSON.stringify(out));'''
o = subprocess.run(['node', '-e', js, os.path.join(H, 'fx_vectors.json'), os.path.join(US, 'fx_component.js')], capture_output=True, text=True)
if o.returncode: print(o.stderr); rec('js', 'harness', False)
else:
    for n, ok in json.loads(o.stdout): rec('js', n, ok)
# PHP: date policy via XML Tarih + KANUS_FX_NOW
X = open(os.path.join(H, 'tcmb_today_20261106.xml'), 'rb').read()
for name, date, now, acc in V['date']:
    root = tempfile.mkdtemp(); pub = os.path.join(root, 'public_html'); fxd = os.path.join(pub, '_fx'); os.makedirs(fxd)
    shutil.copy(os.path.join(US, 'fx_cron.php'), fxd); y, m, d = date.split('-')
    src = os.path.join(root, 's.xml'); open(src, 'wb').write(X.replace(b'Tarih="06.11.2026"', ('Tarih="%s.%s.%s"' % (d, m, y)).encode()))
    r = subprocess.run([PHP, os.path.join(fxd, 'fx_cron.php')], env=dict(os.environ, KANUS_FX_SRC=src, KANUS_FX_NOW=now), capture_output=True, text=True)
    rec('php', name, (r.returncode == 0) == acc and (os.path.exists(os.path.join(pub, 'fx.json')) == acc)); shutil.rmtree(root)
f = res.count(False); print('TOTAL %d PASS / %d FAIL' % (len(res) - f, f)); sys.exit(1 if f else 0)
