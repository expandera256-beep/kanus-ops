import json
import sys
from dataclasses import replace
from pathlib import Path
import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from hook_watchdog import Config, Health, Login, Metrics, Watchdog, p95, check_health


class Clock:
    value = 1793945700.
    def __call__(self): return self.value


@pytest.fixture(autouse=True)
def no_network(monkeypatch):
    import socket
    def forbidden(*a, **k): raise AssertionError('network forbidden')
    monkeypatch.setattr(socket, 'socket', forbidden)
    monkeypatch.setattr(socket, 'create_connection', forbidden)


@pytest.fixture
def env(tmp_path):
    c = Clock()
    cfg = Config(alarm_outbox_path=str(tmp_path / 'outbox.jsonl'))
    return c, cfg, Watchdog(cfg, c)


def frame(c, **kwargs):
    m = Metrics(c(), Health(200, .1), tuple(Login(c(), .2) for _ in range(20)), 2, .5, 0, 0, 0)
    return replace(m, **kwargs)


def tick(env, **kwargs):
    c, cfg, w = env
    r = w.evaluate(frame(c, **kwargs))
    c.value += cfg.poll_seconds
    return r


def mails(env):
    p = Path(env[1].alarm_outbox_path)
    return [json.loads(x) for x in p.read_text().splitlines()] if p.exists() else []


def test_health_n_and_recovery(env):
    for _ in range(2): tick(env, health=Health(503, .1))
    assert not mails(env)
    tick(env, health=Health(503, .1))
    assert len(mails(env)) == 1 and env[2].degraded
    for _ in range(3): tick(env)
    assert len(mails(env)) == 2 and not env[2].degraded
    assert 'recovered' in mails(env)[-1]['body']


@pytest.mark.parametrize('h', [Health(500, .1), Health(200, 2.01), Health(None, 2), Health(200, float('nan'))])
def test_health_failure_types(env, h):
    for _ in range(3): tick(env, health=h)
    assert len(mails(env)) == 1


def test_exact_health_timeout_allowed(env):
    assert tick(env, health=Health(200, 2))['healthz'] == 'ok'


def test_injected_health_callback():
    seen = []
    assert check_health(lambda t: seen.append(t) or Health(200, .1), 2).status_code == 200
    assert seen == [2]
    def fail(t): raise TimeoutError()
    assert check_health(fail, 2).status_code is None
    assert check_health(lambda t: 'invalid', 2).status_code is None


@pytest.mark.parametrize('values,expected', [([],None), ([1],1), (list(range(1,21)),19), ([2]*20,2)])
def test_nearest_rank(values, expected): assert p95(values) == expected


@pytest.mark.parametrize('bad', [-1,float('inf'),float('nan')])
def test_bad_durations(bad):
    with pytest.raises(ValueError): p95([bad])


@pytest.mark.parametrize('seconds,severity', [(2,'ok'),(2.1,'warning'),(5,'warning'),(5.1,'critical')])
def test_login_thresholds(env, seconds, severity):
    c, _, _ = env
    xs = tuple(Login(c(), seconds) for _ in range(20))
    assert tick(env, logins=xs)['login_p95'] == severity


def test_login_timeout_retained_and_failed_counted(env):
    c,_,_ = env
    xs = tuple(Login(c(), 8, failed=True, timed_out=True) for _ in range(20))
    r = tick(env, logins=xs)
    assert r['login_detail']['p95_seconds'] == 8
    assert r['login_detail']['failed'] == r['login_detail']['timeouts'] == 20


def test_login_n_consecutive(env):
    for _ in range(3):
        tick(env, logins=tuple(Login(env[0](), 6) for _ in range(20)))
    assert len(mails(env)) == 1


@pytest.mark.parametrize('kind', ['none','empty','old','future','invalid'])
def test_insufficient_samples_unknown(env, kind):
    c,_,_ = env
    xs = {'none':None,'empty':(), 'old':tuple(Login(c()-300,6) for _ in range(20)),
          'future':tuple(Login(c()+1,6) for _ in range(20)),
          'invalid':tuple(Login(c(),-1) for _ in range(20))}[kind]
    assert tick(env, logins=xs)['login_p95'] == 'unknown'


@pytest.mark.parametrize('kwargs,key', [({'load1':10},'load'),({'cpu_psi_some_avg60':60},'cpu_psi'),
    ({'memory_psi_full_avg60':2},'memory_psi'),({'io_psi_full_avg60':11},'io_psi')])
def test_pressure_requires_n_and_hold(env, kwargs, key):
    for _ in range(10): tick(env, **kwargs)
    assert not mails(env)
    tick(env, **kwargs)
    assert len(mails(env)) == 1
    assert 'hook.'+key in mails(env)[0]['body']


@pytest.mark.parametrize('kwargs,key,severity', [({'load1':4},'load','ok'),({'load1':8},'load','warning'),
    ({'load1':8.1},'load','critical'),({'cpu_psi_some_avg60':20},'cpu_psi','ok'),
    ({'cpu_psi_some_avg60':50},'cpu_psi','warning'),({'memory_psi_full_avg60':1},'memory_psi','ok'),
    ({'io_psi_full_avg60':10},'io_psi','ok'),({'ncpu':0},'load','unknown'),
    ({'cpu_psi_some_avg60':101},'cpu_psi','unknown')])
def test_pressure_boundaries(env, kwargs, key, severity):
    assert tick(env, **kwargs)[key] == severity


def test_cooldown_repeat_at_boundary(env):
    for _ in range(3): tick(env, health=Health(500,.1))
    assert len(mails(env)) == 1
    for _ in range(29): tick(env, health=Health(500,.1))
    assert len(mails(env)) == 1
    tick(env, health=Health(500,.1))
    assert len(mails(env)) == 2


def test_escalation_bypasses_cooldown(env):
    for _ in range(3): tick(env, logins=tuple(Login(env[0](),3) for _ in range(20)))
    tick(env, logins=tuple(Login(env[0](),6) for _ in range(20)))
    assert len(mails(env)) == 2


def test_success_breaks_violation_streak(env):
    tick(env, health=Health(500,.1)); tick(env, health=Health(500,.1)); tick(env)
    tick(env, health=Health(500,.1))
    assert not mails(env)


def test_missing_does_not_recover_active_condition(env):
    for _ in range(3): tick(env, health=Health(500,.1))
    for _ in range(3): tick(env, health=None)
    assert env[2].states['healthz'].active == 'critical'
    assert len(mails(env)) == 2  # health plus missing-data alarm


def test_independent_states(env):
    for _ in range(6): tick(env, logins=tuple(Login(env[0](),6) for _ in range(20)))
    assert env[2].states['login_p95'].active == 'critical'
    assert env[2].states['healthz'].active is None


def test_stale_and_gap_do_not_complete_hold(env):
    for _ in range(10): tick(env, load1=10)
    env[0].value += 100
    tick(env, load1=10)
    assert not mails(env)
    assert env[2].states['load'].bad == 1


def test_stale_unknown(env):
    assert env[2].evaluate(frame(env[0], timestamp=env[0]()-200))['telemetry'] == 'stale_or_invalid'


def test_duplicate_does_not_advance_streak(env):
    c,_,w = env
    m=frame(c,health=Health(500,.1))
    w.evaluate(m)
    for _ in range(10): w.evaluate(m)
    assert w.states['healthz'].bad == 1 and not mails(env)


def test_outbox_only_and_existing_alarm_reused(env):
    from hook_watchdog import ExistingAlarm
    for _ in range(3): tick(env,health=Health(500,.1))
    mail=mails(env)[0]
    assert isinstance(env[2].alarm, ExistingAlarm)
    assert mail['from']=='support@kanus.org' and mail['to']=='demirtumen08@gmail.com'
    assert mail['sent'] is False and mail['send_enabled'] is False
    assert '+03:00' in mail['body']


def test_outbox_error_and_retry(env):
    c,cfg,w=env
    cfg.alarm_outbox_path += '/missing/outbox'
    # Existing alarm has its own config; point it at the deliberately absent parent.
    w.alarm.cfg.alarm_outbox_path=cfg.alarm_outbox_path
    for _ in range(3): tick(env,health=Health(500,.1))
    assert w.degraded and w.outbox_error=='FileNotFoundError'
    assert w.states['healthz'].last_sent is None
    w.alarm.cfg.alarm_outbox_path=str(Path(cfg.alarm_outbox_path).parents[2]/'fixed.jsonl')
    tick(env,health=Health(500,.1))
    assert w.outbox_error is None and w.states['healthz'].active=='critical'


@pytest.mark.parametrize('kwargs',[{'alarm_send_enabled':True},{'consecutive_violations':0},
    {'consecutive_violations':2.5},{'cooldown_seconds':float('nan')},{'login_min_samples':0},
    {'login_warning_seconds':6},{'cpu_psi_critical_percent':101}])
def test_config_validation(kwargs):
    with pytest.raises(ValueError): Config(**kwargs)


def test_config_file_and_custom_thresholds(tmp_path):
    p=tmp_path/'config.json'; p.write_text(json.dumps({'consecutive_violations':1,'health_timeout_seconds':1}))
    cfg=Config.from_file(p)
    assert cfg.consecutive_violations==1 and cfg.health_timeout_seconds==1
    p.write_text('{"unknown":1}')
    with pytest.raises(TypeError): Config.from_file(p)


def test_repeated_stale_calls_do_not_increment(env):
    c,_,w=env
    m=frame(c,timestamp=c()-200)
    w.evaluate(m)
    for _ in range(10): w.evaluate(m)
    assert w.states['telemetry'].bad==1 and not mails(env)


def test_consecutive_and_hold_are_configurable(env):
    c, cfg, w=env
    cfg.consecutive_violations=1
    cfg.pressure_hold_seconds=30
    tick(env,load1=10)
    assert not mails(env)
    tick(env,load1=10)
    assert len(mails(env))==1


def test_recurrence_within_cooldown_still_degraded(env):
    for _ in range(3): tick(env,health=Health(500,.1))
    for _ in range(3): tick(env)
    assert not env[2].degraded and len(mails(env))==2
    for _ in range(3): tick(env,health=Health(500,.1))
    assert env[2].degraded and len(mails(env))==2
