import json, os, re, tempfile, unittest
import i18n_mail as m

HERE = os.path.dirname(os.path.abspath(__file__))
GO_LOCALES = os.environ.get('US_GO_LOCALES', '')


class MailI18nTest(unittest.TestCase):
    def test_default_is_turkish(self):
        s, b, lang = m.render_license_email(order='#1001')
        self.assertEqual(lang, 'tr')
        self.assertEqual(s, 'US Panel lisans anahtarın')
        self.assertIn('#1001', b)

    def test_subject_decision_brand_a(self):
        # Decision 06.11.2026: subject per BRAND_A "US Panel lisans anahtarın".
        # BRAND_A has no EN subject; EN follows its naming ("US Panel") and tone.
        self.assertEqual(m.t('email.license.subject', 'tr'), 'US Panel lisans anahtarın')
        self.assertEqual(m.t('email.license.subject', 'en'), 'Your US Panel license key')
        self.assertNotEqual(m.t('email.license.subject', 'tr'), 'US lisans anahtarınız')

    def test_delivery_not_wired(self):
        # The loader must not be connected to sending yet.
        with open(os.path.join(HERE, 'delivery.py'), encoding='utf-8') as fh:
            self.assertNotIn('i18n_mail', fh.read())

    def test_english_and_placeholder(self):
        s, b, lang = m.render_license_email('en-GB', order='#7')
        self.assertEqual((s, lang), ('Your US Panel license key', 'en'))
        self.assertIn('order #7', b)
        self.assertNotIn('{order}', b)

    def test_unknown_lang_and_missing_param(self):
        self.assertEqual(m.resolve_lang('xx'), 'tr')
        self.assertEqual(m.resolve_lang('', 'de;q=0.9,en;q=0.5'), 'de')
        self.assertEqual(m.resolve_lang('', 'ja;q=0.9,en;q=0.5'), 'en')
        self.assertEqual(m.resolve_lang('', 'en;q=0'), 'tr')
        self.assertIn('{order}', m.t('email.license.body', 'en'))
        self.assertEqual(m.t('no.such.key', 'en'), 'no.such.key')

    def test_fallback_and_rtl(self):
        with tempfile.TemporaryDirectory() as d:
            for name in ('tr', 'en'):
                with open(os.path.join(HERE, 'locales', name + '.json'), encoding='utf-8') as src, open(os.path.join(d, name + '.json'), 'w', encoding='utf-8') as dst:
                    dst.write(src.read())
            with open(os.path.join(d, 'ar.json'), 'w', encoding='utf-8') as fh:
                json.dump({'_meta': {'name': 'test', 'dir': 'rtl'}, 'email.license.subject': 'AR'}, fh)
            self.assertEqual(m.t('email.license.subject', 'ar', d), 'AR')
            self.assertEqual(m.t('email.license.body', 'ar', d, order='9').split('\n')[0], 'Hello,')
            self.assertEqual(m.load(d)['ar']['meta']['dir'], 'rtl')

    def test_bad_locale_rejected(self):
        with tempfile.TemporaryDirectory() as d:
            for n in ('tr', 'en'):
                with open(os.path.join(d, n + '.json'), 'w') as fh:
                    json.dump({'k': 1}, fh)
            with self.assertRaises(ValueError):
                m.load(d)

    def test_keys_and_placeholders_equal(self):
        cats = m.load()
        tr, en = cats['tr']['msgs'], cats['en']['msgs']
        self.assertEqual(set(tr), set(en))
        ph = re.compile(r'\{[a-z_]+\}')
        for k in tr:
            self.assertEqual(sorted(ph.findall(tr[k])), sorted(ph.findall(en[k])), k)

    def test_four_languages_parity(self):
        cats = m.load()
        self.assertTrue({'tr', 'en', 'de', 'ru'} <= set(cats))
        ref = cats['tr']['msgs']
        ph = re.compile(r'\{[a-z_]+\}')
        for code in ('en', 'de', 'ru'):
            self.assertEqual(set(cats[code]['msgs']), set(ref), code)
            for k in ref:
                self.assertEqual(sorted(ph.findall(cats[code]['msgs'][k])), sorted(ph.findall(ref[k])), (code, k))
        de = m.render_license_email('de', order='#1001')
        ru = m.render_license_email('ru-RU', order='#1001')
        self.assertIn('#1001', str(de))
        self.assertIn('#1001', str(ru))
        self.assertIn('Lizenzschlüssel', str(de))
        self.assertIn('лицензионный ключ', str(ru).lower())

    def test_forbidden_slogans_absent(self):
        bad = ['yerli teknoloji', 'küresel vizyon', 'built in türkiye', 'designed for the world', 'homegrown', 'türkiye merkezli', 'yerli ve milli']
        for code, c in m.load().items():
            for k, v in c['msgs'].items():
                for p in bad:
                    self.assertNotIn(p, v.lower(), f'{code}/{k}')

    @unittest.skipUnless(GO_LOCALES, 'US_GO_LOCALES not set')
    def test_in_sync_with_go_locales(self):
        go = m.load(GO_LOCALES)
        for code, c in m.load().items():
            for k, v in c['msgs'].items():
                self.assertEqual(go[code]['msgs'].get(k), v, f'{code}/{k} drifted from Go locales')


if __name__ == '__main__':
    unittest.main()
