import { afterEach, describe, expect, it } from "vitest";

import { createServer, type Server } from "node:http";

import { createApiClient, getVersion } from "./client";

describe("api client", () => {
  let server: Server | undefined;

  afterEach(async () => {
    if (!server) {
      return;
    }

    await new Promise<void>((resolve, reject) => {
      server?.close((error) => {
        if (error) {
          reject(error);
          return;
        }
        resolve();
      });
    });

    server = undefined;
  });

  it("calls generated endpoint with typed request/response", async () => {
    const observed = {
      auth: "",
      requestId: "",
    };

    server = createServer((req, res) => {
      observed.auth = req.headers.authorization ?? "";
      observed.requestId = req.headers["x-request-id"]?.toString() ?? "";

      res.writeHead(200, { "content-type": "application/json" });
      res.end(
        JSON.stringify({
          service: "senju-api",
          version: "0.1.0",
          commit: "abc123",
          build_time: "2026-04-23T00:00:00Z",
        })
      );
    });

    await new Promise<void>((resolve, reject) => {
      server?.once("error", reject);
      server?.listen(0, "127.0.0.1", () => resolve());
    });

    const address = server.address();
    if (!address || typeof address === "string") {
      throw new Error("Failed to acquire test server address");
    }

    const client = createApiClient({
      baseUrl: `http://127.0.0.1:${address.port}`,
      getAuthToken: () => "test-token",
    });

    const payload = await getVersion(client);

    expect(payload.service).toBe("senju-api");
    expect(payload.version).toBe("0.1.0");
    expect(observed.auth).toBe("Bearer test-token");
    expect(observed.requestId).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
    );
  });

  it("normalizes problem+json errors", async () => {
    server = createServer((_req, res) => {
      res.writeHead(400, {
        "content-type": "application/problem+json",
        "x-request-id": "request-123",
      });
      res.end(
        JSON.stringify({
          type: "https://senju.dev/problems/invalid-request",
          title: "Invalid request",
          status: 400,
        })
      );
    });

    await new Promise<void>((resolve, reject) => {
      server?.once("error", reject);
      server?.listen(0, "127.0.0.1", () => resolve());
    });

    const address = server.address();
    if (!address || typeof address === "string") {
      throw new Error("Failed to acquire test server address");
    }

    const client = createApiClient({
      baseUrl: `http://127.0.0.1:${address.port}`,
    });

    await expect(getVersion(client)).rejects.toMatchObject({
      status: 400,
      requestId: "request-123",
      problem: {
        title: "Invalid request",
      },
    });
  });
});
