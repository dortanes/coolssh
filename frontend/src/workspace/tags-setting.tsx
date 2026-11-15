import { XIcon } from 'lucide-react';
import { useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { m } from '@/paraglide/messages.js';

/**
 * A host's tags as chips, each removable, and a field that adds what is typed on Enter or a comma;
 * Backspace in the empty field removes the last tag.
 */
export function TagsSetting({
  labelId,
  tags,
  onChange,
}: {
  labelId: string;
  tags: readonly string[];
  onChange: (tags: string[]) => void;
}) {
  const [draft, setDraft] = useState('');
  const add = () => {
    if (draft.trim() !== '') {
      onChange([...tags, draft]);
      setDraft('');
    }
  };
  return (
    <div className="flex w-60 shrink-0 flex-col gap-2">
      {tags.length > 0 && (
        <ul aria-labelledby={labelId} className="flex flex-wrap gap-1">
          {tags.map((tag) => (
            <li key={tag}>
              <Badge variant="secondary" className="pr-0.5">
                {tag}
                <button
                  type="button"
                  aria-label={m.editor_tag_remove({ tag })}
                  title={m.editor_tag_remove({ tag })}
                  className="flex size-4 items-center justify-center rounded-full outline-none hover:bg-foreground/10 focus-visible:ring-2 focus-visible:ring-ring"
                  onClick={() => onChange(tags.filter((each) => each !== tag))}
                >
                  <XIcon aria-hidden className="size-3" />
                </button>
              </Badge>
            </li>
          ))}
        </ul>
      )}
      <Input
        aria-labelledby={labelId}
        value={draft}
        placeholder={m.editor_tags_placeholder()}
        onChange={(event) => {
          const typed = event.target.value;
          if (typed.endsWith(',')) {
            onChange([...tags, typed.slice(0, -1)]);
            setDraft('');
          } else {
            setDraft(typed);
          }
        }}
        onBlur={add}
        onKeyDown={(event) => {
          if (event.key === 'Enter') {
            event.preventDefault();
            add();
          } else if (event.key === 'Backspace' && draft === '' && tags.length > 0) {
            onChange(tags.slice(0, -1));
          }
        }}
      />
    </div>
  );
}
