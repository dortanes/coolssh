import { SquareTerminalIcon } from 'lucide-react';
import { useEffect, useState } from 'react';
import type { Snapshot } from '#bindings/desktop/models.ts';
import { Action, type Terminal as TerminalSettings } from '#bindings/preferences/models.ts';
import { Status } from '#bindings/session/models.ts';
import { m } from '@/paraglide/messages.js';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import type { Selection } from '../library/selection.ts';
import type { SourcesStore } from '../library/sources-store.ts';
import { DisconnectDialog } from '../sessions/disconnect-dialog.tsx';
import { PaneLayout } from '../sessions/pane-layout.tsx';
import { type Orientation, sessionsIn } from '../sessions/pane-tree.ts';
import type { SessionActions, Split } from '../sessions/session-actions.ts';
import { SessionPane } from '../sessions/session-pane.tsx';
import { SessionTerminals } from '../sessions/session-terminals.ts';
import type { SessionsStore } from '../sessions/sessions-store.ts';
import { type Tab, TabStrip } from '../sessions/tab-strip.tsx';
import type { Tabs } from '../sessions/tabs.ts';
import { TitleBar } from '../shell/title-bar.tsx';
import { useShortcuts } from '../shell/use-shortcuts.ts';
import { useStore } from '../use-store.ts';
import { FolderEditor } from './folder-editor.tsx';
import { HostEditor } from './host-editor.tsx';

type WorkspaceProps = {
  library: LibraryStore;
  selection: Selection;
  sessions: SessionsStore;
  sources: SourcesStore;
  tabs: Tabs;
  actions: SessionActions;
  platform: string;
  terminal: TerminalSettings;
  /** Whether closing a connected session's tab asks first. */
  confirmDisconnect: boolean;
  shortcuts: Snapshot['shortcuts'];
  /** Asks what opens in a new pane beside a session's. */
  onSplit: (split: Split) => void;
};

/**
 * The right side of the main window: the session tabs, and the host's or folder's editor while
 * one is being edited, else the active session, else what opens here.
 */
export function Workspace({
  library,
  selection,
  sessions,
  sources,
  tabs,
  actions,
  platform,
  terminal,
  confirmDisconnect,
  shortcuts,
  onSplit,
}: WorkspaceProps) {
  const view = new LibraryView(useStore(library).library);
  const { id, editing, fresh } = useStore(selection);
  const open = useStore(sessions);
  const { id: focused, layouts } = useStore(tabs);
  const [terminals] = useState(() => new SessionTerminals(sessions, platform, terminal));
  useEffect(() => terminals.configure(terminal), [terminals, terminal]);
  const host = id && editing ? view.host(id) : undefined;
  const folder = id && editing ? view.folder(id) : undefined;
  const editor = host ?? folder;
  const byId = new Map(open.map((session) => [session.id, session]));
  const tabList: Tab[] = layouts.flatMap((layout) => {
    const inTab = sessionsIn(layout).flatMap((each) => byId.get(each) ?? []);
    return inTab[0] ? [{ key: inTab[0].id, sessions: inTab }] : [];
  });
  const shownLayout = editor
    ? undefined
    : layouts.find((layout) => focused && sessionsIn(layout).includes(focused));
  const activeKey = shownLayout
    ? tabList.find((tab) => tab.sessions.some((session) => session.id === focused))?.key
    : undefined;
  const [closing, setClosing] = useState<string[]>([]);
  const close = (ids: string[]) => {
    const connected = ids.filter((each) => sessions.session(each)?.status === Status.Connected);
    if (confirmDisconnect && connected.length > 0) {
      setClosing(ids);
    } else {
      void actions.closeAll(ids);
    }
  };
  const focus = (target: string | undefined) => {
    if (target) {
      actions.show(target);
    }
  };
  const split = (orientation: Orientation) => {
    if (focused) {
      onSplit({ beside: focused, orientation });
    }
  };
  useShortcuts(shortcuts, {
    [Action.ClosePaneAction]: () => focused && close([focused]),
    [Action.NextTabAction]: () => focus(tabs.beside(1)),
    [Action.PreviousTabAction]: () => focus(tabs.beside(-1)),
    [Action.NextPaneAction]: () => focus(tabs.paneBeside(1)),
    [Action.PreviousPaneAction]: () => focus(tabs.paneBeside(-1)),
    [Action.SplitRightAction]: () => split('horizontal'),
    [Action.SplitDownAction]: () => split('vertical'),
  });
  return (
    <main className="flex h-full flex-col overflow-hidden bg-background">
      <TitleBar>
        {open.length > 0 && (
          <TabStrip
            tabs={tabList}
            active={activeKey}
            onShow={(key) => {
              const layout = layouts.find((each) => sessionsIn(each).includes(key));
              focus(layout && tabs.focusIn(layout));
            }}
            onClose={close}
            onSplit={(key, orientation) => {
              const layout = layouts.find((each) => sessionsIn(each).includes(key));
              const beside = layout && tabs.focusIn(layout);
              if (beside) {
                actions.show(beside);
                onSplit({ beside, orientation });
              }
            }}
          />
        )}
      </TitleBar>
      {host && (
        <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
          <HostEditor key={host.id} host={host} fresh={fresh} library={library} sources={sources} />
        </div>
      )}
      {folder && (
        <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
          <FolderEditor
            key={folder.id}
            folder={folder}
            fresh={fresh}
            library={library}
            sources={sources}
          />
        </div>
      )}
      {shownLayout && (
        <div className="min-h-0 flex-1">
          <PaneLayout
            node={shownLayout}
            pane={(sessionId) => {
              const session = byId.get(sessionId);
              return (
                session && (
                  <SessionPane
                    session={session}
                    terminal={terminals.terminal(sessionId)}
                    focused={sessionId === focused}
                    split={'children' in shownLayout}
                    actions={actions}
                    onClose={(each) => close([each])}
                    onFocus={(each) => each !== focused && actions.show(each)}
                    onSplit={(orientation) => onSplit({ beside: sessionId, orientation })}
                  />
                )
              );
            }}
          />
        </div>
      )}
      {!editor && !shownLayout && (
        <div className="flex flex-1 flex-col items-center justify-center gap-4 px-8 in-data-[platform=darwin]:pb-13">
          <span className="flex size-14 items-center justify-center rounded-2xl bg-accent text-muted-foreground">
            <SquareTerminalIcon aria-hidden className="size-7" strokeWidth={1.5} />
          </span>
          <p className="max-w-xs text-center text-sm text-muted-foreground">
            {m.workspace_empty()}
          </p>
        </div>
      )}
      <DisconnectDialog
        hosts={closing.flatMap((each) => {
          const session = sessions.session(each);
          return session?.status === Status.Connected ? [session.title] : [];
        })}
        onConfirm={() => void actions.closeAll(closing)}
        onClose={() => setClosing([])}
      />
    </main>
  );
}
