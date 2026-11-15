import type { Availability, Link, State } from '#bindings/hostsource/models.ts';
import { Broadcasts } from '../broadcasts.ts';
import { Observed } from '../observed.ts';

/** Where a window's host sources come from: Go's bound service and its sources event. */
export type SourcesSource = {
  available(): Promise<Availability[] | null>;
  states(): Promise<{ [folder in string]?: State } | null>;
  /** Calls apply with every set of states Go broadcasts and returns the call that stops listening. */
  listen(apply: (states: { [folder in string]?: State }) => void): () => void;
};

type SourcesState = {
  /** Every source, in the order offered, with whether its client is installed. */
  available: readonly Availability[];
  /** The sync state of each synced folder, by folder id. */
  states: { [folder in string]?: State };
};

/** How this device reaches a synced host's peer, as its folder last synced. */
export function linkOf(
  states: SourcesState['states'],
  synced: { folder: string; peer: string } | undefined,
): Link | undefined {
  return synced && states[synced.folder]?.links?.[synced.peer];
}

/** The host sources a folder can sync from and how each synced folder's sync stands. */
export class SourcesStore extends Observed<SourcesState> {
  private constructor(state: SourcesState) {
    super(state);
  }

  static async open(source: SourcesSource): Promise<SourcesStore> {
    const broadcasts = new Broadcasts(source.listen);
    const [available, states] = await Promise.all([source.available(), source.states()]);
    const store = new SourcesStore({
      available: available ?? [],
      states: broadcasts.latestOr(states ?? {}),
    });
    broadcasts.follow((next) => store.update({ ...store.current(), states: next }));
    return store;
  }
}
