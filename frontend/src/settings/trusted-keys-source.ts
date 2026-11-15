import { Events } from '@wailsio/runtime';
import { ForgetTrustedKey, TrustedKeys } from '#bindings/desktop/service.ts';
import type { TrustedKeysSource } from './trusted-keys-store.ts';

/** The trusted host keys as Go keeps them: the bound service and the event it broadcasts. */
export const goTrustedKeys: TrustedKeysSource = {
  trusted: TrustedKeys,
  forget: ForgetTrustedKey,
  listen: (apply) => Events.On('trusted-keys', (event) => apply(event.data ?? [])),
};
