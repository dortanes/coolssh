import { useEffect, useRef } from 'react';
import { tinykeys } from 'tinykeys';
import type { Snapshot } from '#bindings/desktop/models.ts';
import { Action } from '#bindings/preferences/models.ts';

/**
 * Runs each handler on its action's shortcut anywhere in the window. The listener captures, so a
 * shortcut reaches the window before a focused terminal, which then never sees the key.
 */
export function useShortcuts(
  shortcuts: Snapshot['shortcuts'],
  handlers: { [_ in Action]?: () => void },
): void {
  const current = useRef(handlers);
  current.current = handlers;
  const handled = Object.values(Action)
    .filter((action) => handlers[action] !== undefined)
    .join();
  // biome-ignore lint/correctness/useExhaustiveDependencies: handled names the actions the ref's handlers cover
  useEffect(() => {
    const bindings: Record<string, (event: KeyboardEvent) => void> = {};
    for (const action of Object.values(Action)) {
      const shortcut = shortcuts?.[action];
      if (shortcut && current.current[action]) {
        bindings[shortcut] = (event) => {
          event.preventDefault();
          event.stopPropagation();
          current.current[action]?.();
        };
      }
    }
    // Every shortcut holds Ctrl, Alt or Meta, so none is typing into a field.
    return tinykeys(window, bindings, { capture: true, ignore: () => false });
  }, [shortcuts, handled]);
}
