import type { ReactNode } from 'react';

/**
 * A pane's top strip: on macOS an empty drag area as tall as the unified toolbar, and on every
 * platform a 40 px row ruled off below while it holds children.
 */
export function TitleBar({ children }: { children?: ReactNode }) {
  if (!children) {
    return (
      <div className="hidden h-13 shrink-0 in-data-[platform=darwin]:block [--wails-draggable:drag]" />
    );
  }
  return (
    <div className="flex h-10 shrink-0 items-stretch border-b in-data-[platform=darwin]:[--wails-draggable:drag]">
      {children}
    </div>
  );
}
