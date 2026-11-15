import { Events } from '@wailsio/runtime';
import { SavePreferences, Snapshot } from '#bindings/desktop/service.ts';
import type { PreferencesSource } from './preferences-store.ts';

/** The preferences as Go keeps them: the bound service and the event it broadcasts. */
export const goPreferences: PreferencesSource = {
  snapshot: Snapshot,
  save: SavePreferences,
  listen: (apply) => Events.On('preferences', (event) => apply(event.data)),
};
