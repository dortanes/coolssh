import { Observed } from '../observed.ts';

type Selected = {
  id: string | undefined;
  /** Whether the workspace shows the item's settings; a click only selects it. */
  editing: boolean;
  /** Whether the item was just created, so its editor puts the cursor in its name. */
  fresh: boolean;
};

/** The host or folder selected in the main window, and whether its settings are open. */
export class Selection extends Observed<Selected> {
  constructor() {
    super({ id: undefined, editing: false, fresh: false });
  }

  select(id: string | undefined): void {
    this.set({ id, editing: false, fresh: false });
  }

  /** Selects an item and opens its settings; fresh puts the cursor in its name. */
  edit(id: string, fresh = false): void {
    this.set({ id, editing: true, fresh });
  }

  private set(next: Selected): void {
    const now = this.current();
    if (next.id !== now.id || next.editing !== now.editing || next.fresh !== now.fresh) {
      this.update(next);
    }
  }
}
