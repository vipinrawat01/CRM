import type { Lead, NewLeadInput, NoteKind } from "./types";

// Every function does exactly one HTTP call and throws a plain Error with a
// message pulled from the API's { error } JSON body, so callers can show it
// directly instead of a generic "something went wrong".
async function handle<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let msg = `Request failed (${res.status})`;
    try {
      const body = await res.json();
      if (body.error) msg = body.error;
    } catch {
      // response body wasn't JSON; keep the generic message
    }
    throw new Error(msg);
  }
  if (res.status === 204) return null as T;
  return res.json();
}

export const api = {
  listLeads(q: string): Promise<Lead[]> {
    const url = q ? `/api/leads?q=${encodeURIComponent(q)}` : "/api/leads";
    return fetch(url).then((r) => handle<Lead[]>(r));
  },
  getLead(id: string): Promise<Lead> {
    return fetch(`/api/leads/${id}`).then((r) => handle<Lead>(r));
  },
  createLead(payload: NewLeadInput): Promise<Lead> {
    return fetch("/api/leads", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }).then((r) => handle<Lead>(r));
  },
  setStage(id: string, stage: string): Promise<Lead> {
    return fetch(`/api/leads/${id}/stage`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ stage }),
    }).then((r) => handle<Lead>(r));
  },
  addNote(id: string, kind: NoteKind, content: string): Promise<Lead> {
    return fetch(`/api/leads/${id}/notes`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ kind, content }),
    }).then((r) => handle<Lead>(r));
  },
  addTask(id: string, title: string, dueDate: string): Promise<Lead> {
    return fetch(`/api/leads/${id}/tasks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, dueDate: dueDate || null }),
    }).then((r) => handle<Lead>(r));
  },
  setTaskDone(id: string, taskId: string, done: boolean): Promise<Lead> {
    return fetch(`/api/leads/${id}/tasks/${taskId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ done }),
    }).then((r) => handle<Lead>(r));
  },
  summarize(id: string): Promise<Lead> {
    return fetch(`/api/leads/${id}/summarize`, { method: "POST" }).then((r) => handle<Lead>(r));
  },
};

export function fmtDate(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

export function fmtMoney(n?: number): string {
  if (!n) return "";
  return "₹" + n.toLocaleString("en-IN");
}
