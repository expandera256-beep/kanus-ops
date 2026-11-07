"""Real PHP tests for fx_cron.php (Option A layout public_html/_fx/). PHP binary: $PHP."""
import fcntl, http.server, json, os, re, shutil, socket, subprocess, sys, tempfile, threading, time, unittest, urllib.request, urllib.error
H = os.path.dirname(os.path.abspath(__file__)); US = os.path.join(H, '..'); sys.path.insert(0, US); import fx_snapshot
PHP = os.environ.get('PHP', '/workspace/tools/php/php'); XML = os.path.join(H, 'tcmb_today_20261106.xml')
SECRET_RE = re.compile(r'secret|token|passw|api[_-]?key|authorization|bearer|u346856246', re.I)
def tree(root):
    out = {}
    for r, ds, fs in os.walk(root):
        for f in fs: p = os.path.join(r, f); out[os.path.relpath(p, root)] = open(p, 'rb').read()
    return out
class Base(unittest.TestCase):
    def setUp(self):
        self.root = tempfile.mkdtemp(); self.pub = os.path.join(self.root, 'public_html'); self.fx = os.path.join(self.pub, '_fx')
        os.makedirs(self.fx); shutil.copy(os.path.join(US, 'fx_cron.php'), self.fx); shutil.copy(os.path.join(US, 'fx_htaccess'), os.path.join(self.fx, '.htaccess'))
        open(os.path.join(self.pub, 'index.html'), 'w').write('X'); open(os.path.join(self.pub, 'fx.json'), 'w').write('OLD')
    def tearDown(self): shutil.rmtree(self.pub)
    def run_cron(self, src=None, timeout=None, raw=False):
        env = dict(os.environ); env.pop('KANUS_FX_SRC', None)
        if src: env['KANUS_FX_SRC'] = src
        if timeout: env['KANUS_FX_TIMEOUT'] = str(timeout)
        r = subprocess.run([PHP, os.path.join(self.fx, 'fx_cron.php')], capture_output=True, text=True, env=env, timeout=120)
        self.assertNotRegex(r.stdout + r.stderr, SECRET_RE); return r
    def assert_untouched(self, before, r, code=2):
        self.assertEqual(r.returncode, code, r.stderr); self.assertEqual(open(os.path.join(self.pub, 'fx.json')).read(), 'OLD')
        after = tree(self.pub); after.pop('_fx/fx_cron.lock', None); self.assertEqual(after, before)
    def before(self): b = tree(self.pub); b.pop('_fx/fx_cron.lock', None); return b
    def xml(self, data): p = os.path.join(self.root, 'src.xml'); open(p, 'wb').write(data); return p
class T(Base):
    def test_01_real_tcmb_live(self):
        b = self.before(); r = self.run_cron(); self.assertEqual(r.returncode, 0, r.stderr)
        d = json.load(open(os.path.join(self.pub, 'fx.json'))); fx_snapshot.validate(d, __import__('datetime').date.today())
        after = tree(self.pub); changed = {k for k in after if after[k] != b.get(k)} - {'_fx/fx_cron.lock'}
        self.assertEqual(changed, {'fx.json'}); self.assertEqual(oct(os.stat(os.path.join(self.pub, 'fx.json')).st_mode & 0o777), '0o644')
        self.assertFalse([f for f in os.listdir(self.fx) if 'tmp' in f]); print('\n  live:', d['tcmb_tarih'], d['rates'], file=sys.stderr)
    def test_02_same_schema_as_python(self):
        r = self.run_cron(src=self.xml(open(XML, 'rb').read()))
        if r.returncode == 2 and 'stale' in r.stderr: self.skipTest('fixture older than 5 days')
        self.assertEqual(r.returncode, 0, r.stderr); php = json.load(open(os.path.join(self.pub, 'fx.json')))
        py = fx_snapshot.parse(open(XML, 'rb').read()); php.pop('generated_at'); self.assertEqual(php, py); self.assertEqual(list(php), list(py))
        self.assertIn('"source_url": "https://www.tcmb.gov.tr/kurlar/today.xml"', open(os.path.join(self.pub, 'fx.json')).read())
    def test_03_network_error(self): b = self.before(); self.assert_untouched(b, self.run_cron(src='http://127.0.0.1:9/today.xml', timeout=3))
    def test_04_bad_xml(self):
        b = self.before()
        for data in (b'<html>500</html>', b'not xml at all', b'<!DOCTYPE x [<!ENTITY a "b">]>' + open(XML, 'rb').read(), b''):
            self.assert_untouched(b, self.run_cron(src=self.xml(data)))
    def test_05_missing_rate(self):
        b = self.before(); X = open(XML, 'rb').read()
        self.assert_untouched(b, self.run_cron(src=self.xml(X.replace(b'CurrencyCode="EUR"', b'CurrencyCode="EUX"'))))
        self.assert_untouched(b, self.run_cron(src=self.xml(X.replace(b'<ForexSelling>50.0724', b'<ForexSelling>'))))
    def test_06_holiday_stale(self):
        b = self.before(); X = open(XML, 'rb').read().replace(b'Tarih="06.11.2026"', b'Tarih="20.10.2026"')
        r = self.run_cron(src=self.xml(X)); self.assert_untouched(b, r); self.assertIn('stale', r.stderr)
    def test_07_lock(self):
        b = self.before(); f = open(os.path.join(self.fx, 'fx_cron.lock'), 'a'); fcntl.flock(f, fcntl.LOCK_EX)
        try: t = time.time(); r = self.run_cron(src=self.xml(open(XML, 'rb').read())); self.assert_untouched(b, r, 3); self.assertLess(time.time() - t, 5)
        finally: f.close()
    def test_08_timeout(self):
        class S(http.server.BaseHTTPRequestHandler):
            def log_message(self, *a): pass
            def do_GET(self): time.sleep(15)
        srv = http.server.ThreadingHTTPServer(('127.0.0.1', 0), S); threading.Thread(target=srv.serve_forever, daemon=True).start()
        b = self.before(); t = time.time(); r = self.run_cron(src='http://127.0.0.1:%d/' % srv.server_port, timeout=2); el = time.time() - t; srv.shutdown()
        self.assert_untouched(b, r); self.assertLess(el, 8)
    def test_09_atomic_concurrent_reader(self):
        src = self.xml(open(XML, 'rb').read()); seen = set(); stop = False; bad = []
        def reader():
            while not stop:
                try: s = open(os.path.join(self.pub, 'fx.json')).read()
                except FileNotFoundError: bad.append('missing'); continue
                if s != 'OLD':
                    try: json.loads(s); seen.add('new')
                    except ValueError: bad.append(s[:40])
        th = threading.Thread(target=reader); th.start()
        for _ in range(15):
            r = self.run_cron(src=src)
            if r.returncode != 0: break
        stop = True; th.join()
        if r.returncode == 2 and 'stale' in r.stderr: self.skipTest('fixture stale')
        self.assertEqual(bad, []); self.assertIn('new', seen); self.assertFalse([f for f in os.listdir(self.fx) if 'tmp' in f])
        src_code = open(os.path.join(self.fx, 'fx_cron.php')).read(); self.assertIn('rename($tmp, $out)', src_code); self.assertIn("tempnam($dir", src_code)
    def test_10_http_forbidden_php_S(self):
        s = socket.socket(); s.bind(('127.0.0.1', 0)); port = s.getsockname()[1]; s.close()
        p = subprocess.Popen([PHP, '-S', '127.0.0.1:%d' % port, '-t', self.pub], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL); time.sleep(1)
        try:
            b = self.before()
            try: urllib.request.urlopen('http://127.0.0.1:%d/_fx/fx_cron.php' % port, timeout=10); code = 200
            except urllib.error.HTTPError as e: code, body = e.code, e.read().decode()
            self.assertEqual(code, 403); self.assertEqual(body.strip(), 'Forbidden')
            after = tree(self.pub); after.pop('_fx/fx_cron.lock', None); self.assertEqual(after, b)
            self.assertFalse(os.path.exists(os.path.join(self.fx, 'fx_cron.lock')))
        finally: p.kill(); p.wait()
    def test_11_wrong_location_refused(self):
        os.rename(self.fx, os.path.join(self.pub, 'other')); self.fx = os.path.join(self.pub, 'other'); b = self.before()
        self.assert_untouched(b, self.run_cron(src=self.xml(open(XML, 'rb').read())))
if __name__ == '__main__': unittest.main(verbosity=2)
