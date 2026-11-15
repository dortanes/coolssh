import type { State } from '#bindings/session/models.ts';
import { Observed } from '../observed.ts';
import type { Observable } from '../use-store.ts';
import {
  type Orientation,
  type PaneNode,
  sessionsIn,
  withoutSession,
  withSessionBeside,
} from './pane-tree.ts';

type TabsState = {
  /** The session whose pane has the focus, in the active tab; undefined while no tab is open. */
  id: string | undefined;
  /** Each tab's panes, in the order the tabs show. */
  layouts: readonly PaneNode[];
};

/** The main window's tabs, each one session's pane or several split; a new session opens its own. */
export class Tabs extends Observed<TabsState> {
  /** The sessions the last list held; a session placed before any list holds it is kept. */
  private known: Set<string>;
  /** Sessions most recently focused first, so a tab shown again focuses the pane it last did. */
  private recent: string[] = [];

  constructor(sessions: Observable<readonly State[]>) {
    const ids = sessions.current().map((session) => session.id);
    super({ id: undefined, layouts: ids.map((session) => ({ session })) });
    this.known = new Set(ids);
    sessions.subscribe(() => this.follow(sessions.current()));
  }

  /** Focuses a session's pane, showing its tab. */
  show(id: string): void {
    this.recent = [id, ...this.recent.filter((each) => each !== id)];
    if (id !== this.current().id) {
      this.update({ ...this.current(), id });
    }
  }

  /** The sessions in the tab holding id. */
  tabOf(id: string | undefined): string[] {
    const layout = this.current().layouts.find((each) => id && sessionsIn(each).includes(id));
    return layout ? sessionsIn(layout) : [];
  }

  /** The pane a tab focuses: the one it last focused, else its first. */
  focusIn(layout: PaneNode): string | undefined {
    const ids = sessionsIn(layout);
    return this.recent.find((id) => ids.includes(id)) ?? ids[0];
  }

  /** The session to focus offset tabs from the active one, wrapping around the ends. */
  beside(offset: number): string | undefined {
    const { layouts, id } = this.current();
    const index = layouts.findIndex((each) => id && sessionsIn(each).includes(id));
    const layout = layouts[wrap(index + offset, layouts.length)];
    return layout && this.focusIn(layout);
  }

  /** The session offset panes from the focused one in its tab, wrapping around the ends. */
  paneBeside(offset: number): string | undefined {
    const ids = this.tabOf(this.current().id);
    return ids[wrap(ids.indexOf(this.current().id ?? '') + offset, ids.length)];
  }

  /** Moves a session's pane, or a session not placed yet, beside another's and focuses it. */
  place(id: string, beside: string, orientation: Orientation): void {
    if (id === beside) {
      return;
    }
    const layouts = this.current().layouts.flatMap((layout) => {
      const rest = withoutSession(layout, id);
      return rest ? [withSessionBeside(rest, beside, id, orientation)] : [];
    });
    this.update({ ...this.current(), layouts });
    this.show(id);
  }

  /** Moves a session's pane out of its split into a tab of its own, after the one it was in. */
  separate(id: string): void {
    const { layouts } = this.current();
    const index = layouts.findIndex((each) => sessionsIn(each).includes(id));
    const rest = layouts[index] && withoutSession(layouts[index], id);
    if (!rest) {
      return;
    }
    this.update({
      ...this.current(),
      layouts: [...layouts.slice(0, index), rest, { session: id }, ...layouts.slice(index + 1)],
    });
    this.show(id);
  }

  /**
   * When the focused session goes, a pane left in its tab takes the focus, else the tab that took
   * its place, else the one before it.
   */
  private follow(sessions: readonly State[]): void {
    const ids = new Set(sessions.map((session) => session.id));
    const gone = [...this.known].filter((id) => !ids.has(id));
    this.known = ids;
    this.recent = this.recent.filter((id) => !gone.includes(id));
    const before = this.current();
    const mapped = before.layouts.map((layout) =>
      gone.reduce<PaneNode | undefined>((rest, id) => rest && withoutSession(rest, id), layout),
    );
    const kept = mapped.filter((layout) => layout !== undefined);
    const placed = new Set(kept.flatMap(sessionsIn));
    const layouts = [
      ...kept,
      ...[...ids].filter((id) => !placed.has(id)).map((session) => ({ session })),
    ];
    let { id } = before;
    if (id !== undefined && gone.includes(id)) {
      const index = before.layouts.findIndex((layout) => sessionsIn(layout).includes(id ?? ''));
      const sameTab = mapped[index];
      const keptBefore = mapped.slice(0, index).filter((layout) => layout !== undefined).length;
      const layout = sameTab ?? layouts[Math.min(keptBefore, layouts.length - 1)];
      id = layout && this.focusIn(layout);
    }
    this.update({ id, layouts });
  }
}

function wrap(index: number, count: number): number {
  return count === 0 ? -1 : ((index % count) + count) % count;
}
