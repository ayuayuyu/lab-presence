"use client";

import { useAuth } from "@/lib/auth";
import LoginPage from "./LoginPage";
import type { ReactNode } from "react";

export default function AuthGate({ children }: { children: ReactNode }) {
  const { user, isLoading, isOutsider } = useAuth();

  if (isLoading) {
    return (
      <div className="login-page">
        <p style={{ color: "var(--muted)" }}>読み込み中...</p>
      </div>
    );
  }

  if (!user || isOutsider) {
    return <LoginPage />;
  }

  return <>{children}</>;
}
