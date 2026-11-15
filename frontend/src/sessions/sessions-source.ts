import { Events } from '@wailsio/runtime';
import { Sessions } from '#bindings/desktop/service.ts';
import type { SessionsSource } from './sessions-store.ts';

/** The sessions as Go keeps them: the bound service and the event it broadcasts. */
export const goSessions: SessionsSource = {
  sessions: Sessions,
  listen: (apply) => Events.On('sessions', (event) => apply(event.data ?? [])),
};
