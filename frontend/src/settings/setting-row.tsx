import type { ReactNode } from 'react';

/** One setting: its name, and below it a hint when given, on the left; its control on the right. */
export function SettingRow({
  labelId,
  label,
  hint,
  children,
}: {
  labelId: string;
  label: string;
  hint?: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-6 px-4 py-3">
      <div className="flex min-w-0 flex-col gap-0.5">
        <span id={labelId} className="text-sm">
          {label}
        </span>
        {hint && <span className="text-xs text-muted-foreground">{hint}</span>}
      </div>
      {children}
    </div>
  );
}
