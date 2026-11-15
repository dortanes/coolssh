import './styles/tailwind.css';
import type { ComponentType } from 'react';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { DocumentLook } from './document-look.ts';
import { goLibrary } from './library/library-source.ts';
import { LibraryStore } from './library/library-store.ts';
import { followProblems, Notices } from './notices.tsx';
import { goPreferences } from './preferences-source.ts';
import { PreferencesStore } from './preferences-store.ts';

/** The stores every window reads, each kept in step with Go. */
export type WindowStores = {
  preferences: PreferencesStore;
  library: LibraryStore;
};

/**
 * Opens the stores, applies the preferences' look to the document and renders a window's view with
 * its toasts.
 */
export async function mountWindow(View: ComponentType<WindowStores>): Promise<void> {
  const [preferences, library] = await Promise.all([
    PreferencesStore.open(goPreferences),
    LibraryStore.open(goLibrary),
  ]);
  new DocumentLook(
    preferences,
    document.documentElement,
    matchMedia('(prefers-color-scheme: dark)'),
  ).start();
  followProblems(library);
  const container = document.getElementById('root');
  if (!container) {
    throw new Error('the page has no #root element');
  }
  createRoot(container).render(
    <StrictMode>
      <View preferences={preferences} library={library} />
      <Notices />
    </StrictMode>,
  );
}
