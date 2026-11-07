# DRAFT — requires counsel review; not published

**Subject:** KVKK Art. 9 cross-border transfer clause (privacy notice addendum draft)  
**Date:** 2026-11-06 Europe/Istanbul  
**No definitive legal conclusion.** Options only, based on technical inventory.  
**Companion (TR):** `M9_YURTDISI_AKTARIM_TASLAK.md`

---

## Notice

Do not paste into Shopify/site until counsel + Demir complete `<DOLDUR>` / `[TO BE VERIFIED]` fields. Do not claim “no transfers.”

---

## 1. Draft clause (skeleton)

Personal data may be transferred to or processed by the recipient categories below, limited to stated purposes. Country/region and Art. 9 legal basis are **not finally selected** in this draft.

---

## 2. Recipient inventory (draft)

| Category / brand | Role (draft) | Data categories (technical) | Purpose | Country / region | Evidence |
|---|---|---|---|---|---|
| **Shopify** | E-commerce platform; controller/processor split `<DOLDUR>` | Order id/name, customer email, product/plan | Checkout / orders | Shopify residency **[TO BE VERIFIED]**; storefront showed `x-dc: gcp-us-east1` (not alone proof of Google transfer) | LIVE_STORE_AUDIT; HTTP 2026-11-06 |
| **Payment provider** | `<DOLDUR>` | Card/payment data | Collection | `[TO BE VERIFIED]` — live audit **UNKNOWN** | LIVE_STORE_AUDIT §1 |
| **Hostinger** (VPS + SMTP) | Hosting / email; legal entity `<DOLDUR>` | Store file (email, password hash, licenses, outbox); SMTP To/subject/body | Hook hosting, license mail | SMTP host known; country **[TO BE VERIFIED]** | PROD_PLAN, SMTP report, mail.go |
| **Google Cloud** | Cloud / GPU / Cloud Run | Model prompts/responses (proxy); gateway traffic scope **[TO BE VERIFIED]** | Inference, watchdog, commerce | Technical: **europe-west4**; also us-central1 ops VMs | SRE_AUDIT |
| **Model upstream** | Default model name `deepseek-v4.1-flash`; actual `US_MODEL_UPSTREAM` | Prompts / project content | AI fulfilment | Upstream country **[TO BE VERIFIED]**; DeepSeek DPA `<DOLDUR>` | main.go default |
| Public authorities | Legal duty | As required | Compliance | TR / relevant | — |

---

## 3. KVKK Art. 9 — **options** (no selection)

| Option | Summary | Open question |
|---|---|---|
| **A. Adequacy decision** | Transfer to a country/org with an adequacy finding | Is the destination on the list? `[TO BE VERIFIED]` |
| **B. Appropriate safeguards — standard contract + Board notification** | Board-published standard clauses + notification | Signed DPA/SCC with each vendor? `<DOLDUR>` |
| **C. Binding corporate rules** | Approved BCR | Intra-group only typically |
| **D. Explicit consent for occasional transfers** | Narrow statutory consent path | Risky as sole basis for continuous SaaS/hosting — counsel |
| **E. Other statutory exceptions** | E.g. contract performance where the law allows | May be too narrow for continuous platform transfers |

Post-2024 Art. 9 text must be re-checked at publication; do not rely on the old “consent-only” framing.

---

## 4. GDPR Chapter V note

If GDPR applies to EU/EEA users (`[TO BE VERIFIED]`):

- **SCCs** (EU Commission modules) + transfer impact assessment  
- **EU–US Data Privacy Framework (DPF)** where the importer is certified (`[TO BE VERIFIED]` for Shopify/Google/etc.)  
- **BCRs / Art. 49** derogations — narrow; not a default for continuous transfers  

KVKK Art. 9 ≠ automatic GDPR Chapter V equivalence; dual analysis may be needed.

---

## 5. Questions for counsel

Same as TR file §6 (roles, DPAs, europe-west4 characterization, DeepSeek country, payment provider, storefront edge headers, SCC/DPF, consent box use).

