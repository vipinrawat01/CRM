import { useState, type FormEvent } from "react";
import type { Lead, Task } from "../types";
import { api, fmtDate } from "../api";

interface Props {
  lead: Lead;
  notify: (msg: string) => void;
  onMutated: (lead: Lead) => void;
}

export function TasksSection({ lead, notify, onMutated }: Props) {
  const [title, setTitle] = useState("");
  const [due, setDue] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    setSubmitting(true);
    api
      .addTask(lead.id, title.trim(), due)
      .then((l) => {
        onMutated(l);
        setTitle("");
        setDue("");
      })
      .catch((e: Error) => notify(e.message))
      .finally(() => setSubmitting(false));
  };

  const toggle = (task: Task) => {
    api
      .setTaskDone(lead.id, task.id, !task.done)
      .then(onMutated)
      .catch((e: Error) => notify(e.message));
  };

  return (
    <div className="section">
      <h3>Follow-up tasks</h3>
      {lead.tasks.length === 0 && <div style={{ color: "var(--ink-faint)" }}>No tasks yet.</div>}
      {lead.tasks.map((t) => (
        <div className="task-row" key={t.id}>
          <input type="checkbox" checked={t.done} onChange={() => toggle(t)} />
          <span className={"title" + (t.done ? " done" : "")}>{t.title}</span>
          {t.dueDate && <span className="due mono">due {fmtDate(t.dueDate)}</span>}
        </div>
      ))}
      <form className="inline-form" onSubmit={submit}>
        <input
          type="text"
          placeholder="New follow-up task…"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
        <input type="date" value={due} onChange={(e) => setDue(e.target.value)} />
        <button className="btn btn-sm" disabled={submitting}>
          Add
        </button>
      </form>
    </div>
  );
}
