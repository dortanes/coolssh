import { goSources } from './library/sources-source.ts';
import { SourcesStore } from './library/sources-store.ts';
import { mountWindow } from './mount-window.tsx';
import { goSessions } from './sessions/sessions-source.ts';
import { SessionsStore } from './sessions/sessions-store.ts';
import { MainWindow } from './shell/main-window-view.tsx';

void Promise.all([SessionsStore.open(goSessions), SourcesStore.open(goSources)]).then(
  ([sessions, sources]) =>
    mountWindow((stores) => <MainWindow {...stores} sessions={sessions} sources={sources} />),
);
