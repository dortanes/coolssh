import { PlusIcon, SettingsIcon } from 'lucide-react';
import { useEffect, useState } from 'react';
import { OpenSettings } from '#bindings/desktop/service.ts';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { m } from '@/paraglide/messages.js';
import { DeleteDialog } from '../library/delete-dialog.tsx';
import { HostTree } from '../library/host-tree.tsx';
import type { LibraryActions } from '../library/library-actions.ts';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import type { Selection } from '../library/selection.ts';
import type { SourcesStore } from '../library/sources-store.ts';
import { useStore } from '../use-store.ts';

type SidebarProps = {
  library: LibraryStore;
  sources: SourcesStore;
  selection: Selection;
  actions: LibraryActions;
  onConnect: (hostId: string) => void;
  onInstallKey: (hostId: string) => void;
};

export function Sidebar({
  library,
  sources,
  selection,
  actions,
  onConnect,
  onInstallKey,
}: SidebarProps) {
  const { library: snapshot } = useStore(library);
  const { id: selected } = useStore(selection);
  const [deleting, setDeleting] = useState<string | undefined>();
  useEffect(() => {
    const view = new LibraryView(snapshot);
    if (selected && !view.host(selected) && !view.folder(selected)) {
      selection.select(undefined);
    }
  }, [snapshot, selected, selection]);
  return (
    <div className="h-full">
      <aside
        data-surface="sidebar"
        className="flex h-full flex-col bg-sidebar text-sidebar-foreground"
      >
        {/* On macOS this row is the title bar: it drags the window, the window's buttons sit at its left. */}
        <div
          data-surface="toolbar"
          className="flex h-10 shrink-0 items-center justify-end gap-1.5 px-2 [--wails-draggable:drag] in-data-[platform=darwin]:h-13"
        >
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={m.sidebar_settings()}
            title={m.sidebar_settings()}
            className="[--wails-draggable:no-drag]"
            onClick={() => void OpenSettings()}
          >
            <SettingsIcon aria-hidden />
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={m.library_add()}
                  title={m.library_add()}
                  className="[--wails-draggable:no-drag]"
                />
              }
            >
              <PlusIcon aria-hidden />
            </DropdownMenuTrigger>
            <DropdownMenuContent data-surface="menu" align="end">
              <DropdownMenuItem onClick={() => void actions.newHost('')}>
                {m.library_new_host()}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => void actions.newFolder('')}>
                {m.library_new_folder()}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <HostTree
          library={library}
          sources={sources}
          selection={selection}
          actions={actions}
          onConnect={onConnect}
          onInstallKey={onInstallKey}
          onDelete={setDeleting}
        />
      </aside>
      <DeleteDialog target={deleting} library={library} onClose={() => setDeleting(undefined)} />
    </div>
  );
}
