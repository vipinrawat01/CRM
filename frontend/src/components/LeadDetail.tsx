import { useState } from "react";
import type { Lead, Stage } from "../types";
import { STAGES } from "../types";
import { api, fmtDate, fmtMoney } from "../api";
import { Field } from "./Field";
import { SummarySection } from "./SummarySection";
import { TasksSection } from "./TasksSection";
import { NotesSection } from "./NotesSection";

interface Props {
  lead: Lead;
  notify: (msg: string) => void;
  onMutated: (lead: Lead) => void;
}

export function LeadDetail({ lead, notify, onMutated }: Props) {
  const [changingStage, setChangingStage] = useState(false);

  const changeStage = (stage: Stage) => {
    if (stage === lead.stage) return;
    setChangingStage(true);
    api
      .setStage(lead.id, stage)
      .then(onMutated)
      .catch((e: Error) => notify(e.message))
      .finally(() => setChangingStage(false));
  };

  return (
    <div>
      <div className="detail-header">
        <div>
          <h2>{lead.name}</h2>
          <p className="detail-sub">
            {lead.title ? lead.title + " · " : ""}
            {lead.company}
          </p>
        </div>
        {!!lead.dealValue && (
          <div className="mono" style={{ fontSize: 18, fontWeight: 600 }}>
            {fmtMoney(lead.dealValue)}
          </div>
        )}
      </div>

      <div className="stage-picker">
        {STAGES.map((s) => (
          <button
            key={s}
            className={"stage-chip" + (s === lead.stage ? " current" : "")}
            disabled={changingStage}
            onClick={() => changeStage(s)}
          >
            {s}
          </button>
        ))}
      </div>

      <div className="field-grid">
        <Field label="Email" value={lead.email} />
        <Field label="Phone" value={lead.phone} />
        <Field label="Source" value={lead.source} />
        <Field label="Lead since" value={fmtDate(lead.createdAt)} />
      </div>

      <SummarySection lead={lead} notify={notify} onMutated={onMutated} />
      <TasksSection lead={lead} notify={notify} onMutated={onMutated} />
      <NotesSection lead={lead} notify={notify} onMutated={onMutated} />
    </div>
  );
}
