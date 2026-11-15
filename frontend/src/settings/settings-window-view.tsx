import { cn } from 'cn';
import {
  InfoIcon,
  KeyboardIcon,
  KeyRoundIcon,
  type LucideIcon,
  NetworkIcon,
  SlidersHorizontalIcon,
  SquareTerminalIcon,
} from 'lucide-react';
import { useState } from 'react';
import type { Settings } from '#bindings/library/models.ts';
import type { Product } from '#bindings/product/models.ts';
import { m } from '@/paraglide/messages.js';
import type { WindowStores } from '../mount-window.tsx';
import { TitleBar } from '../shell/title-bar.tsx';
import { useStore } from '../use-store.ts';
import { AboutSection } from './about-section.tsx';
import { ConnectionsSection } from './connections-section.tsx';
import { GeneralSection } from './general-section.tsx';
import { KeysSection } from './keys-section.tsx';
import { ShortcutsSection } from './shortcuts-section.tsx';
import { TerminalSection } from './terminal-section.tsx';
import type { TrustedKeysStore } from './trusted-keys-store.ts';

type Section = 'general' | 'connections' | 'terminal' | 'shortcuts' | 'keys' | 'about';

/** Settings: the sections on the left, the chosen one on the right. */
export function SettingsWindow({
  preferences,
  library,
  product,
  builtInDefaults,
  trustedKeys,
}: WindowStores & {
  product: Product;
  builtInDefaults: Settings;
  trustedKeys: TrustedKeysStore;
}) {
  const [section, setSection] = useState<Section>('general');
  const { snapshot } = useStore(preferences);
  const sections: { id: Section; label: string; Glyph: LucideIcon }[] = [
    { id: 'general', label: m.settings_general(), Glyph: SlidersHorizontalIcon },
    { id: 'connections', label: m.settings_connections(), Glyph: NetworkIcon },
    { id: 'terminal', label: m.settings_terminal(), Glyph: SquareTerminalIcon },
    { id: 'shortcuts', label: m.settings_shortcuts(), Glyph: KeyboardIcon },
    { id: 'keys', label: m.settings_keys(), Glyph: KeyRoundIcon },
    { id: 'about', label: m.settings_about(), Glyph: InfoIcon },
  ];
  return (
    <div className="flex h-full">
      <div className="w-52 shrink-0">
        <nav
          aria-label={m.settings_sections()}
          data-surface="sidebar"
          className="flex h-full flex-col bg-sidebar px-2"
        >
          <TitleBar />
          <ul className="flex flex-col gap-0.5 pt-2 in-data-[platform=darwin]:pt-0">
            {sections.map(({ id, label, Glyph }) => (
              <li key={id}>
                <button
                  type="button"
                  aria-current={section === id ? 'page' : undefined}
                  className={cn(
                    'flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-left text-sm outline-none hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-sidebar-ring',
                    section === id &&
                      'bg-sidebar-accent font-medium text-sidebar-accent-foreground',
                  )}
                  onClick={() => setSection(id)}
                >
                  <Glyph aria-hidden className="size-4 shrink-0 text-muted-foreground" />
                  {label}
                </button>
              </li>
            ))}
          </ul>
        </nav>
      </div>
      <main className="flex flex-1 flex-col overflow-y-auto bg-background">
        <TitleBar />
        {section === 'general' && <GeneralSection preferences={preferences} />}
        {section === 'connections' && (
          <ConnectionsSection library={library} builtInDefaults={builtInDefaults} />
        )}
        {section === 'terminal' && (
          <TerminalSection preferences={preferences} platform={snapshot.platform} />
        )}
        {section === 'shortcuts' && <ShortcutsSection preferences={preferences} />}
        {section === 'keys' && <KeysSection library={library} trustedKeys={trustedKeys} />}
        {section === 'about' && <AboutSection product={product} />}
      </main>
    </div>
  );
}
