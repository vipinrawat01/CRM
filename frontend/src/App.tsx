import { useState, useEffect, useCallback, useRef } from "react";
import type { Lead } from "./types";
import { api } from "./api";
import { Sidebar } from "./components/Sidebar";
import { LeadDetail } from "./components/LeadDetail";
import { NewLeadModal } from "./components/NewLeadModal";

export function App() {
  const [leads, setLeads] = useState<Lead[]>([]);
  const [query, setQuery] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [selectedLead, setSelectedLead] = useState<Lead | null>(null);
  const [loadingList, setLoadingList] = useState(true);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [showNewLead, setShowNewLead] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
  const toastTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const notify = useCallback((msg: string) => {
    setToast(msg);
    clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(null), 3500);
  }, []);

  const refreshList = useCallback(
    (q: string, keepSelection?: boolean) => {
      setLoadingList(true);
      api
        .listLeads(q)
        .then((data) => {
          setLeads(data);
          if (!keepSelection && data.length && !selectedId) {
            setSelectedId(data[0].id);
          }
        })
        .catch((e: Error) => notify(e.message))
        .finally(() => setLoadingList(false));
    },
    [selectedId, notify]
  );

  // initial load
  useEffect(() => {
    refreshList("");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // debounce search
  useEffect(() => {
    const t = setTimeout(() => refreshList(query, true), 220);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const loadDetail = useCallback(
    (id: string | null) => {
      if (!id) return;
      setLoadingDetail(true);
      api
        .getLead(id)
        .then(setSelectedLead)
        .catch((e: Error) => notify(e.message))
        .finally(() => setLoadingDetail(false));
    },
    [notify]
  );

  useEffect(() => {
    loadDetail(selectedId);
  }, [selectedId, loadDetail]);

  const onLeadMutated = useCallback((updatedLead: Lead) => {
    setSelectedLead(updatedLead);
    // keep the sidebar's stage badge / name in sync without a full refetch
    setLeads((prev) => prev.map((l) => (l.id === updatedLead.id ? { ...l, ...updatedLead } : l)));
  }, []);

  return (
    <div className="app">
      <Sidebar
        leads={leads}
        loading={loadingList}
        query={query}
        setQuery={setQuery}
        selectedId={selectedId}
        onSelect={setSelectedId}
        onNewLead={() => setShowNewLead(true)}
      />
      <main className="main">
        <div className="main-inner">
          {loadingDetail && <div className="placeholder">Loading…</div>}
          {!loadingDetail && selectedLead && (
            <LeadDetail lead={selectedLead} notify={notify} onMutated={onLeadMutated} />
          )}
          {!loadingDetail && !selectedLead && (
            <div className="placeholder">Select a lead, or create a new one.</div>
          )}
        </div>
      </main>
      {showNewLead && (
        <NewLeadModal
          onClose={() => setShowNewLead(false)}
          onCreated={(lead) => {
            setShowNewLead(false);
            refreshList(query, true);
            setSelectedId(lead.id);
            notify("Lead created");
          }}
        />
      )}
      {toast && <div className="toast">{toast}</div>}
    </div>
  );
}
