export interface ErrorEnvelope {
  error: string;
  code: string;
  message: string;
  request_id: string;
}

export function parseEnvelope(data: unknown): ErrorEnvelope | undefined {
  if (!data || typeof data !== "object") return undefined;
  const o = data as Record<string, unknown>;
  if (!("error" in o) && !("code" in o)) return undefined;
  return {
    error: String(o.error ?? ""),
    code: String(o.code ?? ""),
    message: String(o.message ?? o.error ?? ""),
    request_id: String(o.request_id ?? ""),
  };
}

export class KanusError extends Error {
  status: number;
  envelope?: ErrorEnvelope;
  /** Raw body — never included in message/toString. */
  raw: string;
  retryAfter?: number;

  constructor(opts: {
    status: number;
    envelope?: ErrorEnvelope;
    raw?: string;
    retryAfter?: number;
  }) {
    const code = opts.envelope?.code ?? "";
    super(`KanusError(status=${opts.status}, code=${JSON.stringify(code)})`);
    this.name = "KanusError";
    this.status = opts.status;
    this.envelope = opts.envelope;
    this.raw = opts.raw ?? "";
    this.retryAfter = opts.retryAfter;
  }

  override toString(): string {
    const code = this.envelope?.code ?? "";
    const rid = this.envelope?.request_id ?? "";
    return `KanusError(status=${this.status}, code=${JSON.stringify(code)}, request_id=${JSON.stringify(rid)}, raw=<redacted len=${this.raw.length}>)`;
  }
}

export interface SSEEvent {
  data: string;
  event?: string;
  id?: string;
}
