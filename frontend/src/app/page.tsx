"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { getPresence, getLastSeen } from "@/lib/api";
import { LastSeen, Presence } from "@/lib/types";

const FALLBACK_POLL_INTERVAL = 60_000; // WebSocket未接続時のフォールバック間隔

function getWsUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const base = process.env.NEXT_PUBLIC_API_URL || window.location.origin;
  const url = new URL(base);
  return `${protocol}//${url.host}/ws/presence`;
}

export default function Dashboard() {
  const [members, setMembers] = useState<Presence[]>([]);
  const [lastSeenList, setLastSeenList] = useState<LastSeen[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchPresence = useCallback(async () => {
    try {
      const [presenceData, lastSeenData] = await Promise.all([
        getPresence(),
        getLastSeen(),
      ]);
      setMembers(presenceData);
      setLastSeenList(lastSeenData);
      setLastUpdated(new Date());
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "取得に失敗しました");
    } finally {
      setLoading(false);
    }
  }, []);

  const stopPolling = useCallback(() => {
    if (pollTimerRef.current) {
      clearInterval(pollTimerRef.current);
      pollTimerRef.current = null;
    }
  }, []);

  const startPolling = useCallback(() => {
    stopPolling();
    pollTimerRef.current = setInterval(fetchPresence, FALLBACK_POLL_INTERVAL);
  }, [fetchPresence, stopPolling]);

  const connectWs = useCallback(() => {
    // 既存接続があれば閉じる
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    const ws = new WebSocket(getWsUrl());
    wsRef.current = ws;

    ws.onopen = () => {
      // WebSocket接続中はポーリング停止
      stopPolling();
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setMembers(data.presence ?? []);
        setLastSeenList(data.last_seen ?? []);
        setLastUpdated(new Date());
        setError("");
        setLoading(false);
      } catch {
        // JSON解析エラーは無視
      }
    };

    ws.onclose = () => {
      wsRef.current = null;
      // タブが表示中の場合のみフォールバックポーリングを開始
      if (document.visibilityState === "visible") {
        startPolling();
      }
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [stopPolling, startPolling]);

  const disconnectWs = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    stopPolling();
  }, [stopPolling]);

  useEffect(() => {
    // 初回データ取得
    fetchPresence();

    // WebSocket接続
    connectWs();

    // visibilitychange: タブ非表示時はWS切断＆ポーリング停止、表示時に再接続
    const handleVisibility = () => {
      if (document.visibilityState === "visible") {
        fetchPresence();
        connectWs();
      } else {
        disconnectWs();
      }
    };

    document.addEventListener("visibilitychange", handleVisibility);

    return () => {
      document.removeEventListener("visibilitychange", handleVisibility);
      disconnectWs();
    };
  }, [fetchPresence, connectWs, disconnectWs]);

  return (
    <div>
      {/* Header: 2-line layout */}
      <div className="mb-6">
        <div className="flex items-center gap-3 mb-1">
          <h1 className="text-3xl font-bold">在室状況</h1>
          {!loading && (
            <span
              className="text-xs font-semibold px-2.5 py-1 rounded-full"
              style={{
                backgroundColor:
                  members.length > 0
                    ? "var(--success-light)"
                    : "var(--accent-light)",
                color:
                  members.length > 0 ? "var(--success)" : "var(--accent)",
              }}
            >
              {members.length}人
            </span>
          )}
        </div>
        <div className="flex items-center gap-3">
          {lastUpdated && (
            <span className="text-xs" style={{ color: "var(--muted)" }}>
              最終更新: {lastUpdated.toLocaleTimeString("ja-JP")}
            </span>
          )}
          <button
            onClick={fetchPresence}
            className="text-sm px-3 py-1.5 rounded-lg border border-[var(--card-border)] hover:bg-[var(--accent-light)] transition-colors"
          >
            更新
          </button>
        </div>
      </div>

      {error && (
        <div
          className="mb-4 p-3 rounded-lg text-sm"
          style={{
            backgroundColor: "var(--danger-light)",
            color: "var(--danger)",
          }}
        >
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-16" style={{ color: "var(--muted)" }}>
          <div className="inline-block w-6 h-6 border-2 border-current border-t-transparent rounded-full animate-spin mb-3" />
          <p>読み込み中...</p>
        </div>
      ) : members.length === 0 ? (
        <div
          className="text-center py-16 rounded-xl border border-dashed border-[var(--card-border)]"
          style={{ color: "var(--muted)" }}
        >
          <p className="text-4xl mb-3">🏠</p>
          <p className="text-lg mb-1">今は誰もいないようです</p>
          <p className="text-sm">
            メンバーが来室すると、ここにカードが表示されます
          </p>
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {members.map((m) => (
            <MemberCard key={m.user_id} member={m} />
          ))}
        </div>
      )}

      {!loading && lastSeenList.length > 0 && (
        <LastSeenSection
          lastSeenList={lastSeenList}
          presentUserIds={new Set(members.map((m) => m.user_id))}
        />
      )}
    </div>
  );
}

function LastSeenSection({
  lastSeenList,
  presentUserIds,
}: {
  lastSeenList: LastSeen[];
  presentUserIds: Set<number>;
}) {
  return (
    <div className="mt-8">
      <h2 className="text-lg font-bold mb-4">最終来室記録</h2>
      <div
        className="rounded-xl border overflow-hidden card-shadow-static"
        style={{
          backgroundColor: "var(--card-bg)",
          borderColor: "var(--card-border)",
        }}
      >
        <table className="w-full text-sm">
          <thead>
            <tr
              className="border-b text-left"
              style={{
                borderColor: "var(--card-border)",
                backgroundColor: "var(--accent-light)",
              }}
            >
              <th className="px-4 py-3 font-semibold">ユーザー</th>
              <th className="px-4 py-3 font-semibold">最終来室日</th>
            </tr>
          </thead>
          <tbody>
            {lastSeenList.map((ls, index) => {
              const isPresent = presentUserIds.has(ls.user_id);
              const date = new Date(ls.last_seen_at);
              const daysAgo = formatDaysAgo(date);
              return (
                <tr
                  key={ls.user_id}
                  className="border-b last:border-b-0"
                  style={{
                    borderColor: "var(--card-border)",
                    backgroundColor:
                      index % 2 === 1 ? "var(--accent-light)" : undefined,
                  }}
                >
                  <td className="px-4 py-3 flex items-center gap-2">
                    {ls.user_picture ? (
                      <img
                        src={ls.user_picture}
                        alt={ls.user_name}
                        className="w-6 h-6 rounded-full shrink-0"
                        referrerPolicy="no-referrer"
                      />
                    ) : (
                      <div
                        className="w-6 h-6 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0"
                        style={{ backgroundColor: "var(--accent)" }}
                      >
                        {ls.user_name.charAt(0)}
                      </div>
                    )}
                    {ls.user_name}
                    {isPresent && (
                      <span
                        className="text-xs font-bold px-2.5 py-0.5 rounded-full"
                        style={{
                          backgroundColor: "var(--success)",
                          color: "white",
                        }}
                      >
                        在室中
                      </span>
                    )}
                  </td>
                  <td
                    className="px-4 py-3"
                    style={{ color: "var(--muted)" }}
                  >
                    {date.getMonth() + 1}月{date.getDate()}日 {date.getHours()}:{String(date.getMinutes()).padStart(2, '0')}
                    <span className="ml-2 text-xs">({daysAgo})</span>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function MemberCard({ member }: { member: Presence }) {
  const detectedAt = new Date(member.detected_at);
  const ago = formatRelative(detectedAt);

  return (
    <div
      className="rounded-xl border p-5 flex items-center gap-4 card-shadow"
      style={{
        backgroundColor: "var(--card-bg)",
        borderColor: "var(--card-border)",
      }}
    >
      {member.user_picture ? (
        <img
          src={member.user_picture}
          alt={member.user_name}
          className="w-12 h-12 rounded-full shrink-0 object-cover"
          referrerPolicy="no-referrer"
        />
      ) : (
        <div
          className="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-base shrink-0"
          style={{ backgroundColor: "var(--success)" }}
        >
          {member.user_name.charAt(0)}
        </div>
      )}
      <div className="min-w-0">
        <p className="font-semibold text-base truncate">{member.user_name}</p>
        <p className="text-sm truncate" style={{ color: "var(--muted)" }}>
          {member.device_label || member.mac_address}
        </p>
        <p className="text-xs mt-0.5" style={{ color: "var(--muted)" }}>
          {ago}
        </p>
      </div>
    </div>
  );
}

function formatRelative(date: Date): string {
  const diff = Math.floor((Date.now() - date.getTime()) / 1000);
  if (diff < 60) return `${diff}秒前`;
  if (diff < 3600) return `${Math.floor(diff / 60)}分前`;
  return `${Math.floor(diff / 3600)}時間前`;
}

function formatDaysAgo(date: Date): string {
  const now = new Date();
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const dateStart = new Date(date.getFullYear(), date.getMonth(), date.getDate());
  const diffDays = Math.floor(
    (todayStart.getTime() - dateStart.getTime()) / (1000 * 60 * 60 * 24)
  );
  if (diffDays === 0) return "今日";
  if (diffDays === 1) return "昨日";
  return `${diffDays}日前`;
}
