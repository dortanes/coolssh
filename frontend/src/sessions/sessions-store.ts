import type { State } from '#bindings/session/models.ts';
import { Broadcasts } from '../broadcasts.ts';
import { Observed } from '../observed.ts';

/** Where a window's sessions come from: Go's bound service and its sessions event. */
export type SessionsSource = {
  sessions(): Promise<State[] | null>;
  /** Calls apply with every list of sessions Go broadcasts and returns the call that stops listening. */
  listen(apply: (sessions: State[]) => void): () => void;
};

/** The open sessions, in the order they opened, kept in step with Go. */
export class SessionsStore extends Observed<readonly State[]> {
  private constructor(sessions: readonly State[]) {
    super(sessions);
  }

  /** Reads the sessions and follows every later change. */
  static async open(source: SessionsSource): Promise<SessionsStore> {
    const broadcasts = new Broadcasts<readonly State[]>(source.listen);
    const store = new SessionsStore(broadcasts.latestOr((await source.sessions()) ?? []));
    broadcasts.follow((sessions) => store.update(sessions));
    return store;
  }

  session(id: string | undefined): State | undefined {
    return this.current().find((session) => session.id === id);
  }
}
