import { type LucideIcon, PlugIcon, SquareTerminalIcon } from 'lucide-react';
import { useState } from 'react';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from '@/components/ui/command';
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@/components/ui/dialog';
import { m } from '@/paraglide/messages.js';
import { hostChoice } from '../library/library-icons.tsx';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import type { Split } from '../sessions/session-actions.ts';
import type { SessionsStore } from '../sessions/sessions-store.ts';
import { useStore } from '../use-store.ts';
import { shortcutLabel } from './shortcut.ts';

/** Something the palette runs, with the shortcut that runs it too. */
export type PaletteAction = {
  id: string;
  label: string;
  Glyph: LucideIcon;
  shortcut: string | undefined;
  run: () => void;
};

/** Text that could be `[user@]host[:port]`: one word with at most one @ and a host after it. */
const addressLike = /^([^\s@]+@)?[^\s@]+$/;

/** Matches every word of the search somewhere in an entry's keywords, ignoring case. */
function filter(_value: string, search: string, keywords?: string[]): number {
  const text = (keywords ?? []).join(' ').toLowerCase();
  return search
    .toLowerCase()
    .split(/\s+/)
    .every((word) => text.includes(word))
    ? 1
    : 0;
}

/** The folders holding id, outermost first, joined for display. */
function folderPath(view: LibraryView, id: string): string {
  const names: string[] = [];
  for (let folder = view.parentOf(id); folder !== ''; folder = view.parentOf(folder)) {
    names.unshift(view.folder(folder)?.name ?? '');
  }
  return names.join(' / ');
}

/**
 * Finds a host, an open session or an action by typing, and connects to a typed address without
 * saving it.
 */
export function CommandPalette({
  open,
  onOpenChange,
  library,
  sessions,
  actions,
  platform,
  split,
  onConnect,
  onConnectAddress,
  onShowSession,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  library: LibraryStore;
  sessions: SessionsStore;
  actions: PaletteAction[];
  platform: string;
  /** Where a chosen host, address or session opens; undefined to open in a tab of its own. */
  split: Split | undefined;
  onConnect: (hostId: string) => void;
  onConnectAddress: (target: string) => void;
  onShowSession: (id: string) => void;
}) {
  const view = new LibraryView(useStore(library).library);
  const running = useStore(sessions);
  const [search, setSearch] = useState('');
  const target = search.trim();
  const placeholder = split ? m.palette_split_placeholder() : m.palette_placeholder();
  const openChange = (next: boolean) => {
    onOpenChange(next);
    if (!next) {
      setSearch('');
    }
  };
  const choose = (run: () => void) => {
    openChange(false);
    run();
  };
  return (
    <Dialog open={open} onOpenChange={openChange}>
      <DialogContent
        data-surface="dialog"
        showCloseButton={false}
        className="top-1/4 translate-y-0 overflow-hidden p-0 sm:max-w-lg"
      >
        <DialogTitle className="sr-only">{m.palette_title()}</DialogTitle>
        <DialogDescription className="sr-only">{placeholder}</DialogDescription>
        <Command filter={filter} className="bg-transparent">
          <CommandInput value={search} onValueChange={setSearch} placeholder={placeholder} />
          <CommandList>
            <CommandEmpty>{m.palette_empty()}</CommandEmpty>
            <CommandGroup heading={m.palette_hosts()}>
              {view.hosts().map((host) => {
                const { Glyph } = hostChoice(view.hostIconIn(view.parentOf(host.id)));
                const path = folderPath(view, host.id);
                return (
                  <CommandItem
                    key={host.id}
                    value={`host-${host.id}`}
                    keywords={[host.name, host.address, path, ...(host.tags ?? [])]}
                    onSelect={() => choose(() => onConnect(host.id))}
                  >
                    <Glyph aria-hidden className="text-muted-foreground" />
                    <span className="truncate">{host.name}</span>
                    <span className="ml-auto truncate text-xs text-muted-foreground">
                      {[...(host.tags ?? []).map((tag) => `#${tag}`), path, host.address]
                        .filter(Boolean)
                        .join(' · ')}
                    </span>
                  </CommandItem>
                );
              })}
            </CommandGroup>
            {running.length > 0 && (
              <CommandGroup heading={m.palette_sessions()}>
                {running
                  .filter((session) => session.id !== split?.beside)
                  .map((session) => (
                    <CommandItem
                      key={session.id}
                      value={`session-${session.id}`}
                      keywords={[session.title]}
                      onSelect={() => choose(() => onShowSession(session.id))}
                    >
                      <SquareTerminalIcon aria-hidden className="text-muted-foreground" />
                      <span className="truncate">{session.title}</span>
                    </CommandItem>
                  ))}
              </CommandGroup>
            )}
            {!split && (
              <CommandGroup heading={m.palette_actions()}>
                {actions.map(({ id, label, Glyph, shortcut, run }) => (
                  <CommandItem
                    key={id}
                    value={`action-${id}`}
                    keywords={[label]}
                    onSelect={() => choose(run)}
                  >
                    <Glyph aria-hidden className="text-muted-foreground" />
                    {label}
                    {shortcut && (
                      <CommandShortcut>{shortcutLabel(shortcut, platform)}</CommandShortcut>
                    )}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
            {addressLike.test(target) && (
              <CommandGroup forceMount heading={m.palette_address()}>
                <CommandItem
                  forceMount
                  value="address"
                  onSelect={() => choose(() => onConnectAddress(target))}
                >
                  <PlugIcon aria-hidden className="text-muted-foreground" />
                  <span className="truncate">{m.palette_connect({ target })}</span>
                </CommandItem>
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </DialogContent>
    </Dialog>
  );
}
