import { useId, useState } from 'react';
import { type Prompt, PromptKind } from '#bindings/session/models.ts';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { m } from '@/paraglide/messages.js';
import { FormDialog } from '../form-dialog.tsx';
import type { SessionActions } from './session-actions.ts';

type PromptDialogProps = {
  sessionId: string;
  prompt: Prompt;
  actions: SessionActions;
};

/** What a session asks the user, over its tab: a host key, a password, a passphrase or answers. */
export function PromptDialog({ sessionId, prompt, actions }: PromptDialogProps) {
  const questions =
    prompt.kind === PromptKind.QuestionsPrompt
      ? (prompt.questions ?? [])
      : [{ text: '', echo: false }];
  const [values, setValues] = useState(() => questions.map(() => ''));
  const [remember, setRemember] = useState(false);
  const ids = useId();
  const reply = (accept: boolean) =>
    void actions.answer(sessionId, prompt.id, { accept, values, remember });
  const { title, description, label, retry, submit } = wording(prompt);
  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && reply(false)}
      title={title}
      description={description}
      className="gap-5 p-5"
      submit={submit}
      submitAutoFocus={prompt.kind === PromptKind.HostKeyPrompt}
      onSubmit={() => reply(true)}
    >
      {prompt.retry && <p className="text-sm text-destructive">{retry}</p>}
      {prompt.kind === PromptKind.HostKeyPrompt ? (
        <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5 text-sm">
          <dt className="text-muted-foreground">{m.session_host_key_type()}</dt>
          <dd className="font-mono">{prompt.keyType}</dd>
          <dt className="text-muted-foreground">{m.session_host_key_fingerprint()}</dt>
          <dd className="font-mono break-all select-text">{prompt.fingerprint}</dd>
        </dl>
      ) : (
        questions.map((question, index) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: a prompt's questions never change while it is shown, and two may read the same
          <div key={index} className="flex flex-col gap-1.5">
            <Label htmlFor={`${ids}-${index}`}>{question.text || label}</Label>
            <Input
              id={`${ids}-${index}`}
              type={question.echo ? 'text' : 'password'}
              autoFocus={index === 0}
              autoComplete="off"
              value={values[index] ?? ''}
              onChange={(event) =>
                setValues(values.map((value, at) => (at === index ? event.target.value : value)))
              }
            />
          </div>
        ))
      )}
      {prompt.kind === PromptKind.PasswordPrompt && prompt.savable && (
        <Label className="flex items-center gap-2 font-normal">
          <Checkbox
            checked={remember}
            onCheckedChange={(checked) => setRemember(checked === true)}
          />
          {m.session_password_remember()}
        </Label>
      )}
    </FormDialog>
  );
}

/** The prompt's title, its explanation, its field's label, what a retry says, and its button. */
function wording(prompt: Prompt) {
  const host = prompt.host;
  switch (prompt.kind) {
    case PromptKind.HostKeyPrompt:
      return {
        title: m.session_host_key_title({ host }),
        description: m.session_host_key_body({ host }),
        label: '',
        retry: '',
        submit: m.session_host_key_trust(),
      };
    case PromptKind.PassphrasePrompt:
      return {
        title: m.session_passphrase_title({ key: prompt.key ?? '' }),
        description: m.session_passphrase_body({ host }),
        label: m.session_passphrase_label(),
        retry: m.session_passphrase_retry(),
        submit: m.session_password_connect(),
      };
    case PromptKind.QuestionsPrompt:
      return {
        title: prompt.name || m.session_questions_title({ host }),
        description: prompt.instruction ?? '',
        label: '',
        retry: m.session_questions_retry(),
        submit: m.session_questions_continue(),
      };
    default:
      return {
        title: m.session_password_title({ host }),
        description: '',
        label: m.session_password_label(),
        retry: m.session_password_retry(),
        submit: m.session_password_connect(),
      };
  }
}
