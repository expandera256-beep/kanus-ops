import { Auth } from "./auth.js";
import { ErrorEnvelope, KanusError, SSEEvent, parseEnvelope } from "./errors.js";

export type SleepFn = (ms: number) => Promise<void>;

const defaultSleep: SleepFn = (ms) => new Promise((r) => setTimeout(r, ms));

export const DEFAULT_MAX_RETRY_AFTER_SEC = 60;
export const DEFAULT_MAX_RETRIES = 3;
export const DEFAULT_SSE_MAX_BUFFER = 1 << 20;
export const DEFAULT_SSE_MAX_LINE = 64 * 1024;

/** Credit reserve/settle do not honor Idempotency-Key (server verified). Never auto-retry those POSTs. */

export interface KanusClientOptions {
  baseUrl: string;
  auth: Auth;
  timeoutMs?: number;
  maxRetries?: number;
  maxRetryAfterSec?: number;
  allowInsecure?: boolean;
  sseMaxBuffer?: number;
  sseMaxLine?: number;
  sleep?: SleepFn;
  fetchImpl?: typeof fetch;
}

function validateBaseUrl(baseUrl: string, allowInsecure: boolean): string {
  const u = new URL(baseUrl);
  if (u.protocol === "https:") return baseUrl.replace(/\/$/, "");
  if (u.protocol === "http:") {
    const host = u.hostname.toLowerCase();
    if (allowInsecure || host === "127.0.0.1" || host === "localhost" || host === "::1") {
      return baseUrl.replace(/\/$/, "");
    }
    throw new Error("base_url must use https (or localhost http / allowInsecure)");
  }
  throw new Error("base_url must be http(s)");
}

export class KanusClient {
  readonly baseUrl: string;
  readonly auth: Auth;
  readonly timeoutMs: number;
  readonly maxRetries: number;
  readonly maxRetryAfterSec: number;
  readonly sseMaxBuffer: number;
  readonly sseMaxLine: number;
  private sleep: SleepFn;
  private fetchImpl: typeof fetch;

  constructor(opts: KanusClientOptions) {
    this.baseUrl = validateBaseUrl(opts.baseUrl, !!opts.allowInsecure);
    this.auth = opts.auth;
    this.timeoutMs = opts.timeoutMs ?? 30_000;
    this.maxRetries = opts.maxRetries ?? DEFAULT_MAX_RETRIES;
    this.maxRetryAfterSec = opts.maxRetryAfterSec ?? DEFAULT_MAX_RETRY_AFTER_SEC;
    this.sseMaxBuffer = opts.sseMaxBuffer ?? DEFAULT_SSE_MAX_BUFFER;
    this.sseMaxLine = opts.sseMaxLine ?? DEFAULT_SSE_MAX_LINE;
    this.sleep = opts.sleep ?? defaultSleep;
    this.fetchImpl = opts.fetchImpl ?? fetch;
  }

  private headers(idempotencyKey?: string, accept = "application/json"): Record<string, string> {
    const h: Record<string, string> = {
      Authorization: this.auth.headerValue(),
      Accept: accept,
      "X-Kanus-Error-Envelope": "1",
      "User-Agent": "kanus-sdk-ts/0.1.1",
    };
    if (idempotencyKey) h["Idempotency-Key"] = idempotencyKey;
    return h;
  }

  /** Exported for tests. */
  capRetryAfter(retryAfter: number | undefined): number {
    const wait = retryAfter === undefined ? 1 : retryAfter;
    const w = wait < 0 ? 0 : wait;
    if (w > this.maxRetryAfterSec) {
      throw new KanusError({
        status: 429,
        envelope: {
          error: "rate",
          code: "retry_after_capped",
          message: "Retry-After exceeds maxRetryAfterSec",
          request_id: "",
        },
        retryAfter: w,
      });
    }
    return w;
  }

  private shouldRetry(method: string): boolean {
    const m = method.toUpperCase();
    return m === "GET" || m === "HEAD" || m === "OPTIONS";
  }

  private async request(
    method: string,
    path: string,
    body?: unknown,
    idempotencyKey?: string,
  ): Promise<{ status: number; headers: Headers; text: string }> {
    let attempt = 0;
    for (;;) {
      attempt++;
      const headers = this.headers(idempotencyKey);
      const init: RequestInit = { method, headers };
      if (body !== undefined) {
        headers["Content-Type"] = "application/json";
        init.body = JSON.stringify(body);
      }
      const ctrl = new AbortController();
      const t = setTimeout(() => ctrl.abort(), this.timeoutMs);
      try {
        const res = await this.fetchImpl(this.baseUrl + path, { ...init, signal: ctrl.signal });
        const text = await res.text();
        if (res.status === 429 && attempt <= this.maxRetries && this.shouldRetry(method)) {
          const ra = Number(res.headers.get("Retry-After") ?? "1");
          const waitSec = this.capRetryAfter(Number.isFinite(ra) ? ra : 1);
          await this.sleep(waitSec * 1000);
          continue;
        }
        if (!res.ok) {
          let envelope: ErrorEnvelope | undefined;
          try {
            envelope = parseEnvelope(JSON.parse(text));
          } catch {
            /* plain */
          }
          const raHdr = res.headers.get("Retry-After");
          throw new KanusError({
            status: res.status,
            envelope,
            raw: text,
            retryAfter: raHdr ? Number(raHdr) : undefined,
          });
        }
        return { status: res.status, headers: res.headers, text };
      } finally {
        clearTimeout(t);
      }
    }
  }

  async creditsStatus(): Promise<unknown> {
    const { text } = await this.request("GET", "/v1/credits/status");
    return JSON.parse(text);
  }

  async creditsReserve(idempotencyKey?: string): Promise<{ id: string }> {
    const { text } = await this.request("POST", "/v1/credits/reserve", {}, idempotencyKey);
    return JSON.parse(text);
  }

  async creditsSettle(
    id: string,
    commit: boolean,
    idempotencyKey?: string,
  ): Promise<{ ok: boolean }> {
    const { text } = await this.request(
      "POST",
      "/v1/credits/settle",
      { id, commit },
      idempotencyKey,
    );
    return JSON.parse(text);
  }

  async usage(): Promise<unknown> {
    const { text } = await this.request("GET", "/v1/usage");
    return JSON.parse(text);
  }

  async *chatSSE(
    payload: Record<string, unknown>,
    opts?: { idempotencyKey?: string; onChunk?: (chunk: Uint8Array) => void },
  ): AsyncGenerator<SSEEvent> {
    const body = { ...payload, stream: true };
    const headers = this.headers(opts?.idempotencyKey, "text/event-stream");
    headers["Content-Type"] = "application/json";
    const ctrl = new AbortController();
    const t = setTimeout(() => ctrl.abort(), this.timeoutMs);
    let res: Response;
    try {
      res = await this.fetchImpl(this.baseUrl + "/v1/chat/completions", {
        method: "POST",
        headers,
        body: JSON.stringify(body),
        signal: ctrl.signal,
      });
    } finally {
      clearTimeout(t);
    }
    if (!res.ok) {
      const text = await res.text();
      let envelope: ErrorEnvelope | undefined;
      try {
        envelope = parseEnvelope(JSON.parse(text));
      } catch {
        /* */
      }
      throw new KanusError({ status: res.status, envelope, raw: text });
    }
    if (!res.body) {
      throw new KanusError({ status: 0, raw: "sse_disconnect:nobody" });
    }
    const reader = res.body.getReader();
    const dec = new TextDecoder();
    let buf = "";
    let total = 0;
    try {
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        if (value) {
          total += value.byteLength;
          if (total > this.sseMaxBuffer) {
            throw new KanusError({ status: 0, raw: "sse_buffer_exceeded" });
          }
          opts?.onChunk?.(value);
          buf += dec.decode(value, { stream: true });
        }
        if (buf.length > this.sseMaxBuffer) {
          throw new KanusError({ status: 0, raw: "sse_buffer_exceeded" });
        }
        for (;;) {
          const idx = buf.indexOf("\n\n");
          if (idx < 0) break;
          const block = buf.slice(0, idx);
          buf = buf.slice(idx + 2);
          const ev: SSEEvent = { data: "" };
          for (const line of block.split("\n")) {
            if (line.length > this.sseMaxLine) {
              throw new KanusError({ status: 0, raw: "sse_line_exceeded" });
            }
            if (line.startsWith("data:")) ev.data += line.slice(5).trimStart();
            else if (line.startsWith("event:")) ev.event = line.slice(6).trimStart();
            else if (line.startsWith("id:")) ev.id = line.slice(3).trimStart();
          }
          if (ev.data === "[DONE]") return;
          if (ev.data) yield ev;
        }
      }
    } catch (e) {
      if (e instanceof KanusError) throw e;
      const name = e instanceof Error ? e.name : "Error";
      throw new KanusError({ status: 0, raw: `sse_disconnect:${name}` });
    }
  }
}
