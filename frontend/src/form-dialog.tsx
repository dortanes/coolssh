import type { ReactNode } from 'react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { m } from '@/paraglide/messages.js';

/** A dialog holding a form: its title, its fields, then Cancel and the button that submits it. */
export function FormDialog({
  open,
  onOpenChange,
  title,
  description,
  className,
  submit,
  submitDisabled = false,
  submitAutoFocus = false,
  onSubmit,
  children,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string | undefined;
  className: string;
  submit: string;
  submitDisabled?: boolean;
  submitAutoFocus?: boolean;
  onSubmit: () => void;
  children: ReactNode;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-surface="dialog" showCloseButton={false} className={className}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <form
          className="flex flex-col gap-3.5"
          onSubmit={(event) => {
            event.preventDefault();
            onSubmit();
          }}
        >
          {children}
          <DialogFooter className="-mx-5 -mb-5 p-5">
            <DialogClose render={<Button type="button" variant="secondary" />}>
              {m.common_cancel()}
            </DialogClose>
            <Button type="submit" disabled={submitDisabled} autoFocus={submitAutoFocus}>
              {submit}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
