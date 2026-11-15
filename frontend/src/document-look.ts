import { isLocale, setLocale } from './paraglide/runtime.js';
import type { PreferencesStore } from './preferences-store.ts';
import { themeFor } from './theme.ts';

/**
 * Keeps the document's theme and language in step with the preferences and the system's appearance;
 * it starts before the window renders, so a render reads its copy in the new language.
 */
export class DocumentLook {
  private readonly store: PreferencesStore;
  private readonly root: HTMLElement;
  private readonly systemDark: MediaQueryList;

  constructor(store: PreferencesStore, root: HTMLElement, systemDark: MediaQueryList) {
    this.store = store;
    this.root = root;
    this.systemDark = systemDark;
  }

  start(): void {
    this.apply();
    this.store.subscribe(() => this.apply());
    this.systemDark.addEventListener('change', () => this.apply());
  }

  private apply(): void {
    const { preferences, locale, look, platform } = this.store.current().snapshot;
    this.root.classList.toggle(
      'dark',
      themeFor(preferences.appearance, this.systemDark.matches) === 'dark',
    );
    this.root.setAttribute('data-style', look.style);
    this.root.setAttribute('data-material', String(look.material));
    this.root.setAttribute('data-platform', platform);
    if (isLocale(locale)) {
      setLocale(locale, { reload: false });
      this.root.lang = locale;
    }
  }
}
