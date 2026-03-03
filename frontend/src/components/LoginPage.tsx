"use client";

import { useAuth } from "@/lib/auth";
import { useEffect, useRef } from "react";

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (response: { credential: string }) => void;
          }) => void;
          renderButton: (
            element: HTMLElement,
            config: { theme: string; size: string; width: number }
          ) => void;
        };
      };
    };
  }
}

export default function LoginPage() {
  const { login, isOutsider } = useAuth();
  const buttonRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;
    if (!clientId) return;

    const script = document.createElement("script");
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.onload = () => {
      window.google?.accounts.id.initialize({
        client_id: clientId,
        callback: (response) => {
          login(response.credential);
        },
      });
      if (buttonRef.current) {
        window.google?.accounts.id.renderButton(buttonRef.current, {
          theme: "outline",
          size: "large",
          width: 300,
        });
      }
    };
    document.head.appendChild(script);

    return () => {
      document.head.removeChild(script);
    };
  }, [login]);

  return (
    <div className="login-page">
      <div className={`login-card ${isOutsider ? "login-card--outsider" : ""}`}>
        {isOutsider ? (
          <>
            <div className="login-icon login-icon--outsider">!</div>
            <h1 className="login-title login-title--outsider">
              あなたは部外者です
            </h1>
            <p className="login-subtitle">
              このシステムは @pluslab.org アカウント専用です
            </p>
            <button
              onClick={() => window.location.reload()}
              className="login-retry-button"
            >
              別のアカウントで試す
            </button>
          </>
        ) : (
          <>
            <div className="login-icon">P</div>
            <h1 className="login-title">Pluslab 滞在管理</h1>
            <p className="login-subtitle">
              @pluslab.org アカウントでログインしてください
            </p>
            <div ref={buttonRef} className="login-google-button" />
          </>
        )}
      </div>
    </div>
  );
}
