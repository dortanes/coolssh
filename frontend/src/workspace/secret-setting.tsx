import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { m } from '@/paraglide/messages.js';

/**
 * A secret kept in the keyring, a password or a token: typed, then saved with the Save button or
 * Enter; once saved, it can be forgotten.
 */
export function SecretSetting({
  labelId,
  saved,
  placeholder,
  onSave,
  onForget,
}: {
  labelId: string;
  /** Whether a secret is saved. */
  saved: boolean;
  placeholder: string;
  /** Saves the secret; resolves true once it is saved. */
  onSave: (secret: string) => Promise<boolean | undefined>;
  onForget: () => void;
}) {
  const [draft, setDraft] = useState('');
  if (saved) {
    return (
      <div className="flex w-60 items-center justify-end gap-3">
        <span className="text-sm text-muted-foreground">{m.editor_password_saved()}</span>
        <Button variant="secondary" size="sm" onClick={onForget}>
          {m.editor_password_forget()}
        </Button>
      </div>
    );
  }
  const save = () => {
    if (draft !== '') {
      void onSave(draft).then((saved) => {
        if (saved) {
          setDraft('');
        }
      });
    }
  };
  return (
    <div className="flex w-60 items-center gap-2">
      <Input
        type="password"
        aria-labelledby={labelId}
        value={draft}
        placeholder={placeholder}
        className="min-w-0 flex-1"
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter') {
            save();
          }
        }}
      />
      {draft !== '' && (
        <Button size="sm" onClick={save}>
          {m.editor_password_save()}
        </Button>
      )}
    </div>
  );
}
