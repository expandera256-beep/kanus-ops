"""SRE watchdog for ds-a100 (on-demand GPU). Core logic only; no network I/O here.

Provider interface (real impl is NOT in this module; tests use a fake):
  get_status() -> str      live GCE instance status (instances.get), no cache
  start() -> None          instances.start (async op is fine)
  stop() -> None           instances.stop (discardLocalSsd decided by provider)
Readiness probe: callable() -> bool  (e.g. GET /v1/models on the VM via tunnel)
"""
from __future__ import annotations
import json, logging, os, threading, time
from dataclasses import dataclass, field
from typing import Callable, Optional

SSD_SCENARIOS = ("pd_standard", "spot_local_ssd", "ondemand_local_ssd")
RUNNING, TERMINATED, STOPPING = "RUNNING", "TERMINATED", "STOPPING"
STARTING_STATES = {"PROVISIONING", "STAGING"}
STARTABLE = {TERMINATED, "STOPPED"}
KNOWN = {
RUNNING, TERMINATED, STOPPING, "STOPPED", "PROVISIONING", "STAGING",
         "SUSPENDING", "SUSPENDED", "REPAIRING"}


LOG = logging.getLogger("ds_watchdog")


class WatchdogError(RuntimeError):
    pass


class ClientWaitTimeout(WatchdogError):
    """User-facing wait budget expired; boot continues in background (not an error of the VM)."""


@dataclass
class Config:
    idle_seconds: float = 600.0          # Demir karari: 10 dk, policy v5
    boot_seconds: float = 3000.0         # gecici; hedef 600 (§4.9), local-SSD olcumu sonrasi dusurulecek
    client_wait_seconds: float = 2700.0  # kullanici bekleme ekrani / SSE; US_MODEL_TIMEOUT_SECONDS ile ayni
    upstream_timeout_seconds: float = 2700.0  # Go hook US_MODEL_TIMEOUT_SECONDS
    sse_keepalive_seconds: float = 15.0
    alarm_threshold: int = 3             # ardisik N hata -> degraded
    alarm_outbox_path: str = "alarm_outbox.jsonl"  # e-posta kuyrugu (dosya); gonderim YOK
    alarm_email_to: Optional[str] = None
    alarm_email_from: Optional[str] = None
    alarm_send_enabled: bool = False     # KAPALI: bu modul asla gondermez; gonderim entegrasyonda acilacak
    billing_project: str = "project-27082a40-6083-4952-9db"
    billing_dataset: str = "billing_export"
    billing_resource_table: Optional[str] = None   # gcp_billing_export_resource_v1_<BILLING_ACCOUNT>; export olusunca
    billing_resource_names: tuple = ("ds-a100",)
    billing_maximum_bytes_billed: int = 1_000_000_000
    billing_lookback_days: int = 3
    billing_export_lag_hours: float = 48.0
    billing_diff_abs_usd: float = 0.50
    billing_diff_pct: float = 0.20
    spend_daily_limit_usd: Optional[float] = None    # None: limit tanimli degil, alarm yok
    spend_monthly_limit_usd: Optional[float] = None
    spend_thresholds: tuple = (0.5, 0.8, 1.0)
    spend_alarm_cooldown_seconds: float = 21600.0
    preserve_local_ssd: bool = False           # discardLocalSsd=false ile mi durduruluyor
    preserved_ssd_gib: float = 1500.0
    # Uc senaryo birlikte raporlanir; hangisinin uygulandigi NOT VERIFIED.
    preserved_ssd_rates_usd_gib_month: Optional[dict] = None  # {"pd_standard":..,"spot_local_ssd":..,"ondemand_local_ssd":..}
    spend_alarm_ssd_scenario: str = "ondemand_local_ssd"      # limit alarmi bunu kullanir (en yuksek = muhafazakar)
    preserved_ssd_verified_rate_usd_gib_month: Optional[float] = None  # export'tan kesin SKU gelince; senaryolarin yerine gecer
    preserved_ssd_verified_sku: Optional[str] = None
    hours_per_month: float = 730.0
    mode: str = "observe"                # observe: karar loglar, start/stop YAPMAZ; active: yonetir
    poll_seconds: float = 5.0
    start_attempts: int = 3
    retry_backoff_seconds: float = 30.0
    stop_wait_seconds: float = 300.0
    price_per_hour_usd: Optional[float] = None  # config'den; bilinmiyorsa None
    price_source: Optional[str] = None
    price_retrieved_at: Optional[str] = None
    price_breakdown: Optional[dict] = None
    cost_log_path: str = "gpu_cost.jsonl"
    instance: str = "ds-a100"

    def __post_init__(self):
        for k in ("idle_seconds", "boot_seconds", "client_wait_seconds", "upstream_timeout_seconds",
                  "sse_keepalive_seconds", "poll_seconds", "stop_wait_seconds"):
            if not getattr(self, k) > 0:
                raise ValueError(f"{k} must be > 0")
        if self.start_attempts < 1 or self.alarm_threshold < 1:
            raise ValueError("start_attempts/alarm_threshold must be >= 1")
        if self.client_wait_seconds > self.upstream_timeout_seconds:
            raise ValueError("client_wait_seconds must be <= upstream_timeout_seconds (SSE/Go timeout)")
        if self.sse_keepalive_seconds >= self.client_wait_seconds:
            raise ValueError("sse_keepalive_seconds must be < client_wait_seconds")
        if self.price_per_hour_usd is not None:
            if self.price_per_hour_usd < 0 or not self.price_source:
                raise ValueError("price_per_hour_usd needs a non-negative value and price_source")
            if self.price_breakdown:
                total = sum(float(c["hourly_usd"]) for c in self.price_breakdown.get("components", []))
                if abs(total - self.price_per_hour_usd) > 1e-4:
                    raise ValueError(f"price_breakdown sums to {total}, not price_per_hour_usd")
        if self.alarm_send_enabled:
            raise ValueError("alarm_send_enabled must stay false in this build (outbox only; sending is an integration step)")
        for k in ("spend_daily_limit_usd", "spend_monthly_limit_usd"):
            v = getattr(self, k)
            if v is not None and not v > 0:
                raise ValueError(f"{k} must be > 0 or null")
        if any(not 0 < t <= 5 for t in self.spend_thresholds):
            raise ValueError("spend_thresholds must be fractions in (0, 5]")
        if self.preserve_local_ssd and self.preserved_ssd_verified_rate_usd_gib_month is None:
            rates = self.preserved_ssd_rates_usd_gib_month or {}
            missing = [k for k in SSD_SCENARIOS if not isinstance(rates.get(k), (int, float))]
            if missing:
                raise ValueError(f"preserved_ssd_rates_usd_gib_month missing {missing}")
        if self.spend_alarm_ssd_scenario not in SSD_SCENARIOS:
            raise ValueError(f"spend_alarm_ssd_scenario must be one of {SSD_SCENARIOS}")
        if self.mode not in ("observe", "active"):
            raise ValueError("mode must be observe|active")

    @classmethod
    def from_file(cls, path: str) -> "Config":
        with open(path, encoding="utf-8") as fh:
            doc = json.load(fh)
        doc = {k: v for k, v in doc.items() if not k.endswith(("_note", "_source", "_decision")) or k in cls.__dataclass_fields__}
        unknown = set(doc) - set(cls.__dataclass_fields__)
        if unknown:
            raise ValueError(f"unknown config keys: {sorted(unknown)}")
        return cls(**doc)


class CostLedger:
    """Append-only JSONL. One 'interval' record per RUNNING span, written when the span closes."""

    def __init__(self, path: str, price_per_hour: Optional[float], clock: Callable[[], float],
                 source: Optional[str] = None, wall_clock: Callable[[], float] = time.time):
        self.path, self.price, self.clock, self.source = path, price_per_hour, clock, source
        self.wall = wall_clock
        self.open_span: Optional[dict] = None
        self._lock = threading.Lock()

    def open(self, trigger: str, reason: str) -> None:
        with self._lock:
            if self.open_span is None:
                self.open_span = {"start": self.clock(), "start_wall": self.wall(),
                                  "trigger": trigger, "start_reason": reason}
                self._append({"event": "open", "ts": self.open_span["start"], "start_wall": self.open_span["start_wall"],
                              "trigger": trigger, "reason": reason})

    def close(self, reason: str) -> Optional[dict]:
        with self._lock:
            if self.open_span is None:
                return None
            end = self.clock()
            dur = max(0.0, end - self.open_span["start"])
            rec = {"event": "interval", "start": self.open_span["start"], "end": end,
                   "start_wall": self.open_span["start_wall"], "end_wall": self.wall(),
                   "duration_seconds": dur, "trigger": self.open_span["trigger"],
                   "start_reason": self.open_span["start_reason"], "stop_reason": reason,
                   "price_per_hour_usd": self.price, "price_source": self.source,
                   "cost_usd": None if self.price is None else round(dur / 3600.0 * self.price, 6)}
            self._append(rec)
            self.open_span = None
            return rec

    def _append(self, rec: dict) -> None:
        with open(self.path, "a", encoding="utf-8") as fh:  # append-only
            fh.write(json.dumps(rec, ensure_ascii=False) + "\n")
            fh.flush()
            os.fsync(fh.fileno())


class Alarm:
    """Consecutive-failure alarm: ERROR log + outbox record (no send) + degraded after N."""

    def __init__(self, cfg: "Config", clock):
        self.cfg, self.clock = cfg, clock
        self.consecutive = 0
        self.last_error: Optional[str] = None
        self._lock = threading.Lock()

    @property
    def degraded(self) -> bool:
        return self.consecutive >= self.cfg.alarm_threshold

    def failure(self, where: str, err: str) -> None:
        with self._lock:
            self.consecutive += 1
            self.last_error = f"{where}: {err}"
            rec = {"ts": self.clock(), "level": "ERROR", "instance": self.cfg.instance,
                   "where": where, "error": err, "consecutive": self.consecutive,
                   "degraded": self.degraded}
            LOG.error(json.dumps(rec, ensure_ascii=False))
            if self.consecutive == self.cfg.alarm_threshold:   # one mail per incident
                mail = {"queued_at": self.clock(), "from": self.cfg.alarm_email_from,
                        "to": self.cfg.alarm_email_to, "sent": False,
                        "send_enabled": self.cfg.alarm_send_enabled,
                        "subject": f"[ds-watchdog] {self.cfg.instance} degraded ({self.consecutive} ardisik hata)",
                        "body": json.dumps(rec, ensure_ascii=False)}
                with open(self.cfg.alarm_outbox_path, "a", encoding="utf-8") as fh:
                    fh.write(json.dumps(mail, ensure_ascii=False) + "\n")

    def notify(self, where: str, err: str, detail: Optional[dict] = None) -> None:
        """Immediate alarm (e.g. billing mismatch): ERROR log + one outbox record per call.
        Does not change the consecutive/degraded counter."""
        rec = {"ts": self.clock(), "level": "ERROR", "instance": self.cfg.instance,
               "where": where, "error": err, "detail": detail or {}}
        LOG.error(json.dumps(rec, ensure_ascii=False))
        mail = {"queued_at": self.clock(), "from": self.cfg.alarm_email_from, "to": self.cfg.alarm_email_to,
                "sent": False, "send_enabled": self.cfg.alarm_send_enabled,
                "subject": f"[ds-watchdog] {self.cfg.instance} {where}", "body": json.dumps(rec, ensure_ascii=False)}
        with open(self.cfg.alarm_outbox_path, "a", encoding="utf-8") as fh:
            fh.write(json.dumps(mail, ensure_ascii=False) + "\n")

    def success(self) -> None:
        with self._lock:
            if self.consecutive:
                LOG.info(json.dumps({"level": "INFO", "event": "recovered", "after": self.consecutive}))
            self.consecutive, self.last_error = 0, None


class _Flight:
    def __init__(self):
        self.done = threading.Event()
        self.error: Optional[BaseException] = None


class Watchdog:
    def __init__(self, provider, ready_probe: Callable[[], bool], config: Config,
                 clock: Callable[[], float] = time.monotonic,
                 sleep: Callable[[float], None] = time.sleep,
                 log: Callable[[str], None] = lambda m: None,
                 wall_clock: Callable[[], float] = time.time):
        self.p, self.ready, self.cfg = provider, ready_probe, config
        self.clock, self.sleep, self.log = clock, sleep, log
        self.cost = CostLedger(config.cost_log_path, config.price_per_hour_usd, clock,
                               source=config.price_source, wall_clock=wall_clock)
        self.alarm = Alarm(config, clock)
        self.legacy_active: Callable[[], bool] = lambda: False
        self.spend = None   # optional spend.SpendTracker; health() includes its summary  # inject: is idle.py/ds-idle.timer on?
        self._lock = threading.Lock()
        self._cv = threading.Condition(self._lock)
        self._flight: Optional[_Flight] = None
        self._stopping = False
        self.in_flight = 0
        self.queue_depth = 0
        self.last_activity = clock()
        self.start_calls = 0
        self.stop_calls = 0

    # ---------- real-request accounting (only these reset the idle timer) ----------
    def enqueue(self) -> None:
        with self._lock:
            self.queue_depth += 1
            self.last_activity = self.clock()

    def dequeue(self) -> None:
        with self._lock:
            self.queue_depth = max(0, self.queue_depth - 1)
            self.last_activity = self.clock()

    def begin_request(self, trigger: str = "request") -> None:
        """Register a real model request, then make sure the GPU is up (shared start)."""
        with self._lock:
            self.in_flight += 1
            self.last_activity = self.clock()
        try:
            self.ensure_running(trigger)
        except BaseException:
            self.end_request()
            raise

    def end_request(self) -> None:
        with self._lock:
            self.in_flight = max(0, self.in_flight - 1)
            self.last_activity = self.clock()

    class _Req:
        def __init__(self, wd, trigger): self.wd, self.trigger = wd, trigger
        def __enter__(self): self.wd.begin_request(self.trigger); return self.wd
        def __exit__(self, *a): self.wd.end_request(); return False

    def request(self, trigger: str = "request"):
        return Watchdog._Req(self, trigger)

    # ---------- health: read-only, never starts, never touches idle ----------
    def health(self) -> dict:
        try:
            status = self.p.get_status()
        except Exception as exc:  # noqa: BLE001
            status, err = "UNKNOWN", type(exc).__name__
            self.alarm.failure("health.get_status", err)
        else:
            err = None
        with self._lock:
            doc = {"instance": self.cfg.instance, "status": status, "in_flight": self.in_flight,
                   "queue_depth": self.queue_depth, "starting": self._flight is not None,
                   "stopping": self._stopping,
                   "idle_for_seconds": round(self.clock() - self.last_activity, 3),
                   "mode": self.cfg.mode, "degraded": self.alarm.degraded,
                   "consecutive_errors": self.alarm.consecutive, "last_error": self.alarm.last_error,
                   "health": "degraded" if self.alarm.degraded else "ok"}
        if err:
            doc["error"] = err
        if self.spend is not None:
            try:
                doc["spend"] = self.spend.summary()
            except Exception as exc:  # noqa: BLE001
                doc["spend"] = {"error": type(exc).__name__}
        return doc

    # ---------- single-flight start ----------
    def ensure_running(self, trigger: str = "request", wait_seconds: Optional[float] = None,
                       on_wait: Optional[Callable[[dict], None]] = None) -> None:
        """Single-flight start. Waits at most wait_seconds (default client_wait_seconds) on the
        injected clock; calls on_wait(health) every sse_keepalive_seconds for SSE keep-alive.
        On ClientWaitTimeout the boot keeps going in the background (up to boot_seconds)."""
        if self.cfg.mode != "active":
            st = self._status()
            if st == RUNNING and self._safe_ready():
                return
            raise WatchdogError(f"observe mode: not starting (status {st})")
        budget = self.cfg.client_wait_seconds if wait_seconds is None else wait_seconds
        t0 = self.clock()
        with self._lock:
            while self._stopping:          # stop in progress: wait, then start again
                self._cv.wait()
            flight = self._flight
            if flight is None:
                flight = self._flight = _Flight()
                threading.Thread(target=self._lead, args=(flight, trigger), daemon=True).start()
        next_ping = t0 + self.cfg.sse_keepalive_seconds
        while not flight.done.wait(0.005):
            now = self.clock()
            if on_wait and now >= next_ping:
                on_wait(self.health()); next_ping = now + self.cfg.sse_keepalive_seconds
            if now - t0 >= budget:
                raise ClientWaitTimeout(f"client wait {budget:.0f}s exceeded; boot continues")
        if flight.error:
            raise flight.error

    def _lead(self, flight: "_Flight", trigger: str) -> None:
        try:
            self._start_with_retries(trigger)
            self.alarm.success()
        except BaseException as exc:  # noqa: BLE001
            flight.error = exc
            self.alarm.failure("start", str(exc))
        finally:
            with self._lock:
                self._flight = None       # next request may retry after a failure
                self._cv.notify_all()
            flight.done.set()

    def adopt_running(self) -> str:
        """Handover step: called once when this watchdog takes ownership from idle.py.
        If the VM is already RUNNING, open a cost span and grant a full idle window
        (no instant stop of a VM started by the legacy system)."""
        st = self._status()
        if st == RUNNING:
            self.cost.open("adopted", "handover")
            with self._lock:
                self.last_activity = self.clock()
            return "adopted:running"
        return f"adopted:{st}"

    def _start_with_retries(self, trigger: str) -> None:
        last: Optional[BaseException] = None
        for attempt in range(1, self.cfg.start_attempts + 1):
            try:
                self._start_once(trigger)
                return
            except WatchdogError as exc:
                last = exc
                self.log(f"start attempt {attempt} failed: {exc}")
                if attempt < self.cfg.start_attempts:
                    self.sleep(self.cfg.retry_backoff_seconds * attempt)
        raise WatchdogError(f"start failed after {self.cfg.start_attempts} attempts: {last}")

    def _start_once(self, trigger: str) -> None:
        deadline = self.clock() + self.cfg.boot_seconds
        status = self._status()
        if status == STOPPING:
            status = self._wait_until(lambda s: s != STOPPING, deadline)
        if status in STARTABLE:
            try:
                self.start_calls += 1
                self.p.start()
            except Exception as exc:  # noqa: BLE001
                raise WatchdogError(f"start api error: {type(exc).__name__}") from exc
        elif status not in STARTING_STATES and status != RUNNING:
            raise WatchdogError(f"instance not startable: {status}")
        self._wait_until(lambda s: s == RUNNING, deadline)
        self.cost.open(trigger, "start")
        while not self._safe_ready():
            if self.clock() >= deadline:
                raise WatchdogError("model readiness timeout")
            if self._status() != RUNNING:
                raise WatchdogError("instance left RUNNING during boot")
            self.sleep(self.cfg.poll_seconds)

    def _safe_ready(self) -> bool:
        try:
            return bool(self.ready())
        except Exception:  # noqa: BLE001
            return False

    def _status(self) -> str:
        try:
            s = self.p.get_status()
        except Exception as exc:  # noqa: BLE001
            raise WatchdogError(f"status api error: {type(exc).__name__}") from exc
        if s not in KNOWN:
            raise WatchdogError(f"unknown instance status: {s}")
        return s

    def _wait_until(self, pred, deadline) -> str:
        while True:
            s = self._status()
            if pred(s):
                return s
            if self.clock() >= deadline:
                raise WatchdogError(f"timeout waiting, last status {s}")
            self.sleep(self.cfg.poll_seconds)

    # ---------- idle tick (call from timer every ~30-60 s) ----------
    def tick(self) -> str:
        """Returns a decision string. Reads live status every time."""
        try:
            status = self._status()
        except WatchdogError as exc:
            self.alarm.failure("reconcile.get_status", str(exc))
            return f"skip:status_error:{exc}"      # fail-safe: never stop on unknown
        with self._lock:
            if self._flight is not None:
                return "skip:starting"
        self.alarm.success()
        if status != RUNNING:
            if status in (TERMINATED, "STOPPED"):
                closed = self.cost.close("observed_" + status.lower())
                return "reconciled:closed" if closed else "noop:not_running"
            return f"noop:{status}"
        if self.cost.open_span is None:          # running but unknown to us (manual/legacy start)
            self.cost.open("external", "observed_running")
        with self._lock:
            if self.in_flight > 0:
                return "skip:active"
            if self.queue_depth > 0:
                return "skip:queued"
            if self.clock() - self.last_activity < self.cfg.idle_seconds:
                return "skip:not_idle"
            if self.cfg.mode != "active":
                return "observe:would_stop"
            try:
                legacy = self.legacy_active()
            except Exception:  # noqa: BLE001
                legacy = True
            if legacy:
                self.alarm.failure("reconcile", "legacy controller (idle.py) still active; refusing dual control")
                return "skip:legacy_controller_active"
            self._stopping = True          # from here new requests block until stop ends
        try:
            self.stop_calls += 1
            self.p.stop()
            deadline = self.clock() + self.cfg.stop_wait_seconds
            try:
                self._wait_until(lambda s: s in (TERMINATED, "STOPPED"), deadline)
            except WatchdogError as exc:
                self.alarm.failure("stop.confirm", str(exc))
                return f"stop_issued:not_confirmed:{exc}"
            self.cost.close("idle")
            return "stopped"
        except Exception as exc:  # noqa: BLE001
            self.alarm.failure("stop", type(exc).__name__)
            return f"stop_error:{type(exc).__name__}"
        finally:
            with self._lock:
                self._stopping = False
                self._cv.notify_all()
