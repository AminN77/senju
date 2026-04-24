import type { components } from "@/lib/api/generated/schema";

export type SessionStatus = "loading" | "authenticated" | "unauthenticated";

export type SessionUser = components["schemas"]["AuthUser"];

export interface SessionState {
  user: SessionUser | null;
  accessToken: string | null;
  status: SessionStatus;
}

export type SessionEnvelope = components["schemas"]["AuthSessionResponse"];
