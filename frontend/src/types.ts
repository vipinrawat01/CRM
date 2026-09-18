export type Stage = "New" | "Contacted" | "Qualified" | "Proposal" | "Won" | "Lost";

export const STAGES: Stage[] = ["New", "Contacted", "Qualified", "Proposal", "Won", "Lost"];

export type NoteKind = "note" | "call" | "email" | "meeting";

export interface Note {
  id: string;
  kind: NoteKind;
  content: string;
  createdAt: string;
}

export interface Task {
  id: string;
  title: string;
  dueDate?: string | null;
  done: boolean;
}

export interface Summary {
  who: string;
  whatMatters: string;
  whatHappened: string;
  missingInfo: string[];
  generatedAt: string;
  source: "model" | "fallback";
  fallbackNote?: string;
}

export interface Lead {
  id: string;
  name: string;
  title?: string;
  company: string;
  email?: string;
  phone?: string;
  source?: string;
  dealValue?: number;
  stage: Stage;
  createdAt: string;
  notes: Note[];
  tasks: Task[];
  summary?: Summary | null;
}

export interface NewLeadInput {
  name: string;
  title: string;
  company: string;
  email: string;
  phone: string;
  source: string;
  dealValue: number;
}
