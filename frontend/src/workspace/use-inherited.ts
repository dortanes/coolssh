import { useEffect, useState } from 'react';
import { Inherited } from '#bindings/desktop/service.ts';
import type { Inherited as InheritedSettings, Library } from '#bindings/library/models.ts';

/**
 * What a host or folder would inherit for each setting, asked of Go again each time the library
 * changes; undefined until Go answers.
 */
export function useInherited(id: string, library: Library): InheritedSettings | undefined {
  const [inherited, setInherited] = useState<InheritedSettings>();
  // biome-ignore lint/correctness/useExhaustiveDependencies: Go answers from the library, so each new library asks again
  useEffect(() => {
    const asked = Inherited(id);
    asked.then(setInherited, () => setInherited(undefined));
    return () => {
      void asked.cancel();
    };
  }, [id, library]);
  return inherited;
}
