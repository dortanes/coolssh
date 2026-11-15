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

/** Asks before closing tabs whose sessions are connected. */
export function DisconnectDialog({
  hosts,
  onConfirm,
  onClose,
}: {
  /** The connected sessions' hosts; the dialog is closed while there are none. */
  hosts: readonly string[];
  onConfirm: () => void;
  onClose: () => void;
}) {
  const [only] = hosts;
  return (
    <AlertDialog open={hosts.length > 0} onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent data-surface="dialog">
        <AlertDialogHeader>
          <AlertDialogTitle>
            {hosts.length === 1 && only !== undefined
              ? m.session_disconnect_title({ host: only })
              : m.session_disconnect_many_title({ count: hosts.length })}
          </AlertDialogTitle>
          <AlertDialogDescription>{m.session_disconnect_body()}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{m.common_cancel()}</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={() => {
              onConfirm();
              onClose();
            }}
          >
            {m.session_disconnect_confirm()}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
