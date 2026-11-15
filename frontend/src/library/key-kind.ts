import { m } from '@/paraglide/messages.js';

/** A key's algorithm as people name it, from the type OpenSSH writes. */
export function keyKind(type: string): string {
  if (type === 'ssh-ed25519') {
    return 'Ed25519';
  }
  if (type === 'ssh-rsa') {
    return 'RSA';
  }
  if (type.startsWith('ecdsa-')) {
    return 'ECDSA';
  }
  if (type.startsWith('sk-')) {
    return m.keys_security_key();
  }
  return type;
}
