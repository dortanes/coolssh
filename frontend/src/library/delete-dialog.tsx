import { Delete } from '#bindings/desktop/service.ts';
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
import { m } from '@/paraglide/messages.js';
import { useStore } from '../use-store.ts';
import type { LibraryStore } from './library-store.ts';
import { LibraryView } from './library-view.ts';

/** Asks before deleting a host, or a folder with what it holds, and deletes it once confirmed. */
export function DeleteDialog({
  target,
  library,
  onClose,
}: {
  /** The host or folder to delete; the dialog is closed while it is undefined. */
  target: string | undefined;
  library: LibraryStore;
  onClose: () => void;
}) {
  const view = new LibraryView(useStore(library).library);
  const folder = target ? view.folder(target) : undefined;
  const name = folder?.name ?? (target ? view.host(target)?.name : undefined) ?? '';
  const inside = folder ? view.countInside(folder.id) : { hosts: 0, folders: 0 };
  return (
    <AlertDialog open={target !== undefined} onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent data-surface="dialog">
        <AlertDialogHeader>
          <AlertDialogTitle>{m.library_delete_title({ name })}</AlertDialogTitle>
          <AlertDialogDescription>
            {folder
              ? m.library_delete_folder_body({ hosts: inside.hosts, folders: inside.folders })
              : m.library_delete_host_body()}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{m.common_cancel()}</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={() => {
              if (target) {
                void library.run(() => Delete(target));
              }
              onClose();
            }}
          >
            {m.library_delete_confirm()}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
