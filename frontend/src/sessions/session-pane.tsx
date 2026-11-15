// jscpd:ignore-start: the tab strip imports the same components; only the imports match
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
import { PromptDialog } from './prompt-dialog.tsx';
import type { SessionActions } from './session-actions.ts';
import type { SessionTerminal } from './session-terminal.ts';
import { SplitItems } from './split-items.tsx';

// jscpd:ignore-end

type SessionPaneProps = {
  session: State;
  terminal: SessionTerminal;
  /** Whether the pane has the focus among the panes of its tab. */
  focused: boolean;
  /** Whether the pane shares its tab with others. */
  split: boolean;
  actions: SessionActions;
  onClose: (id: string) => void;
  onFocus: (id: string) => void;
  /** Splits the pane, asking what opens beside it. */
  onSplit: (orientation: Orientation) => void;
};

/** A session's pane: its terminal, what it asks, and its state while not connected. */
export function SessionPane({
  session,
  terminal,
  focused,
  split,
  actions,
  onClose,
  onFocus,
  onSplit,
}: SessionPaneProps) {
  const element = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (element.current) {
      terminal.attach(element.current);
    }
  }, [terminal]);
  useEffect(() => {
    if (focused) {
      terminal.fitToElement();
      terminal.focus();
    }
  }, [focused, terminal]);
  return (
    <div
      data-focused={split && focused ? '' : undefined}
      className="relative flex h-full min-h-0 flex-col data-focused:shadow-[inset_0_2px_0_var(--ring)]"
      onFocusCapture={() => onFocus(session.id)}
      onPointerDownCapture={() => onFocus(session.id)}
    >
      <ContextMenu>
        <ContextMenuTrigger className="relative min-h-0 flex-1 overflow-hidden">
          {/* Positioned, so the terminal's own size never feeds back into the size it fits. */}
          <div ref={element} className="absolute inset-x-3 top-2 bottom-1 overflow-hidden" />
        </ContextMenuTrigger>
        <ContextMenuContent data-surface="menu">
          <ContextMenuItem onClick={() => void terminal.copy()}>{m.menu_copy()}</ContextMenuItem>
          <ContextMenuItem onClick={() => void terminal.paste()}>{m.menu_paste()}</ContextMenuItem>
          <ContextMenuItem onClick={() => terminal.selectAll()}>
            {m.menu_select_all()}
          </ContextMenuItem>
          <ContextMenuSeparator />
          <ContextMenuItem onClick={() => terminal.clear()}>{m.terminal_clear()}</ContextMenuItem>
          <ContextMenuSeparator />
          <SplitItems onSplit={onSplit} />
          {split && (
            <ContextMenuItem onClick={() => actions.separate(session.id)}>
              {m.session_separate()}
            </ContextMenuItem>
          )}
        </ContextMenuContent>
      </ContextMenu>
      {session.status === Status.Connecting && !session.prompt && (
        <div className="absolute inset-0 flex items-center justify-center gap-2 bg-background text-sm text-muted-foreground">
          <Spinner aria-hidden />
          {m.session_connecting({ host: session.title })}
        </div>
      )}
      {session.status === Status.Closed && (
        <div
          role="status"
          className="flex shrink-0 flex-wrap items-center gap-3 border-t border-border px-4 py-3 text-sm"
        >
          <p className="min-w-0 flex-1 select-text">
            {session.problem || session.outcome || m.session_ended()}
          </p>
          <Button size="sm" variant="secondary" onClick={() => onClose(session.id)}>
            {m.session_close()}
          </Button>
          <Button size="sm" onClick={() => void actions.reconnect(session.id)}>
            {m.session_reconnect()}
          </Button>
        </div>
      )}
      {session.prompt && (
        <PromptDialog
          key={session.prompt.id}
          sessionId={session.id}
          prompt={session.prompt}
          actions={actions}
        />
      )}
    </div>
  );
}
