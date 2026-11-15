import { RotateCcwIcon } from 'lucide-react';
import { useEffect, useId, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Kbd } from '@/components/ui/kbd';
import { m } from '@/paraglide/messages.js';
import type { PreferencesStore } from '../preferences-store.ts';
import { actionLabels, type ShortcutAction, shortcutActions } from '../shell/action-labels.ts';
import { assignShortcut, shortcutLabel, shortcutOf } from '../shell/shortcut.ts';
import { useStore } from '../use-store.ts';
import { SettingRow } from './setting-row.tsx';

/**
 * Every action's shortcut. Choosing one records the next key combination; Backspace removes it and
 * Escape keeps it.
 */
export function ShortcutsSection({ preferences: store }: { preferences: PreferencesStore }) {
  const { snapshot, saveFailed } = useStore(store);
  const own = snapshot.preferences.shortcuts ?? {};
  const inEffect = snapshot.shortcuts ?? {};
  const [recording, setRecording] = useState<ShortcutAction | undefined>();
  const save = (action: ShortcutAction, shortcut: string) =>
    void store.change({ shortcuts: assignShortcut(own, inEffect, action, shortcut) });
  const reset = (action: ShortcutAction) => {
    const { [action]: _, ...rest } = own;
    void store.change({ shortcuts: rest });
  };
  useEffect(() => {
    if (!recording) {
      return;
    }
    // Captured before the window's own Escape handling, which would close it.
    const record = (event: KeyboardEvent) => {
      event.preventDefault();
      event.stopPropagation();
      const plain = !event.ctrlKey && !event.altKey && !event.metaKey && !event.shiftKey;
      if (plain && event.key === 'Escape') {
        setRecording(undefined);
      } else if (plain && (event.key === 'Backspace' || event.key === 'Delete')) {
        setRecording(undefined);
        save(recording, '');
      } else {
        const shortcut = shortcutOf(event);
        if (shortcut) {
          setRecording(undefined);
          save(recording, shortcut);
        }
      }
    };
    window.addEventListener('keydown', record, { capture: true });
    return () => window.removeEventListener('keydown', record, { capture: true });
  });
  return (
    <section aria-labelledby="shortcuts-title" className="w-full max-w-2xl px-8 pb-8">
      <h1 id="shortcuts-title" className="text-lg font-semibold">
        {m.settings_shortcuts()}
      </h1>
      <p className="mt-1 text-sm text-muted-foreground">{m.settings_shortcuts_hint()}</p>
      <div data-surface="group" className="mt-4 divide-y rounded-lg border bg-card">
        {shortcutActions.map((action) => (
          <ShortcutRow
            key={action}
            label={actionLabels[action]()}
            shortcut={inEffect[action] ?? ''}
            platform={snapshot.platform}
            recording={recording === action}
            changed={own[action] !== undefined}
            onRecord={() => setRecording(recording === action ? undefined : action)}
            onReset={() => reset(action)}
          />
        ))}
      </div>
      {saveFailed && (
        <p role="alert" className="mt-3 text-sm text-destructive">
          {m.settings_not_saved()}
        </p>
      )}
    </section>
  );
}

function ShortcutRow({
  label,
  shortcut,
  platform,
  recording,
  changed,
  onRecord,
  onReset,
}: {
  label: string;
  shortcut: string;
  platform: string;
  recording: boolean;
  changed: boolean;
  onRecord: () => void;
  onReset: () => void;
}) {
  const id = useId();
  return (
    <SettingRow labelId={id} label={label}>
      <div className="flex items-center gap-1">
        {changed && (
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={m.settings_shortcut_reset()}
            title={m.settings_shortcut_reset()}
            onClick={onReset}
          >
            <RotateCcwIcon aria-hidden />
          </Button>
        )}
        <Button
          variant="secondary"
          size="sm"
          aria-labelledby={id}
          aria-pressed={recording}
          className="min-w-28 justify-center"
          onClick={onRecord}
        >
          {recording ? (
            m.settings_shortcut_recording()
          ) : shortcut ? (
            <Kbd className="bg-transparent text-foreground">
              {shortcutLabel(shortcut, platform)}
            </Kbd>
          ) : (
            <span className="text-muted-foreground">{m.settings_shortcut_none()}</span>
          )}
        </Button>
      </div>
    </SettingRow>
  );
}
