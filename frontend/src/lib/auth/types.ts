export type SessionStatus = "loading" | "authenticated" | "unauthenticated";

export interface SessionUser {
  id: string;
  email: string;
  role: "uploader" | "runner" | "analyst" | "admin";
}

export interface SessionState {
  user: SessionUser | null;
  accessToken: string | null;
  status: SessionStatus;
}

export interface SessionEnvelope {
  user: SessionUser;
  access_token: string;
}
