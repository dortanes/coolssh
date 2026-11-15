import { ContextMenuItem } from '@/components/ui/context-menu';
import { m } from '@/paraglide/messages.js';
import type { Orientation } from './pane-tree.ts';

/** A menu's items that split a pane, asking what opens beside it. */
export function SplitItems({ onSplit }: { onSplit: (orientation: Orientation) => void }) {
  return (
    <>
      <ContextMenuItem onClick={() => onSplit('horizontal')}>
        {m.shortcuts_split_right()}
      </ContextMenuItem>
      <ContextMenuItem onClick={() => onSplit('vertical')}>
        {m.shortcuts_split_down()}
      </ContextMenuItem>
    </>
  );
}
