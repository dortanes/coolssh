import { Button } from '@/components/ui/button';
import { m } from '@/paraglide/messages.js';
import type { LibraryView } from '../library/library-view.ts';

/**
 * Under a setting's name: where an unset setting comes from, or, once set, the way back to
 * inheriting.
 */
export function InheritedHint({
  set,
  origin,
  view,
  onInherit,
}: {
  set: boolean;
  /** The folder the value comes from, empty for the default. */
  origin: string | undefined;
  view: LibraryView;
  onInherit: () => void;
}) {
  if (set) {
    return (
      <Button variant="link" size="xs" className="h-auto p-0 text-xs" onClick={onInherit}>
        {m.editor_inherit()}
      </Button>
    );
  }
  if (origin === undefined) {
    return null;
  }
  return origin === ''
    ? m.editor_default()
    : m.editor_from({ folder: view.folder(origin)?.name ?? '' });
}
