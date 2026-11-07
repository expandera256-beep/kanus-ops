import http.server, json, os, sys, urllib.parse
ROOT = sys.argv[1]; CSS = sys.argv[2]
class H(http.server.BaseHTTPRequestHandler):
    def log_message(self, *a): pass
    def send(self, code, body, ct):
        b = body if isinstance(body, bytes) else body.encode()
        self.send_response(code); self.send_header('Content-Type', ct); self.send_header('Content-Length', str(len(b))); self.end_headers()
        if self.command != 'HEAD': self.wfile.write(b)
    def do_GET(self):
        u = urllib.parse.urlparse(self.path)
        if u.path == '/api/store.php':
            r = urllib.parse.parse_qs(u.query).get('resource', [''])[0]
            rows = [] if r in ('ledger','balance') else [{"id": "1", "title": "T", "body": "B", "author": "A", "createdAt": "2026-11-06T06:05:00.000Z", "replies": []}]
            return self.send(200, json.dumps(rows if r != 'balance' else {"credits": 0}), 'application/json')
        if u.path.endswith('.css'): return self.send(200, open(CSS, 'rb').read(), 'text/css')
        p = os.path.join(ROOT, u.path.lstrip('/'))
        if os.path.isfile(p): return self.send(200, open(p, 'rb').read(), 'application/javascript' if p.endswith('.js') else 'text/html')
        return self.send(200, open(os.path.join(ROOT, 'index.html'), 'rb').read(), 'text/html')
    def do_HEAD(self): return self.do_GET()
    def do_POST(self):
        self.rfile.read(int(self.headers.get('Content-Length') or 0))
        return self.send(500, json.dumps({"error": "internal"}), 'application/json')
http.server.ThreadingHTTPServer(('127.0.0.1', int(sys.argv[3])), H).serve_forever()
