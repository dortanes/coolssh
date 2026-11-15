import { useState } from 'react';
import { DeleteKey } from '#bindings/desktop/service.ts';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { m } from '@/paraglide/messages.js';
import { GenerateKeyDialog } from '../library/generate-key-dialog.tsx';
import { addKeyFromFile } from '../library/key-actions.ts';
import type { LibraryStore } from '../library/library-store.ts';
import { LibraryView } from '../library/library-view.ts';
import { PasteKeyDialog } from '../library/paste-key-dialog.tsx';
import { useStore } from '../use-store.ts';
import { KeyRow } from './key-row.tsx';
import { TrustedKeys } from './trusted-keys.tsx';
import type { TrustedKeysStore } from './trusted-keys-store.ts';

/** The library's keys and the host keys trusted in the app. */
export function KeysSection({
  library,
  trustedKeys,
}: {
  library: LibraryStore;
  trustedKeys: TrustedKeysStore;
}) {
  const { library: snapshot } = useStore(library);
  const view = new LibraryView(snapshot);
  const [pasting, setPasting] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [deleting, setDeleting] = useState<string | undefined>();
  const deletingKey = view.keys().find((key) => key.id === deleting);
  return (
    <section aria-labelledby="keys-title" className="w-full max-w-2xl px-8 pb-8">
      <h1 id="keys-title" className="text-lg font-semibold">
        {m.settings_keys()}
      </h1>
      <div className="mt-4 flex gap-2">
        <Button variant="secondary" size="sm" onClick={() => setGenerating(true)}>
          {m.keys_generate()}
        </Button>
        <Button variant="secondary" size="sm" onClick={() => void addKeyFromFile(library)}>
          {m.keys_add_file()}
        </Button>
        <Button variant="secondary" size="sm" onClick={() => setPasting(true)}>
          {m.keys_add_paste()}
        </Button>
      </div>
      {view.keys().length === 0 ? (
        <p className="mt-6 text-sm text-muted-foreground">{m.keys_empty()}</p>
      ) : (
        <ul data-surface="group" className="mt-4 divide-y rounded-lg border bg-card">
          {view.keys().map((key) => (
            <KeyRow
              key={key.id}
              keyEntry={key}
              isDefault={key.id === snapshot.defaultKeyId}
              library={library}
              onDelete={() => setDeleting(key.id)}
            />
          ))}
        </ul>
      )}
      <TrustedKeys store={trustedKeys} />
      <PasteKeyDialog open={pasting} onOpenChange={setPasting} />
      <GenerateKeyDialog open={generating} onOpenChange={setGenerating} />
      <AlertDialog
        open={deletingKey !== undefined}
        onOpenChange={(open) => !open && setDeleting(undefined)}
      >
        <AlertDialogContent data-surface="dialog">
          <AlertDialogHeader>
            <AlertDialogTitle>
              {m.keys_delete_title({ name: deletingKey?.name ?? '' })}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {m.keys_delete_body({ count: deletingKey ? view.keyUsers(deletingKey.id) : 0 })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{m.common_cancel()}</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                if (deleting) {
                  void library.run(() => DeleteKey(deleting));
                }
                setDeleting(undefined);
              }}
            >
              {m.library_delete_confirm()}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </section>
  );
}
