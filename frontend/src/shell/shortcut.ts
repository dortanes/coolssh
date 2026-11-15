const modifiers = ['Control', 'Alt', 'Shift', 'Meta'] as const;
type Modifier = (typeof modifiers)[number];

/** macOS draws modifiers as symbols in this order; elsewhere they are words. */
const modifierLabels: Record<Modifier, { mac: string; other: string }> = {
  Control: { mac: '⌃', other: 'Ctrl' },
  Alt: { mac: '⌥', other: 'Alt' },
  Shift: { mac: '⇧', other: 'Shift' },
  Meta: { mac: '⌘', other: 'Win' },
};

const keyLabels: Record<string, string> = {
  BracketLeft: '[',
  BracketRight: ']',
  Comma: ',',
  Period: '.',
  Slash: '/',
  Backslash: '\\',
  Semicolon: ';',
  Quote: "'",
  Backquote: '`',
  Minus: '-',
  Equal: '=',
  Space: 'Space',
  Enter: '↩',
  Tab: '⇥',
  Backspace: '⌫',
  Delete: '⌦',
  ArrowUp: '↑',
  ArrowDown: '↓',
  ArrowLeft: '←',
  ArrowRight: '→',
};

/** Codes of the keys that only modify another. */
const modifierCodes = new Set([
  'ControlLeft',
  'ControlRight',
  'AltLeft',
  'AltRight',
  'ShiftLeft',
  'ShiftRight',
  'MetaLeft',
  'MetaRight',
  'CapsLock',
  'Fn',
]);

/** Keys macOS draws as symbols and other platforms name. */
const namedKeys = new Set(['Enter', 'Tab', 'Backspace', 'Delete']);

function keyLabel(code: string, mac: boolean): string {
  if (!mac && namedKeys.has(code)) {
    return code;
  }
  return keyLabels[code] ?? code.replace(/^(Key|Digit|Numpad)/, '');
}

/** How a shortcut reads on platform, named as Go's runtime.GOOS names it. */
export function shortcutLabel(shortcut: string, platform: string): string {
  const parts = shortcut.split('+');
  const key = parts.at(-1) ?? '';
  const mac = platform === 'darwin';
  const held = modifiers
    .filter((modifier) => parts.includes(modifier))
    .map((modifier) => modifierLabels[modifier][mac ? 'mac' : 'other']);
  return [...held, keyLabel(key, mac)].join(mac ? '' : '+');
}

/**
 * The shortcut a key press makes; undefined until a key other than a modifier is held with Ctrl,
 * Alt or Meta, since Shift alone types a character.
 */
export function shortcutOf(
  event: Pick<KeyboardEvent, 'code' | 'getModifierState'>,
): string | undefined {
  if (modifierCodes.has(event.code)) {
    return undefined;
  }
  const held = modifiers.filter((modifier) => event.getModifierState(modifier));
  return held.some((modifier) => modifier !== 'Shift')
    ? [...held, event.code].join('+')
    : undefined;
}

type ShortcutMap = Partial<Record<string, string>>;

/**
 * The user's shortcuts once action takes shortcut: an action that held it in effect loses it, so
 * no two actions share one. An empty shortcut removes action's.
 */
export function assignShortcut(
  own: ShortcutMap,
  inEffect: ShortcutMap,
  action: string,
  shortcut: string,
): ShortcutMap {
  const next: ShortcutMap = { ...own, [action]: shortcut };
  for (const [other, held] of Object.entries(inEffect)) {
    if (shortcut !== '' && other !== action && held === shortcut) {
      next[other] = '';
    }
  }
  return next;
}
