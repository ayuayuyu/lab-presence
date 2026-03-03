"use client";

import { useAuth } from "@/lib/auth";
import Link from "next/link";
import { usePathname } from "next/navigation";

export default function Header() {
  const { user, logout } = useAuth();
  const pathname = usePathname();

  if (!user) return null;

  const navLinks = [
    { href: "/", label: "ダッシュボード" },
    { href: "/register", label: "デバイス登録" },
  ];

  return (
    <header className="app-header">
      <div className="mx-auto max-w-5xl flex flex-wrap items-center justify-between gap-2 px-4 py-3">
        <Link href="/" className="text-lg font-bold">
          Pluslab滞在管理
        </Link>
        <div className="flex flex-wrap items-center gap-4">
          <nav className="flex gap-1 text-base">
            {navLinks.map((link) => {
              const isActive = pathname === link.href;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={`px-3 py-1.5 rounded-lg transition-colors ${
                    isActive
                      ? "font-bold"
                      : "hover:bg-[var(--accent-light)]"
                  }`}
                  style={{
                    color: isActive ? "var(--accent)" : "var(--muted)",
                    borderBottom: isActive
                      ? "2px solid var(--accent)"
                      : "2px solid transparent",
                  }}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>
          <div className="flex items-center gap-2">
            {user.picture ? (
              <img
                src={user.picture}
                alt={user.name}
                className="w-7 h-7 rounded-full"
                referrerPolicy="no-referrer"
              />
            ) : (
              <div
                className="w-7 h-7 rounded-full flex items-center justify-center text-white text-xs font-bold"
                style={{ backgroundColor: "var(--accent)" }}
              >
                {user.name?.charAt(0) || "?"}
              </div>
            )}
            <span className="text-sm hidden sm:inline">{user.name}</span>
            <button
              onClick={logout}
              className="text-sm px-3 py-1.5 rounded-lg font-medium text-white transition-colors hover:opacity-90"
              style={{ backgroundColor: "var(--danger)" }}
            >
              ログアウト
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}
