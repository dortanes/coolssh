import {
  dragAndDropFeature,
  hotkeysCoreFeature,
  type ItemInstance,
  isOrderedDragTarget,
  renamingFeature,
  type SetStateFn,
  selectionFeature,
  syncDataLoaderFeature,
} from '@headless-tree/core';
import { useTree } from '@headless-tree/react';
import { cn } from 'cn';
import { ChevronRightIcon } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { Move, Rename, Resync, SetExpanded } from '#bindings/desktop/service.ts';
import type { Link, State } from '#bindings/hostsource/models.ts';
import { Icon } from '#bindings/library/models.ts';
import { Button } from '@/components/ui/button';
import { ContextMenu, ContextMenuContent, ContextMenuTrigger } from '@/components/ui/context-menu';
import { Input } from '@/components/ui/input';
import { m } from '@/paraglide/messages.js';
import { useStore } from '../use-store.ts';
import type { LibraryActions } from './library-actions.ts';
import { folderGlyph, hostChoice } from './library-icons.tsx';
import type { LibraryStore } from './library-store.ts';
import { LibraryView } from './library-view.ts';
import { LinkDot } from './peer-link.tsx';
import type { Selection } from './selection.ts';
import { linkOf, type SourcesStore } from './sources-store.ts';
import { SyncGlyph } from './sync-glyph.tsx';
import { TreeMenu } from './tree-menu.tsx';

/** A tree row: source is the kind a synced folder takes its hosts from, empty otherwise. */
type Node = { id: string; folder: boolean; name: string; icon: Icon; source: string };

/** The tree's own root, standing for the library's root, which Go names with an empty id. */
const rootId = 'root';

function libraryId(treeId: string): string {
  return treeId === rootId ? '' : treeId;
}

function nodeOf(view: LibraryView, id: string): Node {
  const folder = view.folder(id);
  if (folder || id === rootId) {
    return {
      id,
      folder: true,
      name: folder?.name ?? '',
      icon: folder?.icon ?? Icon.FolderIcon,
      source: folder?.source?.kind ?? '',
    };
  }
  return {
    id,
    folder: false,
    name: view.host(id)?.name ?? '',
    icon: view.hostIconIn(view.parentOf(id)),
    source: '',
  };
}

function resolve<T>(updater: Parameters<SetStateFn<T>>[0], old: T): T {
  return typeof updater === 'function' ? (updater as (old: T) => T)(old) : updater;
}

type HostTreeProps = {
  library: LibraryStore;
  sources: SourcesStore;
  selection: Selection;
  actions: LibraryActions;
  onConnect: (hostId: string) => void;
  onInstallKey: (hostId: string) => void;
  onDelete: (id: string) => void;
};

/** The library as a tree: folders and hosts in order, with menus, renaming and drag and drop. */
export function HostTree({
  library,
  sources,
  selection,
  actions,
  onConnect,
  onInstallKey,
  onDelete,
}: HostTreeProps) {
  const { library: snapshot } = useStore(library);
  const { states } = useStore(sources);
  const { id: selected } = useStore(selection);
  const view = new LibraryView(snapshot);
  const [menuTarget, setMenuTarget] = useState('');
  // The tree rebuilds whenever its state's arrays change, and a rebuild during render renders
  // again, so the arrays change only with what they hold.
  const expanded = useMemo(
    () =>
      Object.values(snapshot.folders ?? {}).flatMap((folder) =>
        folder?.expanded ? [folder.id] : [],
      ),
    [snapshot],
  );
  const selectedItems = useMemo(() => (selected ? [selected] : []), [selected]);
  const tree = useTree<Node>({
    rootItemId: rootId,
    getItemName: (item) => item.getItemData().name,
    isItemFolder: (item) => item.getItemData().folder,
    dataLoader: {
      getItem: (id) => nodeOf(view, id),
      getChildren: (id) => view.children(libraryId(id)),
    },
    state: { expandedItems: expanded, selectedItems },
    setExpandedItems: (updater) => {
      const next = resolve(updater, expanded);
      for (const id of new Set([...next, ...expanded])) {
        if (next.includes(id) !== expanded.includes(id)) {
          void library.run(() => SetExpanded(id, next.includes(id)));
        }
      }
    },
    setSelectedItems: (updater) => {
      selection.select(resolve(updater, selectedItems)[0]);
    },
    onPrimaryAction: (item) => selection.select(item.getId()),
    onRename: (item, name) => void library.run(() => Rename(item.getId(), name)),
    // A source owns its folder's hosts: they stay put and keep the names it gives them.
    canRename: (item) => item.isFolder() || !view.synced(item.getId()),
    canDrag: (items) => items.every((item) => item.isFolder() || !view.synced(item.getId())),
    canReorder: true,
    canDrop: (_items, target) =>
      target.item.isFolder() && !view.synced(libraryId(target.item.getId())),
    onDrop: async (items, target) => {
      const parent = libraryId(target.item.getId());
      const index = isOrderedDragTarget(target) ? target.childIndex : view.children(parent).length;
      await library.run(() =>
        Move(
          items.map((item) => item.getId()),
          parent,
          index,
        ),
      );
    },
    hotkeys: {
      customConnect: {
        hotkey: 'Enter',
        preventDefault: true,
        isEnabled: (instance) => !instance.isRenamingItem() && !instance.getState().dnd,
        handler: (_event, instance) => {
          const focused = instance.getFocusedItem();
          if (!focused.isFolder()) {
            onConnect(focused.getId());
          }
        },
      },
      customDelete: {
        hotkey: 'Delete',
        preventDefault: true,
        handler: (_event, instance) => {
          const focused = instance.getFocusedItem();
          if (focused.isFolder() || !view.synced(focused.getId())) {
            onDelete(focused.getId());
          }
        },
      },
    },
    indent: 14,
    features: [
      syncDataLoaderFeature,
      selectionFeature,
      hotkeysCoreFeature,
      dragAndDropFeature,
      renamingFeature,
    ],
  });
  // biome-ignore lint/correctness/useExhaustiveDependencies: the tree reads the library through its data loader, so each new snapshot must rebuild it
  useEffect(() => {
    tree.rebuildTree();
  }, [tree, snapshot]);
  const items = tree.getItems();
  return (
    <ContextMenu>
      <ContextMenuTrigger
        className="flex min-h-0 flex-1 flex-col overflow-y-auto"
        onContextMenuCapture={(event) => {
          const row =
            event.target instanceof Element ? event.target.closest('[data-item-id]') : null;
          setMenuTarget(row?.getAttribute('data-item-id') ?? '');
        }}
      >
        {items.length === 0 ? (
          <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
            <p className="text-sm text-muted-foreground">{m.sidebar_empty()}</p>
            <div className="flex flex-wrap justify-center gap-2">
              <Button size="sm" variant="secondary" onClick={() => void actions.newHost('')}>
                {m.library_new_host()}
              </Button>
              <Button size="sm" variant="secondary" onClick={() => void actions.newFolder('')}>
                {m.library_new_folder()}
              </Button>
            </div>
          </div>
        ) : (
          <div
            {...tree.getContainerProps(m.library_tree())}
            className="relative mx-2 mb-2 flex flex-col gap-px outline-none"
          >
            {items.map((item) => (
              <TreeRow
                key={item.getId()}
                item={item}
                sync={states[item.getId()]}
                link={linkOf(states, view.peerOf(item.getId()))}
                onConnect={onConnect}
              />
            ))}
            <div style={tree.getDragLineStyle()} className="h-0.5 rounded-full bg-primary" />
          </div>
        )}
      </ContextMenuTrigger>
      <ContextMenuContent data-surface="menu">
        <TreeMenu
          target={menuTarget}
          view={view}
          library={library}
          actions={actions}
          onConnect={onConnect}
          onInstallKey={onInstallKey}
          onRename={(id) => tree.getItemInstance(id).startRenaming()}
          onDelete={onDelete}
        />
      </ContextMenuContent>
    </ContextMenu>
  );
}

function TreeRow({
  item,
  sync,
  link,
  onConnect,
}: {
  item: ItemInstance<Node>;
  /** A synced folder's sync state. */
  sync: State | undefined;
  /** How this device reaches a synced host's peer. */
  link: Link | undefined;
  onConnect: (hostId: string) => void;
}) {
  const node = item.getItemData();
  const Glyph = node.folder ? folderGlyph(node.icon) : hostChoice(node.icon).Glyph;
  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: item.getProps() makes the row a treeitem with the tree's keyboard handling
    <div
      {...item.getProps()}
      data-item-id={item.getId()}
      onDoubleClick={node.folder ? undefined : () => onConnect(item.getId())}
      className={cn(
        'flex h-7 cursor-default items-center gap-1.5 rounded-md pr-2 text-sm outline-none select-none',
        'aria-selected:bg-sidebar-accent aria-selected:text-sidebar-accent-foreground',
        'focus-visible:ring-2 focus-visible:ring-sidebar-ring',
        item.isDragTarget() && 'bg-sidebar-accent',
      )}
      style={{ paddingInlineStart: `${0.375 + item.getItemMeta().level * 0.875}rem` }}
    >
      {node.folder ? (
        <ChevronRightIcon
          aria-hidden
          className={cn(
            'size-3.5 shrink-0 text-muted-foreground transition-transform',
            item.isExpanded() && 'rotate-90',
          )}
        />
      ) : (
        <span className="size-3.5 shrink-0" />
      )}
      <Glyph aria-hidden className="size-4 shrink-0 text-muted-foreground" />
      {item.isRenaming() ? (
        <Input {...item.getRenameInputProps()} className="h-6 px-1.5 py-0" />
      ) : (
        <span className="truncate">{node.name}</span>
      )}
      {node.source && (
        <SyncGlyph kind={node.source} state={sync} onResync={() => void Resync(item.getId())} />
      )}
      {link && <LinkDot link={link} className="ml-auto" />}
    </div>
  );
}
