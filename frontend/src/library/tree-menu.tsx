import { SetIcon } from '#bindings/desktop/service.ts';
import {
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
} from '@/components/ui/context-menu';
import { m } from '@/paraglide/messages.js';
import type { LibraryActions } from './library-actions.ts';
import { folderIcons } from './library-icons.tsx';
import type { LibraryStore } from './library-store.ts';
import type { LibraryView } from './library-view.ts';

type TreeMenuProps = {
  /** The host or folder right-clicked, empty for the tree's empty space. */
  target: string;
  view: LibraryView;
  library: LibraryStore;
  actions: LibraryActions;
  onConnect: (hostId: string) => void;
  onInstallKey: (hostId: string) => void;
  onRename: (id: string) => void;
  onDelete: (id: string) => void;
};

/** What a right-click in the tree offers: for a host, a folder, or the empty space. */
export function TreeMenu({
  target,
  view,
  library,
  actions,
  onConnect,
  onInstallKey,
  onRename,
  onDelete,
}: TreeMenuProps) {
  const folder = view.folder(target);
  const creating = (
    <>
      <ContextMenuItem onClick={() => void actions.newHost(target)}>
        {m.library_new_host()}
      </ContextMenuItem>
      <ContextMenuItem onClick={() => void actions.newFolder(target)}>
        {m.library_new_folder()}
      </ContextMenuItem>
    </>
  );
  if (target === '' || (!folder && !view.host(target))) {
    return creating;
  }
  // Nothing is created in a synced folder, and its hosts are only connected to and edited.
  const synced = view.synced(target);
  return (
    <>
      {folder && !synced && creating}
      {folder && !synced && <ContextMenuSeparator />}
      {!folder && (
        <>
          <ContextMenuItem onClick={() => onConnect(target)}>{m.library_connect()}</ContextMenuItem>
          <ContextMenuItem onClick={() => onInstallKey(target)}>
            {m.library_install_key()}
          </ContextMenuItem>
          <ContextMenuSeparator />
        </>
      )}
      <ContextMenuItem onClick={() => actions.edit(target)}>{m.library_edit()}</ContextMenuItem>
      {(folder || !synced) && (
        <ContextMenuItem onClick={() => onRename(target)}>{m.library_rename()}</ContextMenuItem>
      )}
      {!folder && !synced && (
        <ContextMenuItem onClick={() => void actions.duplicate(target)}>
          {m.library_duplicate()}
        </ContextMenuItem>
      )}
      {folder && (
        <ContextMenuSub>
          <ContextMenuSubTrigger>{m.library_icon()}</ContextMenuSubTrigger>
          <ContextMenuSubContent data-surface="menu">
            {folderIcons.map(({ icon, Glyph, label }) => (
              <ContextMenuItem
                key={icon}
                onClick={() => void library.run(() => SetIcon(target, icon))}
              >
                <Glyph aria-hidden />
                {label()}
              </ContextMenuItem>
            ))}
          </ContextMenuSubContent>
        </ContextMenuSub>
      )}
      {(folder || !synced) && (
        <>
          <ContextMenuSeparator />
          <ContextMenuItem variant="destructive" onClick={() => onDelete(target)}>
            {m.library_delete()}
          </ContextMenuItem>
        </>
      )}
    </>
  );
}
