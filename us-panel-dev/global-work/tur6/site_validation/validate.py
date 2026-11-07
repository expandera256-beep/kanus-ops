#!/usr/bin/env python3
import os, re, json, sys, glob
from html.parser import HTMLParser
import xml.etree.ElementTree as ET
OUT="out"; fails=[]; BASE="https://kanus.org"
class P(HTMLParser):
    VOID={"meta","link","br","img","input","hr"}
    def __init__(s): super().__init__(); s.stack=[]; s.links=[]; s.meta={}; s.ld=[]; s.inld=False; s.html={}; s.err=[]
    def handle_starttag(s,t,a):
        a=dict(a)
        if t=="html": s.html=a
        if t=="link": s.links.append(a)
        if t=="meta": s.meta[a.get("property") or a.get("name") or a.get("charset") and "charset"]=a.get("content")
        if t=="script" and a.get("type")=="application/ld+json": s.inld=True
        if t not in s.VOID: s.stack.append(t)
    def handle_endtag(s,t):
        if t in s.VOID: return
        if not s.stack or s.stack[-1]!=t: s.err.append(f"mismatch </{t}>")
        else: s.stack.pop()
    def handle_data(s,d):
        if s.inld: s.ld.append(d); s.inld=False
pages={}
for f in glob.glob(f"{OUT}/*/**/index.html",recursive=True):
    p=P(); p.feed(open(f,encoding="utf-8").read()); pages[f]=p
    rel=f[len(OUT):-len("index.html")]; u=BASE+rel
    if p.err or p.stack: fails.append(f"{f}: html structure {p.err} {p.stack}")
    canon=[l["href"] for l in p.links if l.get("rel")=="canonical"]
    if canon!=[u]: fails.append(f"{f}: canonical {canon}")
    alt={l["hreflang"]:l["href"] for l in p.links if l.get("rel")=="alternate"}
    p.alt=alt; p.url=u
    for k in ("tr","en","x-default"):
        if k not in alt: fails.append(f"{f}: hreflang {k} missing")
    for k in ("og:title","og:description","og:url","og:locale","viewport","description"):
        if not p.meta.get(k): fails.append(f"{f}: meta {k} missing")
    if p.meta.get("og:url")!=u: fails.append(f"{f}: og:url")
    if p.html.get("dir") not in ("ltr","rtl") or not p.html.get("lang"): fails.append(f"{f}: html lang/dir")
    try: json.loads(p.ld[0])
    except Exception as e: fails.append(f"{f}: JSON-LD {e}")
    low=open(f,encoding="utf-8").read().lower()
    for bad in ("yerli teknoloji","küresel vizyon","built in türkiye","designed for the world","homegrown","türkiye merkezli","kan.us)"):
        if bad in low: fails.append(f"{f}: forbidden '{bad}'")
    p.noindex = "noindex" in (p.meta.get("robots") or "")
    p.hrefs = re.findall(r'href="(/[^"]*)"', open(f,encoding="utf-8").read())
    for bad in ("taslak","draft","yayinlanmaz","yayınlanmaz","kvkk","gdpr","veri sorumlusu","data controller","mersis","vergi no","tax id","hukuki","legal text"):
        if bad in low: fails.append(f"{f}: draft/legal content '{bad}' in build output")
byurl={p.url:p for p in pages.values()}
for p in pages.values():
    for l,h in p.alt.items():
        q=byurl.get(h)
        if not q: fails.append(f"{p.url}: alternate {h} not a page"); continue
        if p.url not in q.alt.values(): fails.append(f"{p.url}<->{h} not reciprocal")
t=ET.parse(f"{OUT}/sitemap.xml"); ns={"s":"http://www.sitemaps.org/schemas/sitemap/0.9"}
locs={e.text for e in t.getroot().findall("s:url/s:loc",ns)}
indexable={u for u,p in byurl.items() if not p.noindex}
if locs!=indexable: fails.append(f"sitemap mismatch {locs^indexable}")
for u,p in byurl.items():
    if p.noindex and u in locs: fails.append(f"{u}: noindex page in sitemap")
for leg in ("privacy","terms","refund"):
    for l in ("tr","en"):
        if os.path.exists(f"{OUT}/{l}/{leg}"): fails.append(f"legal page {l}/{leg} present in package")
for p in pages.values():
    for h in p.hrefs:
        if BASE+h not in byurl: fails.append(f"{p.url}: internal link to missing page {h}")
    if any(re.search(r"/(privacy|terms|refund)/", h) for h in p.alt.values()): fails.append(f"{p.url}: hreflang to legal page")
for x in glob.glob(f"{OUT}/**/*",recursive=True):
    if os.path.isfile(x) and re.search(r"taslak|draft",open(x,encoding="utf-8",errors="ignore").read(),re.I):
        fails.append(f"{x}: TASLAK/DRAFT marker in output")
print(f"pages={len(pages)} sitemap_urls={len(locs)} noindex={sum(p.noindex for p in pages.values())}")
print("\n".join(fails) or "ALL CHECKS PASS"); sys.exit(1 if fails else 0)
