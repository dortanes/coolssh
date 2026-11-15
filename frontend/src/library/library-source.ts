import { Events } from '@wailsio/runtime';
import { Library } from '#bindings/desktop/service.ts';
import type { LibrarySource } from './library-store.ts';

/** The library as Go keeps it: the bound service and the event it broadcasts. */
export const goLibrary: LibrarySource = {
  library: Library,
  listen: (apply) => Events.On('library', (event) => apply(event.data)),
};
