import { OctagonXIcon } from 'lucide-react';
import {
  Toast,
  ToastClose,
  ToastContent,
  ToastDescription,
  ToastPortal,
  ToastProvider,
  ToastViewport,
  toast,
  useToastManager,
} from '@/components/ui/toast';
import { m } from '@/paraglide/messages.js';
import type { LibraryStore } from './library/library-store.ts';

/** Shows each new problem the library reports as a toast; closing the toast clears the problem. */
export function followProblems(library: LibraryStore): void {
  let shown: string | undefined;
  library.subscribe(() => {
    const { problem } = library.current();
    if (problem && problem !== shown) {
      toast.add({
        description: problem,
        type: 'error',
        onClose: () => {
          if (library.current().problem === problem) {
            library.report(undefined);
          }
        },
      });
    }
    shown = problem;
  });
}

function NoticeList() {
  const { toasts } = useToastManager();
  return toasts.map((notice) => (
    <Toast key={notice.id} toast={notice} data-surface="toast">
      <ToastContent>
        <OctagonXIcon aria-hidden className="size-4 shrink-0 self-start text-destructive" />
        <ToastDescription className="flex-1 text-foreground" />
        <ToastClose aria-label={m.common_dismiss()} className="-my-1 self-start" />
      </ToastContent>
    </Toast>
  ));
}

/** The window's toasts, stacked at its bottom right corner. */
export function Notices() {
  return (
    <ToastProvider toastManager={toast}>
      <ToastPortal>
        <ToastViewport>
          <NoticeList />
        </ToastViewport>
      </ToastPortal>
    </ToastProvider>
  );
}
