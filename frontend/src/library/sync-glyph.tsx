import { CircleAlertIcon, RefreshCwIcon } from 'lucide-react';
import { type State, Status } from '#bindings/hostsource/models.ts';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { m } from '@/paraglide/messages.js';
import { sourceStatusText, troubled } from './source-status.ts';

/**
 * A synced folder's sync state beside its name in the tree, worded for tooltips and readers. While
 * the folder is not connecting, it is a button that reads the source again.
 */
export function SyncGlyph({
  kind,
  state,
  onResync,
}: {
  kind: string;
  state: State | undefined;
  onResync: () => void;
}) {
  const text = sourceStatusText(kind, state);
  if (!troubled(state) && state?.status !== Status.Synced) {
    return (
      <span title={text} className="ml-auto flex size-6 shrink-0 items-center justify-center">
        <Spinner aria-hidden className="size-3.5 text-muted-foreground" />
        <span className="sr-only">{text}</span>
      </span>
    );
  }
  return (
    <Button
      variant="ghost"
      size="icon-xs"
      title={`${text} · ${m.source_resync()}`}
      aria-label={`${m.source_resync()}: ${text}`}
      className="ml-auto"
      // The row would take the click as selecting or opening the folder.
      onClick={(event) => {
        event.stopPropagation();
        onResync();
      }}
    >
      {troubled(state) ? (
        <CircleAlertIcon aria-hidden className="size-3.5 text-destructive" />
      ) : (
        <RefreshCwIcon aria-hidden className="size-3.5 text-muted-foreground" />
      )}
    </Button>
  );
}
