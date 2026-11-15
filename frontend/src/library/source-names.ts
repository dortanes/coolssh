import { m } from '@/paraglide/messages.js';

/** Each source's name as the user reads it, by kind; a new source adds its line here. */
const names: { [kind in string]?: () => string } = {
  netbird: m.source_netbird,
};

export function sourceName(kind: string): string {
  return names[kind]?.() ?? kind;
}
