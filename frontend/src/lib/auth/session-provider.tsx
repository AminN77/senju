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

import { createApiClient } from "@/lib/api/client";
import type { components } from "@/lib/api/generated/schema";

import type { SessionEnvelope, SessionState, SessionUser } from "./types";

type SignInInput = components["schemas"]["AuthLoginRequest"];

interface SessionContextValue {
  user: SessionUser | null;
  status: SessionState["status"];
  signIn: (input: SignInInput) => Promise<void>;
  signOut: () => Promise<void>;
  refresh: () => Promise<void>;
}

const SessionContext = createContext<SessionContextValue | null>(null);

export function SessionProvider({
  children,
  bootstrap = true,
}: {
  children: ReactNode;
  bootstrap?: boolean;
}) {
  const authClient = useMemo(() => createApiClient(), []);

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
    try {
      const result = await authClient.POST("/v1/auth/refresh", {
        credentials: "include",
        headers: { "content-type": "application/json" },
        cache: "no-store",
      });

      if (result.error || !result.data) {
        clearSession();
        return;
      }

      applyAuthenticatedSession(result.data);
    } catch {
      clearSession();
    }
  }, [applyAuthenticatedSession, authClient, clearSession]);

  const signIn = useCallback(
    async (input: SignInInput) => {
      const result = await authClient.POST("/v1/auth/login", {
        body: input,
        credentials: "include",
        headers: { "content-type": "application/json" },
      });

      if (result.error || !result.data) {
        clearSession();
        throw new Error("Authentication failed");
      }

      applyAuthenticatedSession(result.data);
    },
    [applyAuthenticatedSession, authClient, clearSession]
  );

  const signOut = useCallback(async () => {
    try {
      await authClient.POST("/v1/auth/logout", {
        credentials: "include",
        headers: { "content-type": "application/json" },
      });
    } finally {
      clearSession();
    }
  }, [authClient, clearSession]);

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
