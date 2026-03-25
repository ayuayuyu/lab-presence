"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { syncAuthMe } from "./api";

export type AuthUser = {
  id: number;
  email: string;
  name: string;
  picture: string;
  hd: string;
};

type AuthContextType = {
  user: AuthUser | null;
  setUser: (user: AuthUser | null) => void;
  isAdmin: boolean;
  isOutsider: boolean;
  isLoading: boolean;
  login: (idToken: string) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextType>({
  user: null,
  setUser: () => {},
  isAdmin: false,
  isOutsider: false,
  isLoading: true,
  login: () => {},
  logout: () => {},
});

const ALLOWED_DOMAIN = "pluslab.org";
const TOKEN_KEY = "lab_id_token";

const ADMIN_EMAILS: Set<string> = new Set(
  (process.env.NEXT_PUBLIC_ADMIN_EMAILS || "")
    .split(",")
    .map((e) => e.trim())
    .filter(Boolean)
);

function decodeJWTPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    const binary = atob(parts[1].replace(/-/g, "+").replace(/_/g, "/"));
    const bytes = Uint8Array.from(binary, (c) => c.charCodeAt(0));
    const payload = new TextDecoder().decode(bytes);
    return JSON.parse(payload);
  } catch {
    return null;
  }
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

async function syncUser(
  payload: Record<string, unknown>,
  setUser: (u: AuthUser) => void,
  setIsAdmin: (v: boolean) => void
) {
  const email = payload.email as string;
  const googleName = payload.name as string;
  const googlePicture = payload.picture as string;
  const hd = (payload.hd as string) || "";

  try {
    const dbUser = await syncAuthMe(googleName, googlePicture || "");
    setUser({
      id: dbUser.id,
      email,
      name: dbUser.name,
      picture: dbUser.picture || googlePicture || "",
      hd,
    });
  } catch {
    // API失敗時はJWTの情報だけで進む
    setUser({
      id: 0,
      email,
      name: googleName,
      picture: googlePicture || "",
      hd,
    });
  }
  setIsAdmin(ADMIN_EMAILS.has(email));
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isAdmin, setIsAdmin] = useState(false);
  const [isOutsider, setIsOutsider] = useState(false);
  // トークンが存在する場合のみ復元処理が走るため、ない場合は最初からfalse
  const [isLoading, setIsLoading] = useState(() => {
    const token = getToken();
    return !!(token && decodeJWTPayload(token));
  });

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setUser(null);
    setIsAdmin(false);
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
    syncUser(payload, setUser, setIsAdmin);
  }, []);

  // 起動時に localStorage から復元
  useEffect(() => {
    const token = getToken();
    if (token) {
      const payload = decodeJWTPayload(token);
      if (payload) {
        syncUser(payload, setUser, setIsAdmin).finally(() =>
          setIsLoading(false)
        );
        return;
      }
    }
    // isLoading はトークンなしの場合 false で初期化されているため setIsLoading 不要
  }, []);

  return (
    <AuthContext.Provider value={{ user, setUser, isAdmin, isOutsider, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
