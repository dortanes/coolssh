import { useId, useState } from 'react';
import { GenerateKey } from '#bindings/desktop/service.ts';
import { KeyKind } from '#bindings/library/models.ts';
import { Label } from '@/components/ui/label';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { m } from '@/paraglide/messages.js';
import { KeyDialog, KeyField } from './key-dialog-parts.tsx';

const kinds = [
  { value: KeyKind.Ed25519Kind, label: 'Ed25519' },
  { value: KeyKind.RSAKind, label: 'RSA' },
];

/** Makes a new key with a name, an algorithm and, when given, a passphrase. */
export function GenerateKeyDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [name, setName] = useState('');
  const [kind, setKind] = useState(KeyKind.Ed25519Kind);
  const [passphrase, setPassphrase] = useState('');
  const kindId = useId();
  return (
    <KeyDialog
      open={open}
      onOpenChange={onOpenChange}
      onReset={() => {
        setName('');
        setKind(KeyKind.Ed25519Kind);
        setPassphrase('');
      }}
      title={m.keys_generate_title()}
      description={m.keys_generate_body()}
      submit={m.keys_generate_add()}
      ready={name.trim() !== ''}
      onSubmit={async () => {
        await GenerateKey(name, kind, passphrase);
      }}
      className="gap-5 p-5 sm:max-w-md"
    >
      <KeyField
        label={m.keys_paste_name()}
        value={name}
        placeholder={m.keys_paste_name_placeholder()}
        onChange={setName}
      />
      <div className="flex flex-col gap-1.5">
        <Label id={kindId}>{m.keys_generate_kind()}</Label>
        <ToggleGroup
          aria-labelledby={kindId}
          variant="outline"
          data-surface="segmented"
          value={[kind]}
          onValueChange={([pressed]) => {
            const chosen = kinds.find((option) => option.value === pressed)?.value;
            if (chosen) {
              setKind(chosen);
            }
          }}
        >
          {kinds.map((option) => (
            <ToggleGroupItem key={option.value} value={option.value}>
              {option.label}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      </div>
      <KeyField
        label={m.keys_generate_passphrase()}
        value={passphrase}
        secret
        onChange={setPassphrase}
      />
    </KeyDialog>
  );
}
