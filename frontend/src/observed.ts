/** State a window shows and re-renders on: its current value and who to tell when it changes. */
export abstract class Observed<T> {
  private state: T;
  private readonly listeners = new Set<() => void>();

  protected constructor(state: T) {
    this.state = state;
  }

  readonly subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  readonly current = (): T => this.state;

  protected update(state: T): void {
    this.state = state;
    for (const listener of this.listeners) {
      listener();
    }
  }
}
