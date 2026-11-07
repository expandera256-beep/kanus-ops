#!/usr/bin/env python3
"""Static /tr + /en site for kanus.org (local only, never uploads).
Copy: docs/BRAND_A_COPY_20261106.md (Karar A). The slogan lives in ONE key:
content.json -> "slogan": {"tr": ..., "en": ...}. Pages marked "draft": true
(legal drafts) are excluded from the output, nav, hreflang and sitemap.
Usage: python3 build.py [--slogan-tr TEXT] [--slogan-en TEXT]"""
import argparse, json, os, html, shutil
HERE = os.path.dirname(os.path.abspath(__file__))
BASE = "https://kanus.org"
OUT = os.path.join(HERE, "out")
CONTENT = os.path.join(HERE, "content.json")
ALL_PAGES = ["", "product", "pricing", "download", "docs", "api", "support", "privacy", "terms", "refund"]
DEFAULT = "tr"

def url(lang, page): return f"{BASE}/{lang}/" + (page + "/" if page else "")

def render(C, LANGS, PAGES, lang, page):
    c = LANGS[lang]; p = c["pages"][page or "home"]; slogan = C["slogan"][lang]
    alts = "\n".join(f'<link rel="alternate" hreflang="{l}" href="{url(l,page)}">' for l in LANGS)
    alts += f'\n<link rel="alternate" hreflang="x-default" href="{url(DEFAULT,page)}">'
    nav = " ".join(f'<a href="/{lang}/{q+"/" if q else ""}">{html.escape(c["nav"][q or "home"])}</a>' for q in PAGES)
    sw = " | ".join(f'<a href="/{l}/{page+"/" if page else ""}" hreflang="{l}" lang="{l}">{LANGS[l]["meta"]["name"]}</a>' for l in LANGS)
    title = f'{p["title"]} — Kanus' if page else f'Kanus — {slogan.rstrip(".")}'
    ld = {"@context": "https://schema.org", "@type": "WebPage", "name": title, "url": url(lang, page),
          "inLanguage": c["meta"]["bcp47"], "description": p["desc"],
          "isPartOf": {"@type": "WebSite", "name": "Kanus", "url": BASE + "/"}}
    if not page:
        ld["slogan"] = slogan
        ld = [ld, {"@context": "https://schema.org", "@type": "Organization", "name": "Kanus", "url": BASE + "/",
                   "email": "support@kanus.org", "slogan": slogan},
              {"@context": "https://schema.org", "@type": "SoftwareApplication", "name": "US Panel",
               "operatingSystem": "Windows", "applicationCategory": "DeveloperApplication", "description": p["desc"]}]
    e = html.escape
    others = "".join(f'<meta property="og:locale:alternate" content="{LANGS[l]["meta"]["bcp47"].replace("-","_")}">' for l in LANGS if l != lang)
    return f'''<!DOCTYPE html>
<html lang="{c["meta"]["bcp47"]}" dir="{c["meta"]["dir"]}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{e(title)}</title>
<meta name="description" content="{e(p["desc"])}">
<link rel="canonical" href="{url(lang,page)}">
{alts}
<meta property="og:type" content="website">
<meta property="og:site_name" content="Kanus">
<meta property="og:title" content="{e(title)}">
<meta property="og:description" content="{e(p["desc"])}">
<meta property="og:url" content="{url(lang,page)}">
<meta property="og:locale" content="{c["meta"]["bcp47"].replace("-","_")}">
{others}
<script type="application/ld+json">{json.dumps(ld, ensure_ascii=False)}</script>
<script>try{{localStorage.setItem("kanus.lang","{lang}")}}catch(e){{}}</script>
<style>body{{font-family:system-ui,sans-serif;max-width:60rem;margin:auto;padding:1rem;line-height:1.5}}nav a{{margin-inline-end:.75rem}}</style>
</head>
<body>
<header><nav>{nav}</nav><p>{sw}</p></header>
<main><h1>{e(p["title"] if page else slogan)}</h1><p>{e(p["body"])}</p></main>
<footer><p>support@kanus.org · <a href="https://shop.kanus.org">shop.kanus.org</a></p></footer>
</body></html>
'''

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--slogan-tr"); ap.add_argument("--slogan-en")
    a = ap.parse_args()
    C = json.load(open(CONTENT, encoding="utf-8"))
    if a.slogan_tr or a.slogan_en:
        if a.slogan_tr: C["slogan"]["tr"] = a.slogan_tr.strip()
        if a.slogan_en: C["slogan"]["en"] = a.slogan_en.strip()
        json.dump(C, open(CONTENT, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    LANGS = {k: v for k, v in C.items() if k != "slogan"}
    PAGES = [q for q in ALL_PAGES if not any(LANGS[l]["pages"][q or "home"].get("draft") for l in LANGS)]
    shutil.rmtree(OUT, ignore_errors=True)
    for lang in LANGS:
        for page in PAGES:
            d = os.path.join(OUT, lang, page); os.makedirs(d, exist_ok=True)
            open(os.path.join(d, "index.html"), "w", encoding="utf-8").write(render(C, LANGS, PAGES, lang, page))
    sm = ['<?xml version="1.0" encoding="UTF-8"?>',
          '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">']
    for page in PAGES:
        for lang in LANGS:
            sm.append(f"<url><loc>{url(lang,page)}</loc>")
            sm += [f'<xhtml:link rel="alternate" hreflang="{l}" href="{url(l,page)}"/>' for l in LANGS]
            sm.append(f'<xhtml:link rel="alternate" hreflang="x-default" href="{url(DEFAULT,page)}"/></url>')
    sm.append("</urlset>")
    open(os.path.join(OUT, "sitemap.xml"), "w").write("\n".join(sm) + "\n")
    open(os.path.join(OUT, "robots.txt"), "w").write(f"User-agent: *\nAllow: /\nSitemap: {BASE}/sitemap.xml\n")
    print(f"built {len(PAGES)*len(LANGS)} pages, excluded: {[q for q in ALL_PAGES if q not in PAGES]}")
main()
