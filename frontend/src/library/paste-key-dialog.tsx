import { useId, useState } from 'react';
import { AddKeyText } from '#bindings/desktop/service.ts';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { m } from '@/paraglide/messages.js';
import { KeyDialog, KeyField } from './key-dialog-parts.tsx';

/** Adds a pasted private key with a name and, for an encrypted one, its passphrase. */
export function PasteKeyDialog({
  open,
  onOpenChange,
  onAdded,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdded?: (id: string) => void;
}) {
  const [name, setName] = useState('');
  const [text, setText] = useState('');
  const [passphrase, setPassphrase] = useState('');
  const textId = useId();
  return (
    <KeyDialog
      open={open}
      onOpenChange={onOpenChange}
      onReset={() => {
        setName('');
        setText('');
        setPassphrase('');
      }}
      title={m.keys_paste_title()}
      submit={m.keys_paste_add()}
      ready={text.trim() !== ''}
      onSubmit={async () => onAdded?.(await AddKeyText(name, text, passphrase))}
      className="gap-5 p-5 sm:max-w-lg"
    >
      <KeyField
        label={m.keys_paste_name()}
        value={name}
        placeholder={m.keys_paste_name_placeholder()}
        onChange={setName}
      />
      <div className="flex flex-col gap-1.5">
        <Label htmlFor={textId}>{m.keys_paste_text()}</Label>
        <Textarea
          id={textId}
          value={text}
          spellCheck={false}
          className="h-36 resize-none font-mono text-xs [field-sizing:fixed] md:text-xs"
          placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
          onChange={(event) => setText(event.target.value)}
        />
      </div>
      <KeyField
        label={m.keys_paste_passphrase()}
        value={passphrase}
        secret
        onChange={setPassphrase}
      />
    </KeyDialog>
  );
}
