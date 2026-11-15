import assert from 'node:assert/strict';
import { test } from 'node:test';
import { type State, Status } from '#bindings/session/models.ts';
import { Observed } from '../observed.ts';
import { Tabs } from './tabs.ts';

class Sessions extends Observed<readonly State[]> {
  constructor(...ids: string[]) {
    super(ids.map(session));
  }

  open(...ids: string[]): void {
    this.update(ids.map(session));
  }
}

function session(id: string): State {
  return { id, hostId: id, title: id, status: Status.Connected, prompt: null, problem: '' };
}

test('closing the active tab activates the one that takes its place', () => {
  const sessions = new Sessions('a', 'b', 'c');
  const tabs = new Tabs(sessions);
  tabs.show('b');
  sessions.open('a', 'c');
  assert.equal(tabs.current().id, 'c');
});

test('closing the last tab activates the one before it', () => {
  const sessions = new Sessions('a', 'b');
  const tabs = new Tabs(sessions);
  tabs.show('b');
  sessions.open('a');
  assert.equal(tabs.current().id, 'a');
});

test('closing the only tab leaves none active', () => {
  const sessions = new Sessions('a');
  const tabs = new Tabs(sessions);
  tabs.show('a');
  sessions.open();
  assert.equal(tabs.current().id, undefined);
});

test('closing another tab keeps the active one', () => {
  const sessions = new Sessions('a', 'b');
  const tabs = new Tabs(sessions);
  tabs.show('b');
  sessions.open('b');
  assert.equal(tabs.current().id, 'b');
});

test('the tabs beside the active one wrap around the ends', () => {
  const sessions = new Sessions('a', 'b', 'c');
  const tabs = new Tabs(sessions);
  tabs.show('c');
  assert.equal(tabs.beside(1), 'a');
  assert.equal(tabs.beside(-1), 'b');
  assert.equal(new Tabs(new Sessions()).beside(1), undefined);
});

test('a session placed beside another shares its tab, and its tab goes', () => {
  const sessions = new Sessions('a', 'b');
  const tabs = new Tabs(sessions);
  tabs.place('b', 'a', 'horizontal');
  assert.deepEqual(tabs.current().layouts, [
    { orientation: 'horizontal', children: [{ session: 'a' }, { session: 'b' }] },
  ]);
  assert.equal(tabs.current().id, 'b');
  assert.equal(tabs.paneBeside(1), 'a');
});

test('a new session placed before the sessions list holds it stays in its split', () => {
  const sessions = new Sessions('a');
  const tabs = new Tabs(sessions);
  tabs.place('b', 'a', 'vertical');
  sessions.open('a', 'b');
  assert.deepEqual(tabs.tabOf('a'), ['a', 'b']);
  assert.equal(tabs.current().layouts.length, 1);
});

test('closing a focused pane focuses a pane left in its tab', () => {
  const sessions = new Sessions('a', 'b', 'c');
  const tabs = new Tabs(sessions);
  tabs.place('b', 'a', 'horizontal');
  tabs.show('a');
  tabs.show('c');
  tabs.show('b');
  sessions.open('a', 'c');
  assert.equal(tabs.current().id, 'a');
});

test('a pane separated from its split gets a tab after it', () => {
  const sessions = new Sessions('a', 'b', 'c');
  const tabs = new Tabs(sessions);
  tabs.place('b', 'a', 'horizontal');
  tabs.separate('a');
  assert.deepEqual(tabs.current().layouts, [{ session: 'b' }, { session: 'a' }, { session: 'c' }]);
});
