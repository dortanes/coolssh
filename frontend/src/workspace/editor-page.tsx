import type { ReactNode } from 'react';

/** The column an editor sits in, under the item's name as its heading. */
export function EditorPage({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section aria-label={title} className="mx-auto w-full max-w-2xl px-8 pb-10">
      <h1 className="mb-4 truncate text-lg font-semibold">{title}</h1>
      {children}
    </section>
  );
}
