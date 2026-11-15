import { useState } from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { m } from '@/paraglide/messages.js';
import { addKeyFromFile } from '../library/key-actions.ts';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import { PasteKeyDialog } from '../library/paste-key-dialog.tsx';
import { useStore } from '../use-store.ts';

/** The options that are not keys: inheriting, and adding a key from a file or by pasting. */
const inherit = 'inherit';
const chooseFile = 'choose-file';
const paste = 'paste';

/**
 * The key a host or folder uses: inherited, none, one of the library's, or a new one added from a
 * file or pasted, which it then uses.
 */
export function KeySetting({
  labelId,
  value,
  inherited,
  library,
  onChoose,
}: {
  labelId: string;
  /** The key set here: an id, empty for none, null or absent to inherit. */
  value: string | null | undefined;
  /** The key inherited: an id, empty for none. */
  inherited: string;
  library: LibraryStore;
  onChoose: (keyId: string | null) => void;
}) {
  const view = new LibraryView(useStore(library).library);
  const [pasting, setPasting] = useState(false);
  const keyName = (id: string) =>
    id === '' ? m.editor_no_key() : (view.keys().find((key) => key.id === id)?.name ?? '');
  const items = [
    { value: inherit, label: m.editor_inherit_value({ value: keyName(inherited) }) },
    { value: '', label: m.editor_no_key() },
    ...view.keys().map((key) => ({ value: key.id, label: key.name })),
    { value: chooseFile, label: m.editor_choose_file() },
    { value: paste, label: m.editor_paste_key() },
  ];
  const choose = async (picked: unknown) => {
    if (picked === chooseFile) {
      const added = await addKeyFromFile(library);
      if (added) {
        onChoose(added);
      }
    } else if (picked === paste) {
      setPasting(true);
    } else if (typeof picked === 'string') {
      onChoose(picked === inherit ? null : picked);
    }
  };
  return (
    <>
      <Select
        items={items}
        value={value ?? inherit}
        onValueChange={(picked) => void choose(picked)}
      >
        <SelectTrigger aria-labelledby={labelId} className="min-w-44">
          <SelectValue />
        </SelectTrigger>
        <SelectContent data-surface="menu">
          {items.slice(0, -2).map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
          <SelectSeparator />
          <SelectItem value={chooseFile}>{m.editor_choose_file()}</SelectItem>
          <SelectItem value={paste}>{m.editor_paste_key()}</SelectItem>
        </SelectContent>
      </Select>
      <PasteKeyDialog open={pasting} onOpenChange={setPasting} onAdded={onChoose} />
    </>
  );
}
