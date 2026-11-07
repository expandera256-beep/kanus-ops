import datetime as dt, json, os, sys
import pytest
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
from watchdog import Config, Alarm
from billing_reconcile import build_query, estimated_by_day, reconcile, TZ

TABLE = "gcp_billing_export_resource_v1_000000_000000_000000"   # sahte ad


class FakeBQ:
    def __init__(self, rows=(), scan=10_485_760, error=None):
        self.rows, self.scan, self.error, self.calls = list(rows), scan, error, []
    def query(self, sql, params, maximum_bytes_billed, dry_run):
        self.calls.append({"sql": sql, "params": params, "max": maximum_bytes_billed, "dry": dry_run})
        if self.error: raise self.error
        return self.scan if dry_run else self.rows


def ts(y, m, d, h=0):
    return dt.datetime(y, m, d, h, tzinfo=TZ).timestamp()


def setup(tmp_path, **kw):
    cfg = Config(cost_log_path=str(tmp_path / "cost.jsonl"), alarm_outbox_path=str(tmp_path / "out.jsonl"),
                 billing_resource_table=TABLE, alarm_email_to="demirtumen08@gmail.com", **kw)
    return cfg, Alarm(cfg, lambda: 0.0)


def write_interval(cfg, start, end, cost):
    with open(cfg.cost_log_path, "a") as fh:
        fh.write(json.dumps({"event": "interval", "start_wall": start, "end_wall": end, "cost_usd": cost}) + "\n")


def outbox(tmp_path):
    p = tmp_path / "out.jsonl"
    return [json.loads(l) for l in open(p)] if p.exists() else []


def test_query_uses_config_names_date_filter_and_partition(tmp_path):
    cfg, _ = setup(tmp_path, billing_dataset="billing_export")
    sql, params = build_query(cfg, dt.date(2026, 11, 3), dt.date(2026, 11, 7))
    assert f"`project-27082a40-6083-4952-9db.billing_export.{TABLE}`" in sql
    assert "_PARTITIONTIME >=" in sql and "usage_start_time <" in sql and "global_name" in sql
    assert params == {"start_day": "2026-11-03", "end_day": "2026-11-07", "names": ["ds-a100"]}

def test_query_rejects_unconfigured_or_bad_identifiers(tmp_path):
    cfg = Config()
    with pytest.raises(ValueError, match="not configured"): build_query(cfg, dt.date(2026,11,1), dt.date(2026,11,2))
    cfg, _ = setup(tmp_path, billing_dataset="x`; DROP")
    with pytest.raises(ValueError, match="identifier"): build_query(cfg, dt.date(2026,11,1), dt.date(2026,11,2))
    cfg = Config(billing_resource_table="gcp_billing_export_v1_X")
    with pytest.raises(ValueError, match="detailed"): build_query(cfg, dt.date(2026,11,1), dt.date(2026,11,2))

def test_not_configured_returns_status_without_query(tmp_path):
    cfg = Config(cost_log_path=str(tmp_path / "c"), alarm_outbox_path=str(tmp_path / "o"))
    bq = FakeBQ()
    r = reconcile(cfg, bq, Alarm(cfg, lambda: 0), dt.date(2026, 11, 6), ts(2026, 11, 6))
    assert r["status"] == "not_configured" and bq.calls == []

def test_dry_run_first_with_maximum_bytes_billed(tmp_path):
    cfg, al = setup(tmp_path, billing_maximum_bytes_billed=5_000_000_000)
    bq = FakeBQ()
    reconcile(cfg, bq, al, dt.date(2026, 11, 6), ts(2026, 11, 6))
    assert [c["dry"] for c in bq.calls] == [True, False] and all(c["max"] == 5_000_000_000 for c in bq.calls)

def test_scan_over_limit_skips_real_query_and_alarms(tmp_path):
    cfg, al = setup(tmp_path, billing_maximum_bytes_billed=1000)
    bq = FakeBQ(scan=10_000)
    r = reconcile(cfg, bq, al, dt.date(2026, 11, 6), ts(2026, 11, 6))
    assert r["status"] == "skipped_scan_limit" and len(bq.calls) == 1 and len(outbox(tmp_path)) == 1

def test_estimate_split_across_istanbul_midnight(tmp_path):
    cfg, _ = setup(tmp_path)
    write_interval(cfg, ts(2026, 11, 4, 23), ts(2026, 11, 5, 1), 2.0)   # 1 h each side
    est = estimated_by_day(cfg.cost_log_path, dt.date(2026, 11, 1), dt.date(2026, 11, 7))
    assert est == pytest.approx({"2026-11-04": 1.0, "2026-11-05": 1.0})

def test_match_within_threshold_no_alarm_and_sku_breakdown(tmp_path):
    cfg, al = setup(tmp_path)
    write_interval(cfg, ts(2026, 11, 3, 10), ts(2026, 11, 3, 11), 1.848238)
    rows = [{"day": "2026-11-03", "sku_id": "9BBC-6855-1C9A", "sku_description": "A100 Spot", "cost": 1.36, "credits": 0, "currency": "USD"},
            {"day": "2026-11-03", "sku_id": "8D0B-EA58-9E35", "sku_description": "SSD Spot", "cost": 0.6, "credits": -0.1, "currency": "USD"}]
    r = reconcile(cfg, FakeBQ(rows), al, dt.date(2026, 11, 6), ts(2026, 11, 6))
    d = r["days"][0]
    assert d["status"] == "ok" and d["actual_usd"] == pytest.approx(1.86) and len(d["skus"]) == 2
    assert outbox(tmp_path) == []

def test_mismatch_over_threshold_writes_outbox(tmp_path):
    cfg, al = setup(tmp_path)
    write_interval(cfg, ts(2026, 11, 3, 10), ts(2026, 11, 3, 11), 1.848238)
    rows = [{"day": "2026-11-03", "sku_id": "9BBC-6855-1C9A", "sku_description": "A100", "cost": 5.0, "credits": 0}]
    r = reconcile(cfg, FakeBQ(rows), al, dt.date(2026, 11, 6), ts(2026, 11, 6))
    assert r["days"][0]["status"] == "mismatch"
    m = outbox(tmp_path); assert len(m) == 1 and m[0]["sent"] is False and "billing_reconcile" in m[0]["subject"]

def test_recent_day_is_provisional_not_alarm(tmp_path):
    cfg, al = setup(tmp_path)
    write_interval(cfg, ts(2026, 11, 6, 1), ts(2026, 11, 6, 2), 1.85)
    rows = [{"day": "2026-11-06", "sku_id": "x", "sku_description": "x", "cost": 0.2, "credits": 0}]
    r = reconcile(cfg, FakeBQ(rows), al, dt.date(2026, 11, 6), ts(2026, 11, 6, 8))
    assert r["days"][0]["status"] == "provisional" and outbox(tmp_path) == []

def test_missing_billing_rows_after_lag_alarms_but_pending_before(tmp_path):
    cfg, al = setup(tmp_path)
    write_interval(cfg, ts(2026, 11, 3, 10), ts(2026, 11, 3, 11), 1.85)
    write_interval(cfg, ts(2026, 11, 6, 1), ts(2026, 11, 6, 2), 1.85)
    r = reconcile(cfg, FakeBQ([]), al, dt.date(2026, 11, 6), ts(2026, 11, 6, 8))
    st = {d["day"]: d["status"] for d in r["days"]}
    assert st == {"2026-11-03": "missing_billing_data", "2026-11-06": "pending"} and len(outbox(tmp_path)) == 1

def test_query_error_counts_as_failure(tmp_path):
    cfg, al = setup(tmp_path, alarm_threshold=1)
    r = reconcile(cfg, FakeBQ(error=RuntimeError("403")), al, dt.date(2026, 11, 6), ts(2026, 11, 6))
    assert r["status"] == "query_error" and al.degraded

def test_ledger_writes_wall_clock(tmp_path):
    from tests.test_watchdog import make
    wd, g, clk = make(tmp_path, status="RUNNING", price_per_hour_usd=1.0, price_source="t")
    wd.tick(); clk.advance(60); g.status = "TERMINATED"; wd.tick()
    iv = [json.loads(l) for l in open(tmp_path / "cost.jsonl") if '"interval"' in l][0]
    assert iv["start_wall"] > 1.7e9 and iv["end_wall"] >= iv["start_wall"]
