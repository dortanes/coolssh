import {
  Answer as AnswerPrompt,
  CloseSession,
  Connect,
  ConnectAddress,
  InstallKey,
  Reconnect,
} from '#bindings/desktop/service.ts';
import type { Answer } from '#bindings/session/models.ts';
import type { LibraryStore } from '../library/library-store.ts';
import type { Selection } from '../library/selection.ts';
import type { Orientation } from './pane-tree.ts';
import type { SessionsStore } from './sessions-store.ts';
import type { Tabs } from './tabs.ts';

/** Where a new session goes: beside a session's pane, split along orientation. */
export type Split = { beside: string; orientation: Orientation };

/** What the tree and the tabs do to sessions; a refusal shows as the library's problem. */
export class SessionActions {
  private readonly library: LibraryStore;
  private readonly sessions: SessionsStore;
  private readonly tabs: Tabs;
  private readonly selection: Selection;

  constructor(library: LibraryStore, sessions: SessionsStore, tabs: Tabs, selection: Selection) {
    this.library = library;
    this.sessions = sessions;
    this.tabs = tabs;
    this.selection = selection;
  }

  /** Opens a session to a host in a new tab, or beside a pane when split, and shows it. */
  async connect(hostId: string, split?: Split): Promise<void> {
    const id = await this.library.run(() => Connect(hostId));
    if (id) {
      this.open(id, split);
      this.selection.select(hostId);
    }
  }

  /** Opens a session to a typed [user@]host[:port] without adding a host. */
  async connectAddress(target: string, split?: Split): Promise<void> {
    const id = await this.library.run(() => ConnectAddress(target));
    if (id) {
      this.open(id, split);
      this.selection.select(undefined);
    }
  }

  /** Opens a session that adds the key a host uses to its authorized_keys, in a tab of its own. */
  async installKey(hostId: string): Promise<void> {
    const id = await this.library.run(() => InstallKey(hostId));
    if (id) {
      this.open(id, undefined);
    }
  }

  /** Focuses a session's pane, closing any editor, and selects its host. */
  show(id: string): void {
    this.tabs.show(id);
    this.selection.select(this.sessions.session(id)?.hostId || undefined);
  }

  /** Moves an open session's pane beside another's. */
  place(id: string, split: Split): void {
    this.tabs.place(id, split.beside, split.orientation);
    this.show(id);
  }

  /** Moves a session's pane out of its split into a tab of its own. */
  separate(id: string): void {
    this.tabs.separate(id);
  }

  /** Ends the sessions and closes their panes. */
  async closeAll(ids: readonly string[]): Promise<void> {
    for (const id of ids) {
      await this.library.run(() => CloseSession(id));
    }
  }

  async reconnect(id: string): Promise<void> {
    await this.library.run(() => Reconnect(id));
  }

  async answer(id: string, prompt: number, answer: Answer): Promise<void> {
    await this.library.run(() => AnswerPrompt(id, prompt, answer));
  }

  private open(id: string, split: Split | undefined): void {
    if (split) {
      this.tabs.place(id, split.beside, split.orientation);
    } else {
      this.tabs.show(id);
    }
  }
}
