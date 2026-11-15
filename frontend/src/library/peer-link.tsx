import { cn } from 'cn';
import { type Link, LinkStatus } from '#bindings/hostsource/models.ts';
import { m } from '@/paraglide/messages.js';

/** How this device reaches a peer, in words: its state, then direct or relayed and the latency. */
export function linkText(link: Link): string {
  if (link.status === LinkStatus.LinkConnecting) {
    return m.source_link_connecting();
  }
  if (link.status !== LinkStatus.LinkConnected) {
    return m.source_link_idle();
  }
  const parts = [
    m.source_link_connected(),
    link.relayed ? m.source_link_relayed() : m.source_link_direct(),
  ];
  if (link.latencyMs > 0) {
    parts.push(m.source_link_latency({ ms: link.latencyMs }));
  }
  return parts.join(' · ');
}

/** A dot for a peer's link: filled while connected, hollow otherwise; its words for readers. */
export function LinkDot({ link, className }: { link: Link; className?: string }) {
  const connected = link.status === LinkStatus.LinkConnected;
  return (
    <span title={linkText(link)} className={cn('flex shrink-0 items-center', className)}>
      <span
        aria-hidden
        className={cn(
          'size-2 rounded-full border',
          connected ? 'border-online bg-online' : 'border-muted-foreground',
        )}
      />
      <span className="sr-only">{linkText(link)}</span>
    </span>
  );
}
