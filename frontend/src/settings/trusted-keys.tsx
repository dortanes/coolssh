import { ServerIcon } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { m } from '@/paraglide/messages.js';
import { keyKind } from '../library/key-kind.ts';
import { useStore } from '../use-store.ts';
import type { TrustedKeysStore } from './trusted-keys-store.ts';

/** The host keys trusted in the app, each forgettable so the next connection asks again. */
export function TrustedKeys({ store }: { store: TrustedKeysStore }) {
  const { trusted, problem } = useStore(store);
  return (
    <section aria-labelledby="trusted-keys-title" className="mt-8">
      <h2 id="trusted-keys-title" className="text-sm font-semibold">
        {m.keys_trusted_title()}
      </h2>
      <p className="mt-1 text-sm text-muted-foreground">{m.keys_trusted_hint()}</p>
      {trusted.length === 0 ? (
        <p className="mt-4 text-sm text-muted-foreground">{m.keys_trusted_empty()}</p>
      ) : (
        <ul data-surface="group" className="mt-4 divide-y rounded-lg border bg-card">
          {trusted.map((key) => (
            <li
              key={`${key.host} ${key.fingerprint}`}
              className="flex items-center gap-3 px-4 py-3"
            >
              <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-accent text-muted-foreground">
                <ServerIcon aria-hidden className="size-4" />
              </span>
              <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                <span className="flex items-baseline gap-2">
                  <span className="truncate text-sm font-medium select-text">{key.host}</span>
                  <span className="shrink-0 text-xs text-muted-foreground">
                    {keyKind(key.type)}
                  </span>
                </span>
                <span
                  className="truncate font-mono text-[0.6875rem] text-muted-foreground select-text"
                  title={key.fingerprint}
                >
                  {key.fingerprint}
                </span>
              </div>
              <Button variant="secondary" size="sm" onClick={() => void store.forget(key)}>
                {m.keys_trusted_forget()}
              </Button>
            </li>
          ))}
        </ul>
      )}
      {problem && (
        <p role="alert" className="mt-3 text-sm text-destructive">
          {problem}
        </p>
      )}
    </section>
  );
}
