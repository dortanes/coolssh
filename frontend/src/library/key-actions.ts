import { Clipboard } from '@wailsio/runtime';
import { AddKeyFile, PickKeyFile, PublicKey } from '#bindings/desktop/service.ts';
import type { LibraryStore } from './library-store.ts';

/** Asks for a private key file and adds it; undefined when the user cancels or Go refuses it. */
export async function addKeyFromFile(library: LibraryStore): Promise<string | undefined> {
  const path = await library.run(() => PickKeyFile());
  return path ? library.run(() => AddKeyFile(path, '')) : undefined;
}

/** Puts a key's public half on the clipboard, ready for a host's authorized_keys. */
export async function copyPublicKey(library: LibraryStore, id: string): Promise<void> {
  const line = await library.run(() => PublicKey(id));
  if (line) {
    await Clipboard.SetText(line);
  }
}
