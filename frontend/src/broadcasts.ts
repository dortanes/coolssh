/**
 * What Go broadcasts to a store, listened to before the store reads its first value, so a broadcast
 * sent while that read is pending is not lost.
 */
export class Broadcasts<T> {
  private latest: { value: T } | undefined;
  private apply: ((value: T) => void) | undefined;

  constructor(listen: (apply: (value: T) => void) => unknown) {
    listen((value) => {
      if (this.apply) {
        this.apply(value);
      } else {
        this.latest = { value };
      }
    });
  }

  /** The value last broadcast before follow, else read. */
  latestOr(read: T): T {
    return this.latest ? this.latest.value : read;
  }

  /** Passes every later broadcast to apply. */
  follow(apply: (value: T) => void): void {
    this.apply = apply;
  }
}
