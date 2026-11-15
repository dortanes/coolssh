import assert from 'node:assert/strict';
import { test } from 'node:test';
import { Broadcasts } from './broadcasts.ts';

test('a broadcast sent before the store follows wins over the value it read', () => {
  let send: (value: string) => void = () => {};
  const broadcasts = new Broadcasts<string>((apply) => {
    send = apply;
  });
  send('broadcast');
  assert.equal(broadcasts.latestOr('read'), 'broadcast');
  const followed: string[] = [];
  broadcasts.follow((value) => followed.push(value));
  send('later');
  assert.deepEqual(followed, ['later']);
});

test('without a broadcast the value read stands', () => {
  const broadcasts = new Broadcasts<string>(() => {});
  assert.equal(broadcasts.latestOr('read'), 'read');
});
