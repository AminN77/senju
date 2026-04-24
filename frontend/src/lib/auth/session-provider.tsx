"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import type { SessionEnvelope, SessionState, SessionUser } from "./types";

interface SignInInput {
  email: string;
  password: string;
}

interface SessionContextValue {
  user: SessionUser | null;
  status: SessionState["status"];
  signIn: (input: SignInInput) => Promise<void>;
  signOut: () => Promise<void>;
  refresh: () => Promise<void>;
}

const SessionContext = createContext<SessionContextValue | null>(null);

const AUTH_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export function SessionProvider({
  children,
  bootstrap = true,
}: {
  children: ReactNode;
  bootstrap?: boolean;
}) {
  const [session, setSession] = useState<SessionState>({
    user: null,
    accessToken: null,
    status: bootstrap ? "loading" : "unauthenticated",
  });

  const applyAuthenticatedSession = useCallback((payload: SessionEnvelope) => {
    setSession({
      user: payload.user,
      accessToken: payload.access_token,
      status: "authenticated",
    });
  }, []);

  const clearSession = useCallback(() => {
    setSession({
      user: null,
      accessToken: null,
      status: "unauthenticated",
    });
  }, []);

  const refresh = useCallback(async () => {
    const response = await fetch(`${AUTH_BASE_URL}/v1/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { "content-type": "application/json" },
      cache: "no-store",
    });

    if (!response.ok) {
      clearSession();
      return;
    }

    const payload = (await response.json()) as SessionEnvelope;
    applyAuthenticatedSession(payload);
  }, [applyAuthenticatedSession, clearSession]);

  const signIn = useCallback(
    async (input: SignInInput) => {
      const response = await fetch(`${AUTH_BASE_URL}/v1/auth/login`, {
        method: "POST",
        credentials: "include",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(input),
      });

      if (!response.ok) {
        clearSession();
        throw new Error("Authentication failed");
      }

      const payload = (await response.json()) as SessionEnvelope;
      applyAuthenticatedSession(payload);
    },
    [applyAuthenticatedSession, clearSession]
  );

  const signOut = useCallback(async () => {
    try {
      await fetch(`${AUTH_BASE_URL}/v1/auth/logout`, {
        method: "POST",
        credentials: "include",
        headers: { "content-type": "application/json" },
      });
    } finally {
      clearSession();
    }
  }, [clearSession]);

  useEffect(() => {
    if (!bootstrap) {
      return;
    }

    const handle = window.setTimeout(() => {
      void refresh();
    }, 0);

    return () => {
      window.clearTimeout(handle);
    };
  }, [bootstrap, refresh]);

  const value = useMemo<SessionContextValue>(
    () => ({
      user: session.user,
      status: session.status,
      signIn,
      signOut,
      refresh,
    }),
    [refresh, session.status, session.user, signIn, signOut]
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionContextValue {
  const value = useContext(SessionContext);
  if (!value) {
    throw new Error("useSession must be used inside SessionProvider");
  }
  return value;
}
