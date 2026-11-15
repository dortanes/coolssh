import assert from 'node:assert/strict';
import { test } from 'node:test';
import { Icon, type Library } from '#bindings/library/models.ts';
import { type LibrarySource, LibraryStore } from './library-store.ts';
import { LibraryView } from './library-view.ts';

const empty: Library = {
  version: 1,
  root: [],
  folders: {},
  hosts: {},
  keys: [],
  defaults: {},
};

class FakeGo implements LibrarySource {
  private listener: ((library: Library) => void) | undefined;

  library = async (): Promise<Library> => empty;

  listen = (apply: (library: Library) => void): (() => void) => {
    this.listener = apply;
    return () => {
      this.listener = undefined;
    };
  };

  broadcast(library: Library): void {
    this.listener?.(library);
  }
}

test('a refused change shows Go’s wording until a change succeeds', async () => {
  const store = await LibraryStore.open(new FakeGo());
  const refused = await store.run(() => Promise.reject(new Error('Enter a name.')));
  assert.equal(refused, undefined);
  assert.equal(store.current().problem, 'Enter a name.');
  const id = await store.run(() => Promise.resolve('new-id'));
  assert.equal(id, 'new-id');
  assert.equal(store.current().problem, undefined);
});

test('a library broadcast by Go replaces the one shown', async () => {
  const go = new FakeGo();
  const store = await LibraryStore.open(go);
  const changed = { ...empty, root: ['a'] };
  go.broadcast(changed);
  assert.deepEqual(store.current().library, changed);
});

const nested: Library = {
  version: 1,
  root: ['prod', 'bastion'],
  folders: {
    prod: {
      id: 'prod',
      name: 'prod',
      icon: Icon.FolderIcon,
      expanded: true,
      children: ['db', 'web'],
      settings: {},
    },
    db: {
      id: 'db',
      name: 'db',
      icon: Icon.DatabaseIcon,
      expanded: false,
      children: ['primary'],
      settings: { keyId: 'k' },
    },
  },
  hosts: {
    primary: { id: 'primary', name: 'primary', address: '', settings: { keyId: 'k' } },
    web: { id: 'web', name: 'web', address: '', settings: {} },
    bastion: { id: 'bastion', name: 'bastion', address: '', settings: {} },
  },
  keys: [],
  defaults: {},
};

test('the view counts what a folder holds and who uses a key', () => {
  const view = new LibraryView(nested);
  assert.deepEqual(view.countInside('prod'), { hosts: 2, folders: 1 });
  assert.equal(view.keyUsers('k'), 2);
  assert.equal(view.parentOf('primary'), 'db');
  assert.equal(view.parentOf('bastion'), '');
});

test('hosts take the host icon of the nearest folder that sets one', () => {
  const prod = nested.folders?.['prod'];
  assert.ok(prod);
  const view = new LibraryView({
    ...nested,
    folders: { ...nested.folders, prod: { ...prod, hostIcon: Icon.CloudIcon } },
  });
  assert.equal(view.hostIconIn('db'), Icon.CloudIcon);
  assert.equal(view.hostIconIn(''), Icon.MonitorIcon);
  assert.equal(new LibraryView(nested).hostIconIn('db'), Icon.MonitorIcon);
});
