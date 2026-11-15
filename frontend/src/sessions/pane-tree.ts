/** Side by side (horizontal) or one above another (vertical). */
export type Orientation = 'horizontal' | 'vertical';

/** A tab's panes: a session's pane, or panes split along one orientation. */
export type PaneNode =
  | { session: string }
  | { orientation: Orientation; children: readonly PaneNode[] };

/** The sessions in a tree, in reading order. */
export function sessionsIn(node: PaneNode): string[] {
  return 'session' in node ? [node.session] : node.children.flatMap(sessionsIn);
}

/** The tree without a session's pane; a split left with one pane becomes that pane. */
export function withoutSession(node: PaneNode, id: string): PaneNode | undefined {
  if ('session' in node) {
    return node.session === id ? undefined : node;
  }
  const children = node.children.flatMap((child) => withoutSession(child, id) ?? []);
  if (children.length === 0) {
    return undefined;
  }
  return children.length === 1 && children[0] ? children[0] : { ...node, children };
}

/**
 * The tree with a pane for id after beside's, split along orientation. A split already along it
 * takes the pane as a sibling, so three panes side by side share one level.
 */
export function withSessionBeside(
  node: PaneNode,
  beside: string,
  id: string,
  orientation: Orientation,
): PaneNode {
  const pane = { session: id };
  if ('session' in node) {
    return node.session === beside ? { orientation, children: [node, pane] } : node;
  }
  const index = node.children.findIndex((child) => 'session' in child && child.session === beside);
  if (index >= 0 && node.orientation === orientation) {
    return {
      ...node,
      children: [...node.children.slice(0, index + 1), pane, ...node.children.slice(index + 1)],
    };
  }
  return {
    ...node,
    children: node.children.map((child) => withSessionBeside(child, beside, id, orientation)),
  };
}
