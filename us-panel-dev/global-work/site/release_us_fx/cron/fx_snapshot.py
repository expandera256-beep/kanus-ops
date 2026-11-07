#!/usr/bin/env python3
"""TCMB today.xml -> fx.json snapshot (EUR, USD; ForexSelling = TRY per 1 unit).
Browser never calls TCMB; it reads only /fx.json from kanus.org.
Usage: fx_snapshot.py OUT.json [--xml FILE] [--now YYYY-MM-DD]
Exit 0 = written; 2 = invalid/stale input (OUT untouched -> site keeps hiding or old file)."""
import datetime as dt, json, os, re, sys, tempfile, urllib.request
import xml.etree.ElementTree as ET
URL = 'https://www.tcmb.gov.tr/kurlar/today.xml'
SCHEMA = 'kanus.fx/1'; MAX_AGE_DAYS = 5
RANGE = {'USD': (1.0, 1000.0), 'EUR': (1.0, 1000.0)}

class Bad(Exception): pass

def parse(xml_bytes):
    if len(xml_bytes) > 512 * 1024: raise Bad('too large')
    if b'<!DOCTYPE' in xml_bytes or b'<!ENTITY' in xml_bytes: raise Bad('doctype/entity not allowed')
    try: root = ET.fromstring(xml_bytes)
    except ET.ParseError as e: raise Bad('xml: %s' % e)
    if root.tag != 'Tarih_Date': raise Bad('root %s' % root.tag)
    tarih = root.get('Tarih', ''); m = re.fullmatch(r'(\d\d)\.(\d\d)\.(\d{4})', tarih)
    if not m: raise Bad('Tarih %r' % tarih)
    date = dt.date(int(m[3]), int(m[2]), int(m[1]))
    rates = {}
    for c in root.findall('Currency'):
        k = c.get('CurrencyCode')
        if k in RANGE:
            unit = (c.findtext('Unit') or '').strip(); s = (c.findtext('ForexSelling') or '').strip()
            if unit != '1' or not re.fullmatch(r'\d{1,4}\.\d{1,6}', s): raise Bad('%s unit/selling %r/%r' % (k, unit, s))
            rates[k] = float(s)
    return {'schema': SCHEMA, 'source': 'TCMB today.xml ForexSelling (döviz satış)', 'source_url': URL,
            'tcmb_tarih': tarih, 'bulten_no': root.get('Bulten_No', ''), 'date': date.isoformat(),
            'base': 'TRY', 'rates': rates}

def validate(d, today):
    if not isinstance(d, dict) or d.get('schema') != SCHEMA or d.get('base') != 'TRY': raise Bad('schema')
    if not re.fullmatch(r'\d{4}-\d\d-\d\d', str(d.get('date', ''))): raise Bad('date')
    age = (today - dt.date.fromisoformat(d['date'])).days
    if age < -1 or age > MAX_AGE_DAYS: raise Bad('stale/future: %d days' % age)
    r = d.get('rates')
    if not isinstance(r, dict) or set(r) != set(RANGE): raise Bad('rates keys %r' % (sorted(r) if isinstance(r, dict) else r))
    for k, (lo, hi) in RANGE.items():
        v = r[k]
        if isinstance(v, bool) or not isinstance(v, (int, float)) or not lo <= v <= hi: raise Bad('%s=%r' % (k, v))
    if not 0.8 <= r['EUR'] / r['USD'] <= 1.6: raise Bad('EUR/USD cross %.3f' % (r['EUR'] / r['USD']))
    return d

def main(a):
    out = a[0]; xmlf = a[a.index('--xml') + 1] if '--xml' in a else None
    today = dt.date.fromisoformat(a[a.index('--now') + 1]) if '--now' in a else dt.date.today()
    try:
        if xmlf: raw = open(xmlf, 'rb').read()
        else:
            req = urllib.request.Request(URL, headers={'User-Agent': 'kanus-fx/1'})
            with urllib.request.urlopen(req, timeout=20) as r: raw = r.read(512 * 1024 + 1)
        d = validate(parse(raw), today)
    except (Bad, OSError, ValueError) as e:
        print('INVALID: %s' % e, file=sys.stderr); return 2
    d['generated_at'] = dt.datetime.now(dt.timezone.utc).replace(microsecond=0).isoformat()
    fd, tmp = tempfile.mkstemp(dir=os.path.dirname(os.path.abspath(out)) or '.')
    with os.fdopen(fd, 'w') as f: json.dump(d, f, ensure_ascii=False, indent=1); f.write('\n')
    os.chmod(tmp, 0o644); os.replace(tmp, out); print(json.dumps(d, ensure_ascii=False)); return 0

if __name__ == '__main__': sys.exit(main(sys.argv[1:]))
