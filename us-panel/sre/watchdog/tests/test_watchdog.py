import json, threading, time, sys, os
import pytest
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
from watchdog import Watchdog, Config, WatchdogError, ClientWaitTimeout


class Clock:
    def __init__(self): self.t = 1000.0; self.lock = threading.Lock()
    def __call__(self):
        with self.lock: return self.t
    def advance(self, s):
        with self.lock: self.t += s


class FakeGCE:
    """Fake instance: start -> STAGING -> RUNNING after `boot_polls` status reads."""
    def __init__(self, status="TERMINATED", boot_polls=2):
        self.status, self.boot_polls = status, boot_polls
        self.starts = self.stops = self.gets = 0
        self.start_gate = None; self.start_error = None; self.get_error = None
        self._pending = 0; self.lock = threading.Lock()
    def get_status(self):
        with self.lock:
            self.gets += 1
            if self.get_error: raise self.get_error
            if self.status == "STAGING":
                self._pending -= 1
                if self._pending <= 0: self.status = "RUNNING"
            elif self.status == "STOPPING":
                self.status = "TERMINATED"
            return self.status if self.status != "TERMINATED_AFTER" else "TERMINATED"
    def start(self):
        if self.start_gate: self.start_gate.wait(5)
        with self.lock:
            self.starts += 1
            if self.start_error: raise self.start_error
            if self.status == "TERMINATED":
                self.status = "STAGING"; self._pending = self.boot_polls
    def stop(self):
        with self.lock:
            self.stops += 1
            if self.status == "RUNNING": self.status = "STOPPING"


def make(tmp_path, status="TERMINATED", ready=lambda: True, **cfg):
    clk = Clock(); g = FakeGCE(status)
    cfg.setdefault("mode", "active")
    c = Config(cost_log_path=str(tmp_path / "cost.jsonl"), alarm_outbox_path=str(tmp_path / "outbox.jsonl"),
               poll_seconds=1, retry_backoff_seconds=1, **cfg)
    wd = Watchdog(g, ready, c, clock=clk, sleep=lambda s: (clk.advance(s), time.sleep(0.0005)))
    return wd, g, clk


# ---- §4.3 single-flight ----
def test_parallel_requests_share_single_start(tmp_path):
    wd, g, _ = make(tmp_path)
    g.start_gate = threading.Event()
    errs = []
    def worker():
        try:
            with wd.request(): pass
        except Exception as e: errs.append(e)
    ts = [threading.Thread(target=worker) for _ in range(20)]
    for t in ts: t.start()
    time.sleep(0.2); g.start_gate.set()
    for t in ts: t.join(5)
    assert errs == [] and g.starts == 1 and g.status == "RUNNING"

def test_start_error_propagates_to_all_waiters_then_retry_possible(tmp_path):
    wd, g, _ = make(tmp_path, start_attempts=1)
    g.start_gate = threading.Event(); g.start_error = RuntimeError("quota")
    errs = []
    def worker():
        try: wd.ensure_running()
        except WatchdogError as e: errs.append(str(e))
    ts = [threading.Thread(target=worker) for _ in range(8)]
    for t in ts: t.start()
    time.sleep(0.2); g.start_gate.set()
    for t in ts: t.join(5)
    assert len(errs) == 8 and g.starts == 1
    g.start_error = None
    wd.ensure_running()
    assert g.starts == 2 and g.status == "RUNNING"

# ---- §4.10 controlled retry ----
def test_start_retries_with_backoff_then_succeeds(tmp_path):
    wd, g, _ = make(tmp_path, start_attempts=3)
    calls = {"n": 0}; orig = g.start
    def flaky():
        calls["n"] += 1
        if calls["n"] < 3: raise RuntimeError("ZONE_RESOURCE_POOL_EXHAUSTED")
        orig()
    g.start = flaky
    wd.ensure_running()
    assert calls["n"] == 3 and g.status == "RUNNING"

def test_start_gives_up_after_max_attempts(tmp_path):
    wd, g, _ = make(tmp_path, start_attempts=2)
    g.start_error = RuntimeError("x")
    with pytest.raises(WatchdogError, match="after 2 attempts"): wd.ensure_running()

def test_readiness_timeout(tmp_path):
    wd, g, _ = make(tmp_path, ready=lambda: False, boot_seconds=60, start_attempts=1)
    with pytest.raises(WatchdogError, match="readiness timeout"): wd.ensure_running()

def test_running_vm_no_start_call(tmp_path):
    wd, g, _ = make(tmp_path, status="RUNNING")
    wd.ensure_running(); assert g.starts == 0

def test_stopping_vm_waits_then_starts(tmp_path):
    wd, g, _ = make(tmp_path, status="STOPPING")
    wd.ensure_running(); assert g.starts == 1 and g.status == "RUNNING"

# ---- §4.5 / §4.6 active / queued ----
def test_no_stop_while_request_in_flight(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    wd.begin_request(); clk.advance(5000)
    assert wd.tick() == "skip:active" and g.stops == 0
    wd.end_request()

def test_no_stop_while_queue_nonempty(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    wd.enqueue(); clk.advance(5000)
    assert wd.tick() == "skip:queued" and g.stops == 0

def test_race_request_arriving_during_stop_restarts_gpu(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    clk.advance(601)
    entered = threading.Event(); release = threading.Event(); orig = g.stop
    def slow_stop():
        orig(); entered.set(); release.wait(5)
    g.stop = slow_stop
    res = {}
    t = threading.Thread(target=lambda: res.setdefault("tick", wd.tick())); t.start()
    entered.wait(5)
    r = threading.Thread(target=lambda: wd.ensure_running("late")); r.start()
    time.sleep(0.2)
    assert g.starts == 0           # late request did not start on a stopping VM
    release.set(); t.join(5); r.join(5)
    assert res["tick"] == "stopped" and g.starts == 1 and g.status == "RUNNING"

def test_decision_is_atomic_with_new_request(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    clk.advance(601)
    wd.begin_request()             # arrives before tick takes the lock
    assert wd.tick() == "skip:active" and g.stops == 0

# ---- idle 10 min, only real requests reset ----
def test_default_idle_is_10_minutes():
    assert Config().idle_seconds == 600

def test_idle_boundary_599_vs_600(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    clk.advance(599); assert wd.tick() == "skip:not_idle"
    clk.advance(1); assert wd.tick() == "stopped" and g.stops == 1

def test_health_does_not_reset_idle(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    for _ in range(10):
        clk.advance(90); wd.health()
    assert wd.tick() == "stopped"

def test_real_request_resets_idle(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    clk.advance(500)
    with wd.request(): pass
    clk.advance(500); assert wd.tick() == "skip:not_idle"

# ---- §4.15 health never starts ----
@pytest.mark.parametrize("st", ["TERMINATED", "STOPPING", "RUNNING", "STAGING"])
def test_health_reports_live_status_without_start(tmp_path, st):
    wd, g, _ = make(tmp_path, status=st)
    h = wd.health()
    assert h["status"] in (st, "RUNNING", "TERMINATED") and g.starts == 0 and g.gets == 1

def test_health_api_error_returns_unknown_not_start(tmp_path):
    wd, g, _ = make(tmp_path)
    g.get_error = RuntimeError("403")
    h = wd.health(); assert h["status"] == "UNKNOWN" and g.starts == 0

# ---- §4.11 live state, fail-safe ----
def test_tick_reads_live_status_every_time(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    wd.tick(); wd.tick(); assert g.gets >= 2

def test_tick_status_error_never_stops(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    clk.advance(5000); g.get_error = RuntimeError("503")
    assert wd.tick().startswith("skip:status_error") and g.stops == 0

def test_externally_stopped_vm_reconciles_cost(tmp_path):   # spot preemption / manual stop
    wd, g, clk = make(tmp_path, status="RUNNING")
    wd.tick(); clk.advance(60); g.status = "TERMINATED"
    assert wd.tick() == "reconciled:closed" and g.stops == 0

def test_unknown_status_rejected(tmp_path):
    wd, g, _ = make(tmp_path, status="WEIRD", start_attempts=1)
    with pytest.raises(WatchdogError): wd.ensure_running()

def test_stop_idempotent_when_already_terminated(tmp_path):
    wd, g, clk = make(tmp_path, status="TERMINATED")
    clk.advance(5000); assert wd.tick() == "noop:not_running" and g.stops == 0

def test_stop_not_confirmed_reported(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", stop_wait_seconds=3)
    clk.advance(601); g.stop = lambda: None
    assert wd.tick().startswith("stop_issued:not_confirmed")

def test_tick_skips_while_starting(tmp_path):
    wd, g, clk = make(tmp_path)
    g.start_gate = threading.Event()
    def run():
        try: wd.ensure_running()
        except ClientWaitTimeout: pass   # fake clock jump > client_wait; boot continues
    t = threading.Thread(target=run); t.start(); time.sleep(0.1)
    clk.advance(5000); assert wd.tick() == "skip:starting"
    g.start_gate.set(); t.join(5)

# ---- §4.14 cost ledger ----
def test_cost_ledger_records_interval_price_null(tmp_path):
    wd, g, clk = make(tmp_path)
    with wd.request("api"): pass
    clk.advance(600); assert wd.tick() == "stopped"
    recs = [json.loads(l) for l in open(tmp_path / "cost.jsonl")]
    iv = [r for r in recs if r["event"] == "interval"][0]
    assert iv["trigger"] == "api" and iv["stop_reason"] == "idle"
    assert iv["duration_seconds"] > 0 and iv["price_per_hour_usd"] is None and iv["cost_usd"] is None

def test_cost_ledger_uses_config_price_and_appends(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", price_per_hour_usd=2.0, price_source="test")
    wd.tick(); clk.advance(1800); g.status = "TERMINATED"; wd.tick()
    wd.tick()
    g.status = "RUNNING"; wd.enqueue(); wd.tick(); wd.dequeue(); clk.advance(3600); g.status = "TERMINATED"; wd.tick()
    ivs = [json.loads(l) for l in open(tmp_path / "cost.jsonl") if '"interval"' in l]
    assert [i["cost_usd"] for i in ivs] == [1.0, 2.0] and ivs[0]["trigger"] == "external"

def test_config_file_rejects_unknown_keys(tmp_path):
    p = tmp_path / "c.json"; p.write_text('{"idle_seconds": 600, "bogus": 1}')
    with pytest.raises(ValueError): Config.from_file(str(p))
    p.write_text(open(os.path.join(os.path.dirname(__file__), "..", "config.example.json")).read())
    assert Config.from_file(str(p)).price_per_hour_usd == 1.848238


# ---- config: no hard-coded durations ----
def test_defaults_idle_600_boot_3000_client_2700():
    c = Config(); assert (c.idle_seconds, c.boot_seconds, c.client_wait_seconds) == (600, 3000, 2700)
    assert c.mode == "observe"

def test_idle_and_boot_come_from_config(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", idle_seconds=120)
    clk.advance(120); assert wd.tick() == "stopped"
    wd, g, clk = make(tmp_path, ready=lambda: False, boot_seconds=600, start_attempts=1, client_wait_seconds=2700)
    t0 = clk()
    with pytest.raises(WatchdogError, match="readiness timeout"): wd.ensure_running()
    assert 600 <= clk() - t0 < 700

@pytest.mark.parametrize("bad", [dict(idle_seconds=0), dict(boot_seconds=-1),
    dict(client_wait_seconds=3000, upstream_timeout_seconds=2700), dict(sse_keepalive_seconds=2700),
    dict(mode="yolo"), dict(alarm_threshold=0)])
def test_config_validation(bad):
    with pytest.raises(ValueError): Config(**bad)

# ---- user wait screen / SSE 2700 s ----
def test_client_wait_timeout_2700_boot_continues(tmp_path):
    wd, g, clk = make(tmp_path, ready=lambda: clk() - t0 >= 2900)   # model ready at +2900 s
    t0 = clk()
    with pytest.raises(ClientWaitTimeout): wd.ensure_running()
    assert 2700 <= clk() - t0
    for _ in range(500):                       # background boot finishes within boot_seconds=3000
        if wd._flight is None: break
        time.sleep(0.01)
    assert wd._flight is None and g.starts == 1
    wd.ensure_running(); assert g.starts == 1  # next request: VM ready, no new start

def test_sse_keepalive_callback_with_health(tmp_path):
    pings = []
    wd, g, clk = make(tmp_path, ready=lambda: clk() - t0 >= 100, sse_keepalive_seconds=15)
    t0 = clk()
    wd.ensure_running(on_wait=pings.append)
    assert len(pings) >= 2 and all("status" in p and "health" in p for p in pings)

# ---- alarms ----
def test_reconcile_error_logs_structured_error(tmp_path, caplog):
    wd, g, clk = make(tmp_path, status="RUNNING"); g.get_error = RuntimeError("503")
    with caplog.at_level("ERROR", logger="ds_watchdog"): wd.tick()
    rec = json.loads(caplog.records[-1].getMessage())
    assert rec["level"] == "ERROR" and rec["where"] == "reconcile.get_status" and rec["consecutive"] == 1

def test_outbox_written_once_at_threshold_no_send(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", alarm_threshold=3, alarm_email_to="ops@example.invalid")
    g.get_error = RuntimeError("503")
    for _ in range(2): wd.tick()
    assert not (tmp_path / "outbox.jsonl").exists()
    for _ in range(3): wd.tick()
    mails = [json.loads(l) for l in open(tmp_path / "outbox.jsonl")]
    assert len(mails) == 1 and mails[0]["sent"] is False and "degraded" in mails[0]["subject"]

def test_health_degraded_after_n_and_recovers(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", alarm_threshold=2)
    g.get_error = RuntimeError("503"); wd.tick()
    g.get_error = None; assert wd.health()["health"] == "ok"
    g.get_error = RuntimeError("503"); wd.tick(); wd.tick()
    g.get_error = None
    h = wd.health(); assert h["health"] == "degraded" and h["consecutive_errors"] >= 2
    wd.tick(); assert wd.health()["health"] == "ok"

def test_start_failure_counts_as_alarm(tmp_path):
    wd, g, _ = make(tmp_path, start_attempts=1, alarm_threshold=1); g.start_error = RuntimeError("quota")
    with pytest.raises(WatchdogError): wd.ensure_running()
    assert wd.health()["degraded"] is True

# ---- ownership / handover ----
def test_observe_mode_never_starts_or_stops(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", mode="observe")
    clk.advance(5000); assert wd.tick() == "observe:would_stop" and g.stops == 0
    g.status = "TERMINATED"
    with pytest.raises(WatchdogError, match="observe"): wd.ensure_running()
    assert g.starts == 0

def test_refuses_dual_control_while_legacy_idle_active(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING"); wd.legacy_active = lambda: True
    clk.advance(5000); assert wd.tick() == "skip:legacy_controller_active" and g.stops == 0

def test_legacy_check_error_is_treated_as_active(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    def boom(): raise OSError("systemctl")
    wd.legacy_active = boom; clk.advance(5000)
    assert wd.tick() == "skip:legacy_controller_active" and g.stops == 0

def test_adopt_running_grants_full_idle_window(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING")
    clk.advance(5000); assert wd.adopt_running() == "adopted:running"
    assert wd.tick() == "skip:not_idle"
    clk.advance(600); assert wd.tick() == "stopped"
    iv = [json.loads(l) for l in open(tmp_path / "cost.jsonl") if '"interval"' in l][0]
    assert iv["trigger"] == "adopted"


def test_example_config_alarm_addresses_and_send_disabled(tmp_path):
    c = Config.from_file(os.path.join(os.path.dirname(__file__), "..", "config.example.json"))
    assert c.alarm_email_from == "support@kanus.org" and c.alarm_email_to == "demirtumen08@gmail.com"
    assert c.alarm_send_enabled is False and c.price_per_hour_usd == 1.848238

def test_send_flag_cannot_be_enabled_in_core():
    with pytest.raises(ValueError, match="alarm_send_enabled"): Config(alarm_send_enabled=True)

def test_outbox_record_has_from_to_and_not_sent(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", alarm_threshold=1,
                      alarm_email_from="support@kanus.org", alarm_email_to="demirtumen08@gmail.com")
    g.get_error = RuntimeError("503"); wd.tick()
    m = json.loads(open(tmp_path / "outbox.jsonl").readline())
    assert (m["from"], m["to"], m["sent"], m["send_enabled"]) == ("support@kanus.org", "demirtumen08@gmail.com", False, False)


def test_example_price_breakdown_from_catalog_skus():
    c = Config.from_file(os.path.join(os.path.dirname(__file__), "..", "config.example.json"))
    comps = {x["skuId"]: x for x in c.price_breakdown["components"]}
    assert set(comps) == {"E540-EA28-2318", "D847-3FF3-B717", "9BBC-6855-1C9A", "8D0B-EA58-9E35"}
    assert c.price_breakdown["model"] == "SPOT" and c.price_breakdown["hours_per_month"] == 730
    assert abs(comps["8D0B-EA58-9E35"]["hourly_usd"] - 1500 * 0.0528 / 730) < 1e-6
    assert c.price_source.startswith("Cloud Billing Catalog") and c.price_retrieved_at.startswith("2026-11-06")

def test_price_requires_source_and_consistent_breakdown():
    with pytest.raises(ValueError): Config(price_per_hour_usd=1.0)
    with pytest.raises(ValueError):
        Config(price_per_hour_usd=1.0, price_source="x", price_breakdown={"components": [{"hourly_usd": 0.5}]})
    Config(price_per_hour_usd=1.0, price_source="x", price_breakdown={"components": [{"hourly_usd": 1.0}]})

def test_ledger_records_price_source(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", price_per_hour_usd=1.848238, price_source="catalog")
    wd.tick(); clk.advance(3600); g.status = "TERMINATED"; wd.tick()
    iv = [json.loads(l) for l in open(tmp_path / "cost.jsonl") if '"interval"' in l][0]
    assert iv["price_source"] == "catalog" and iv["cost_usd"] == 1.848238


def test_idle_599_no_alarm_601_plus_stop_wait_alarm(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", stop_wait_seconds=3, alarm_threshold=1)
    g.stop = lambda: None  # fake stop never reaches TERMINATED
    clk.advance(599)
    assert wd.tick() == "skip:not_idle"
    assert not wd.alarm.degraded
    assert not os.path.exists(wd.cfg.alarm_outbox_path)
    clk.advance(2)
    before = clk()
    assert wd.tick().startswith("stop_issued:not_confirmed")
    assert clk() >= before + wd.cfg.stop_wait_seconds
    assert wd.alarm.degraded
    assert wd.alarm.last_error
    assert os.path.exists(wd.cfg.alarm_outbox_path)


def test_idle_600_cost_uses_elapsed_time(tmp_path):
    wd, g, clk = make(tmp_path, status="RUNNING", price_per_hour_usd=1.848238, price_source="test")
    wd.tick()
    clk.advance(599)
    assert wd.tick() == "skip:not_idle"
    clk.advance(1)
    assert wd.tick() == "stopped"
    intervals = [json.loads(line) for line in open(wd.cfg.cost_log_path) if '"interval"' in line]
    assert len(intervals) == 1
    assert intervals[0]["duration_seconds"] == 600
    assert intervals[0]["cost_usd"] == round(1.848238 * 600 / 3600, 6)
