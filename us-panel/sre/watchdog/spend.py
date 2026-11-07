"""GPU spend estimate + spend-limit alarms for ds-a100. Not deployed.

Sources, in priority order per Europe/Istanbul day:
  1. settled real cost from billing_reconcile (detailed export)  -> replaces the estimate
  2. watchdog JSONL ledger intervals x price_per_hour_usd (running time)
  3. separate line: preserved local SSD while the VM is stopped (scenario from config, NOT VERIFIED price)
"""
from __future__ import annotations
import calendar
import datetime as dt
import json
import time
from typing import Callable, Optional

from billing_reconcile import TZ, estimated_by_day
from watchdog import Alarm, Config, SSD_SCENARIOS

NOT_DEFINED = "limit tanımlı değil"


def _split_by_day(start: float, end: float, rate_per_s: float, out: dict) -> None:
    cur = start
    while cur < end:
        d = dt.datetime.fromtimestamp(cur, TZ).date()
        nxt = dt.datetime.combine(d + dt.timedelta(days=1), dt.time(), TZ).timestamp()
        seg = min(end, nxt)
        out[d.isoformat()] = out.get(d.isoformat(), 0.0) + rate_per_s * (seg - cur)
        cur = seg


class SpendTracker:
    def __init__(self, cfg: Config, alarm: Alarm, wall_clock: Callable[[], float] = time.time,
                 open_span: Callable[[], Optional[dict]] = lambda: None):
        self.cfg, self.alarm, self.now = cfg, alarm, wall_clock
        self.open_span = open_span          # e.g. lambda: watchdog.cost.open_span
        self.actual: dict = {}              # day -> settled real USD (billing export)
        self._last_alarm: dict = {}         # (period, period_id, threshold) -> wall ts

    # ---- inputs ----
    def apply_reconcile(self, result: dict) -> int:
        n = 0
        for d in result.get("days", []):
            if d.get("settled") and d.get("actual_usd") is not None:
                self.actual[d["day"]] = float(d["actual_usd"]); n += 1
        return n

    def _ledger(self):
        try:
            with open(self.cfg.cost_log_path, encoding="utf-8") as fh:
                return [json.loads(l) for l in fh if l.strip()]
        except FileNotFoundError:
            return []

    # ---- estimates ----
    def running_by_day(self, start_day: dt.date, end_day: dt.date) -> dict:
        out = estimated_by_day(self.cfg.cost_log_path, start_day, end_day)
        span = self.open_span()
        price = self.cfg.price_per_hour_usd
        if span and price is not None and span.get("start_wall") is not None:
            tmp: dict = {}
            _split_by_day(span["start_wall"], self.now(), price / 3600.0, tmp)
            for k, v in tmp.items():
                if start_day.isoformat() <= k < end_day.isoformat():
                    out[k] = out.get(k, 0.0) + v
        return out

    def _ssd_rates(self) -> dict:
        """scenario -> USD/GiB-month. A verified (export) rate replaces all scenarios."""
        if not self.cfg.preserve_local_ssd:
            return {}
        v = self.cfg.preserved_ssd_verified_rate_usd_gib_month
        if v is not None:
            return {"verified": float(v)}
        r = self.cfg.preserved_ssd_rates_usd_gib_month or {}
        return {k: float(r[k]) for k in SSD_SCENARIOS if k in r}

    def ssd_hours_by_day(self, start_day: dt.date, end_day: dt.date) -> dict:
        """Hours the VM sat stopped with preserved SSD (idle stops only), split by TSI day."""
        out: dict = {}
        if not self.cfg.preserve_local_ssd:
            return out
        recs = self._ledger()
        ivs = sorted((r for r in recs if r.get("event") == "interval" and r.get("end_wall")), key=lambda r: r["end_wall"])
        opens = sorted(r["start_wall"] for r in recs if r.get("start_wall") is not None)
        for iv in ivs:
            if iv.get("stop_reason") != "idle":          # preemption/external stop: SSD lost, no charge
                continue
            nxt = [t for t in opens if t > iv["end_wall"]]
            end = nxt[0] if nxt else self.now()
            tmp: dict = {}
            _split_by_day(iv["end_wall"], end, 1 / 3600.0, tmp)
            for k, v in tmp.items():
                if start_day.isoformat() <= k < end_day.isoformat():
                    out[k] = out.get(k, 0.0) + v
        return out

    def summary(self) -> dict:
        now = self.now()
        today = dt.datetime.fromtimestamp(now, TZ).date()
        m0 = today.replace(day=1)
        days_in_month = calendar.monthrange(today.year, today.month)[1]
        end = today + dt.timedelta(days=1)
        run, ssd_h = self.running_by_day(m0, end), self.ssd_hours_by_day(m0, end)
        rates = self._ssd_rates()
        verified = "verified" in rates
        alarm_sc = ("verified" if verified else self.cfg.spend_alarm_ssd_scenario) if self.cfg.preserve_local_ssd else None
        per_h = {k: v * self.cfg.preserved_ssd_gib / self.cfg.hours_per_month for k, v in rates.items()}
        days, mtd_gpu, ssd_mtd = [], 0.0, {k: 0.0 for k in per_h}
        mtd_alarm = 0.0
        for i in range((today - m0).days + 1):
            d = (m0 + dt.timedelta(days=i)).isoformat()
            if d in self.actual:                         # export already includes SSD for that day
                gpu, ssd_d, src = self.actual[d], {k: 0.0 for k in per_h}, "billing_export"
            else:
                gpu, src = run.get(d, 0.0), "estimate"
                ssd_d = {k: ssd_h.get(d, 0.0) * r for k, r in per_h.items()}
            mtd_gpu += gpu
            for k in ssd_d: ssd_mtd[k] += ssd_d[k]
            alarm_total = gpu + ssd_d.get(alarm_sc, 0.0)
            mtd_alarm += alarm_total
            days.append({"day": d, "usd": round(alarm_total, 6), "gpu_or_export_usd": round(gpu, 6),
                         "ssd_usd_by_scenario": {k: round(v, 6) for k, v in ssd_d.items()}, "source": src})
        month_start = dt.datetime.combine(m0, dt.time(), TZ).timestamp()
        factor = (days_in_month * 86400) / max(1.0, now - month_start)
        today_doc = days[-1]
        scen = {k: {"rate_usd_gib_month": rates[k], "usd_per_hour_stopped": round(per_h[k], 6),
                    "month_to_date_usd": round(ssd_mtd[k], 6),
                    "month_total_with_gpu_usd": round(mtd_gpu + ssd_mtd[k], 6),
                    "month_end_projection_usd": round((mtd_gpu + ssd_mtd[k]) * factor, 6)} for k in per_h}
        return {
            "currency": "USD", "as_of": dt.datetime.fromtimestamp(now, TZ).isoformat(timespec="seconds"),
            "price_per_hour_usd": self.cfg.price_per_hour_usd, "price_source": self.cfg.price_source,
            "today_usd": today_doc["usd"], "today_source": today_doc["source"],
            "month_to_date_gpu_usd": round(mtd_gpu, 6),
            "month_to_date_usd": round(mtd_alarm, 6),
            "month_end_projection_usd": round(mtd_alarm * factor, 6),
            "preserved_ssd": {
                "enabled": self.cfg.preserve_local_ssd, "gib": self.cfg.preserved_ssd_gib,
                "price_verified": verified,
                "note": ("SSD korunmuyor, kapalıyken 0" if not self.cfg.preserve_local_ssd else
                         ("export SKU " + str(self.cfg.preserved_ssd_verified_sku)) if verified else "fiyat doğrulanmadı"),
                "stopped_hours_month": round(sum(ssd_h.values()), 4),
                "scenarios": scen},
            "alarm_ssd_scenario": alarm_sc,
            "limits": self._limits_doc(today_doc["usd"], mtd_alarm),
            "days": days,
        }

    def _limits_doc(self, today_usd: float, mtd: float) -> dict:
        def one(limit, spent):
            if limit is None:
                return {"limit_usd": None, "status": NOT_DEFINED}
            return {"limit_usd": limit, "spent_usd": round(spent, 6), "ratio": round(spent / limit, 4)}
        return {"daily": one(self.cfg.spend_daily_limit_usd, today_usd),
                "monthly": one(self.cfg.spend_monthly_limit_usd, mtd)}

    # ---- alarms ----
    def check_limits(self) -> list:
        s = self.summary()
        now = self.now()
        today = dt.datetime.fromtimestamp(now, TZ).date()
        fired = []
        for period, limit, spent, pid in (("daily", self.cfg.spend_daily_limit_usd, s["today_usd"], today.isoformat()),
                                          ("monthly", self.cfg.spend_monthly_limit_usd, s["month_to_date_usd"], today.strftime("%Y-%m"))):
            if limit is None:
                continue
            crossed = [t for t in sorted(self.cfg.spend_thresholds) if spent >= t * limit]
            if not crossed:
                continue
            t = crossed[-1]                       # highest threshold reached
            key = (period, pid, t)
            last = self._last_alarm.get(key)
            if last is not None and now - last < self.cfg.spend_alarm_cooldown_seconds:
                continue
            self._last_alarm[key] = now
            self.alarm.notify("spend_limit", f"{period} spend {spent:.2f} USD >= {int(t * 100)}% of {limit:.2f} USD",
                              {"period": period, "period_id": pid, "threshold": t, "limit_usd": limit,
                               "spent_usd": round(spent, 6), "projection_usd": s["month_end_projection_usd"]})
            fired.append({"period": period, "threshold": t})
        return fired
