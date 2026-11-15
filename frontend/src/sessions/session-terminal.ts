import '@xterm/xterm/css/xterm.css';
import { Browser, Clipboard, Stream } from '@wailsio/runtime';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { WebglAddon } from '@xterm/addon-webgl';
import { type ITerminalOptions, type ITheme, Terminal } from '@xterm/xterm';
import { Resize } from '#bindings/desktop/service.ts';
import { Cursor, type Terminal as TerminalSettings } from '#bindings/preferences/models.ts';
import { darkPalette, lightPalette } from './terminal-palette.ts';

/** xterm.js's name for each cursor shape. */
const cursorStyles: Record<Cursor, NonNullable<ITerminalOptions['cursorStyle']>> = {
  [Cursor.$zero]: 'block',
  [Cursor.BlockCursor]: 'block',
  [Cursor.BarCursor]: 'bar',
  [Cursor.UnderlineCursor]: 'underline',
};

/** The scrollbar's width, px: a 6 px thumb inside a 2 px margin, like macOS's overlay scroller. */
const scrollbarWidth = 10;

/** The stream Go serves each terminal on; its first frame names the session. */
const terminalStream = 'terminal';

/** A colour as the canvas writes it, `#rrggbb` or `rgba(…)`, which xterm.js parses. */
function canvasColour(css: string): string {
  const context = document.createElement('canvas').getContext('2d');
  if (!context) {
    return css;
  }
  context.fillStyle = css;
  return context.fillStyle;
}

/** `#rrggbb` with an alpha. */
function translucent(hex: string, alpha: number): string {
  const channels = [1, 3, 5].map((start) => Number.parseInt(hex.slice(start, start + 2), 16));
  return `rgba(${channels.join(', ')}, ${alpha})`;
}

/**
 * The terminal's colours from the document's theme: its foreground and ANSI palette, and its
 * background when opaque, none over a translucent one (Mica's), which the pane paints.
 */
function documentTheme(style: CSSStyleDeclaration): ITheme {
  const dark = document.documentElement.classList.contains('dark');
  const foreground = canvasColour(style.getPropertyValue('--foreground').trim());
  const background = canvasColour(style.getPropertyValue('--background').trim());
  return {
    ...(dark ? darkPalette : lightPalette),
    background: background.startsWith('#') ? background : '#00000000',
    foreground,
    cursor: foreground,
    // The ruler only sets the scrollbar's width; xterm.js does not read the keyword transparent.
    overviewRulerBorder: '#00000000',
    cursorAccent: dark ? '#000000' : '#ffffff',
    ...(foreground.startsWith('#') && { selectionBackground: translucent(foreground, 0.3) }),
  };
}

/**
 * The terminal's colours and type as the document's style sets them (`--terminal-*`), with the
 * user's settings over them.
 */
function terminalLook(settings: TerminalSettings): ITerminalOptions {
  const style = getComputedStyle(document.documentElement);
  const token = (name: string) => style.getPropertyValue(`--terminal-${name}`).trim();
  const theme = documentTheme(style);
  return {
    // Only a translucent background, Mica's, needs transparency. It comes before the theme: a theme
    // applied while it is off has its background made opaque, black for a transparent one.
    allowTransparency: theme.background === '#00000000',
    theme,
    fontFamily: token('font'),
    fontSize: settings.fontSize || Number.parseFloat(token('font-size')),
    lineHeight: Number.parseFloat(token('line-height')),
    cursorStyle: cursorStyles[settings.cursor],
    cursorBlink: settings.cursorBlink,
    scrollback: settings.scrollback,
    macOptionIsMeta: settings.optionAsMeta,
  };
}

/**
 * One session's xterm.js terminal on the session's stream, in an element of its own that moves
 * between panes, so the terminal and its scrollback outlive any one pane.
 */
export class SessionTerminal {
  private readonly id: string;
  private readonly element = document.createElement('div');
  private readonly terminal: Terminal;
  private readonly fit = new FitAddon();
  private readonly socket: ReturnType<typeof Stream>;
  private readonly resizing: ResizeObserver;
  private readonly theming: MutationObserver;
  private settings: TerminalSettings;
  private webgl: WebglAddon | undefined;
  private opened = false;

  /** platform is Go's name of the operating system, which picks the clipboard shortcuts. */
  constructor(id: string, platform: string, settings: TerminalSettings) {
    this.id = id;
    this.settings = settings;
    this.element.style.cssText = 'width: 100%; height: 100%;';
    this.terminal = new Terminal({
      // The scrollbar is as wide as the overview ruler, 14 px unless it is set.
      overviewRuler: { width: scrollbarWidth },
      ...terminalLook(settings),
    });
    this.terminal.loadAddon(this.fit);
    this.terminal.loadAddon(new WebLinksAddon((_event, uri) => void Browser.OpenURL(uri)));
    // macOS copies and pastes through the Edit menu, whose ⌘C and ⌘V reach the terminal as
    // clipboard events; elsewhere Ctrl+C and Ctrl+V belong to the shell.
    if (platform !== 'darwin') {
      this.terminal.attachCustomKeyEventHandler((event) => this.shortcut(event));
    }
    this.socket = Stream(terminalStream);
    this.socket.binaryType = 'arraybuffer';
    this.socket.onopen = () => this.socket.send(id);
    this.socket.onmessage = (event: MessageEvent<ArrayBuffer>) =>
      this.terminal.write(new Uint8Array(event.data));
    const encoder = new TextEncoder();
    this.terminal.onData((data) => this.send(encoder.encode(data)));
    this.terminal.onBinary((data) => this.send(Uint8Array.from(data, (c) => c.charCodeAt(0))));
    this.terminal.onResize(({ cols, rows }) => void Resize(id, cols, rows));
    this.terminal.onSelectionChange(() => {
      if (this.settings.copyOnSelect) {
        void this.copy();
      }
    });
    this.resizing = new ResizeObserver(() => this.fitToElement());
    this.resizing.observe(this.element);
    this.theming = new MutationObserver(() => this.restyle());
    this.theming.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class', 'data-style'],
    });
  }

  /**
   * Shows the terminal in parent in place of what it showed, since a pane switching sessions keeps
   * its element; the first attach opens xterm.js, which measures its font in the document.
   */
  attach(parent: HTMLElement): void {
    parent.replaceChildren(this.element);
    if (!this.opened) {
      this.opened = true;
      this.terminal.open(this.element);
      this.render();
      this.fitToElement();
      void Resize(this.id, this.terminal.cols, this.terminal.rows);
    }
  }

  /** Applies the user's terminal settings to the open terminal. */
  configure(settings: TerminalSettings): void {
    this.settings = settings;
    this.restyle();
  }

  private restyle(): void {
    Object.assign(this.terminal.options, terminalLook(this.settings));
    if (this.opened) {
      this.render();
      this.fitToElement();
    }
  }

  /**
   * Draws with WebGL while hardware acceleration is on and the background opaque, since WebGL cannot
   * antialias over a transparent one; else, or once WebGL's context is lost, with the DOM renderer.
   */
  private render(): void {
    if (!this.settings.hardwareAcceleration || this.terminal.options.allowTransparency) {
      this.webgl?.dispose();
      this.webgl = undefined;
      return;
    }
    if (this.webgl) {
      return;
    }
    try {
      const webgl = new WebglAddon();
      webgl.onContextLoss(() => {
        webgl.dispose();
        this.webgl = undefined;
      });
      this.terminal.loadAddon(webgl);
      this.webgl = webgl;
    } catch {
      // Without WebGL the DOM renderer draws the terminal.
    }
  }

  private send(data: Uint8Array<ArrayBuffer>): void {
    if (this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(data);
    }
  }

  /** Ctrl+Shift+C and Ctrl+Shift+V copy and paste; false keeps the key from the shell. */
  private shortcut(event: KeyboardEvent): boolean {
    if (event.type !== 'keydown' || !event.ctrlKey || !event.shiftKey) {
      return true;
    }
    if (event.code === 'KeyC') {
      void this.copy();
      return false;
    }
    if (event.code === 'KeyV') {
      void this.paste();
      return false;
    }
    return true;
  }

  /** Fits the terminal to its element while the element is shown. */
  fitToElement(): void {
    if (this.element.clientWidth > 0 && this.element.clientHeight > 0) {
      this.fit.fit();
    }
  }

  focus(): void {
    this.terminal.focus();
  }

  hasSelection(): boolean {
    return this.terminal.hasSelection();
  }

  async copy(): Promise<void> {
    if (this.terminal.hasSelection()) {
      await Clipboard.SetText(this.terminal.getSelection());
    }
  }

  /** Types the clipboard's text, bracketed when the program asked for bracketed paste. */
  async paste(): Promise<void> {
    this.terminal.paste(await Clipboard.Text());
    this.focus();
  }

  selectAll(): void {
    this.terminal.selectAll();
  }

  /** Clears the scrollback, keeping the line the cursor is on. */
  clear(): void {
    this.terminal.clear();
    this.focus();
  }

  dispose(): void {
    this.resizing.disconnect();
    this.theming.disconnect();
    this.socket.close();
    this.terminal.dispose();
  }
}
