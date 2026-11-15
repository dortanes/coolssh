import { type State, Status } from '#bindings/hostsource/models.ts';
import { m } from '@/paraglide/messages.js';
import { sourceName } from './source-names.ts';

/** Whether a sync state needs the user's attention: the source is unreachable or refuses. */
export function troubled(state: State | undefined): boolean {
  return (
    state !== undefined && state.status !== Status.Synced && state.status !== Status.Connecting
  );
}

/** What a synced folder says about its sync. */
export function sourceStatusText(kind: string, state: State | undefined): string {
  const source = sourceName(kind);
  switch (state?.status) {
    case Status.Synced:
      return m.source_status_synced({ count: state.hosts });
    case Status.NotRunning:
      return m.source_status_not_running({ source });
    case Status.Disconnected:
      return m.source_status_disconnected({ source });
    case Status.Unauthorized:
      return m.source_status_unauthorized({ source });
    case Status.GroupsUnavailable:
      return m.source_status_groups_unavailable();
    default:
      return m.source_status_connecting({ source });
  }
}
