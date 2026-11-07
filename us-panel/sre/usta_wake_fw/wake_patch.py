p="/opt/usta-wake.py"; s=open(p).read()
helpers='''
A100_TOKEN_FILE = "/etc/usta-wake/a100.token"
A100_ALLOWED_IPS = {"31.220.63.191"}


def a100_token_ok(header: str) -> bool:
    import hmac
    try:
        want = open(A100_TOKEN_FILE, encoding="utf-8").read().strip()
    except OSError:
        return False
    got = (header or "").strip()
    if not want or not got.startswith("Bearer "):
        return False
    return hmac.compare_digest(got[7:].strip().encode(), want.encode())


def a100_start_async() -> None:
    def run() -> None:
        worker = WORKERS["a100"]
        try:
            with LOCKS["a100"]:
                if power_verb(instance_status(worker)) == "start":
                    stop_other_gpus(worker)
                    power(worker, "start")
        except Exception as exc:  # noqa: BLE001
            print("a100 start failed", type(exc).__name__, flush=True)
    threading.Thread(target=run, daemon=True).start()


class Door(BaseHTTPRequestHandler):'''
s=s.replace("\nclass Door(BaseHTTPRequestHandler):",helpers,1)
old='''    def forward(self) -> None:
        if not worth_waking(self.path):'''
new='''    def a100_control(self) -> bool:
        bare = self.path.split("?", 1)[0]
        if bare not in ("/a100/start", "/a100/status"):
            return False
        def reply(code: int, doc: dict) -> None:
            body = json.dumps(doc).encode() + b"\\n"
            self.send_response(code)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.send_header("Connection", "close")
            self.end_headers()
            if self.command != "HEAD":
                self.wfile.write(body)
        if self.client_address[0] not in A100_ALLOWED_IPS or not a100_token_ok(self.headers.get("Authorization", "")):
            reply(403, {"ok": False})
            return True
        worker = WORKERS["a100"]
        if bare == "/a100/status" and self.command == "GET":
            try:
                status = instance_status(worker)
            except Exception:  # noqa: BLE001
                status = "UNKNOWN"
            reply(200, {"ok": True, "instance": status, "model_ready": status == "RUNNING" and worker_awake(worker)})
            return True
        if bare == "/a100/start" and self.command == "POST":
            length = int(self.headers.get("Content-Length", "0") or "0")
            if length:
                self.rfile.read(min(length, 4096))
            try:
                status = instance_status(worker)
            except Exception:  # noqa: BLE001
                status = "UNKNOWN"
            if power_verb(status) == "start":
                a100_start_async()
                reply(202, {"ok": True, "instance": status, "action": "start_requested"})
            else:
                reply(200, {"ok": True, "instance": status, "action": "none"})
            return True
        reply(405, {"ok": False})
        return True

    def forward(self) -> None:
        if self.a100_control():
            return
        if not worth_waking(self.path):'''
assert s.count(old)==1
s=s.replace(old,new,1)
open(p,"w").write(s)
