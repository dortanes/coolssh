import assert from 'node:assert/strict';
import { test } from 'node:test';
import { Appearance } from '#bindings/preferences/models.ts';
import { themeFor } from './theme.ts';

test('a chosen theme holds whatever the system shows', () => {
  assert.equal(themeFor(Appearance.LightAppearance, true), 'light');
  assert.equal(themeFor(Appearance.DarkAppearance, false), 'dark');
});

test('the system appearance follows the system', () => {
  assert.equal(themeFor(Appearance.SystemAppearance, true), 'dark');
  assert.equal(themeFor(Appearance.SystemAppearance, false), 'light');
});
