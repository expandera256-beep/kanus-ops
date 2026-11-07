import http from "node:http";

export type MockState = {
  force429: number;
  seenIdempotency: string[];
  lastAuthPrefix: string;
  sseMode: "ok" | "drop";
};

export function startMock(): Promise<{ baseUrl: string; state: MockState; close: () => Promise<void> }> {
  const state: MockState = {
    force429: 0,
    seenIdempotency: [],
    lastAuthPrefix: "",
    sseMode: "ok",
  };

  const server = http.createServer((req, res) => {
    const auth = req.headers.authorization ?? "";
    state.lastAuthPrefix = auth.slice(0, 20);
    const idem = req.headers["idempotency-key"];
    if (typeof idem === "string") state.seenIdempotency.push(idem);

    const sendErr = (status: number, code: string, msg: string) => {
      const body = JSON.stringify({
        error: msg,
        code,
        message: msg,
        request_id: "req_ts_1",
      });
      res.writeHead(status, {
        "Content-Type": "application/json",
        "X-Request-Id": "req_ts_1",
        ...(status === 429 ? { "Retry-After": "0" } : {}),
      });
      res.end(body);
    };

    const readBody = (cb: (j: unknown) => void) => {
      const chunks: Buffer[] = [];
      req.on("data", (c) => chunks.push(c));
      req.on("end", () => {
        const raw = Buffer.concat(chunks).toString("utf8");
        cb(raw ? JSON.parse(raw) : {});
      });
    };

    if (state.force429 > 0 && req.url !== "/v1/chat/completions") {
      state.force429--;
      return sendErr(429, "rate_limited", "rate");
    }

    if (req.method === "GET" && req.url === "/v1/credits/status") {
      if (state.force429 > 0) {
        state.force429--;
        return sendErr(429, "rate_limited", "rate");
      }
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ kind: "account", remaining: 7, available: 7 }));
      return;
    }

    if (req.method === "POST" && req.url === "/v1/credits/reserve") {
      if (auth.includes("usa1.") || auth.includes("usk_")) {
        return sendErr(401, "unauthorized", "license");
      }
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ id: "res_ts" }));
      return;
    }

    if (req.method === "POST" && req.url === "/v1/credits/settle") {
      return readBody(() => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ ok: true }));
      });
    }

    if (req.method === "POST" && req.url === "/v1/chat/completions") {
      if (state.force429 > 0) {
        state.force429--;
        return sendErr(429, "rate_limited", "rate");
      }
      res.writeHead(200, { "Content-Type": "text/event-stream" });
      if (state.sseMode === "drop") {
        res.write("data: {\"partial\":true}\n");
        res.end();
        return;
      }
      // fragmented writes
      res.write("data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n");
      res.write("\n");
      res.write("data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n");
      res.write("data: [DONE]\n\n");
      res.end();
      return;
    }

    sendErr(404, "not_found", "not found");
  });

  return new Promise((resolve) => {
    server.listen(0, "127.0.0.1", () => {
      const addr = server.address();
      if (!addr || typeof addr === "string") throw new Error("addr");
      resolve({
        baseUrl: `http://127.0.0.1:${addr.port}`,
        state,
        close: () =>
          new Promise((r, j) => server.close((e) => (e ? j(e) : r()))),
      });
    });
  });
}
