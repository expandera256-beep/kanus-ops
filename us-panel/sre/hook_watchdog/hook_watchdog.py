"""Offline hook alarm draft. No network, daemon, deployment or process control.
Collectors supply timestamped observations. Optional health callback is injected;
it must enforce timeout itself (this module cannot interrupt arbitrary callbacks).
"""
from __future__ import annotations
import importlib.util
import json
import logging
import math
import sys
import time
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo

# Load the existing local module without colliding with third-party 'watchdog'.
_NAME = '_sre_hook_existing_watchdog'
if _NAME not in sys.modules:
    _spec = importlib.util.spec_from_file_location(_NAME, Path(__file__).resolve().parents[1] / 'watchdog' / 'watchdog.py')
    _mod = importlib.util.module_from_spec(_spec)
    sys.modules[_NAME] = _mod
    _spec.loader.exec_module(_mod)
ExistingAlarm = sys.modules[_NAME].Alarm
ExistingConfig = sys.modules[_NAME].Config
LOG = logging.getLogger('hook_watchdog')


@dataclass
class Config:
    poll_seconds: float = 30
    health_timeout_seconds: float = 2
    consecutive_violations: int = 3
    recovery_successes: int = 3
    cooldown_seconds: float = 900
    stale_seconds: float = 90
    login_window_seconds: float = 300
    login_min_samples: int = 20
    login_warning_seconds: float = 2
    login_critical_seconds: float = 5
    load_warning_per_cpu: float = 2
    load_critical_per_cpu: float = 4
    cpu_psi_warning_percent: float = 20
    cpu_psi_critical_percent: float = 50
    memory_psi_warning_percent: float = 1
    io_psi_warning_percent: float = 10
    pressure_hold_seconds: float = 300
    alarm_outbox_path: str = 'hook_alarm_outbox.jsonl'
    alarm_email_from: str = 'support@kanus.org'
    alarm_email_to: str = 'demirtumen08@gmail.com'
    alarm_send_enabled: bool = False

    def __post_init__(self):
        for k, v in vars(self).items():
            if isinstance(v, (float, int)) and not isinstance(v, bool):
                if not math.isfinite(v) or v <= 0:
                    raise ValueError(f'{k} must be finite and > 0')
        for k in ('consecutive_violations', 'recovery_successes', 'login_min_samples'):
            if type(getattr(self, k)) is not int:
                raise ValueError(f'{k} must be an integer')
        for w, c in [('login_warning_seconds','login_critical_seconds'), ('load_warning_per_cpu','load_critical_per_cpu'), ('cpu_psi_warning_percent','cpu_psi_critical_percent')]:
            if getattr(self, w) >= getattr(self, c):
                raise ValueError(f'{w} must be below {c}')
        for k in ('cpu_psi_warning_percent','cpu_psi_critical_percent','memory_psi_warning_percent','io_psi_warning_percent'):
            if getattr(self, k) > 100:
                raise ValueError(f'{k} exceeds 100')
        if self.alarm_send_enabled is not False:
            raise ValueError('outbox only: sending must remain false')

    @classmethod
    def from_file(cls, path):
        return cls(**json.loads(Path(path).read_text()))


@dataclass(frozen=True)
class Health:
    status_code: int | None
    elapsed_seconds: float


@dataclass(frozen=True)
class Login:
    timestamp: float
    seconds: float
    failed: bool = False
    timed_out: bool = False  # seconds is an observed lower bound, retained in p95


@dataclass(frozen=True)
class Metrics:
    timestamp: float
    health: Health | None = None
    logins: tuple[Login, ...] | None = None  # None: unavailable; (): observed empty
    ncpu: int | None = None
    load1: float | None = None
    cpu_psi_some_avg60: float | None = None
    memory_psi_full_avg60: float | None = None
    io_psi_full_avg60: float | None = None


@dataclass
class State:
    bad: int = 0
    good: int = 0
    since: float | None = None
    active: str | None = None
    last_sent: float | None = None


def finite(value, minimum=0, maximum=None):
    return isinstance(value, (float, int)) and not isinstance(value, bool) and math.isfinite(value) and value >= minimum and (maximum is None or value <= maximum)


def p95(values):
    """Nearest-rank percentile; empty input is unknown, not zero."""
    if not values:
        return None
    if not all(finite(v) for v in values):
        raise ValueError('invalid login duration')
    return sorted(values)[math.ceil(.95 * len(values)) - 1]


def check_health(probe, timeout_seconds):
    """Injected probe(timeout)->Health. No real probe implementation is provided."""
    try:
        result = probe(timeout_seconds)
        if not isinstance(result, Health):
            raise TypeError('probe must return Health')
        return result
    except Exception:
        return Health(None, timeout_seconds)


class Watchdog:
    def __init__(self, cfg: Config, clock=time.time, alarm=None):
        self.cfg, self.clock = cfg, clock
        self.alarm = alarm if alarm is not None else ExistingAlarm(ExistingConfig(
            instance='us-license-hook', alarm_outbox_path=cfg.alarm_outbox_path,
            alarm_email_from=cfg.alarm_email_from, alarm_email_to=cfg.alarm_email_to,
            alarm_send_enabled=False), clock)
        self.states: dict[str, State] = {}
        self.last_timestamp = None
        self.last_evaluated = None
        self.outbox_error = None

    @property
    def degraded(self):
        return self.outbox_error is not None or any(s.active for s in self.states.values())

    def _emit(self, key, severity, now, detail):
        try:
            self.alarm.notify('hook.' + key, severity, {
                'time_tsi': datetime.fromtimestamp(now, ZoneInfo('Europe/Istanbul')).isoformat(),
                'severity': severity, **detail})
        except OSError as exc:
            self.outbox_error = type(exc).__name__
            LOG.error('hook alarm outbox write failed: %s', self.outbox_error)
            return False
        self.outbox_error = None
        return True

    def _step(self, key, severity, now, detail, hold=0):
        st = self.states.setdefault(key, State())
        if severity == 'unknown':
            st.bad = st.good = 0
            st.since = None  # gap cannot count as sustained violation or recovery
            return
        if severity == 'ok':
            st.bad, st.since = 0, None
            st.good += 1
            if st.active and st.good >= self.cfg.recovery_successes:
                if self._emit(key, 'recovered', now, detail):
                    st.active, st.last_sent = None, now
            return
        st.good = 0
        st.bad += 1
        if st.since is None:
            st.since = now
        if st.bad < self.cfg.consecutive_violations or now - st.since < hold:
            return
        escalation = st.active == 'warning' and severity == 'critical'
        due = st.last_sent is None or now - st.last_sent >= self.cfg.cooldown_seconds
        # Reminders retain the highest severity until recovery, not silent downgrades.
        target = 'critical' if st.active == 'critical' else severity
        st.active = target  # health state is independent of notification cooldown/delivery
        if escalation or due:
            if self._emit(key, target, now, detail):
                st.active, st.last_sent = target, now

    def evaluate(self, m: Metrics):
        now = self.clock()
        result = {}
        if self.last_evaluated is not None and now - self.last_evaluated < self.cfg.poll_seconds:
            return {'telemetry': 'duplicate_or_too_soon', 'degraded': self.degraded}
        self.last_evaluated = now
        if not finite(m.timestamp) or m.timestamp > now or now - m.timestamp > self.cfg.stale_seconds:
            for key in list(self.states):
                if key != 'telemetry':
                    self._step(key, 'unknown', now, {})
            self._step('telemetry', 'warning', now, {'reason': 'stale_or_invalid'})
            return {'telemetry': 'stale_or_invalid', 'degraded': self.degraded}
        if self.last_timestamp is not None and m.timestamp - self.last_timestamp < self.cfg.poll_seconds:
            return {'telemetry': 'duplicate_or_too_soon', 'degraded': self.degraded}
        if self.last_timestamp is not None and m.timestamp - self.last_timestamp > self.cfg.stale_seconds:
            for key in list(self.states):
                self._step(key, 'unknown', now, {})
        self.last_timestamp = m.timestamp
        missing = []

        def record(key, severity, detail=None, hold=0):
            result[key] = severity
            self._step(key, severity, now, detail or {}, hold)
            if severity == 'unknown':
                missing.append(key)

        h = m.health
        record('healthz', 'unknown' if h is None else (
            'ok' if h.status_code == 200 and finite(h.elapsed_seconds) and h.elapsed_seconds <= self.cfg.health_timeout_seconds else 'critical'))
        if m.logins is None:
            record('login_p95', 'unknown', {'reason': 'missing_log_data'})
        else:
            valid = all(finite(x.timestamp) and x.timestamp <= m.timestamp and finite(x.seconds) for x in m.logins)
            xs = [x for x in m.logins if m.timestamp - self.cfg.login_window_seconds < x.timestamp <= m.timestamp] if valid else []
            if len(xs) < self.cfg.login_min_samples:
                record('login_p95', 'unknown', {'reason': 'insufficient_or_invalid_samples'})
            else:
                value = p95([x.seconds for x in xs])
                detail = {'p95_seconds': value, 'samples': len(xs), 'failed': sum(x.failed for x in xs),
                          'timeouts': sum(x.timed_out for x in xs), 'window_seconds': self.cfg.login_window_seconds}
                record('login_p95', self._severity(value, self.cfg.login_warning_seconds, self.cfg.login_critical_seconds), detail)
                result['login_detail'] = detail
        ratio = m.load1 / m.ncpu if finite(m.load1) and type(m.ncpu) is int and m.ncpu > 0 else None
        pressure = [
            ('load', ratio, self.cfg.load_warning_per_cpu, self.cfg.load_critical_per_cpu, None),
            ('cpu_psi', m.cpu_psi_some_avg60, self.cfg.cpu_psi_warning_percent, self.cfg.cpu_psi_critical_percent, 100),
            ('memory_psi', m.memory_psi_full_avg60, self.cfg.memory_psi_warning_percent, None, 100),
            ('io_psi', m.io_psi_full_avg60, self.cfg.io_psi_warning_percent, None, 100)]
        for key, value, warning, critical, maximum in pressure:
            record(key, self._severity(value, warning, critical) if finite(value, maximum=maximum) else 'unknown',
                   {'value': value, 'warning': warning, 'critical': critical}, self.cfg.pressure_hold_seconds)
        record('telemetry', 'warning' if missing else 'ok', {'missing': missing})
        result['degraded'] = self.degraded
        return result

    @staticmethod
    def _severity(value, warning, critical):
        return 'critical' if critical is not None and value > critical else ('warning' if value > warning else 'ok')
