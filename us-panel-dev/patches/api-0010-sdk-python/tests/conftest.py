import json
import threading
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse

import pytest


class MockState:
    def __init__(self):
        self.reserve_calls = 0
        self.seen_idempotency = []
        self.force_429_remaining = 0
        self.sse_mode = "ok"  # ok | fragment | drop
        self.last_auth = None


def make_handler(state: MockState):
    class H(BaseHTTPRequestHandler):
        def log_message(self, fmt, *args):
            return  # silence; never log Authorization

        def _read_json(self):
            n = int(self.headers.get("Content-Length") or 0)
            if n == 0:
                return {}
            return json.loads(self.rfile.read(n))

        def _auth(self):
            a = self.headers.get("Authorization", "")
            state.last_auth = a[:20] + "…" if len(a) > 20 else a
            return a

        def _error(self, code, status, msg):
            body = json.dumps(
                {"error": msg, "code": code, "message": msg, "request_id": "req_test_1"}
            ).encode()
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("X-Request-Id", "req_test_1")
            if status == 429:
                self.send_header("Retry-After", "0")
            self.end_headers()
            self.wfile.write(body)

        def do_GET(self):
            path = urlparse(self.path).path
            auth = self._auth()
            if state.force_429_remaining > 0:
                state.force_429_remaining -= 1
                return self._error("rate_limited", 429, "rate")
            if path == "/v1/credits/status":
                if not auth.startswith("Bearer "):
                    return self._error("unauthorized", 401, "license")
                body = json.dumps({"kind": "account", "remaining": 10, "available": 10}).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(body)
                return
            if path == "/v1/usage":
                body = json.dumps({"events": 1, "by_route": {"authorize": 1}}).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(body)
                return
            self._error("not_found", 404, "not found")

        def do_POST(self):
            path = urlparse(self.path).path
            auth = self._auth()
            idem = self.headers.get("Idempotency-Key")
            if idem:
                state.seen_idempotency.append(idem)
            if state.force_429_remaining > 0 and path != "/v1/chat/completions":
                state.force_429_remaining -= 1
                return self._error("rate_limited", 429, "rate")
            if path == "/v1/credits/reserve":
                if "usk_" in auth or "usa1." in auth:
                    return self._error("unauthorized", 401, "license")
                state.reserve_calls += 1
                body = json.dumps({"id": "res_1"}).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(body)
                return
            if path == "/v1/credits/settle":
                payload = self._read_json()
                body = json.dumps({"ok": True, "id": payload.get("id")}).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(body)
                return
            if path == "/v1/chat/completions":
                if state.force_429_remaining > 0:
                    state.force_429_remaining -= 1
                    return self._error("rate_limited", 429, "rate")
                self.send_response(200)
                self.send_header("Content-Type", "text/event-stream")
                self.end_headers()
                if state.sse_mode == "drop":
                    self.wfile.write(b"data: {\"partial\":true}\n")
                    self.wfile.flush()
                    # close without finishing → client sees disconnect / incomplete
                    return
                # fragmented SSE: write tiny pieces
                parts = [
                    b"data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n",
                    b"\n",
                    b"data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n",
                    b"data: [DONE]\n\n",
                ]
                for p in parts:
                    self.wfile.write(p)
                    self.wfile.flush()
                return
            self._error("not_found", 404, "not found")

    return H


@pytest.fixture
def mock_api():
    state = MockState()
    server = HTTPServer(("127.0.0.1", 0), make_handler(state))
    t = threading.Thread(target=server.serve_forever, daemon=True)
    t.start()
    host, port = server.server_address
    yield f"http://{host}:{port}", state
    server.shutdown()
