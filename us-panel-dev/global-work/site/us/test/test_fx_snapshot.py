import datetime as dt, json, os, subprocess, sys, tempfile, unittest
H = os.path.dirname(os.path.abspath(__file__)); sys.path.insert(0, os.path.join(H, '..')); import fx_snapshot as fx
X = open(os.path.join(H, 'tcmb_today_20261106.xml'), 'rb').read(); T = dt.date(2026, 11, 8)
class T1(unittest.TestCase):
    def ok(self, raw, today=T): return fx.validate(fx.parse(raw), today)
    def bad(self, raw, today=T): self.assertRaises(fx.Bad, self.ok, raw, today)
    def test_real(self):
        d = self.ok(X); self.assertEqual((d['date'], d['tcmb_tarih'], d['rates']), ('2026-11-06', '06.11.2026', {'USD': 50.0724, 'EUR': 55.9609}))
    def test_stale(self): self.bad(X, dt.date(2026, 11, 12))
    def test_future(self): self.bad(X, dt.date(2026, 11, 4))
    def test_garbage(self): self.bad(b'<html>nope</html>'); self.bad(b'not xml')
    def test_entity(self): self.bad(b'<!DOCTYPE x [<!ENTITY a "b">]>' + X)
    def test_missing_eur(self): self.bad(X.replace(b'CurrencyCode="EUR"', b'CurrencyCode="EUX"'))
    def test_bad_number(self): self.bad(X.replace(b'<ForexSelling>50.0724', b'<ForexSelling>abc'))
    def test_empty_selling(self): self.bad(X.replace(b'<ForexSelling>55.9609</ForexSelling>', b'<ForexSelling></ForexSelling>'))
    def test_cross(self): self.bad(X.replace(b'<ForexSelling>55.9609', b'<ForexSelling>155.9609'))
    def test_bad_date(self): self.bad(X.replace(b'Tarih="06.11.2026"', b'Tarih="2026-11-06"'))
    def test_json_validate_types(self):
        d = self.ok(X); d['rates']['USD'] = True; self.assertRaises(fx.Bad, fx.validate, d, T)
    def test_cli_invalid_keeps_old(self):
        with tempfile.TemporaryDirectory() as t:
            out = os.path.join(t, 'fx.json'); open(out, 'w').write('OLD'); bad = os.path.join(t, 'b.xml'); open(bad, 'wb').write(b'<x/>')
            r = subprocess.run([sys.executable, os.path.join(H, '..', 'fx_snapshot.py'), out, '--xml', bad, '--now', '2026-11-08'], capture_output=True)
            self.assertEqual((r.returncode, open(out).read()), (2, 'OLD'))
            good = os.path.join(t, 'g.xml'); open(good, 'wb').write(X)
            r = subprocess.run([sys.executable, os.path.join(H, '..', 'fx_snapshot.py'), out, '--xml', good, '--now', '2026-11-08'], capture_output=True)
            self.assertEqual(r.returncode, 0); self.assertEqual(json.load(open(out))['rates']['EUR'], 55.9609)
            self.assertEqual(oct(os.stat(out).st_mode & 0o777), '0o644')
if __name__ == '__main__': unittest.main(verbosity=1)
