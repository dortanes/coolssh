import { useId } from 'react';
import { ClearPassword, SetPassword, SetSettings } from '#bindings/desktop/service.ts';
import type { Settings } from '#bindings/library/models.ts';
import { m } from '@/paraglide/messages.js';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import { ChoiceSelect } from '../settings/choice-select.tsx';
import { SettingRow } from '../settings/setting-row.tsx';
import { saveWholeNumber } from '../settings/whole-number.ts';
import { useStore } from '../use-store.ts';
import { InheritedHint } from './inherited-hint.tsx';
import { KeySetting } from './key-setting.tsx';
import { SecretSetting } from './secret-setting.tsx';
import { TextSetting } from './text-setting.tsx';
import { useInherited } from './use-inherited.ts';

/** The select option standing for a setting left to inherit. */
const inherit = 'inherit';

/** A host's or folder's connection settings, each showing what it inherits while left unset. */
export function ConnectionFields({
  id,
  settings,
  library,
}: {
  id: string;
  settings: Settings;
  library: LibraryStore;
}) {
  const { library: snapshot } = useStore(library);
  const view = new LibraryView(snapshot);
  const inherited = useInherited(id, snapshot);
  const ids = {
    user: useId(),
    port: useId(),
    key: useId(),
    password: useId(),
    jump: useId(),
    agent: useId(),
    keepalive: useId(),
  };
  const save = (change: Partial<Settings>) =>
    void library.run(() => SetSettings(id, { ...settings, ...change }));
  const saveNumber = saveWholeNumber(library, save);
  const hint = (field: keyof Settings) => (
    <InheritedHint
      set={settings[field] != null}
      origin={inherited?.origins[field]}
      view={view}
      onInherit={() => save({ [field]: null })}
    />
  );
  const values = inherited?.values;
  const jumpHosts = view.hosts().filter((host) => host.id !== id);
  const inheritedJump = values?.jumpHostId ? view.host(values.jumpHostId)?.name : m.editor_none();
  return (
    <div data-surface="group" className="divide-y rounded-lg border bg-card">
      <SettingRow labelId={ids.user} label={m.editor_user()} hint={hint('user')}>
        <TextSetting
          labelId={ids.user}
          value={settings.user ?? ''}
          placeholder={values?.user ?? undefined}
          onCommit={(user) => save({ user: user.trim() === '' ? null : user.trim() })}
        />
      </SettingRow>
      <SettingRow labelId={ids.port} label={m.editor_port()} hint={hint('port')}>
        <TextSetting
          labelId={ids.port}
          value={settings.port?.toString() ?? ''}
          placeholder={values?.port?.toString()}
          inputMode="numeric"
          onCommit={(port) => saveNumber('port', port, m.library_port_range())}
        />
      </SettingRow>
      <SettingRow labelId={ids.key} label={m.editor_key()} hint={hint('keyId')}>
        <KeySetting
          labelId={ids.key}
          value={settings.keyId}
          inherited={values?.keyId ?? ''}
          library={library}
          onChoose={(keyId) => save({ keyId })}
        />
      </SettingRow>
      <SettingRow labelId={ids.password} label={m.editor_password()} hint={hint('password')}>
        <SecretSetting
          labelId={ids.password}
          saved={settings.password === true}
          placeholder={m.editor_password_placeholder()}
          onSave={(password) => library.run(() => SetPassword(id, password).then(() => true))}
          onForget={() => void library.run(() => ClearPassword(id))}
        />
      </SettingRow>
      <SettingRow labelId={ids.jump} label={m.editor_jump_host()} hint={hint('jumpHostId')}>
        <ChoiceSelect
          labelId={ids.jump}
          value={settings.jumpHostId ?? inherit}
          choices={[
            { value: inherit, label: m.editor_inherit_value({ value: inheritedJump ?? '' }) },
            { value: '', label: m.editor_none() },
            ...jumpHosts.map((host) => ({ value: host.id, label: host.name })),
          ]}
          onChoose={(jump) => save({ jumpHostId: jump === inherit ? null : jump })}
        />
      </SettingRow>
      <SettingRow
        labelId={ids.agent}
        label={m.editor_agent_forwarding()}
        hint={hint('agentForwarding')}
      >
        <ChoiceSelect
          labelId={ids.agent}
          value={settings.agentForwarding == null ? inherit : String(settings.agentForwarding)}
          choices={[
            {
              value: inherit,
              label: m.editor_inherit_value({
                value: values?.agentForwarding ? m.editor_on() : m.editor_off(),
              }),
            },
            { value: 'true', label: m.editor_on() },
            { value: 'false', label: m.editor_off() },
          ]}
          onChoose={(agent) =>
            save({ agentForwarding: agent === inherit ? null : agent === 'true' })
          }
        />
      </SettingRow>
      <SettingRow
        labelId={ids.keepalive}
        label={m.editor_keepalive()}
        hint={hint('keepaliveSeconds')}
      >
        <TextSetting
          labelId={ids.keepalive}
          value={settings.keepaliveSeconds?.toString() ?? ''}
          placeholder={values?.keepaliveSeconds?.toString()}
          inputMode="numeric"
          onCommit={(seconds) =>
            saveNumber('keepaliveSeconds', seconds, m.library_keepalive_range())
          }
        />
      </SettingRow>
    </div>
  );
}
