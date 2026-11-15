import { Events } from '@wailsio/runtime';
import { SourceStates, Sources } from '#bindings/desktop/service.ts';
import type { SourcesSource } from './sources-store.ts';

/** The host sources as Go keeps them: the bound service and the event it broadcasts. */
export const goSources: SourcesSource = {
  available: Sources,
  states: SourceStates,
  listen: (apply) => Events.On('sources', (event) => apply(event.data ?? {})),
};
