import type { ITheme } from '@xterm/xterm';

/** The 16 ANSI colours, from black to bright white. */
export type Palette = Required<
  Pick<
    ITheme,
    | 'black'
    | 'red'
    | 'green'
    | 'yellow'
    | 'blue'
    | 'magenta'
    | 'cyan'
    | 'white'
    | 'brightBlack'
    | 'brightRed'
    | 'brightGreen'
    | 'brightYellow'
    | 'brightBlue'
    | 'brightMagenta'
    | 'brightCyan'
    | 'brightWhite'
  >
>;

/** Each chromatic colour and bright black keep 4.5:1 against every style's light background. */
export const lightPalette: Palette = {
  black: '#1d1d1f',
  red: '#c41a16',
  green: '#1a7f37',
  yellow: '#7d5d00',
  blue: '#0b57d0',
  magenta: '#a2299f',
  cyan: '#0f6f7a',
  white: '#d1d1d6',
  brightBlack: '#6b6b70',
  brightRed: '#d1242f',
  brightGreen: '#1c7a35',
  brightYellow: '#8a6600',
  brightBlue: '#0969da',
  brightMagenta: '#8250df',
  brightCyan: '#0a7a85',
  brightWhite: '#f5f5f7',
};

/** Each chromatic colour and bright black keep 4.5:1 against every style's dark background. */
export const darkPalette: Palette = {
  black: '#3a3a3c',
  red: '#ff6b64',
  green: '#5fd068',
  yellow: '#e5c04a',
  blue: '#5aa9ff',
  magenta: '#d48cf6',
  cyan: '#4fc8d8',
  white: '#d1d1d6',
  brightBlack: '#9a9aa0',
  brightRed: '#ff8a84',
  brightGreen: '#7ee787',
  brightYellow: '#ffd866',
  brightBlue: '#79c0ff',
  brightMagenta: '#e2a8ff',
  brightCyan: '#7ad8e6',
  brightWhite: '#ffffff',
};

/** The colours text is drawn in: every chromatic colour, plain and bright, and bright black. */
export const textColours = [
  'red',
  'green',
  'yellow',
  'blue',
  'magenta',
  'cyan',
  'brightBlack',
  'brightRed',
  'brightGreen',
  'brightYellow',
  'brightBlue',
  'brightMagenta',
  'brightCyan',
] as const satisfies readonly (keyof Palette)[];
