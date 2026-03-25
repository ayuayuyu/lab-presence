"use client";

import { useState, useRef, useEffect } from "react";
import { useAuth } from "@/lib/auth";
import { updateMyName } from "@/lib/api";
import Link from "next/link";
import { usePathname } from "next/navigation";

export default function Header() {
  const { user, setUser, isAdmin, logout } = useAuth();
  const pathname = usePathname();
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState("");
  const [saving, setSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editing]);

  if (!user) return null;

  const navLinks = [
    { href: "/", label: "ダッシュボード" },
    { href: "/register", label: "デバイス登録" },
  ];

  const startEdit = () => {
    setEditName(user.name);
    setEditing(true);
  };

  const cancelEdit = () => {
    setEditing(false);
  };

  const saveEdit = async () => {
    const trimmed = editName.trim();
    if (!trimmed || trimmed === user.name) {
      setEditing(false);
      return;
    }
    setSaving(true);
    try {
      const updated = await updateMyName(trimmed);
      setUser({ ...user, name: updated.name });
      setEditing(false);
    } catch {
      // エラー時は編集状態を維持
    } finally {
      setSaving(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      saveEdit();
    } else if (e.key === "Escape") {
      cancelEdit();
    }
  };

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
                className="w-7 h-7 rounded-full cursor-pointer"
                referrerPolicy="no-referrer"
                onClick={startEdit}
                title="クリックして表示名を変更"
              />
            ) : (
              <div
                className="w-7 h-7 rounded-full flex items-center justify-center text-white text-xs font-bold cursor-pointer"
                style={{ backgroundColor: "var(--accent)" }}
                onClick={startEdit}
                title="クリックして表示名を変更"
              >
                {user.name?.charAt(0) || "?"}
              </div>
            )}
            {editing ? (
              <div className="flex items-center gap-1">
                <input
                  ref={inputRef}
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  onKeyDown={handleKeyDown}
                  onBlur={cancelEdit}
                  disabled={saving}
                  className="text-sm px-2 py-1 rounded-lg border border-[var(--card-border)] bg-transparent focus:outline-none focus:ring-2 focus:ring-[var(--accent)] w-32"
                />
              </div>
            ) : (
              <button
                onClick={startEdit}
                className="text-sm hidden sm:inline hover:underline cursor-pointer"
                title="クリックして表示名を変更"
              >
                {user.name}
              </button>
            )}
            {isAdmin && (
              <span
                className="text-xs font-bold px-1.5 py-0.5 rounded"
                style={{ backgroundColor: "var(--accent-light)", color: "var(--accent)" }}
              >
                管理者
              </span>
            )}
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
