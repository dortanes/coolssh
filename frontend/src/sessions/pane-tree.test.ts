import assert from 'node:assert/strict';
import { test } from 'node:test';
import { type PaneNode, sessionsIn, withoutSession, withSessionBeside } from './pane-tree.ts';

test('a pane splits beside another, sharing a split already along the same orientation', () => {
  let tree: PaneNode = { session: 'a' };
  tree = withSessionBeside(tree, 'a', 'b', 'horizontal');
  tree = withSessionBeside(tree, 'a', 'c', 'horizontal');
  assert.deepEqual(tree, {
    orientation: 'horizontal',
    children: [{ session: 'a' }, { session: 'c' }, { session: 'b' }],
  });
  tree = withSessionBeside(tree, 'b', 'd', 'vertical');
  assert.deepEqual(sessionsIn(tree), ['a', 'c', 'b', 'd']);
  assert.deepEqual(tree, {
    orientation: 'horizontal',
    children: [
      { session: 'a' },
      { session: 'c' },
      { orientation: 'vertical', children: [{ session: 'b' }, { session: 'd' }] },
    ],
  });
});

test('removing a pane collapses a split left with one', () => {
  const tree: PaneNode = {
    orientation: 'horizontal',
    children: [
      { session: 'a' },
      { orientation: 'vertical', children: [{ session: 'b' }, { session: 'c' }] },
    ],
  };
  assert.deepEqual(withoutSession(tree, 'c'), {
    orientation: 'horizontal',
    children: [{ session: 'a' }, { session: 'b' }],
  });
  assert.deepEqual(withoutSession(withoutSession(tree, 'c') ?? tree, 'a'), { session: 'b' });
  assert.equal(withoutSession({ session: 'a' }, 'a'), undefined);
});
