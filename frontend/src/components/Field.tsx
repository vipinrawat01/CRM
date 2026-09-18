interface Props {
  label: string;
  value?: string | null;
}

export function Field({ label, value }: Props) {
  return (
    <div className="field">
      <label>{label}</label>
      <div className={"value" + (value ? "" : " empty")}>{value || "not provided"}</div>
    </div>
  );
}
