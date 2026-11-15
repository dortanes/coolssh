import type { Library } from '#bindings/library/models.ts';
import { Broadcasts } from '../broadcasts.ts';
import { Observed } from '../observed.ts';

/** Where a window's library comes from: Go's bound service and its library event. */
export type LibrarySource = {
  library(): Promise<Library>;
  /** Calls apply with every library Go broadcasts and returns the call that stops listening. */
  listen(apply: (library: Library) => void): () => void;
};

type LibraryState = {
  library: Library;
  /** Why the last change was refused, worded by Go; cleared by the next change that succeeds. */
  problem: string | undefined;
};

/** The library a window shows, kept in step with every other window through Go. */
export class LibraryStore extends Observed<LibraryState> {
  private constructor(library: Library) {
    super({ library, problem: undefined });
  }

  /** Reads the library and follows every later change. */
  static async open(source: LibrarySource): Promise<LibraryStore> {
    const broadcasts = new Broadcasts(source.listen);
    const store = new LibraryStore(broadcasts.latestOr(await source.library()));
    broadcasts.follow((library) => store.update({ ...store.current(), library }));
    return store;
  }

  /**
   * Runs a call into Go. Its result comes back on success; a refusal becomes the problem the
   * window shows, and the result is undefined.
   */
  async run<T>(call: () => Promise<T>): Promise<T | undefined> {
    try {
      const result = await call();
      this.report(undefined);
      return result;
    } catch (error) {
      this.report(refusal(error));
      return undefined;
    }
  }

  /** Shows a problem found before anything reached Go, or clears it. */
  report(problem: string | undefined): void {
    if (problem !== this.current().problem) {
      this.update({ ...this.current(), problem });
    }
  }
}

/** Go's wording of why a call was refused. */
export function refusal(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
