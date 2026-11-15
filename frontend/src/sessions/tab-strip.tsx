import { cn } from 'cn';
import { Columns2Icon, XIcon } from 'lucide-react';
import { useEffect, useRef } from 'react';
import { type State, Status } from '#bindings/session/models.ts';
import { Button } from '@/components/ui/button';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { Spinner } from '@/components/ui/spinner';
import { m } from '@/paraglide/messages.js';
import type { Orientation } from './pane-tree.ts';
import { SplitItems } from './split-items.tsx';

/** A tab: the sessions in its panes, in reading order, under the key of its first. */
export type Tab = { key: string; sessions: readonly State[] };

type TabStripProps = {
  tabs: readonly Tab[];
  /** The active tab's key, undefined while the workspace shows something else. */
  active: string | undefined;
  onShow: (key: string) => void;
  /** Ends the given sessions, closing their panes. */
  onClose: (ids: string[]) => void;
  /** Splits a tab's focused pane, asking what opens beside it. */
  onSplit: (key: string, orientation: Orientation) => void;
};

const statusText: Record<Status, () => string> = {
  [Status.$zero]: () => '',
  [Status.Connecting]: m.session_status_connecting,
  [Status.Connected]: m.session_status_connected,
  [Status.Closed]: m.session_status_closed,
};

/** A tab's state: connecting while any session is, closed once all are, else connected. */
function tabStatus(sessions: readonly State[]): Status {
  if (sessions.some((session) => session.status === Status.Connecting)) {
    return Status.Connecting;
  }
  return sessions.every((session) => session.status === Status.Closed)
    ? Status.Closed
    : Status.Connected;
}

/**
 * A tab per group of panes, named after its hosts, with its state and a close button; tabs share the
 * width evenly down to a minimum, then the strip scrolls without a scrollbar.
 */
export function TabStrip({ tabs, active, onShow, onClose, onSplit }: TabStripProps) {
  const strip = useRef<HTMLDivElement>(null);
  // The active tab, a new session's among them, scrolls into view.
  useEffect(() => {
    if (active) {
      strip.current
        ?.querySelector('[data-active]')
        ?.scrollIntoView({ inline: 'nearest', block: 'nearest', behavior: 'smooth' });
    }
  }, [active]);
  return (
    <div
      ref={strip}
      role="tablist"
      aria-label={m.session_tabs()}
      data-surface="tabs"
      className="flex min-w-0 flex-1 items-stretch overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      // A mouse wheel turns vertically; the strip only scrolls sideways.
      onWheel={(event) => {
        if (Math.abs(event.deltaY) > Math.abs(event.deltaX)) {
          event.currentTarget.scrollLeft += event.deltaY;
        }
      }}
    >
      {tabs.map((tab, index) => {
        const ids = (from: number, to: number) =>
          tabs.slice(from, to).flatMap((each) => each.sessions.map((session) => session.id));
        const own = ids(index, index + 1);
        const selected = tab.key === active;
        const status = tabStatus(tab.sessions);
        const title = tab.sessions.map((session) => session.title).join(', ');
        return (
          <ContextMenu key={tab.key}>
            <ContextMenuTrigger
              render={<div />}
              data-active={selected || undefined}
              className={cn(
                'group flex min-w-32 flex-1 basis-0 items-center border-r pr-1.5 text-sm last:border-r-0 [--wails-draggable:no-drag]',
                selected
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground',
              )}
            >
              <button
                type="button"
                role="tab"
                aria-selected={selected}
                title={`${title} · ${statusText[status]()}`}
                className="flex min-w-0 flex-1 items-center gap-1.5 self-stretch truncate pl-3 text-left outline-none focus-visible:underline"
                onClick={() => onShow(tab.key)}
              >
                {status === Status.Connecting && <Spinner aria-hidden className="size-3" />}
                {tab.sessions.length > 1 && (
                  <Columns2Icon aria-hidden className="size-3.5 shrink-0 opacity-70" />
                )}
                <span className={cn('truncate', status === Status.Closed && 'line-through')}>
                  {title}
                </span>
                <span className="sr-only">{statusText[status]()}</span>
              </button>
              <Button
                variant="ghost"
                size="icon-xs"
                className={cn(
                  'opacity-0 group-hover:opacity-100 focus-visible:opacity-100',
                  selected && 'opacity-100',
                )}
                aria-label={m.session_close()}
                title={m.session_close()}
                onClick={() => onClose(own)}
              >
                <XIcon aria-hidden />
              </Button>
            </ContextMenuTrigger>
            <ContextMenuContent data-surface="menu">
              <SplitItems onSplit={(orientation) => onSplit(tab.key, orientation)} />
              <ContextMenuSeparator />
              <ContextMenuItem onClick={() => onClose(own)}>{m.session_close()}</ContextMenuItem>
              <ContextMenuItem disabled={index === 0} onClick={() => onClose(ids(0, index))}>
                {m.session_close_left()}
              </ContextMenuItem>
              <ContextMenuItem
                disabled={index === tabs.length - 1}
                onClick={() => onClose(ids(index + 1, tabs.length))}
              >
                {m.session_close_right()}
              </ContextMenuItem>
              <ContextMenuItem onClick={() => onClose(ids(0, tabs.length))}>
                {m.session_close_all()}
              </ContextMenuItem>
            </ContextMenuContent>
          </ContextMenu>
        );
      })}
    </div>
  );
}
