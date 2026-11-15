import { type HTMLAttributes, useEffect, useRef, useState } from 'react';
import { Input } from '@/components/ui/input';

/**
 * A text field that saves when it loses focus or on Enter, only if it changed; Escape puts the
 * saved value back.
 */
export function TextSetting({
  labelId,
  value,
  placeholder,
  focus,
  inputMode,
  onCommit,
}: {
  labelId: string;
  value: string;
  placeholder?: string | undefined;
  /** Puts the cursor in the field when it appears. */
  focus?: boolean | undefined;
  inputMode?: HTMLAttributes<HTMLInputElement>['inputMode'];
  onCommit: (value: string) => void;
}) {
  const [draft, setDraft] = useState(value);
  const field = useRef<HTMLInputElement>(null);
  useEffect(() => setDraft(value), [value]);
  useEffect(() => {
    if (focus) {
      field.current?.select();
    }
  }, [focus]);
  const commit = () => {
    if (draft !== value) {
      onCommit(draft);
    }
  };
  return (
    <Input
      ref={field}
      aria-labelledby={labelId}
      value={draft}
      placeholder={placeholder}
      inputMode={inputMode}
      spellCheck={false}
      autoCapitalize="off"
      autoCorrect="off"
      className="w-60"
      onChange={(event) => setDraft(event.target.value)}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          commit();
        } else if (event.key === 'Escape') {
          event.preventDefault();
          setDraft(value);
        }
      }}
    />
  );
}
