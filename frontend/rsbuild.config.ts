import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginTailwindcss } from '@rsbuild/plugin-tailwindcss';

export default defineConfig(({ command }) => ({
  plugins: [pluginReact(), pluginTailwindcss()],
  // The import aliases come from tsconfig.json's paths.
  source: {
    entry: { index: './src/main-window.tsx', settings: './src/settings-window.tsx' },
  },
  server: {
    host: '127.0.0.1',
    // The port is the one wails3 dev chose and waits for.
    // biome-ignore lint/complexity/useLiteralKeys: noPropertyAccessFromIndexSignature requires brackets
    port: Number(process.env['WAILS_VITE_PORT'] ?? 9245),
    strictPort: true,
  },
  dev: {
    // The window loads the page from the Wails asset server, so the reload socket names the dev
    // server itself rather than the page's origin.
    client: { protocol: 'ws', host: '127.0.0.1', port: '<port>' },
  },
  html: {
    template: './src/index.html',
    title: 'CoolSSH',
    meta: {
      'content-security-policy': {
        'http-equiv': 'Content-Security-Policy',
        content: [
          "default-src 'self'",
          "script-src 'self'",
          "style-src 'self' 'unsafe-inline'",
          "img-src 'self' data: blob:",
          command === 'dev' ? "connect-src 'self' ws://127.0.0.1:*" : "connect-src 'self'",
          "font-src 'self' data:",
          "object-src 'none'",
          "frame-src 'none'",
          "base-uri 'none'",
          "form-action 'none'",
        ].join('; '),
      },
    },
  },
  output: {
    distPath: { root: 'dist' },
    assetPrefix: './',
  },
}));
