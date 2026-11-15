import { useId } from 'react';
import { Appearance, Language, Style } from '#bindings/preferences/models.ts';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { m } from '@/paraglide/messages.js';
import type { PreferencesStore } from '../preferences-store.ts';
import { useStore } from '../use-store.ts';
import { type Choice, ChoiceSelect } from './choice-select.tsx';
import { SettingRow } from './setting-row.tsx';

export function GeneralSection({ preferences: store }: { preferences: PreferencesStore }) {
  const { snapshot, saveFailed } = useStore(store);
  const languageId = useId();
  const appearanceId = useId();
  const styleId = useId();
  const languages: Choice<Language>[] = [
    { value: Language.SystemLanguage, label: m.settings_language_system() },
    { value: Language.English, label: m.settings_language_en() },
    { value: Language.Russian, label: m.settings_language_ru() },
  ];
  const appearances: Choice<Appearance>[] = [
    { value: Appearance.SystemAppearance, label: m.settings_appearance_system() },
    { value: Appearance.LightAppearance, label: m.settings_appearance_light() },
    { value: Appearance.DarkAppearance, label: m.settings_appearance_dark() },
  ];
  const styles: Choice<Style>[] = [
    { value: Style.AutomaticStyle, label: m.settings_style_auto() },
    { value: Style.LiquidGlassStyle, label: m.settings_style_liquid_glass() },
    { value: Style.MicaStyle, label: m.settings_style_mica() },
    { value: Style.ShadcnStyle, label: m.settings_style_shadcn() },
  ];
  return (
    <section aria-labelledby="general-title" className="w-full max-w-2xl px-8 pb-8">
      <h1 id="general-title" className="text-lg font-semibold">
        {m.settings_general()}
      </h1>
      <div data-surface="group" className="mt-4 divide-y rounded-lg border bg-card">
        <SettingRow labelId={languageId} label={m.settings_language()}>
          <ChoiceSelect
            labelId={languageId}
            choices={languages}
            value={snapshot.preferences.language}
            onChoose={(language) => void store.change({ language })}
          />
        </SettingRow>
        <SettingRow labelId={appearanceId} label={m.settings_appearance()}>
          <ToggleGroup
            aria-labelledby={appearanceId}
            variant="outline"
            data-surface="segmented"
            value={[snapshot.preferences.appearance]}
            onValueChange={([pressed]) => {
              const appearance = appearances.find((option) => option.value === pressed)?.value;
              if (appearance) {
                void store.change({ appearance });
              }
            }}
          >
            {appearances.map((appearance) => (
              <ToggleGroupItem key={appearance.value} value={appearance.value}>
                {appearance.label}
              </ToggleGroupItem>
            ))}
          </ToggleGroup>
        </SettingRow>
        <SettingRow labelId={styleId} label={m.settings_style()}>
          <ChoiceSelect
            labelId={styleId}
            choices={styles}
            value={snapshot.preferences.style}
            onChoose={(style) => void store.change({ style })}
          />
        </SettingRow>
      </div>
      {saveFailed && (
        <p role="alert" className="mt-3 text-sm text-destructive">
          {m.settings_not_saved()}
        </p>
      )}
    </section>
  );
}
