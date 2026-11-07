"""Bounded, thread-safe instance-local token buckets (not a global quota)."""
import json
import math
import threading
import time

DEFAULT_LIMITS = {'health': (2.0, 10), 'read': (5.0, 30), 'api': (10.0, 60)}
HEALTH_PATHS = frozenset(('/health', '/healthz', '/ready', '/readyz'))
READ_PATHS = frozenset(('/v1/models', '/account/balance', '/account/requests', '/account/workers'))

class TokenBuckets:
    def __init__(self, limits=None, clock=time.monotonic):
        self.clock = clock
        self.lock = threading.Lock()
        self.limits = dict(DEFAULT_LIMITS if limits is None else limits)
        if set(self.limits) != {'health', 'read', 'api'}:
            raise ValueError('three bucket groups required')
        for rate, capacity in self.limits.values():
            if not math.isfinite(rate) or rate <= 0 or not math.isfinite(capacity) or capacity < 1:
                raise ValueError('positive finite rate/capacity required')
        now = clock()
        self.state = {k: [float(capacity), now] for k, (_, capacity) in self.limits.items()}

    def take(self, group):
        with self.lock:
            rate, capacity = self.limits[group]
            tokens, stamp = self.state[group]
            now = max(stamp, self.clock())  # Clock rollback cannot mint tokens.
            tokens = min(float(capacity), tokens + (now - stamp) * rate)
            if tokens >= 1:
                self.state[group] = [tokens - 1, now]
                return True, 0
            self.state[group] = [tokens, now]
            return False, max(1, math.ceil((1 - tokens) / rate))

def rate_limit_app(app, *, service, buckets=None):
    buckets = TokenBuckets() if buckets is None else buckets
    def limited(env, start):
        method, path = env.get('REQUEST_METHOD', ''), env.get('PATH_INFO', '')
        # Cloud Run IAM, not a caller-controlled header, authenticates this service.
        # Only the IAM-private watchdog's scheduler POST bypasses the limiter.
        if service == 'watchdog' and method == 'POST' and path == '/internal/reconcile':
            return app(env, start)
        group = 'health' if path in HEALTH_PATHS else 'read' if path in READ_PATHS else 'api'
        ok, retry = buckets.take(group)
        if ok:
            return app(env, start)
        data = json.dumps({'error': 'rate limit exceeded'}).encode()
        start('429 Too Many Requests', [('Content-Type', 'application/json'),
              ('Cache-Control', 'no-store'), ('Retry-After', str(retry)),
              ('Content-Length', str(len(data)))])
        return [data]
    return limited
