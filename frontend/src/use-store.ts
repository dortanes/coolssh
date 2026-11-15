import { useSyncExternalStore } from 'react';

export type Observable<T> = {
  subscribe(listener: () => void): () => void;
  current(): T;
};

/** The current state of a store, re-rendering when it changes. */
export function useStore<T>(store: Observable<T>): T {
  return useSyncExternalStore(store.subscribe, store.current);
}
