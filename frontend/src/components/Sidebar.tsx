import type { Lead } from "../types";
import { fmtMoney } from "../api";

interface Props {
  leads: Lead[];
  loading: boolean;
  query: string;
  setQuery: (q: string) => void;
  selectedId: string | null;
  onSelect: (id: string) => void;
  onNewLead: () => void;
}

export function Sidebar({ leads, loading, query, setQuery, selectedId, onSelect, onNewLead }: Props) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <h1>Northstar CRM</h1>
        <p>B2B sales, lead tracking &amp; AI summaries</p>
      </div>
      <div className="sidebar-controls">
        <input
          className="search-input"
          placeholder="Search leads or companies…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <button className="btn btn-primary" onClick={onNewLead}>
          + New
        </button>
      </div>
      <div className="lead-list">
        {loading && <div className="empty-note">Loading leads…</div>}
        {!loading && leads.length === 0 && <div className="empty-note">No leads match.</div>}
        {!loading &&
          leads.map((l) => (
            <div
              key={l.id}
              className={"lead-row" + (l.id === selectedId ? " active" : "")}
              onClick={() => onSelect(l.id)}
            >
              <div className="name">{l.name}</div>
              <div className="company">{l.company}</div>
              <div className="meta">
                <span className={"badge badge-" + l.stage}>{l.stage}</span>
                {!!l.dealValue && (
                  <span className="mono" style={{ fontSize: 11.5, color: "var(--ink-faint)" }}>
                    {fmtMoney(l.dealValue)}
                  </span>
                )}
              </div>
            </div>
          ))}
      </div>
    </aside>
  );
}
