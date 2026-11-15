import assert from 'node:assert/strict';
import { test } from 'node:test';
import { assignShortcut, shortcutLabel, shortcutOf } from './shortcut.ts';

function press(code: string, ...held: string[]) {
  return { code, getModifierState: (modifier: string) => held.includes(modifier) };
}

test('a shortcut reads as symbols on macOS and words elsewhere', () => {
  assert.equal(shortcutLabel('Meta+Shift+BracketRight', 'darwin'), '⇧⌘]');
  assert.equal(shortcutLabel('Shift+Control+KeyK', 'linux'), 'Ctrl+Shift+K');
  assert.equal(shortcutLabel('Control+Tab', 'windows'), 'Ctrl+Tab');
  assert.equal(shortcutLabel('Control+Tab', 'darwin'), '⌃⇥');
});

test('a key press makes a shortcut only with Ctrl, Alt or Meta held', () => {
  assert.equal(shortcutOf(press('KeyK', 'Shift', 'Meta')), 'Shift+Meta+KeyK');
  assert.equal(shortcutOf(press('KeyK', 'Shift')), undefined);
  assert.equal(shortcutOf(press('KeyK')), undefined);
  assert.equal(shortcutOf(press('MetaLeft', 'Meta')), undefined);
});

test('a shortcut moves to the action that takes it', () => {
  const inEffect = { newHost: 'Meta+KeyN', newFolder: 'Meta+Shift+KeyN' };
  assert.deepEqual(assignShortcut({}, inEffect, 'newHost', 'Meta+Shift+KeyN'), {
    newHost: 'Meta+Shift+KeyN',
    newFolder: '',
  });
  assert.deepEqual(assignShortcut({ newHost: 'Meta+KeyH' }, inEffect, 'newHost', ''), {
    newHost: '',
  });
});
