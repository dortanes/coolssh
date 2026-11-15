import { useId } from 'react';
import { SetDefaults } from '#bindings/desktop/service.ts';
import type { Settings } from '#bindings/library/models.ts';
import { m } from '@/paraglide/messages.js';
import type { LibraryStore } from '../library/library-store.ts';
import { useStore } from '../use-store.ts';
import { TextSetting } from '../workspace/text-setting.tsx';
import { ChoiceSelect } from './choice-select.tsx';
import { SettingRow } from './setting-row.tsx';
import { saveWholeNumber } from './whole-number.ts';

/** The select option standing for the built-in value. */
const builtIn = 'built-in';

/**
 * The defaults every host falls back to when neither it nor a folder above it sets a value; each
 * shows the built-in value until the user sets one, and an emptied field returns to it.
 */
export function ConnectionsSection({
  library,
  builtInDefaults,
}: {
  library: LibraryStore;
  builtInDefaults: Settings;
}) {
  const { library: snapshot } = useStore(library);
  const defaults = snapshot.defaults;
  const ids = { user: useId(), port: useId(), agent: useId(), keepalive: useId() };
  const save = (change: Partial<Settings>) =>
    void library.run(() => SetDefaults({ ...defaults, ...change }));
  const saveNumber = saveWholeNumber(library, save);
  const onOff = (value: boolean | null | undefined) => (value ? m.editor_on() : m.editor_off());
  return (
    <section aria-labelledby="connections-title" className="w-full max-w-2xl px-8 pb-8">
      <h1 id="connections-title" className="text-lg font-semibold">
        {m.settings_connections()}
      </h1>
      <p className="mt-1 text-sm text-muted-foreground">{m.settings_connections_hint()}</p>
      <div data-surface="group" className="mt-4 divide-y rounded-lg border bg-card">
        <SettingRow labelId={ids.user} label={m.editor_user()}>
          <TextSetting
            labelId={ids.user}
            value={defaults.user ?? ''}
            placeholder={builtInDefaults.user ?? undefined}
            onCommit={(user) => save({ user: user.trim() === '' ? null : user.trim() })}
          />
        </SettingRow>
        <SettingRow labelId={ids.port} label={m.editor_port()}>
          <TextSetting
            labelId={ids.port}
            value={defaults.port?.toString() ?? ''}
            placeholder={builtInDefaults.port?.toString()}
            inputMode="numeric"
            onCommit={(port) => saveNumber('port', port, m.library_port_range())}
          />
        </SettingRow>
        <SettingRow labelId={ids.agent} label={m.editor_agent_forwarding()}>
          <ChoiceSelect
            labelId={ids.agent}
            value={defaults.agentForwarding == null ? builtIn : String(defaults.agentForwarding)}
            choices={[
              {
                value: builtIn,
                label: m.settings_built_in({ value: onOff(builtInDefaults.agentForwarding) }),
              },
              { value: 'true', label: m.editor_on() },
              { value: 'false', label: m.editor_off() },
            ]}
            onChoose={(agent) =>
              save({ agentForwarding: agent === builtIn ? null : agent === 'true' })
            }
          />
        </SettingRow>
        <SettingRow labelId={ids.keepalive} label={m.editor_keepalive()}>
          <TextSetting
            labelId={ids.keepalive}
            value={defaults.keepaliveSeconds?.toString() ?? ''}
            placeholder={builtInDefaults.keepaliveSeconds?.toString()}
            inputMode="numeric"
            onCommit={(seconds) =>
              saveNumber('keepaliveSeconds', seconds, m.library_keepalive_range())
            }
          />
        </SettingRow>
      </div>
    </section>
  );
}
