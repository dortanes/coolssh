import { useId } from 'react';
import { Rename, SetAddress, SetTags } from '#bindings/desktop/service.ts';
import type { Host } from '#bindings/library/models.ts';
import { m } from '@/paraglide/messages.js';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import { LinkDot, linkText } from '../library/peer-link.tsx';
import { linkOf, type SourcesStore } from '../library/sources-store.ts';
import { SettingRow } from '../settings/setting-row.tsx';
import { useStore } from '../use-store.ts';
import { ConnectionFields } from './connection-fields.tsx';
import { EditorPage } from './editor-page.tsx';
import { TagsSetting } from './tags-setting.tsx';
import { TextSetting } from './text-setting.tsx';

/**
 * A host's name, address and connection settings, saved as they change. A synced host's name and
 * address come from its source and are shown, not edited.
 */
export function HostEditor({
  host,
  fresh,
  library,
  sources,
}: {
  host: Host;
  fresh: boolean;
  library: LibraryStore;
  sources: SourcesStore;
}) {
  const ids = { name: useId(), address: useId(), link: useId(), tags: useId() };
  const synced = Boolean(host.peer);
  const { states } = useStore(sources);
  const link = linkOf(states, new LibraryView(useStore(library).library).peerOf(host.id));
  return (
    <EditorPage title={host.name}>
      <div data-surface="group" className="divide-y rounded-lg border bg-card">
        <SettingRow
          labelId={ids.name}
          label={m.editor_name()}
          hint={synced ? m.library_synced() : undefined}
        >
          {synced ? (
            <span className="w-60 shrink-0 truncate text-sm select-text">{host.name}</span>
          ) : (
            <TextSetting
              labelId={ids.name}
              value={host.name}
              focus={fresh}
              onCommit={(name) => void library.run(() => Rename(host.id, name))}
            />
          )}
        </SettingRow>
        <SettingRow labelId={ids.address} label={m.editor_address()}>
          {synced ? (
            <span className="w-60 shrink-0 truncate font-mono text-sm select-text">
              {host.address}
            </span>
          ) : (
            <TextSetting
              labelId={ids.address}
              value={host.address}
              placeholder={m.editor_address_placeholder()}
              onCommit={(address) => void library.run(() => SetAddress(host.id, address))}
            />
          )}
        </SettingRow>
        <SettingRow labelId={ids.tags} label={m.editor_tags()} hint={m.editor_tags_hint()}>
          <TagsSetting
            labelId={ids.tags}
            tags={host.tags ?? []}
            onChange={(tags) => void library.run(() => SetTags(host.id, tags))}
          />
        </SettingRow>
        {link && (
          <SettingRow labelId={ids.link} label={m.source_link()}>
            <span className="flex w-60 shrink-0 items-center gap-2 text-sm">
              <LinkDot link={link} />
              <span className="truncate">{linkText(link)}</span>
            </span>
          </SettingRow>
        )}
      </div>
      <h2 className="mt-8 mb-3 text-sm font-semibold">{m.editor_connection()}</h2>
      <ConnectionFields id={host.id} settings={host.settings} library={library} />
    </EditorPage>
  );
}
