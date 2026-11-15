import type { Terminal as TerminalSettings } from '#bindings/preferences/models.ts';
import type { State } from '#bindings/session/models.ts';
import type { Observable } from '../use-store.ts';
import { SessionTerminal } from './session-terminal.ts';

/**
 * Every open session's terminal, kept while its session is open whichever pane shows it, and
 * disposed once the session goes.
 */
export class SessionTerminals {
  private readonly terminals = new Map<string, SessionTerminal>();
  private readonly platform: string;
  private settings: TerminalSettings;

  constructor(
    sessions: Observable<readonly State[]>,
    platform: string,
    settings: TerminalSettings,
  ) {
    this.platform = platform;
    this.settings = settings;
    sessions.subscribe(() => {
      const open = new Set(sessions.current().map((session) => session.id));
      for (const [id, terminal] of this.terminals) {
        if (!open.has(id)) {
          terminal.dispose();
          this.terminals.delete(id);
        }
      }
    });
  }

  /** A session's terminal, opened on first use. */
  terminal(id: string): SessionTerminal {
    let terminal = this.terminals.get(id);
    if (!terminal) {
      terminal = new SessionTerminal(id, this.platform, this.settings);
      this.terminals.set(id, terminal);
    }
    return terminal;
  }

  /** Applies the user's terminal settings to every terminal. */
  configure(settings: TerminalSettings): void {
    this.settings = settings;
    for (const terminal of this.terminals.values()) {
      terminal.configure(settings);
    }
  }
}
