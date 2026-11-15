import type { TrustedKey } from '#bindings/session/models.ts';
import { Broadcasts } from '../broadcasts.ts';
import { refusal } from '../library/library-store.ts';
import { Observed } from '../observed.ts';

/** Where a window's trusted host keys come from: Go's bound service and its event. */
export type TrustedKeysSource = {
  trusted(): Promise<TrustedKey[] | null>;
  forget(host: string, fingerprint: string): Promise<void>;
  /** Calls apply with every list Go broadcasts and returns the call that stops listening. */
  listen(apply: (trusted: TrustedKey[]) => void): () => void;
};

type TrustedKeysState = {
  trusted: readonly TrustedKey[];
  /** Why the last change failed, worded; undefined once one succeeds. */
  problem: string | undefined;
};

/** The host keys trusted in the app, kept in step with Go. */
export class TrustedKeysStore extends Observed<TrustedKeysState> {
  private readonly source: TrustedKeysSource;

  private constructor(source: TrustedKeysSource, trusted: readonly TrustedKey[]) {
    super({ trusted, problem: undefined });
    this.source = source;
  }

  static async open(source: TrustedKeysSource): Promise<TrustedKeysStore> {
    const broadcasts = new Broadcasts<readonly TrustedKey[]>(source.listen);
    const store = new TrustedKeysStore(source, broadcasts.latestOr((await source.trusted()) ?? []));
    broadcasts.follow((trusted) => store.update({ trusted, problem: undefined }));
    return store;
  }

  /** Stops trusting a host's key; Go broadcasts the list that remains. */
  async forget(key: TrustedKey): Promise<void> {
    try {
      await this.source.forget(key.host, key.fingerprint);
    } catch (error) {
      this.update({ ...this.current(), problem: refusal(error) });
    }
  }
}
