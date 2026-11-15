import { cn } from 'cn';
import { ChevronDownIcon } from 'lucide-react';
import { useId, useState } from 'react';
import {
  ClearSourceToken,
  SetSource,
  SetSourceToken,
  SourceGroups,
} from '#bindings/desktop/service.ts';
import type { Group } from '#bindings/hostsource/models.ts';
import type { Folder, Source } from '#bindings/library/models.ts';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { m } from '@/paraglide/messages.js';
import { type LibraryStore, refusal } from '../library/library-store.ts';
import { sourceName } from '../library/source-names.ts';
import { sourceStatusText, troubled } from '../library/source-status.ts';
import type { SourcesStore } from '../library/sources-store.ts';
import { type Choice, ChoiceSelect } from '../settings/choice-select.tsx';
import { SettingRow } from '../settings/setting-row.tsx';
import { useStore } from '../use-store.ts';
import { SecretSetting } from './secret-setting.tsx';

/** The select option standing for a folder the user fills. */
const off = 'off';

type SourceSettingsProps = {
  folder: Folder;
  library: LibraryStore;
  sources: SourcesStore;
};

/** Where a folder takes its hosts from, the token and groups that narrow them, and the sync's state. */
export function SourceSettings({ folder, library, sources }: SourceSettingsProps) {
  const { available, states } = useStore(sources);
  const source = folder.source ?? undefined;
  const ids = { source: useId(), token: useId(), groups: useId() };
  const choices: Choice<string>[] = [
    { value: off, label: m.source_off() },
    ...available.map(({ kind, installed }) => ({
      value: kind,
      label: installed ? sourceName(kind) : m.source_not_installed({ source: sourceName(kind) }),
      disabled: !installed && kind !== source?.kind,
    })),
  ];
  const state = states[folder.id];
  return (
    <>
      <div data-surface="group" className="divide-y rounded-lg border bg-card">
        <SettingRow labelId={ids.source} label={m.source_title()} hint={m.source_hint()}>
          <ChoiceSelect
            labelId={ids.source}
            choices={choices}
            value={source?.kind ?? off}
            onChoose={(kind) =>
              void library.run(() => SetSource(folder.id, kind === off ? '' : kind, null))
            }
          />
        </SettingRow>
        {source && (
          <SettingRow labelId={ids.token} label={m.source_token()} hint={m.source_token_hint()}>
            <SecretSetting
              labelId={ids.token}
              saved={source.token === true}
              placeholder={m.source_token_placeholder()}
              onSave={(token) =>
                library.run(() => SetSourceToken(folder.id, token).then(() => true))
              }
              onForget={() => void library.run(() => ClearSourceToken(folder.id))}
            />
          </SettingRow>
        )}
        {source && (
          <SettingRow
            labelId={ids.groups}
            label={m.source_groups()}
            hint={source.token ? undefined : m.source_groups_need_token()}
          >
            {source.token && (
              <GroupPicker
                labelId={ids.groups}
                folder={folder.id}
                source={source}
                library={library}
              />
            )}
          </SettingRow>
        )}
      </div>
      {source && (
        <p
          role="status"
          className={cn(
            'mt-2 px-1 text-xs',
            troubled(state) ? 'text-destructive' : 'text-muted-foreground',
          )}
        >
          {sourceStatusText(source.kind, state)}
        </p>
      )}
    </>
  );
}

/** The groups a synced folder is narrowed to, chosen from the source's groups when opened. */
function GroupPicker({
  labelId,
  folder,
  source,
  library,
}: {
  labelId: string;
  folder: string;
  source: Source;
  library: LibraryStore;
}) {
  const [groups, setGroups] = useState<Group[] | undefined>();
  const [problem, setProblem] = useState<string | undefined>();
  const chosen = source.groups ?? [];
  const load = async () => {
    try {
      setGroups((await SourceGroups(folder)) ?? []);
      setProblem(undefined);
    } catch (error) {
      setProblem(refusal(error));
    }
  };
  const choose = (next: string[]) =>
    void library.run(() => SetSource(folder, source.kind, next.length > 0 ? next : null));
  return (
    <DropdownMenu onOpenChange={(open) => open && void load()}>
      <DropdownMenuTrigger
        render={
          <Button
            variant="secondary"
            aria-labelledby={labelId}
            className="min-w-44 justify-between"
          />
        }
      >
        {chosen.length === 0
          ? m.source_groups_all()
          : m.source_groups_count({ count: chosen.length })}
        <ChevronDownIcon aria-hidden />
      </DropdownMenuTrigger>
      <DropdownMenuContent data-surface="menu" align="end">
        {problem && <DropdownMenuItem disabled>{problem}</DropdownMenuItem>}
        {groups?.map((group) => (
          <DropdownMenuCheckboxItem
            key={group.id}
            checked={chosen.includes(group.id)}
            onCheckedChange={(checked) =>
              choose(checked ? [...chosen, group.id] : chosen.filter((id) => id !== group.id))
            }
          >
            {group.name}
          </DropdownMenuCheckboxItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
