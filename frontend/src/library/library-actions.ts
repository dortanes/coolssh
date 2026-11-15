import { CreateFolder, CreateHost, Duplicate, SetExpanded } from '#bindings/desktop/service.ts';
import { m } from '@/paraglide/messages.js';
import type { LibraryStore } from './library-store.ts';
import { LibraryView } from './library-view.ts';
import type { Selection } from './selection.ts';

/** What the tree and its menus do to the library, then show in the workspace. */
export class LibraryActions {
  private readonly library: LibraryStore;
  private readonly selection: Selection;

  constructor(library: LibraryStore, selection: Selection) {
    this.library = library;
    this.selection = selection;
  }

  /** Adds a host inside parent, the root when empty, opens it and puts the cursor in its name. */
  async newHost(parent: string): Promise<void> {
    await this.reveal(parent);
    const id = await this.library.run(() => CreateHost(parent, m.library_new_host()));
    if (id) {
      this.selection.edit(id, true);
    }
  }

  /** Adds a folder inside parent, the root when empty, opens it and puts the cursor in its name. */
  async newFolder(parent: string): Promise<void> {
    await this.reveal(parent);
    const id = await this.library.run(() => CreateFolder(parent, m.library_new_folder()));
    if (id) {
      this.selection.edit(id, true);
    }
  }

  /** Opens a host's or folder's settings in the workspace. */
  edit(id: string): void {
    this.selection.edit(id);
  }

  async duplicate(id: string): Promise<void> {
    const host = new LibraryView(this.library.current().library).host(id);
    if (!host) {
      return;
    }
    const copied = await this.library.run(() =>
      Duplicate(id, m.library_copy_name({ name: host.name })),
    );
    if (copied) {
      this.selection.select(copied);
    }
  }

  /** Expands a collapsed folder, so what goes into it is seen. */
  private async reveal(folder: string): Promise<void> {
    const shown = new LibraryView(this.library.current().library).folder(folder);
    if (shown && !shown.expanded) {
      await this.library.run(() => SetExpanded(folder, true));
    }
  }
}
