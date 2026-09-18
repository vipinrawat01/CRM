import { useState, type FormEvent } from "react";
import type { Lead, NoteKind } from "../types";
import { api, fmtDate } from "../api";

interface Props {
  lead: Lead;
  notify: (msg: string) => void;
  onMutated: (lead: Lead) => void;
}

export function NotesSection({ lead, notify, onMutated }: Props) {
  const [kind, setKind] = useState<NoteKind>("note");
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;
    setSubmitting(true);
    api
      .addNote(lead.id, kind, content.trim())
      .then((l) => {
        onMutated(l);
        setContent("");
      })
      .catch((e: Error) => notify(e.message))
      .finally(() => setSubmitting(false));
  };

  return (
    <div className="section">
      <h3>Notes &amp; activity</h3>
      <form className="inline-form" style={{ marginBottom: 14 }} onSubmit={submit}>
        <select value={kind} onChange={(e) => setKind(e.target.value as NoteKind)}>
          <option value="note">Note</option>
          <option value="call">Call</option>
          <option value="email">Email</option>
          <option value="meeting">Meeting</option>
        </select>
        <textarea
          placeholder="Log a note or activity…"
          value={content}
          onChange={(e) => setContent(e.target.value)}
        />
        <button className="btn btn-sm" disabled={submitting}>
          Log
        </button>
      </form>
      {lead.notes.length === 0 && <div style={{ color: "var(--ink-faint)" }}>No activity logged yet.</div>}
      {lead.notes.map((n) => (
        <div className="timeline-item" key={n.id}>
          <span className="kind">{n.kind}</span>
          <span className="date">{fmtDate(n.createdAt)}</span>
          <div className="content">{n.content}</div>
        </div>
      ))}
    </div>
  );
}
