"""Cloud Logging JSON request records; inputs are allowlisted, never dumped."""
import ipaddress
import json
import os
import re
import sys
import time
import uuid

KNOWN_PATHS = frozenset(('/health', '/healthz', '/ready', '/readyz', '/v1/models',
    '/v1/chat/completions', '/account/register', '/account/login', '/account/redeem',
    '/account/balance', '/account/requests', '/account/workers', '/webhooks/shopify',
    '/internal/reconcile', '/'))

def _stdout(record):
    sys.stdout.write(json.dumps(record, separators=(',', ':')) + '\n')
    sys.stdout.flush()

def logging_app(app, *, project=None, emit=None, clock=time.monotonic):
    emit = _stdout if emit is None else emit
    project = project or os.environ.get('GOOGLE_CLOUD_PROJECT') or os.environ.get('GCP_PROJECT')
    project = project if isinstance(project, str) and re.fullmatch(r'[a-z][a-z0-9-]{4,62}', project) else None
    def logged(env, start):
        begin = clock()
        state = {'status': 500, 'bytes': 0, 'started': False, 'failed': False}
        request_id = uuid.uuid4().hex  # Never log caller-controlled request IDs.
        method = env.get('REQUEST_METHOD', '')
        method = method if method in ('GET','POST','PUT','PATCH','DELETE','HEAD','OPTIONS') else 'UNKNOWN'
        path = env.get('PATH_INFO', '')
        path = path if path in KNOWN_PATHS else '[unmatched]'
        trace = env.get('HTTP_X_CLOUD_TRACE_CONTEXT', '')
        valid_trace = re.fullmatch(r'([0-9a-fA-F]{32})(?:/[0-9]{1,20})?(?:;o=[01])?', trace)
        trace_id = valid_trace.group(1).lower() if valid_trace else uuid.uuid4().hex
        def response(status, headers, exc_info=None):
            state['status'] = int(status.split(' ',1)[0])
            state['started'] = True
            # Preserve existing headers; add a server-generated correlation ID.
            outgoing = list(headers) + [('X-Request-ID', request_id)]
            return start(status, outgoing, exc_info) if exc_info is not None else start(status, outgoing)
        def failure():
            if not state['started']:
                data = b'{"error":"service unavailable"}'
                response('500 Internal Server Error', [('Content-Type','application/json'),
                         ('Content-Length', str(len(data))), ('Cache-Control','no-store')])
                return [data]
            return []
        def stream():
            body = None
            try:
                try:
                    body = app(env, response)
                except Exception:
                    state['failed'] = True
                    body = failure()
                try:
                    for part in body:
                        state['bytes'] += len(part)
                        yield part
                except Exception:
                    state['failed'] = True
                    for part in failure():
                        state['bytes'] += len(part)
                        yield part
            finally:
                if body is not None and callable(getattr(body, 'close', None)):
                    try:
                        body.close()
                    except Exception:
                        state['failed'] = True
                status = state['status']
                http = {'requestMethod': method, 'requestUrl': path, 'status': status,
                        'responseSize': str(state['bytes']), 'latency': format(max(0, clock()-begin), '.6f')+'s'}
                try:
                    http['remoteIp'] = str(ipaddress.ip_address(env.get('REMOTE_ADDR', '')))
                except ValueError:
                    pass
                record = {'severity': 'ERROR' if state['failed'] or status >= 500 else 'WARNING' if status >= 400 else 'INFO',
                          'message': 'HTTP request failed' if state['failed'] else 'HTTP request completed',
                          'httpRequest': http, 'request_id': request_id}
                if project:
                    record['logging.googleapis.com/trace'] = 'projects/'+project+'/traces/'+trace_id
                try:
                    emit(record)
                except Exception:
                    pass  # Logging failures cannot fail an otherwise valid request.
        return stream()
    return logged
