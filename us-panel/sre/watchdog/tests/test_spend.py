import datetime as dt, json, os, sys
import pytest
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
from watchdog import Config, Alarm
from spend import SpendTracker, NOT_DEFINED
from billing_reconcile import TZ

RATES = {"pd_standard": 0.044, "spot_local_ssd": 0.0528, "ondemand_local_ssd": 0.088}


def ts(y, m, d, h=0, mi=0):
    return dt.datetime(y, m, d, h, mi, tzinfo=TZ).timestamp()


class Wall:
    def __init__(self, t): self.t = t
    def __call__(self): return self.t


def mk(tmp_path, now, **kw):
    kw.setdefault("price_per_hour_usd", 2.0); kw.setdefault("price_source", "test")
    cfg = Config(cost_log_path=str(tmp_path / "cost.jsonl"), alarm_outbox_path=str(tmp_path / "out.jsonl"),
                 alarm_email_from="support@kanus.org", alarm_email_to="demirtumen08@gmail.com",
                 preserved_ssd_rates_usd_gib_month=RATES, **kw)
    w = Wall(now)
    return cfg, SpendTracker(cfg, Alarm(cfg, w), wall_clock=w), w


def iv(cfg, start, end, cost, reason="idle"):
    with open(cfg.cost_log_path, "a") as fh:
        fh.write(json.dumps({"event": "open", "start_wall": start}) + "\n")
        fh.write(json.dumps({"event": "interval", "start_wall": start, "end_wall": end, "cost_usd": cost, "stop_reason": reason}) + "\n")


def outbox(tmp_path):
    p = tmp_path / "out.jsonl"
    return [json.loads(l) for l in open(p)] if p.exists() else []


def test_daily_monthly_and_projection(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 11))          # 10 full days elapsed in a 30-day month
    iv(cfg, ts(2026, 11, 2, 10), ts(2026, 11, 2, 11), 2.0)
    iv(cfg, ts(2026, 11, 10, 10), ts(2026, 11, 10, 12), 4.0)
    s = sp.summary()
    assert s["month_to_date_usd"] == pytest.approx(6.0)
    assert s["month_end_projection_usd"] == pytest.approx(18.0)
    assert s["today_usd"] == 0 and s["currency"] == "USD"

def test_open_span_counts_running_cost(tmp_path):
    cfg, sp, w = mk(tmp_path, ts(2026, 11, 6, 9))
    sp.open_span = lambda: {"start_wall": ts(2026, 11, 6, 8)}
    assert sp.summary()["today_usd"] == pytest.approx(2.0)

def test_settled_billing_replaces_estimate(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 6, 9))
    iv(cfg, ts(2026, 11, 3, 10), ts(2026, 11, 3, 11), 2.0)
    n = sp.apply_reconcile({"days": [{"day": "2026-11-03", "actual_usd": 1.5, "settled": True},
                                     {"day": "2026-11-05", "actual_usd": 9.0, "settled": False}]})
    s = sp.summary(); d = {x["day"]: x for x in s["days"]}
    assert n == 1 and d["2026-11-03"]["usd"] == 1.5 and d["2026-11-03"]["source"] == "billing_export"
    assert d["2026-11-05"]["source"] == "estimate" and s["month_to_date_usd"] == pytest.approx(1.5)

def test_preserved_ssd_three_scenarios_shown_together(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3, 0), preserve_local_ssd=True)
    iv(cfg, ts(2026, 11, 1, 0), ts(2026, 11, 1, 1), 2.0)     # stopped (idle) 01:00 -> now = 47 h
    s = sp.summary(); ps = s["preserved_ssd"]
    assert set(ps["scenarios"]) == {"pd_standard", "spot_local_ssd", "ondemand_local_ssd"}
    assert ps["price_verified"] is False and ps["note"] == "fiyat doğrulanmadı" and ps["stopped_hours_month"] == pytest.approx(47)
    for k, r in RATES.items():
        assert ps["scenarios"][k]["month_to_date_usd"] == pytest.approx(1500 * r / 730 * 47, rel=1e-6)
        assert ps["scenarios"][k]["month_total_with_gpu_usd"] == pytest.approx(2.0 + 1500 * r / 730 * 47, rel=1e-6)

def test_alarm_total_uses_conservative_scenario_by_default(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3, 0), preserve_local_ssd=True)
    iv(cfg, ts(2026, 11, 1, 0), ts(2026, 11, 1, 1), 2.0)
    s = sp.summary()
    assert cfg.spend_alarm_ssd_scenario == "ondemand_local_ssd" and s["alarm_ssd_scenario"] == "ondemand_local_ssd"
    assert s["month_to_date_usd"] == pytest.approx(2.0 + 1500 * 0.088 / 730 * 47, rel=1e-6)
    assert s["month_to_date_gpu_usd"] == pytest.approx(2.0)

def test_alarm_scenario_configurable_and_validated(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3, 0), preserve_local_ssd=True, spend_alarm_ssd_scenario="pd_standard")
    iv(cfg, ts(2026, 11, 1, 0), ts(2026, 11, 1, 1), 2.0)
    assert sp.summary()["month_to_date_usd"] == pytest.approx(2.0 + 1500 * 0.044 / 730 * 47, rel=1e-6)
    with pytest.raises(ValueError): Config(spend_alarm_ssd_scenario="cheapest")

def test_verified_export_rate_replaces_scenarios(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3, 0), preserve_local_ssd=True,
                    preserved_ssd_verified_rate_usd_gib_month=0.05, preserved_ssd_verified_sku="TEST-SKU")
    iv(cfg, ts(2026, 11, 1, 0), ts(2026, 11, 1, 1), 2.0)
    s = sp.summary(); ps = s["preserved_ssd"]
    assert list(ps["scenarios"]) == ["verified"] and ps["price_verified"] is True and "TEST-SKU" in ps["note"]
    assert s["alarm_ssd_scenario"] == "verified"
    assert s["month_to_date_usd"] == pytest.approx(2.0 + 1500 * 0.05 / 730 * 47, rel=1e-6)

def test_ssd_not_charged_after_preemption_or_when_discarding(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3), preserve_local_ssd=True)
    iv(cfg, ts(2026, 11, 1, 0), ts(2026, 11, 1, 1), 2.0, reason="observed_terminated")
    assert sp.summary()["preserved_ssd"]["stopped_hours_month"] == 0
    (tmp_path / "x").mkdir()
    cfg, sp, _ = mk(tmp_path / "x", ts(2026, 11, 3))
    iv(cfg, ts(2026, 11, 1, 0), ts(2026, 11, 1, 1), 2.0)
    s = sp.summary()
    assert s["preserved_ssd"]["stopped_hours_month"] == 0 and s["month_to_date_usd"] == pytest.approx(2.0)

def test_ssd_rates_required_when_preserving():
    with pytest.raises(ValueError): Config(preserve_local_ssd=True)

def test_settled_export_day_has_no_extra_ssd(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3, 0), preserve_local_ssd=True)
    iv(cfg, ts(2026, 11, 1, 0), ts(2026, 11, 1, 1), 2.0)
    sp.apply_reconcile({"days": [{"day": "2026-11-01", "actual_usd": 3.0, "settled": True}]})
    d = {x["day"]: x for x in sp.summary()["days"]}["2026-11-01"]
    assert d["source"] == "billing_export" and d["usd"] == 3.0 and all(v == 0 for v in d["ssd_usd_by_scenario"].values())

def test_null_limits_no_alarm_and_health_says_not_defined(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 6, 9))
    iv(cfg, ts(2026, 11, 6, 0), ts(2026, 11, 6, 8), 1000.0)
    assert sp.check_limits() == [] and outbox(tmp_path) == []
    lim = sp.summary()["limits"]
    assert lim["daily"]["status"] == NOT_DEFINED and lim["monthly"]["status"] == NOT_DEFINED

@pytest.mark.parametrize("spent,expected", [(4.9, None), (5.0, 0.5), (8.0, 0.8), (10.0, 1.0), (25.0, 1.0)])
def test_daily_threshold_levels(tmp_path, spent, expected):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 6, 12), spend_daily_limit_usd=10.0)
    iv(cfg, ts(2026, 11, 6, 1), ts(2026, 11, 6, 2), spent)
    f = sp.check_limits()
    assert (f[0]["threshold"] if f else None) == expected

def test_cooldown_suppresses_repeats_then_allows(tmp_path):
    cfg, sp, w = mk(tmp_path, ts(2026, 11, 6, 12), spend_daily_limit_usd=10.0, spend_alarm_cooldown_seconds=3600)
    iv(cfg, ts(2026, 11, 6, 1), ts(2026, 11, 6, 2), 6.0)
    assert len(sp.check_limits()) == 1
    w.t += 600; assert sp.check_limits() == []
    w.t += 3600; assert len(sp.check_limits()) == 1
    assert len(outbox(tmp_path)) == 2

def test_higher_threshold_fires_despite_cooldown_of_lower(tmp_path):
    cfg, sp, w = mk(tmp_path, ts(2026, 11, 6, 12), spend_daily_limit_usd=10.0)
    iv(cfg, ts(2026, 11, 6, 1), ts(2026, 11, 6, 2), 6.0)
    sp.check_limits()
    iv(cfg, ts(2026, 11, 6, 3), ts(2026, 11, 6, 4), 3.0)     # now 9.0 -> 80 %
    w.t += 60
    f = sp.check_limits(); assert f == [{"period": "daily", "threshold": 0.8}]

def test_monthly_limit_alarm_goes_to_outbox_not_sent(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 20), spend_monthly_limit_usd=100.0)
    iv(cfg, ts(2026, 11, 2, 0), ts(2026, 11, 2, 1), 85.0)
    assert sp.check_limits() == [{"period": "monthly", "threshold": 0.8}]
    m = outbox(tmp_path)[0]
    assert (m["from"], m["to"], m["sent"], m["send_enabled"]) == ("support@kanus.org", "demirtumen08@gmail.com", False, False)
    assert "spend_limit" in m["subject"] and "80%" in m["body"]

def test_new_day_resets_daily_alarm_key(tmp_path):
    cfg, sp, w = mk(tmp_path, ts(2026, 11, 6, 23), spend_daily_limit_usd=10.0)
    iv(cfg, ts(2026, 11, 6, 1), ts(2026, 11, 6, 2), 6.0)
    iv(cfg, ts(2026, 11, 7, 0, 10), ts(2026, 11, 7, 1, 10), 6.0)
    assert len(sp.check_limits()) == 1
    w.t = ts(2026, 11, 7, 2); assert len(sp.check_limits()) == 1    # different period id

@pytest.mark.parametrize("bad", [dict(spend_daily_limit_usd=0), dict(spend_monthly_limit_usd=-5), dict(spend_thresholds=(0,))])
def test_limit_config_validation(bad):
    with pytest.raises(ValueError): Config(**bad)

def test_watchdog_health_includes_spend_summary(tmp_path):
    from tests.test_watchdog import make
    wd, g, clk = make(tmp_path, status="RUNNING", price_per_hour_usd=2.0, price_source="t")
    wd.spend = SpendTracker(wd.cfg, wd.alarm, wall_clock=lambda: ts(2026, 11, 6, 9), open_span=lambda: wd.cost.open_span)
    h = wd.health()
    assert h["spend"]["limits"]["daily"]["status"] == NOT_DEFINED and "month_end_projection_usd" in h["spend"]
    assert g.starts == 0

def test_example_config_demir_limits_and_ssd_rates():
    c = Config.from_file(os.path.join(os.path.dirname(__file__), "..", "config.example.json"))
    assert c.spend_daily_limit_usd == 20.0 and c.spend_monthly_limit_usd == 200.0
    assert tuple(c.spend_thresholds) == (0.5, 0.8, 1.0)
    assert c.spend_alarm_ssd_scenario == "ondemand_local_ssd" and c.preserved_ssd_verified_rate_usd_gib_month is None
    assert c.preserved_ssd_rates_usd_gib_month == {"pd_standard": 0.044, "spot_local_ssd": 0.0528, "ondemand_local_ssd": 0.088}

def test_example_config_limits_fire_at_10_16_20_daily(tmp_path):
    c = Config.from_file(os.path.join(os.path.dirname(__file__), "..", "config.example.json"))
    for spent, exp in ((9.99, None), (10.0, 0.5), (16.0, 0.8), (20.0, 1.0)):
        (tmp_path / str(spent)).mkdir()
        cfg, sp, _ = mk(tmp_path / str(spent), ts(2026, 11, 6, 12), spend_daily_limit_usd=c.spend_daily_limit_usd,
                        spend_monthly_limit_usd=c.spend_monthly_limit_usd, spend_thresholds=c.spend_thresholds)
        iv(cfg, ts(2026, 11, 6, 1), ts(2026, 11, 6, 2), spent)
        daily = [f for f in sp.check_limits() if f["period"] == "daily"]
        assert (daily[0]["threshold"] if daily else None) == exp

def test_example_config_monthly_200_thresholds(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 20), spend_monthly_limit_usd=200.0, spend_daily_limit_usd=20.0)
    iv(cfg, ts(2026, 11, 2, 0), ts(2026, 11, 2, 1), 160.0)
    assert {"period": "monthly", "threshold": 0.8} in sp.check_limits()


@pytest.mark.parametrize("preserve", [False, True])
def test_preserve_switch_keeps_running_spot_price(tmp_path, preserve):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3), preserve_local_ssd=preserve,
                    price_per_hour_usd=1.848238)
    iv(cfg, ts(2026, 11, 1), ts(2026, 11, 1, 1), 1.848238)
    s = sp.summary()
    assert s["price_per_hour_usd"] == 1.848238
    assert s["month_to_date_gpu_usd"] == 1.848238
    assert s["preserved_ssd"]["enabled"] is preserve
    if preserve:
        assert len(s["preserved_ssd"]["scenarios"]) == 3
        assert s["month_to_date_usd"] > 1.848238
    else:
        assert s["preserved_ssd"]["scenarios"] == {}
        assert s["preserved_ssd"]["note"] == "SSD korunmuyor, kapalıyken 0"
        assert s["alarm_ssd_scenario"] is None
        assert s["month_to_date_usd"] == 1.848238
        assert all(d["ssd_usd_by_scenario"] == {} for d in s["days"])


def test_discard_hides_even_verified_preservation_rate(tmp_path):
    cfg, sp, _ = mk(tmp_path, ts(2026, 11, 3), preserve_local_ssd=False,
                    preserved_ssd_verified_rate_usd_gib_month=0.05)
    iv(cfg, ts(2026, 11, 1), ts(2026, 11, 1, 1), 2.0)
    s = sp.summary()
    assert s["preserved_ssd"]["scenarios"] == {}
    assert s["preserved_ssd"]["stopped_hours_month"] == 0
    assert s["month_to_date_usd"] == 2.0


def test_preserve_defaults_false_in_config_and_example():
    assert Config().preserve_local_ssd is False
    c = Config.from_file(os.path.join(os.path.dirname(__file__), "..", "config.example.json"))
    assert c.preserve_local_ssd is False
    assert c.price_per_hour_usd == 1.848238
