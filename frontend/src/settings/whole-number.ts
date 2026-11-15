import type { Settings } from '#bindings/library/models.ts';
import type { LibraryStore } from '../library/library-store.ts';

/** A whole number typed into a field: null when empty, undefined when it is not one. */
function wholeNumber(text: string): number | null | undefined {
  const trimmed = text.trim();
  if (trimmed === '') {
    return null;
  }
  return /^\d+$/.test(trimmed) ? Number(trimmed) : undefined;
}

/**
 * Saves a numeric setting typed as text through save: empty clears it, and text that is not a
 * whole number shows refusal instead.
 */
export function saveWholeNumber(
  library: LibraryStore,
  save: (change: Partial<Settings>) => void,
): (field: 'port' | 'keepaliveSeconds', text: string, refusal: string) => void {
  return (field, text, refusal) => {
    const value = wholeNumber(text);
    if (value === undefined) {
      library.report(refusal);
    } else {
      save({ [field]: value });
    }
  };
}
