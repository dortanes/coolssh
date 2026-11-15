import { useId } from 'react';
import { Rename, SetHostIcon, SetIcon } from '#bindings/desktop/service.ts';
import { type Folder, Icon } from '#bindings/library/models.ts';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { m } from '@/paraglide/messages.js';
import { folderIcons, hostChoice, hostIcons } from '../library/library-icons.tsx';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import type { SourcesStore } from '../library/sources-store.ts';
import { ChoiceSelect } from '../settings/choice-select.tsx';
import { SettingRow } from '../settings/setting-row.tsx';
import { useStore } from '../use-store.ts';
import { ConnectionFields } from './connection-fields.tsx';
import { EditorPage } from './editor-page.tsx';
import { SourceSettings } from './source-settings.tsx';
import { TextSetting } from './text-setting.tsx';

const inherit = 'inherit';

/** A folder's name, icon, its hosts' icon and the connection settings what it holds inherits. */
export function FolderEditor({
  folder,
  fresh,
  library,
  sources,
}: {
  folder: Folder;
  fresh: boolean;
  library: LibraryStore;
  sources: SourcesStore;
}) {
  const ids = { name: useId(), icon: useId(), hostIcon: useId() };
  const view = new LibraryView(useStore(library).library);
  const inherited = hostChoice(view.hostIconIn(view.parentOf(folder.id)));
  return (
    <EditorPage title={folder.name}>
      <div data-surface="group" className="divide-y rounded-lg border bg-card">
        <SettingRow labelId={ids.name} label={m.editor_name()}>
          <TextSetting
            labelId={ids.name}
            value={folder.name}
            focus={fresh}
            onCommit={(name) => void library.run(() => Rename(folder.id, name))}
          />
        </SettingRow>
        <SettingRow labelId={ids.icon} label={m.library_icon()}>
          <ToggleGroup
            aria-labelledby={ids.icon}
            value={[folder.icon]}
            className="grid w-60 grid-cols-6 gap-1"
            onValueChange={([pressed]) => {
              const choice = folderIcons.find((option) => option.icon === pressed);
              if (choice) {
                void library.run(() => SetIcon(folder.id, choice.icon));
              }
            }}
          >
            {folderIcons.map(({ icon, Glyph, label }) => (
              <ToggleGroupItem key={icon} value={icon} aria-label={label()} title={label()}>
                <Glyph aria-hidden />
              </ToggleGroupItem>
            ))}
          </ToggleGroup>
        </SettingRow>
        <SettingRow labelId={ids.hostIcon} label={m.editor_host_icon()}>
          <ChoiceSelect<Icon | typeof inherit>
            labelId={ids.hostIcon}
            value={folder.hostIcon || inherit}
            choices={[
              {
                value: inherit,
                label: m.editor_inherit_value({ value: inherited.label() }),
                Glyph: inherited.Glyph,
              },
              ...hostIcons.map(({ icon, Glyph, label }) => ({
                value: icon,
                label: label(),
                Glyph,
              })),
            ]}
            onChoose={(icon) =>
              void library.run(() => SetHostIcon(folder.id, icon === inherit ? Icon.$zero : icon))
            }
          />
        </SettingRow>
      </div>
      <h2 className="mt-8 mb-3 text-sm font-semibold">{m.source_title()}</h2>
      <SourceSettings folder={folder} library={library} sources={sources} />
      <h2 className="mt-8 mb-1 text-sm font-semibold">{m.editor_connection()}</h2>
      <p className="mb-3 text-xs text-muted-foreground">{m.editor_folder_connection()}</p>
      <ConnectionFields id={folder.id} settings={folder.settings} library={library} />
    </EditorPage>
  );
}
