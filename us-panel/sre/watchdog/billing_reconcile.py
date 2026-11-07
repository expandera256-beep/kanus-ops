"""Compare watchdog's estimated GPU cost (JSONL ledger) with real cost from the
Cloud Billing *detailed (resource)* BigQuery export. Read-only. Not deployed.

BigQuery client interface (real adapter below; tests use a fake):
  query(sql, params: dict, maximum_bytes_billed: int, dry_run: bool) -> list[dict] | int(bytes)
"""
from __future__ import annotations
import datetime as dt
import json
import re
from typing import Iterable, Optional
from zoneinfo import ZoneInfo

from watchdog import Alarm, Config

TZ = ZoneInfo("Europe/Istanbul")
_IDENT = re.compile(r"^[A-Za-z0-9_\-]+$")

SQL = """
SELECT
  DATE(usage_start_time, 'Europe/Istanbul') AS day,
  sku.id AS sku_id,
  sku.description AS sku_description,
  SUM(cost) AS cost,
  SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS credits,
  SUM(usage.amount_in_pricing_units) AS usage_pricing_units,
  ANY_VALUE(usage.pricing_unit) AS pricing_unit,
  ANY_VALUE(currency) AS currency
FROM `{project}.{dataset}.{table}`
WHERE _PARTITIONTIME >= TIMESTAMP_SUB(TIMESTAMP(@start_day), INTERVAL 1 DAY)
  AND usage_start_time >= TIMESTAMP(@start_day, 'Europe/Istanbul')
  AND usage_start_time <  TIMESTAMP(@end_day, 'Europe/Istanbul')
  AND service.description = 'Compute Engine'
  AND (resource.name IN UNNEST(@names)
       OR EXISTS(SELECT 1 FROM UNNEST(@names) n WHERE ENDS_WITH(resource.global_name, CONCAT('/', n))))
GROUP BY day, sku_id, sku_description
ORDER BY day, sku_id
"""


def build_query(cfg: Config, start_day: dt.date, end_day: dt.date):
    if not cfg.billing_resource_table:
        raise ValueError("billing_resource_table not configured (export table not present yet)")
    for part in (cfg.billing_project, cfg.billing_dataset, cfg.billing_resource_table):
        if not _IDENT.match(part):
            raise ValueError(f"invalid BigQuery identifier: {part!r}")
    if not cfg.billing_resource_table.startswith("gcp_billing_export_resource_v1_"):
        raise ValueError("billing_resource_table must be the detailed (resource) export table")
    if end_day <= start_day:
        raise ValueError("end_day must be after start_day")
    sql = SQL.format(project=cfg.billing_project, dataset=cfg.billing_dataset, table=cfg.billing_resource_table)
    params = {"start_day": start_day.isoformat(), "end_day": end_day.isoformat(),
              "names": list(cfg.billing_resource_names)}
    return sql, params


def estimated_by_day(ledger_path: str, start_day: dt.date, end_day: dt.date) -> dict:
    """Split each ledger interval across Europe/Istanbul days, pro-rata by seconds."""
    out: dict = {}
    try:
        lines = open(ledger_path, encoding="utf-8").read().splitlines()
    except FileNotFoundError:
        return out
    for line in lines:
        rec = json.loads(line)
        if rec.get("event") != "interval" or rec.get("cost_usd") is None:
            continue
        s, e = rec.get("start_wall"), rec.get("end_wall")
        if s is None or e is None or e <= s:
            continue
        rate = rec["cost_usd"] / (e - s)
        cur = s
        while cur < e:
            d = dt.datetime.fromtimestamp(cur, TZ).date()
            nxt = dt.datetime.combine(d + dt.timedelta(days=1), dt.time(), TZ).timestamp()
            seg_end = min(e, nxt)
            if start_day <= d < end_day:
                out[d.isoformat()] = out.get(d.isoformat(), 0.0) + rate * (seg_end - cur)
            cur = seg_end
    return out


def reconcile(cfg: Config, client, alarm: Alarm, today: dt.date, now_ts: float) -> dict:
    start_day = today - dt.timedelta(days=cfg.billing_lookback_days)
    end_day = today + dt.timedelta(days=1)
    try:
        sql, params = build_query(cfg, start_day, end_day)
    except ValueError as exc:
        return {"status": "not_configured", "error": str(exc)}
    try:
        scanned = client.query(sql, params, maximum_bytes_billed=cfg.billing_maximum_bytes_billed, dry_run=True)
        if scanned > cfg.billing_maximum_bytes_billed:
            alarm.notify("billing_reconcile", "dry-run scan exceeds maximum_bytes_billed",
                         {"bytes": scanned, "max": cfg.billing_maximum_bytes_billed})
            return {"status": "skipped_scan_limit", "bytes": scanned}
        rows = client.query(sql, params, maximum_bytes_billed=cfg.billing_maximum_bytes_billed, dry_run=False)
    except Exception as exc:  # noqa: BLE001
        alarm.failure("billing_reconcile.query", type(exc).__name__)
        return {"status": "query_error", "error": type(exc).__name__}
    actual: dict = {}
    by_sku: dict = {}
    for r in rows:
        day = str(r["day"])
        net = float(r["cost"]) + float(r.get("credits") or 0)
        actual[day] = actual.get(day, 0.0) + net
        by_sku.setdefault(day, []).append({"sku_id": r["sku_id"], "sku_description": r["sku_description"],
                                           "net_cost": round(net, 6), "currency": r.get("currency")})
    est = estimated_by_day(cfg.cost_log_path, start_day, end_day)
    days = []
    for day in sorted(set(actual) | set(est)):
        a, e = actual.get(day), est.get(day, 0.0)
        day_end = dt.datetime.combine(dt.date.fromisoformat(day) + dt.timedelta(days=1), dt.time(), TZ).timestamp()
        settled = now_ts - day_end >= cfg.billing_export_lag_hours * 3600
        item = {"day": day, "estimated_usd": round(e, 6), "actual_usd": None if a is None else round(a, 6),
                "settled": settled, "skus": by_sku.get(day, [])}
        if a is None:
            item["status"] = "missing_billing_data" if (settled and e > 0) else "pending"
            if item["status"] == "missing_billing_data":
                alarm.notify("billing_reconcile", f"no billing rows for {day}", item)
        else:
            diff = a - e
            limit = max(cfg.billing_diff_abs_usd, cfg.billing_diff_pct * max(a, e))
            item["diff_usd"] = round(diff, 6)
            item["status"] = "mismatch" if (settled and abs(diff) > limit) else ("ok" if settled else "provisional")
            if item["status"] == "mismatch":
                alarm.notify("billing_reconcile", f"cost mismatch {day}: actual {a:.4f} vs est {e:.4f}", item)
        days.append(item)
    return {"status": "ok", "bytes_scanned_dry_run": scanned, "days": days}


class BigQueryClientAdapter:  # pragma: no cover - real client, not used in tests
    """Thin adapter over google-cloud-bigquery. Read-only queries only."""

    def __init__(self, project: str):
        from google.cloud import bigquery
        self.bq, self.client = bigquery, bigquery.Client(project=project)

    def query(self, sql, params, maximum_bytes_billed, dry_run):
        bq = self.bq
        qp = [bq.ScalarQueryParameter("start_day", "DATE", params["start_day"]),
              bq.ScalarQueryParameter("end_day", "DATE", params["end_day"]),
              bq.ArrayQueryParameter("names", "STRING", params["names"])]
        job = self.client.query(sql, job_config=bq.QueryJobConfig(
            query_parameters=qp, dry_run=dry_run, use_query_cache=False,
            maximum_bytes_billed=maximum_bytes_billed))
        if dry_run:
            return job.total_bytes_processed
        return [dict(r) for r in job.result()]
