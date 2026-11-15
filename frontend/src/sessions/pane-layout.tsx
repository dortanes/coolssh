import { Fragment, type ReactNode } from 'react';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { type PaneNode, sessionsIn } from './pane-tree.ts';

/** The smallest a pane gets while dragging the line between panes, CSS pixels. */
const minPane = 120;

/** Draws a tab's panes, nested splits as nested groups with a draggable line between panes. */
export function PaneLayout({
  node,
  pane,
}: {
  node: PaneNode;
  pane: (session: string) => ReactNode;
}) {
  if ('session' in node) {
    return pane(node.session);
  }
  return (
    <ResizablePanelGroup data-surface="panes" orientation={node.orientation}>
      {node.children.map((child, index) => {
        const key = sessionsIn(child).join();
        return (
          <Fragment key={key}>
            {index > 0 && <ResizableHandle />}
            <ResizablePanel id={key} minSize={minPane}>
              <PaneLayout node={child} pane={pane} />
            </ResizablePanel>
          </Fragment>
        );
      })}
    </ResizablePanelGroup>
  );
}
