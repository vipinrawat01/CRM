import { useState, type FormEvent } from "react";
import type { Lead } from "../types";
import { api } from "../api";

interface Props {
  onClose: () => void;
  onCreated: (lead: Lead) => void;
}

const emptyForm = {
  name: "",
  title: "",
  company: "",
  email: "",
  phone: "",
  source: "",
  dealValue: "",
};

export function NewLeadModal({ onClose, onCreated }: Props) {
  const [form, setForm] = useState(emptyForm);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const set = (k: keyof typeof emptyForm) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (!form.name.trim() || !form.company.trim()) {
      setError("Name and company are required.");
      return;
    }
    setSubmitting(true);
    setError(null);
    api
      .createLead({
        ...form,
        dealValue: form.dealValue ? parseInt(form.dealValue, 10) : 0,
      })
      .then(onCreated)
      .catch((e: Error) => setError(e.message))
      .finally(() => setSubmitting(false));
  };

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h3>New lead</h3>
        <form onSubmit={submit}>
          <div className="form-row">
            <label>Name *</label>
            <input value={form.name} onChange={set("name")} autoFocus />
          </div>
          <div className="form-row">
            <label>Company *</label>
            <input value={form.company} onChange={set("company")} />
          </div>
          <div className="form-row">
            <label>Title</label>
            <input value={form.title} onChange={set("title")} />
          </div>
          <div className="form-row">
            <label>Email</label>
            <input value={form.email} onChange={set("email")} />
          </div>
          <div className="form-row">
            <label>Phone</label>
            <input value={form.phone} onChange={set("phone")} />
          </div>
          <div className="form-row">
            <label>Source</label>
            <input value={form.source} onChange={set("source")} placeholder="e.g. Referral, Website form" />
          </div>
          <div className="form-row">
            <label>Estimated deal value</label>
            <input value={form.dealValue} onChange={set("dealValue")} inputMode="numeric" />
          </div>
          {error && <div className="form-error">{error}</div>}
          <div className="form-actions">
            <button type="button" className="btn" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? "Creating…" : "Create lead"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
