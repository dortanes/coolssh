import { type ReactNode, useId, useState } from 'react';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { FormDialog } from '../form-dialog.tsx';
import { refusal } from './library-store.ts';

/** A labelled one-line field of a key dialog. */
export function KeyField({
  label,
  value,
  onChange,
  placeholder,
  secret = false,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  /** Hides what is typed, for a passphrase. */
  secret?: boolean;
}) {
  const id = useId();
  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        type={secret ? 'password' : 'text'}
        value={value}
        placeholder={placeholder}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
}

/** A dialog that adds a key: one submission at a time, closed once it succeeds, else saying why. */
export function KeyDialog({
  open,
  onOpenChange,
  onReset,
  title,
  description,
  submit,
  ready,
  onSubmit,
  className,
  children,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Clears the fields as the dialog closes. */
  onReset: () => void;
  title: string;
  description?: string;
  submit: string;
  /** Whether the fields can be submitted. */
  ready: boolean;
  onSubmit: () => Promise<void>;
  className: string;
  children: ReactNode;
}) {
  const [problem, setProblem] = useState<string | undefined>();
  const [working, setWorking] = useState(false);
  const close = (next: boolean) => {
    if (!next) {
      onReset();
      setProblem(undefined);
    }
    onOpenChange(next);
  };
  const run = async () => {
    setWorking(true);
    try {
      await onSubmit();
      close(false);
    } catch (error) {
      setProblem(refusal(error));
    } finally {
      setWorking(false);
    }
  };
  return (
    <FormDialog
      open={open}
      onOpenChange={close}
      title={title}
      description={description}
      className={className}
      submit={submit}
      submitDisabled={working || !ready}
      onSubmit={() => void run()}
    >
      {children}
      {problem && (
        <p role="alert" className="text-sm text-destructive">
          {problem}
        </p>
      )}
    </FormDialog>
  );
}
