import { EllipsisIcon, KeyRoundIcon } from 'lucide-react';
import { useState } from 'react';
import { RenameKey, SetDefaultKey } from '#bindings/desktop/service.ts';
import { type Key, KeySource } from '#bindings/library/models.ts';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { m } from '@/paraglide/messages.js';
import { copyPublicKey } from '../library/key-actions.ts';
import { keyKind } from '../library/key-kind.ts';
import type { LibraryStore } from '../library/library-store.ts';

/** The last part of a path, whichever separator the system writes. */
function fileName(path: string): string {
  return path.split(/[\\/]/).pop() ?? path;
}

/** One key: its name and kind, where it lives, its fingerprint, and a menu of what it can do. */
export function KeyRow({
  keyEntry,
  isDefault,
  library,
  onDelete,
}: {
  keyEntry: Key;
  /** Whether hosts use this key when neither they nor their folders set one. */
  isDefault: boolean;
  library: LibraryStore;
  onDelete: () => void;
}) {
  const [renaming, setRenaming] = useState(false);
  const [draft, setDraft] = useState(keyEntry.name);
  const finishRenaming = (save: boolean) => {
    if (save && draft !== keyEntry.name) {
      void library.run(() => RenameKey(keyEntry.id, draft));
    }
    setDraft(keyEntry.name);
    setRenaming(false);
  };
  const source =
    keyEntry.source === KeySource.FileKey ? fileName(keyEntry.path ?? '') : m.keys_pasted();
  return (
    <li className="flex items-center gap-3 px-4 py-3">
      <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-accent text-muted-foreground">
        <KeyRoundIcon aria-hidden className="size-4" />
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        {renaming ? (
          <Input
            autoFocus
            aria-label={m.keys_name()}
            value={draft}
            className="h-7"
            onChange={(event) => setDraft(event.target.value)}
            onBlur={() => finishRenaming(true)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                finishRenaming(true);
              } else if (event.key === 'Escape') {
                event.preventDefault();
                finishRenaming(false);
              }
            }}
          />
        ) : (
          <span className="flex items-baseline gap-2">
            <span className="truncate text-sm font-medium">{keyEntry.name}</span>
            {keyEntry.type && (
              <span className="shrink-0 text-xs text-muted-foreground">
                {keyKind(keyEntry.type)}
              </span>
            )}
            {isDefault && (
              <span
                title={m.keys_default_hint()}
                className="shrink-0 rounded-full bg-accent px-2 py-px text-[0.6875rem] font-medium"
              >
                {m.keys_is_default()}
              </span>
            )}
          </span>
        )}
        <span
          className="truncate text-xs text-muted-foreground"
          title={keyEntry.source === KeySource.FileKey ? keyEntry.path : undefined}
        >
          {[source, keyEntry.encrypted && m.keys_encrypted()].filter(Boolean).join(' · ')}
        </span>
        {keyEntry.fingerprint && (
          <span
            className="truncate font-mono text-[0.6875rem] text-muted-foreground"
            title={keyEntry.fingerprint}
          >
            {keyEntry.fingerprint}
          </span>
        )}
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={<Button variant="ghost" size="icon-sm" aria-label={m.keys_actions()} />}
        >
          <EllipsisIcon aria-hidden />
        </DropdownMenuTrigger>
        <DropdownMenuContent data-surface="menu" align="end">
          <DropdownMenuItem
            onClick={() => void library.run(() => SetDefaultKey(isDefault ? '' : keyEntry.id))}
          >
            {isDefault ? m.keys_unset_default() : m.keys_make_default()}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => void copyPublicKey(library, keyEntry.id)}>
            {m.keys_copy_public()}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setRenaming(true)}>
            {m.library_rename()}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem variant="destructive" onClick={onDelete}>
            {m.library_delete()}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </li>
  );
}
