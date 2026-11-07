from __future__ import annotations

import json
import time
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from typing import Any, Callable, Iterator, Optional
from urllib.parse import urlparse

from .auth import Auth

# Server (licensesrv CreditHandler) does NOT honor Idempotency-Key on
# /v1/credits/reserve|settle — verified 2026-11-06. Non-idempotent POSTs
# must never be auto-retried on 429.

DEFAULT_MAX_RETRY_AFTER_SEC = 60.0
DEFAULT_MAX_RETRIES = 3
DEFAULT_SSE_MAX_BUFFER = 1 << 20  # 1 MiB aggregate
DEFAULT_SSE_MAX_LINE = 64 * 1024


@dataclass
class ErrorEnvelope:
    error: str
    code: str
    message: str
    request_id: str

    @classmethod
    def from_body(cls, data: dict[str, Any]) -> ErrorEnvelope:
        return cls(
            error=str(data.get("error", "")),
            code=str(data.get("code", "")),
            message=str(data.get("message", data.get("error", ""))),
            request_id=str(data.get("request_id", "")),
        )


@dataclass
class KanusError(Exception):
    status: int
    envelope: Optional[ErrorEnvelope] = None
    raw: str = field(default="", repr=False)
    retry_after: Optional[int] = None

    def _safe_raw(self) -> str:
        if not self.raw:
            return ""
        if len(self.raw) <= 64:
            return "<redacted len=%d>" % len(self.raw)
        return "<redacted len=%d>" % len(self.raw)

    def __str__(self) -> str:
        code = self.envelope.code if self.envelope else ""
        return f"KanusError(status={self.status}, code={code!r})"

    def __repr__(self) -> str:
        code = self.envelope.code if self.envelope else ""
        rid = self.envelope.request_id if self.envelope else ""
        return f"KanusError(status={self.status}, code={code!r}, request_id={rid!r}, raw={self._safe_raw()!r})"


@dataclass
class SSEEvent:
    data: str
    event: str = "message"
    id: str = ""


def _validate_base_url(base_url: str, allow_insecure: bool) -> str:
    base_url = base_url.rstrip("/")
    u = urlparse(base_url)
    if u.scheme == "https":
        return base_url
    if u.scheme == "http":
        host = (u.hostname or "").lower()
        if allow_insecure or host in ("127.0.0.1", "localhost", "::1"):
            return base_url
        raise ValueError("base_url must use https (or localhost http / allow_insecure=True)")
    raise ValueError("base_url must be http(s)")


@dataclass
class KanusClient:
    base_url: str
    auth: Auth
    timeout: float = 30.0
    max_retries: int = DEFAULT_MAX_RETRIES
    max_retry_after: float = DEFAULT_MAX_RETRY_AFTER_SEC
    allow_insecure: bool = False
    sse_max_buffer: int = DEFAULT_SSE_MAX_BUFFER
    sse_max_line: int = DEFAULT_SSE_MAX_LINE
    opener: Any = field(default=None, repr=False)
    sleep: Callable[[float], None] = field(default=time.sleep, repr=False)

    def __post_init__(self) -> None:
        self.base_url = _validate_base_url(self.base_url, self.allow_insecure)
        if self.opener is None:
            self.opener = urllib.request.build_opener()

    def _headers(self, idempotency_key: str | None = None, accept: str = "application/json") -> dict[str, str]:
        h = {
            "Authorization": self.auth.header_value(),
            "Accept": accept,
            "X-Kanus-Error-Envelope": "1",
            "User-Agent": "kanus-sdk-python/0.1.1",
        }
        if idempotency_key:
            h["Idempotency-Key"] = idempotency_key
        return h

    def _cap_retry_after(self, retry_after: Optional[int]) -> float:
        wait = float(1 if retry_after is None else retry_after)
        if wait < 0:
            wait = 0.0
        if wait > self.max_retry_after:
            raise KanusError(
                status=429,
                envelope=ErrorEnvelope("rate", "retry_after_capped", "Retry-After exceeds max_retry_after", ""),
                raw="",
                retry_after=int(wait),
            )
        return wait

    def _should_retry(self, method: str, idempotency_key: str | None) -> bool:
        """GET is safe. POST only if caller supplied Idempotency-Key AND server supports it.
        Credit reserve/settle do NOT support Idempotency-Key → never retry those POSTs.
        """
        m = method.upper()
        if m in ("GET", "HEAD", "OPTIONS"):
            return True
        # Server does not honor Idempotency-Key on mutating credit routes.
        # Even with a key, auto-retry of POST is disabled for safety until server supports it.
        return False

    def _request(
        self,
        method: str,
        path: str,
        body: dict[str, Any] | None = None,
        *,
        idempotency_key: str | None = None,
    ) -> tuple[int, dict[str, str], bytes]:
        url = self.base_url + path
        data = None if body is None else json.dumps(body).encode()
        headers = self._headers(idempotency_key)
        if data is not None:
            headers["Content-Type"] = "application/json"
        attempt = 0
        while True:
            attempt += 1
            req = urllib.request.Request(url, data=data, headers=headers, method=method)
            try:
                resp = self.opener.open(req, timeout=self.timeout)
                raw = resp.read()
                hdrs = {k.lower(): v for k, v in resp.headers.items()}
                return resp.status, hdrs, raw
            except urllib.error.HTTPError as e:
                raw = e.read()
                hdrs = {k.lower(): v for k, v in e.headers.items()} if e.headers else {}
                retry_after = None
                if "retry-after" in hdrs:
                    try:
                        retry_after = int(hdrs["retry-after"])
                    except ValueError:
                        retry_after = 1
                envelope = None
                text = raw.decode("utf-8", "replace")
                try:
                    parsed = json.loads(text)
                    if isinstance(parsed, dict) and ("code" in parsed or "error" in parsed):
                        envelope = ErrorEnvelope.from_body(parsed)
                except json.JSONDecodeError:
                    pass
                if e.code == 429 and attempt <= self.max_retries and self._should_retry(method, idempotency_key):
                    wait = self._cap_retry_after(retry_after)
                    self.sleep(wait)
                    continue
                raise KanusError(status=e.code, envelope=envelope, raw=text, retry_after=retry_after) from None

    def credits_status(self) -> dict[str, Any]:
        _, _, raw = self._request("GET", "/v1/credits/status")
        return json.loads(raw)

    def credits_reserve(self, *, idempotency_key: str | None = None) -> dict[str, Any]:
        # idempotency_key accepted for forward-compat but server ignores it; no 429 retry.
        _, _, raw = self._request("POST", "/v1/credits/reserve", {}, idempotency_key=idempotency_key)
        return json.loads(raw)

    def credits_settle(self, reservation_id: str, commit: bool, *, idempotency_key: str | None = None) -> dict[str, Any]:
        _, _, raw = self._request(
            "POST",
            "/v1/credits/settle",
            {"id": reservation_id, "commit": commit},
            idempotency_key=idempotency_key,
        )
        return json.loads(raw)

    def usage(self) -> dict[str, Any]:
        _, _, raw = self._request("GET", "/v1/usage")
        return json.loads(raw)

    def chat_sse(
        self,
        payload: dict[str, Any],
        *,
        idempotency_key: str | None = None,
        on_chunk: Callable[[bytes], None] | None = None,
    ) -> Iterator[SSEEvent]:
        body = dict(payload)
        body["stream"] = True
        url = self.base_url + "/v1/chat/completions"
        data = json.dumps(body).encode()
        headers = self._headers(idempotency_key, accept="text/event-stream")
        headers["Content-Type"] = "application/json"
        req = urllib.request.Request(url, data=data, headers=headers, method="POST")
        try:
            resp = self.opener.open(req, timeout=self.timeout)
        except urllib.error.HTTPError as e:
            raw = e.read().decode("utf-8", "replace")
            envelope = None
            try:
                envelope = ErrorEnvelope.from_body(json.loads(raw))
            except Exception:
                pass
            raise KanusError(status=e.code, envelope=envelope, raw=raw) from None

        buf = ""
        total = 0
        try:
            while True:
                chunk = resp.read(64)
                if not chunk:
                    break
                total += len(chunk)
                if total > self.sse_max_buffer:
                    raise KanusError(status=0, raw="sse_buffer_exceeded")
                if on_chunk:
                    on_chunk(chunk)
                buf += chunk.decode("utf-8", "replace")
                if len(buf) > self.sse_max_buffer:
                    raise KanusError(status=0, raw="sse_buffer_exceeded")
                while "\n\n" in buf:
                    block, buf = buf.split("\n\n", 1)
                    for line in block.split("\n"):
                        if len(line) > self.sse_max_line:
                            raise KanusError(status=0, raw="sse_line_exceeded")
                    event = SSEEvent(data="")
                    for line in block.split("\n"):
                        if line.startswith("data:"):
                            event.data += line[5:].lstrip()
                        elif line.startswith("event:"):
                            event.event = line[6:].lstrip()
                        elif line.startswith("id:"):
                            event.id = line[3:].lstrip()
                    if event.data == "[DONE]":
                        return
                    if event.data:
                        yield event
        except KanusError:
            raise
        except (BrokenPipeError, ConnectionResetError, TimeoutError) as e:
            raise KanusError(status=0, raw=f"sse_disconnect:{type(e).__name__}") from e
