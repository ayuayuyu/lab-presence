"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import {
  createDevice,
  createUser,
  deleteDevice,
  getDevices,
  getUsers,
  updateDevice,
} from "@/lib/api";
import { Device, User } from "@/lib/types";

export default function RegisterPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [message, setMessage] = useState({ text: "", isError: false });

  const refresh = useCallback(async () => {
    try {
      const [u, d] = await Promise.all([getUsers(), getDevices()]);
      setUsers(u);
      setDevices(d);
    } catch {
      setMessage({ text: "データの取得に失敗しました", isError: true });
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return (
    <div className="space-y-8">
      <h1 className="text-3xl font-bold">デバイス登録</h1>

      {message.text && (
        <div
          className="p-3 rounded-lg text-sm"
          style={{
            backgroundColor: message.isError
              ? "var(--danger-light)"
              : "var(--success-light)",
            color: message.isError ? "var(--danger)" : "var(--success)",
          }}
        >
          {message.text}
        </div>
      )}

      <div className="grid gap-8 md:grid-cols-2">
        <UserForm
          onCreated={(msg) => {
            setMessage({ text: msg, isError: false });
            refresh();
          }}
          onError={(msg) => setMessage({ text: msg, isError: true })}
        />
        <DeviceForm
          users={users}
          onCreated={(msg) => {
            setMessage({ text: msg, isError: false });
            refresh();
          }}
          onError={(msg) => setMessage({ text: msg, isError: true })}
        />
      </div>

      <RegisteredDevices
        users={users}
        devices={devices}
        onUpdated={(msg) => {
          setMessage({ text: msg, isError: false });
          refresh();
        }}
        onError={(msg) => setMessage({ text: msg, isError: true })}
      />
    </div>
  );
}

function UserForm({
  onCreated,
  onError,
}: {
  onCreated: (msg: string) => void;
  onError: (msg: string) => void;
}) {
  const [name, setName] = useState("");
  const [studentId, setStudentId] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    setSubmitting(true);
    try {
      const user = await createUser(name.trim(), studentId.trim());
      onCreated(`ユーザー「${user.name}」を登録しました`);
      setName("");
      setStudentId("");
    } catch (e) {
      onError(e instanceof Error ? e.message : "登録に失敗しました");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-xl border p-5 space-y-4 card-shadow"
      style={{
        backgroundColor: "var(--card-bg)",
        borderColor: "var(--card-border)",
      }}
    >
      <h2 className="text-lg font-semibold">ユーザー登録</h2>
      <div>
        <label className="block text-sm mb-1" style={{ color: "var(--muted)" }}>
          名前 *
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          className="w-full rounded-lg border px-3 py-2 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]"
        />
      </div>
      <div>
        <label className="block text-sm mb-1" style={{ color: "var(--muted)" }}>
          学籍番号
        </label>
        <input
          type="text"
          value={studentId}
          onChange={(e) => setStudentId(e.target.value)}
          className="w-full rounded-lg border px-3 py-2 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]"
        />
      </div>
      <button
        type="submit"
        disabled={submitting}
        className="w-full rounded-lg py-2 text-sm font-medium text-white transition-colors disabled:opacity-50"
        style={{ backgroundColor: "var(--accent)" }}
      >
        {submitting ? "登録中..." : "ユーザーを登録"}
      </button>
    </form>
  );
}

function DeviceForm({
  users,
  onCreated,
  onError,
}: {
  users: User[];
  onCreated: (msg: string) => void;
  onError: (msg: string) => void;
}) {
  const [userId, setUserId] = useState("");
  const [mac, setMac] = useState("");
  const [label, setLabel] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!userId || !mac.trim()) return;
    setSubmitting(true);
    try {
      await createDevice(Number(userId), mac.trim(), label.trim());
      onCreated(`デバイス「${mac.trim()}」を登録しました`);
      setMac("");
      setLabel("");
    } catch (e) {
      onError(e instanceof Error ? e.message : "登録に失敗しました");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-xl border p-5 space-y-4 card-shadow"
      style={{
        backgroundColor: "var(--card-bg)",
        borderColor: "var(--card-border)",
      }}
    >
      <h2 className="text-lg font-semibold">デバイス登録</h2>
      <div>
        <label className="block text-sm mb-1" style={{ color: "var(--muted)" }}>
          ユーザー *
        </label>
        <select
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
          required
          className="w-full rounded-lg border px-3 py-2 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]"
        >
          <option value="">選択してください</option>
          {users.map((u) => (
            <option key={u.id} value={u.id}>
              {u.name} {u.student_id && `(${u.student_id})`}
            </option>
          ))}
        </select>
      </div>
      <div>
        <label className="block text-sm mb-1" style={{ color: "var(--muted)" }}>
          MACアドレス *
        </label>
        <input
          type="text"
          value={mac}
          onChange={(e) => setMac(e.target.value)}
          placeholder="aa:bb:cc:dd:ee:ff"
          required
          className="w-full rounded-lg border px-3 py-2 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)] font-mono"
        />
      </div>
      <div>
        <label className="block text-sm mb-1" style={{ color: "var(--muted)" }}>
          ラベル（例: MacBook Pro）
        </label>
        <input
          type="text"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          className="w-full rounded-lg border px-3 py-2 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]"
        />
      </div>
      <button
        type="submit"
        disabled={submitting || users.length === 0}
        className="w-full rounded-lg py-2 text-sm font-medium text-white transition-colors disabled:opacity-50"
        style={{ backgroundColor: "var(--accent)" }}
      >
        {submitting ? "登録中..." : "デバイスを登録"}
      </button>
    </form>
  );
}

function RegisteredDevices({
  users,
  devices,
  onUpdated,
  onError,
}: {
  users: User[];
  devices: Device[];
  onUpdated: (msg: string) => void;
  onError: (msg: string) => void;
}) {
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editMac, setEditMac] = useState("");
  const [editLabel, setEditLabel] = useState("");
  const [editUserId, setEditUserId] = useState("");
  const [saving, setSaving] = useState(false);

  if (devices.length === 0) return null;

  const userMap = new Map(users.map((u) => [u.id, u]));

  const startEdit = (d: Device) => {
    setEditingId(d.id);
    setEditMac(d.mac_address);
    setEditLabel(d.label);
    setEditUserId(String(d.user_id));
  };

  const cancelEdit = () => {
    setEditingId(null);
  };

  const saveEdit = async (id: number) => {
    if (!editMac.trim() || !editUserId) return;
    setSaving(true);
    try {
      await updateDevice(id, Number(editUserId), editMac.trim(), editLabel.trim());
      setEditingId(null);
      onUpdated("デバイスを更新しました");
    } catch (e) {
      onError(e instanceof Error ? e.message : "更新に失敗しました");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (d: Device) => {
    if (!window.confirm(`デバイス「${d.mac_address}」を削除しますか？`)) return;
    try {
      await deleteDevice(d.id);
      onUpdated("デバイスを削除しました");
    } catch (e) {
      onError(e instanceof Error ? e.message : "削除に失敗しました");
    }
  };

  return (
    <div>
      <h2 className="text-lg font-semibold mb-3">登録済みデバイス</h2>
      <div className="overflow-x-auto">
        <div
          className="rounded-xl border overflow-hidden card-shadow-static"
          style={{ borderColor: "var(--card-border)" }}
        >
          <table className="w-full text-sm">
            <thead>
              <tr
                className="text-left"
                style={{ backgroundColor: "var(--accent-light)" }}
              >
                <th className="px-4 py-3 font-semibold">ユーザー</th>
                <th className="px-4 py-3 font-semibold">MACアドレス</th>
                <th className="px-4 py-3 font-semibold">ラベル</th>
                <th className="px-4 py-3 font-semibold">操作</th>
              </tr>
            </thead>
            <tbody>
              {devices.map((d, index) => (
                <tr
                  key={d.id}
                  className="border-t border-[var(--card-border)]"
                  style={{
                    backgroundColor:
                      index % 2 === 1 ? "var(--accent-light)" : "var(--card-bg)",
                    opacity: index % 2 === 1 ? 0.8 : undefined,
                  }}
                >
                {editingId === d.id ? (
                  <>
                    <td className="px-4 py-2">
                      <select
                        value={editUserId}
                        onChange={(e) => setEditUserId(e.target.value)}
                        className="w-full rounded-lg border px-2 py-1 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]"
                      >
                        {users.map((u) => (
                          <option key={u.id} value={u.id}>
                            {u.name}
                          </option>
                        ))}
                      </select>
                    </td>
                    <td className="px-4 py-2">
                      <input
                        type="text"
                        value={editMac}
                        onChange={(e) => setEditMac(e.target.value)}
                        className="w-full rounded-lg border px-2 py-1 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)] font-mono"
                      />
                    </td>
                    <td className="px-4 py-2">
                      <input
                        type="text"
                        value={editLabel}
                        onChange={(e) => setEditLabel(e.target.value)}
                        className="w-full rounded-lg border px-2 py-1 text-sm bg-transparent border-[var(--card-border)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]"
                      />
                    </td>
                    <td className="px-4 py-2">
                      <div className="flex gap-1">
                        <button
                          onClick={() => saveEdit(d.id)}
                          disabled={saving}
                          className="rounded-lg px-2.5 py-1 text-xs font-medium text-white transition-colors disabled:opacity-50"
                          style={{ backgroundColor: "var(--accent)" }}
                        >
                          {saving ? "..." : "保存"}
                        </button>
                        <button
                          onClick={cancelEdit}
                          disabled={saving}
                          className="rounded-lg px-2.5 py-1 text-xs font-medium transition-colors border border-[var(--card-border)]"
                        >
                          取消
                        </button>
                      </div>
                    </td>
                  </>
                ) : (
                  <>
                    <td className="px-4 py-3">
                      {userMap.get(d.user_id)?.name ?? `ID:${d.user_id}`}
                    </td>
                    <td className="px-4 py-3 font-mono">{d.mac_address}</td>
                    <td className="px-4 py-3" style={{ color: d.label ? undefined : "var(--muted)" }}>
                      {d.label || "-"}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-1">
                        <button
                          onClick={() => startEdit(d)}
                          className="rounded-lg px-2.5 py-1 text-xs font-medium transition-colors border border-[var(--card-border)] hover:bg-[var(--card-bg)]"
                        >
                          編集
                        </button>
                        <button
                          onClick={() => handleDelete(d)}
                          className="rounded-lg px-2.5 py-1 text-xs font-medium transition-colors border"
                          style={{
                            color: "var(--danger)",
                            borderColor: "var(--danger)",
                          }}
                        >
                          削除
                        </button>
                      </div>
                    </td>
                  </>
                )}
              </tr>
            ))}
          </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
