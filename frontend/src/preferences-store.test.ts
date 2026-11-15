import assert from 'node:assert/strict';
import { test } from 'node:test';
import type { Snapshot } from '#bindings/desktop/models.ts';
import {
  Appearance,
  Cursor,
  Language,
  type Preferences,
  Style,
} from '#bindings/preferences/models.ts';
import { type PreferencesSource, PreferencesStore } from './preferences-store.ts';

const english: Snapshot = {
  preferences: {
    language: Language.SystemLanguage,
    appearance: Appearance.SystemAppearance,
    style: Style.AutomaticStyle,
    sidebarWidth: 260,
    terminal: {
      fontSize: 0,
      cursor: Cursor.BlockCursor,
      cursorBlink: true,
      scrollback: 10_000,
      copyOnSelect: false,
      optionAsMeta: false,
      hardwareAcceleration: true,
    },
    confirmDisconnect: true,
  },
  locale: 'en',
  look: { style: Style.LiquidGlassStyle, material: true },
  platform: 'darwin',
  shortcuts: {},
};

/** A Go side that saves what it is given, or refuses every save, and broadcasts to listeners. */
class FakeGo implements PreferencesSource {
  saved: Preferences[] = [];
  private listener: ((snapshot: Snapshot) => void) | undefined;
  private readonly refuse: boolean;

  constructor(refuse = false) {
    this.refuse = refuse;
  }

  snapshot = async (): Promise<Snapshot> => english;

  save = async (preferences: Preferences): Promise<Snapshot> => {
    if (this.refuse) {
      throw new Error('the folder is read-only');
    }
    this.saved.push(preferences);
    return {
      ...english,
      preferences,
      locale: preferences.language === Language.Russian ? 'ru' : 'en',
    };
  };

  listen = (apply: (snapshot: Snapshot) => void): (() => void) => {
    this.listener = apply;
    return () => {
      this.listener = undefined;
    };
  };

  broadcast(snapshot: Snapshot): void {
    this.listener?.(snapshot);
  }
}

test('a change saves the whole preferences with the change applied', async () => {
  const go = new FakeGo();
  const store = await PreferencesStore.open(go);
  await store.change({ language: Language.Russian });
  assert.deepEqual(go.saved, [{ ...english.preferences, language: Language.Russian }]);
  assert.equal(store.current().snapshot.locale, 'ru');
});

test('a change made in another window reaches this one', async () => {
  const go = new FakeGo();
  const store = await PreferencesStore.open(go);
  let notified = 0;
  store.subscribe(() => {
    notified++;
  });
  const dark = {
    ...english,
    preferences: { ...english.preferences, appearance: Appearance.DarkAppearance },
  };
  go.broadcast(dark);
  assert.equal(notified, 1);
  assert.deepEqual(store.current().snapshot, dark);
});

test('a refused save keeps the preferences in effect and says so', async () => {
  const store = await PreferencesStore.open(new FakeGo(true));
  await store.change({ appearance: Appearance.DarkAppearance });
  assert.deepEqual(store.current(), { snapshot: english, saveFailed: true });
});

test('the sidebar width is saved only when a drag changes it', async () => {
  const go = new FakeGo();
  const store = await PreferencesStore.open(go);
  await store.rememberSidebarWidth(260.2);
  assert.equal(go.saved.length, 0);
  await store.rememberSidebarWidth(301.6);
  assert.equal(go.saved.at(-1)?.sidebarWidth, 302);
});
