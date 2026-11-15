import { FolderPlusIcon, PlusIcon, SettingsIcon } from 'lucide-react';
import { useState } from 'react';
import { usePanelRef } from 'react-resizable-panels';
import { OpenSettings } from '#bindings/desktop/service.ts';
import { Action } from '#bindings/preferences/models.ts';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { m } from '@/paraglide/messages.js';
import { LibraryActions } from '../library/library-actions.ts';
import { Selection } from '../library/selection.ts';
import type { SourcesStore } from '../library/sources-store.ts';
import type { WindowStores } from '../mount-window.tsx';
import { SessionActions, type Split } from '../sessions/session-actions.ts';
import type { SessionsStore } from '../sessions/sessions-store.ts';
import { Tabs } from '../sessions/tabs.ts';
import { useStore } from '../use-store.ts';
import { Workspace } from '../workspace/workspace.tsx';
import { actionLabels } from './action-labels.ts';
import { CommandPalette, type PaletteAction } from './command-palette.tsx';
import { Sidebar } from './sidebar.tsx';
import { useShortcuts } from './use-shortcuts.ts';

/** The sidebar's limits in CSS pixels: room for a host's name, and most of the window for work. */
const sidebarWidths = { min: 200, max: 440 };

export function MainWindow({
  preferences,
  library,
  sessions,
  sources,
}: WindowStores & { sessions: SessionsStore; sources: SourcesStore }) {
  const { snapshot } = useStore(preferences);
  const [selection] = useState(() => new Selection());
  const [tabs] = useState(() => new Tabs(sessions));
  const [actions] = useState(() => new SessionActions(library, sessions, tabs, selection));
  const [libraryActions] = useState(() => new LibraryActions(library, selection));
  // Open with a split when the palette picks what opens in a new pane.
  const [palette, setPalette] = useState<{ split: Split | undefined } | undefined>();
  const { shortcuts, platform } = snapshot;
  const newHost = () => void libraryActions.newHost('');
  const newFolder = () => void libraryActions.newFolder('');
  useShortcuts(shortcuts, {
    [Action.CommandPaletteAction]: () =>
      setPalette((shown) => (shown ? undefined : { split: undefined })),
    [Action.NewHostAction]: newHost,
    [Action.NewFolderAction]: newFolder,
  });
  const paletteActions: PaletteAction[] = [
    {
      id: Action.NewHostAction,
      label: actionLabels[Action.NewHostAction](),
      Glyph: PlusIcon,
      shortcut: shortcuts?.[Action.NewHostAction],
      run: newHost,
    },
    {
      id: Action.NewFolderAction,
      label: actionLabels[Action.NewFolderAction](),
      Glyph: FolderPlusIcon,
      shortcut: shortcuts?.[Action.NewFolderAction],
      run: newFolder,
    },
    {
      id: 'settings',
      label: m.sidebar_settings(),
      Glyph: SettingsIcon,
      // The native menu's Settings item owns this shortcut.
      shortcut: platform === 'darwin' ? 'Meta+Comma' : 'Control+Comma',
      run: () => void OpenSettings(),
    },
  ];
  const sidebar = usePanelRef();
  const rememberWidth = () => {
    const size = sidebar.current?.getSize();
    if (size) {
      void preferences.rememberSidebarWidth(size.inPixels);
    }
  };
  return (
    <ResizablePanelGroup orientation="horizontal" onLayoutChanged={rememberWidth}>
      <ResizablePanel
        panelRef={sidebar}
        defaultSize={snapshot.preferences.sidebarWidth}
        minSize={sidebarWidths.min}
        maxSize={sidebarWidths.max}
        groupResizeBehavior="preserve-pixel-size"
      >
        <Sidebar
          library={library}
          sources={sources}
          selection={selection}
          actions={libraryActions}
          onInstallKey={(hostId) => void actions.installKey(hostId)}
          onConnect={(hostId) => void actions.connect(hostId)}
        />
      </ResizablePanel>
      <ResizableHandle aria-label={m.sidebar_resize()} />
      <ResizablePanel>
        <Workspace
          library={library}
          selection={selection}
          sessions={sessions}
          sources={sources}
          tabs={tabs}
          actions={actions}
          platform={platform}
          shortcuts={shortcuts}
          onSplit={(split) => setPalette({ split })}
          terminal={snapshot.preferences.terminal}
          confirmDisconnect={snapshot.preferences.confirmDisconnect}
        />
      </ResizablePanel>
      <CommandPalette
        open={palette !== undefined}
        onOpenChange={(open) => !open && setPalette(undefined)}
        split={palette?.split}
        library={library}
        sessions={sessions}
        actions={paletteActions}
        platform={platform}
        onConnect={(hostId) => void actions.connect(hostId, palette?.split)}
        onConnectAddress={(target) => void actions.connectAddress(target, palette?.split)}
        onShowSession={(id) =>
          palette?.split ? actions.place(id, palette.split) : actions.show(id)
        }
      />
    </ResizablePanelGroup>
  );
}
