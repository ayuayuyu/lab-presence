import { Device, LastSeen, Presence, User } from "./types";
import { getToken } from "./auth";

// Nginx経由のとき（本番）は空文字 → 相対パスで /api/* にアクセス
// ローカル開発時は NEXT_PUBLIC_API_URL=http://localhost:8080 を設定
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

function authHeaders(): Record<string, string> {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function handleUnauthorized(res: Response) {
  if (res.status === 401) {
    localStorage.removeItem("lab_id_token");
    window.location.reload();
  }
}

async function fetcher<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      ...authHeaders(),
      ...init?.headers,
    },
  });
  if (!res.ok) {
    handleUnauthorized(res);
    const text = await res.text();
    throw new Error(`API error ${res.status}: ${text}`);
  }
  return res.json();
}

// 在室者一覧
export function getPresence(): Promise<Presence[]> {
  return fetcher("/api/presence");
}

// 最終来室記録
export function getLastSeen(): Promise<LastSeen[]> {
  return fetcher("/api/presence/last-seen");
}

// ユーザー一覧
export function getUsers(): Promise<User[]> {
  return fetcher("/api/users");
}

// ユーザー作成
export function createUser(name: string, studentId: string): Promise<User> {
  return fetcher("/api/users", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, student_id: studentId }),
  });
}

// デバイス一覧
export function getDevices(): Promise<Device[]> {
  return fetcher("/api/devices");
}

// デバイス登録
export function createDevice(
  userId: number,
  macAddress: string,
  label: string
): Promise<Device> {
  return fetcher("/api/devices", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userId, mac_address: macAddress, label }),
  });
}

// デバイス更新
export function updateDevice(
  id: number,
  userId: number,
  macAddress: string,
  label: string
): Promise<Device> {
  return fetcher(`/api/devices/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userId, mac_address: macAddress, label }),
  });
}

// デバイス削除
export function deleteDevice(id: number): Promise<void> {
  return fetcher(`/api/devices/${id}`, { method: "DELETE" });
}
