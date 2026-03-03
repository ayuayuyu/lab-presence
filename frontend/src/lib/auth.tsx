"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

export type AuthUser = {
  email: string;
  name: string;
  picture: string;
  hd: string;
};

type AuthContextType = {
  user: AuthUser | null;
  isOutsider: boolean;
  isLoading: boolean;
  login: (idToken: string) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextType>({
  user: null,
  isOutsider: false,
  isLoading: true,
  login: () => {},
  logout: () => {},
});

const ALLOWED_DOMAIN = "pluslab.org";
const TOKEN_KEY = "lab_id_token";

function decodeJWTPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    const payload = atob(parts[1].replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(payload);
  } catch {
    return null;
  }
}

function isTokenExpired(payload: Record<string, unknown>): boolean {
  const exp = payload.exp as number | undefined;
  if (!exp) return true;
  return Date.now() / 1000 > exp;
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) return null;

  const payload = decodeJWTPayload(token);
  if (!payload || isTokenExpired(payload)) {
    localStorage.removeItem(TOKEN_KEY);
    return null;
  }
  return token;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isOutsider, setIsOutsider] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setUser(null);
    setIsOutsider(false);
  }, []);

  const login = useCallback((idToken: string) => {
    const payload = decodeJWTPayload(idToken);
    if (!payload) return;

    const hd = (payload.hd as string) || "";
    if (hd !== ALLOWED_DOMAIN) {
      setIsOutsider(true);
      setUser(null);
      return;
    }

    setIsOutsider(false);
    localStorage.setItem(TOKEN_KEY, idToken);
    setUser({
      email: payload.email as string,
      name: payload.name as string,
      picture: payload.picture as string,
      hd,
    });
  }, []);

  // 起動時に localStorage から復元
  useEffect(() => {
    const token = getToken();
    if (token) {
      const payload = decodeJWTPayload(token);
      if (payload && !isTokenExpired(payload)) {
        setUser({
          email: payload.email as string,
          name: payload.name as string,
          picture: payload.picture as string,
          hd: (payload.hd as string) || "",
        });
      }
    }
    setIsLoading(false);
  }, []);

  // トークン期限チェック（1分ごと）
  useEffect(() => {
    const interval = setInterval(() => {
      const token = getToken();
      if (!token && user) {
        logout();
      }
    }, 60_000);
    return () => clearInterval(interval);
  }, [user, logout]);

  return (
    <AuthContext.Provider value={{ user, isOutsider, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
