import type { Snapshot } from '#bindings/desktop/models.ts';
import type { Preferences } from '#bindings/preferences/models.ts';
import { Broadcasts } from './broadcasts.ts';
import { Observed } from './observed.ts';

/** Where a window's preferences come from: Go's bound service and its preferences event. */
export type PreferencesSource = {
  snapshot(): Promise<Snapshot>;
  save(preferences: Preferences): Promise<Snapshot>;
  /** Calls apply with every snapshot Go broadcasts and returns the call that stops listening. */
  listen(apply: (snapshot: Snapshot) => void): () => void;
};

type PreferencesState = {
  snapshot: Snapshot;
  /** Whether the last change failed to save; the snapshot is then the one still in effect. */
  saveFailed: boolean;
};

/** The preferences a window shows, kept in step with every other window through Go. */
export class PreferencesStore extends Observed<PreferencesState> {
  private readonly source: PreferencesSource;

  private constructor(source: PreferencesSource, snapshot: Snapshot) {
    super({ snapshot, saveFailed: false });
    this.source = source;
  }

  /** Reads the preferences in effect and follows every later change. */
  static async open(source: PreferencesSource): Promise<PreferencesStore> {
    const broadcasts = new Broadcasts(source.listen);
    const store = new PreferencesStore(source, broadcasts.latestOr(await source.snapshot()));
    broadcasts.follow((snapshot) => store.update({ snapshot, saveFailed: false }));
    return store;
  }

  /** Saves the preferences with change applied; every window, this one included, follows. */
  async change(change: Partial<Preferences>): Promise<void> {
    const { snapshot } = this.current();
    try {
      this.update({
        snapshot: await this.source.save({ ...snapshot.preferences, ...change }),
        saveFailed: false,
      });
    } catch {
      this.update({ snapshot, saveFailed: true });
    }
  }

  /** Saves the sidebar's width once a drag ends, unless it ends where it started. */
  async rememberSidebarWidth(pixels: number): Promise<void> {
    const sidebarWidth = Math.round(pixels);
    if (sidebarWidth !== this.current().snapshot.preferences.sidebarWidth) {
      await this.change({ sidebarWidth });
    }
  }
}
