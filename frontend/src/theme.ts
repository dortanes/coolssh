import { Appearance } from '#bindings/preferences/models.ts';

type Theme = 'light' | 'dark';

/** The theme a window shows for the chosen appearance while the system is dark or light. */
export function themeFor(appearance: Appearance, systemDark: boolean): Theme {
  switch (appearance) {
    case Appearance.LightAppearance:
      return 'light';
    case Appearance.DarkAppearance:
      return 'dark';
    default:
      return systemDark ? 'dark' : 'light';
  }
}
