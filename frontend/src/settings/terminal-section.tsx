import { type ReactNode, useId } from 'react';
import { Cursor, type Terminal } from '#bindings/preferences/models.ts';
import { Switch } from '@/components/ui/switch';
import { m } from '@/paraglide/messages.js';
import { getLocale } from '@/paraglide/runtime.js';
import type { PreferencesStore } from '../preferences-store.ts';
import { useStore } from '../use-store.ts';
import { type Choice, ChoiceSelect } from './choice-select.tsx';
import { SettingRow } from './setting-row.tsx';

/** The font sizes offered, CSS pixels, within the bounds Go accepts. */
const fontSizes = [10, 11, 12, 13, 14, 15, 16, 18, 20];

/** The scrollback lengths offered, within the bounds Go accepts. */
const scrollbacks = [1_000, 5_000, 10_000, 50_000, 100_000];

function Group({ title, children }: { title: string; children: ReactNode }) {
  const id = useId();
  return (
    <section aria-labelledby={id} className="mt-6 first-of-type:mt-4">
      <h2 id={id} className="px-1 text-xs font-medium text-muted-foreground">
        {title}
      </h2>
      <div data-surface="group" className="mt-2 divide-y rounded-lg border bg-card">
        {children}
      </div>
    </section>
  );
}

function SwitchRow({
  label,
  hint,
  checked,
  onChange,
}: {
  label: string;
  hint?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  const id = useId();
  return (
    <SettingRow labelId={id} label={label} hint={hint}>
      <Switch aria-labelledby={id} checked={checked} onCheckedChange={onChange} />
    </SettingRow>
  );
}

/** How session terminals look and behave, and whether ending a session asks first. */
export function TerminalSection({
  preferences: store,
  platform,
}: {
  preferences: PreferencesStore;
  platform: string;
}) {
  const { snapshot, saveFailed } = useStore(store);
  const { terminal, confirmDisconnect } = snapshot.preferences;
  const ids = { size: useId(), cursor: useId(), scrollback: useId() };
  const change = (next: Partial<Terminal>) =>
    void store.change({ terminal: { ...terminal, ...next } });
  const styleSize = Number.parseFloat(
    getComputedStyle(document.documentElement).getPropertyValue('--terminal-font-size'),
  );
  const sizes: Choice<string>[] = [
    { value: '0', label: m.settings_font_size_default({ size: styleSize }) },
    ...fontSizes.map((size) => ({
      value: String(size),
      label: m.settings_font_size_value({ size }),
    })),
  ];
  const cursors: Choice<Cursor>[] = [
    { value: Cursor.BlockCursor, label: m.settings_cursor_block() },
    { value: Cursor.BarCursor, label: m.settings_cursor_bar() },
    { value: Cursor.UnderlineCursor, label: m.settings_cursor_underline() },
  ];
  const lines = new Intl.NumberFormat(getLocale());
  const scrollbackChoices: Choice<string>[] = scrollbacks.map((count) => ({
    value: String(count),
    label: m.settings_scrollback_lines({ count: lines.format(count) }),
  }));
  return (
    <section aria-labelledby="terminal-title" className="w-full max-w-2xl px-8 pb-8">
      <h1 id="terminal-title" className="text-lg font-semibold">
        {m.settings_terminal()}
      </h1>
      <Group title={m.settings_group_text()}>
        <SettingRow labelId={ids.size} label={m.settings_font_size()}>
          <ChoiceSelect
            labelId={ids.size}
            choices={sizes}
            value={String(terminal.fontSize)}
            onChoose={(size) => change({ fontSize: Number(size) })}
          />
        </SettingRow>
        <SettingRow labelId={ids.cursor} label={m.settings_cursor()}>
          <ChoiceSelect
            labelId={ids.cursor}
            choices={cursors}
            value={terminal.cursor}
            onChoose={(cursor) => change({ cursor })}
          />
        </SettingRow>
        <SwitchRow
          label={m.settings_cursor_blink()}
          checked={terminal.cursorBlink}
          onChange={(cursorBlink) => change({ cursorBlink })}
        />
      </Group>
      <Group title={m.settings_group_behaviour()}>
        <SwitchRow
          label={m.settings_copy_on_select()}
          hint={m.settings_copy_on_select_hint()}
          checked={terminal.copyOnSelect}
          onChange={(copyOnSelect) => change({ copyOnSelect })}
        />
        {platform === 'darwin' && (
          <SwitchRow
            label={m.settings_option_as_meta()}
            hint={m.settings_option_as_meta_hint()}
            checked={terminal.optionAsMeta}
            onChange={(optionAsMeta) => change({ optionAsMeta })}
          />
        )}
        <SettingRow labelId={ids.scrollback} label={m.settings_scrollback()}>
          <ChoiceSelect
            labelId={ids.scrollback}
            choices={scrollbackChoices}
            value={String(terminal.scrollback)}
            onChoose={(count) => change({ scrollback: Number(count) })}
          />
        </SettingRow>
      </Group>
      <Group title={m.settings_group_sessions()}>
        <SwitchRow
          label={m.settings_confirm_disconnect()}
          hint={m.settings_confirm_disconnect_hint()}
          checked={confirmDisconnect}
          onChange={(confirm) => void store.change({ confirmDisconnect: confirm })}
        />
      </Group>
      <Group title={m.settings_group_performance()}>
        <SwitchRow
          label={m.settings_hardware_acceleration()}
          hint={m.settings_hardware_acceleration_hint()}
          checked={terminal.hardwareAcceleration}
          onChange={(hardwareAcceleration) => change({ hardwareAcceleration })}
        />
      </Group>
      {saveFailed && (
        <p role="alert" className="mt-3 text-sm text-destructive">
          {m.settings_not_saved()}
        </p>
      )}
    </section>
  );
}
