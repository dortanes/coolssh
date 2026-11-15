import type { LucideIcon } from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

/** An option; a disabled one is listed and cannot be chosen. */
export type Choice<Value extends string> = {
  value: Value;
  label: string;
  Glyph?: LucideIcon;
  disabled?: boolean;
};

function ChoiceLabel({ label, Glyph }: { label: string; Glyph?: LucideIcon | undefined }) {
  return Glyph ? (
    <>
      <Glyph aria-hidden />
      {label}
    </>
  ) : (
    label
  );
}

/** A setting picked from a list: shows the chosen option and reports only a listed one. */
export function ChoiceSelect<Value extends string>({
  labelId,
  choices,
  value,
  onChoose,
}: {
  labelId: string;
  choices: Choice<Value>[];
  value: Value;
  onChoose: (value: Value) => void;
}) {
  return (
    <Select
      items={choices.map(({ value, label, Glyph }) => ({
        value,
        label: <ChoiceLabel label={label} Glyph={Glyph} />,
      }))}
      value={value}
      onValueChange={(picked) => {
        const choice = choices.find((option) => option.value === picked);
        if (choice) {
          onChoose(choice.value);
        }
      }}
    >
      <SelectTrigger aria-labelledby={labelId} className="min-w-44">
        <SelectValue />
      </SelectTrigger>
      <SelectContent data-surface="menu">
        {choices.map((choice) => (
          <SelectItem key={choice.value} value={choice.value} disabled={choice.disabled ?? false}>
            <ChoiceLabel label={choice.label} Glyph={choice.Glyph} />
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
