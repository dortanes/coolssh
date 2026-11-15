import assert from 'node:assert/strict';
import { test } from 'node:test';
import { darkPalette, lightPalette, type Palette, textColours } from './terminal-palette.ts';

/**
 * The workspace's opaque backgrounds in each theme, from the styles: Liquid Glass, shadcn, and
 * Mica over its material and over a solid window.
 */
const backgrounds = {
  light: ['#ffffff', '#f3f3f3'],
  dark: ['#0a0a0a', '#1e1e1e', '#181818', '#282828'],
};

function luminance(hex: string): number {
  const channels = [1, 3, 5].map((start) => {
    const value = Number.parseInt(hex.slice(start, start + 2), 16) / 255;
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  const [red = 0, green = 0, blue = 0] = channels;
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}

function contrast(first: string, second: string): number {
  const [lighter, darker] = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return ((lighter ?? 0) + 0.05) / ((darker ?? 0) + 0.05);
}

function assertReadable(palette: Palette, against: string[]): void {
  for (const colour of textColours) {
    for (const background of against) {
      const ratio = contrast(palette[colour], background);
      assert.ok(
        ratio >= 4.5,
        `${colour} ${palette[colour]} on ${background} is ${ratio.toFixed(2)}:1`,
      );
    }
  }
}

test('the light palette is readable on every light background', () => {
  assertReadable(lightPalette, backgrounds.light);
});

test('the dark palette is readable on every dark background', () => {
  assertReadable(darkPalette, backgrounds.dark);
});
