"""Localized e-mail text loader. Reads the same locale JSON format as the
Go panel (kiyas/i18n/locales). Rendering only; it never sends mail.
Default language is TR; missing keys fall back to EN, then TR."""
import json, os, re

DEFAULT = 'tr'
FALLBACK = 'en'
_PLACEHOLDER = re.compile(r'\{([a-z_]+)\}')
_cache = {}


def locales_dir():
    return os.environ.get('US_LOCALES_DIR') or os.path.join(os.path.dirname(os.path.abspath(__file__)), 'locales')


def normalize(tag):
    tag = (tag or '').strip().lower().replace('_', '-')
    return tag.split('-', 1)[0]


def load(directory=None):
    directory = directory or locales_dir()
    if directory in _cache:
        return _cache[directory]
    catalogs = {}
    for name in sorted(os.listdir(directory)):
        if not name.endswith('.json'):
            continue
        with open(os.path.join(directory, name), encoding='utf-8') as fh:
            doc = json.load(fh)
        meta = doc.pop('_meta', {})
        if meta.get('dir', 'ltr') not in ('ltr', 'rtl'):
            raise ValueError(f'{name}: bad dir')
        for k, v in doc.items():
            if not isinstance(v, str):
                raise ValueError(f'{name}: {k} is not a string')
        catalogs[normalize(name[:-5])] = {'meta': meta, 'msgs': doc}
    for need in (DEFAULT, FALLBACK):
        if need not in catalogs:
            raise ValueError(f'required locale {need} missing')
    _cache[directory] = catalogs
    return catalogs


def resolve_lang(preferred=None, accept_language='', directory=None):
    cats = load(directory)
    p = normalize(preferred)
    if p in cats:
        return p
    cands = []
    for i, part in enumerate((accept_language or '').split(',')):
        bits = part.strip().split(';')
        code = normalize(bits[0])
        q = 1.0
        for b in bits[1:]:
            b = b.strip()
            if b.startswith('q='):
                try:
                    q = float(b[2:])
                except ValueError:
                    pass
        if code and code != '*' and q > 0:
            cands.append((-q, i, code))
    for _, _, code in sorted(cands):
        if code in cats:
            return code
    return DEFAULT


def t(key, lang=None, directory=None, **params):
    cats = load(directory)
    lang = normalize(lang) if normalize(lang) in cats else DEFAULT
    for code in (lang, FALLBACK, DEFAULT):
        text = cats[code]['msgs'].get(key)
        if text is not None:
            break
    else:
        return key
    return _PLACEHOLDER.sub(lambda m: str(params[m.group(1)]) if m.group(1) in params else m.group(0), text)


def render_license_email(lang=None, order='', directory=None):
    """Return (subject, body, lang) for the license delivery mail."""
    code = resolve_lang(lang, directory=directory)
    return t('email.license.subject', code, directory), t('email.license.body', code, directory, order=order), code
